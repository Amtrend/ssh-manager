package handlers

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"ssh_manager/internal/services"
	"ssh_manager/internal/utils"
	"strconv"
	"time"

	"github.com/pkg/sftp"
)

// GetFilesHandler Returns a list of files for the file manager.
func (h *Handlers) GetFilesHandler(w http.ResponseWriter, r *http.Request) {
	hostID, err := strconv.Atoi(r.URL.Query().Get("host_id"))
	if err != nil {
		utils.SendJSONResponse(w, false, "Invalid host ID", nil)
		return
	}

	path := r.URL.Query().Get("path")

	session, _ := h.Store.Get(r, utils.SessionName)
	userID, _ := session.Values[utils.UserIDKey].(int)

	as, err := h.SSHService.GetSession(userID, hostID, r.Context())
	if err != nil {
		utils.LogErrorf("Failed to get SSH session for SFTP", err, "host_id", hostID)
		utils.SendJSONResponse(w, false, "SSH connection failed: "+err.Error(), nil)
		return
	}

	if path == "" {
		path = "/"
	}

	// Using a client from an active session to work with SFTP.
	as.Mu.Lock()
	sftpClient := as.SFTPClient
	as.Mu.Unlock()

	if sftpClient == nil {
		utils.SendJSONResponse(w, false, "SFTP service is not available for this session", nil)
		return
	}

	files, err := services.ListDirectory(sftpClient, path)
	if err != nil {
		utils.SendJSONResponse(w, false, "SFTP Error: "+err.Error(), nil)
		return
	}

	utils.SendJSONResponse(w, true, "Success", map[string]interface{}{
		"current_path": path,
		"files":        files,
	})
}

// DownloadFileHandler downloads file when clicked.
func (h *Handlers) DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	hostID, _ := strconv.Atoi(query.Get("host_id"))
	remotePath := query.Get("path")

	session, _ := h.Store.Get(r, utils.SessionName)
	userID, _ := session.Values[utils.UserIDKey].(int)
	as, err := h.SSHService.GetSession(userID, hostID, r.Context())
	if err != nil {
		http.Error(w, "SSH session failed", http.StatusInternalServerError)
		return
	}

	as.Mu.Lock()
	sftpClient := as.SFTPClient
	as.Mu.Unlock()

	if sftpClient == nil {
		utils.SendJSONResponse(w, false, "SFTP service is not available for this session", nil)
		return
	}

	// Open a file on a remote server.
	file, err := sftpClient.Open(remotePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	// Getting file information for headers.
	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Stat failed", http.StatusInternalServerError)
		return
	}

	// Headlines: Force the browser to download, not open.
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", stat.Name()))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))

	// Copying data from sftp directly into the http response
	_, err = io.Copy(w, file)
	if err != nil {
		utils.LogErrorf("Error streaming file", err)
	}
}

// DownloadZipHandler downloads selected files as a zip archive.
func (h *Handlers) DownloadZipHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	hostIDStr := query.Get("host_id")
	parentPath := query.Get("parent_path")
	filesJSON := query.Get("files")
	receivedCSRF := query.Get("csrf_token")

	session, _ := h.Store.Get(r, utils.SessionName)
	expectedCSRF, _ := session.Values["csrf_token"].(string)

	if receivedCSRF == "" || receivedCSRF != expectedCSRF {
		utils.LogErrorf("CSRF mismatch in DownloadZip", nil, "received", receivedCSRF)
		http.Error(w, "Invalid security token", http.StatusForbidden)
		return
	}

	hostID, err := strconv.Atoi(hostIDStr)
	if err != nil {
		http.Error(w, "Invalid host ID", http.StatusBadRequest)
		return
	}

	var fileNames []string
	if err := json.Unmarshal([]byte(filesJSON), &fileNames); err != nil {
		http.Error(w, "Invalid files list format", http.StatusBadRequest)
		return
	}

	userID, _ := session.Values[utils.UserIDKey].(int)
	as, err := h.SSHService.GetSession(userID, hostID, r.Context())
	if err != nil {
		utils.LogErrorf("Failed to get SSH session for ZIP", err, "host_id", hostID)
		http.Error(w, "SSH connection failed", http.StatusInternalServerError)
		return
	}

	as.Mu.Lock()
	sftpClient := as.SFTPClient
	as.Mu.Unlock()

	if sftpClient == nil {
		utils.SendJSONResponse(w, false, "SFTP service is not available for this session", nil)
		return
	}

	zipName := fmt.Sprintf("archive_%d.zip", time.Now().Unix())
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", zipName))
	w.Header().Set("Content-Type", "application/zip")
	// We remove caching so that the mobile phone doesn't slip in an old file.
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	// Initiating a zip stream.
	zipWriter := zip.NewWriter(w)

	for _, name := range fileNames {
		remoteFullPath := path.Join(parentPath, name)

		// Recursively adding files/folders.
		err := h.addSftpToZip(sftpClient, zipWriter, remoteFullPath, "")
		if err != nil {
			utils.LogErrorf("Error adding to zip", err, "path", remoteFullPath)
		}
	}

	err = zipWriter.Close()
	if err != nil {
		utils.LogErrorf("Failed to close zip writer", err)
		return
	}
}

// UploadHandler uploads selected files to the host.
func (h *Handlers) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(100 << 20); err != nil { // 100 MB limit
		utils.SendJSONResponse(w, false, "Files too large", nil)
		return
	}

	session, _ := h.Store.Get(r, utils.SessionName)
	expectedToken, _ := session.Values["csrf_token"].(string)
	actualToken := r.FormValue("csrf_token")

	if actualToken == "" || actualToken != expectedToken {
		utils.SendJSONResponse(w, false, "Invalid security token (CSRF)", nil)
		return
	}

	hostID, _ := strconv.Atoi(r.FormValue("host_id"))
	remotePath := r.FormValue("remote_path")

	userID, _ := session.Values[utils.UserIDKey].(int)
	as, err := h.SSHService.GetSession(userID, hostID, r.Context())
	if err != nil {
		utils.SendJSONResponse(w, false, "SSH session failed", nil)
		return
	}

	as.Mu.Lock()
	sftpClient := as.SFTPClient
	as.Mu.Unlock()

	if sftpClient == nil {
		utils.SendJSONResponse(w, false, "SFTP service is not available for this session", nil)
		return
	}

	files := r.MultipartForm.File["files"]
	for _, fileHeader := range files {
		src, err := fileHeader.Open()
		if err != nil {
			continue
		}
		defer src.Close()

		safeFileName := path.Base(fileHeader.Filename)
		dstPath := path.Join(remotePath, safeFileName)
		dst, err := sftpClient.Create(dstPath)
		if err != nil {
			utils.LogErrorf("Failed to create file on SFTP", err, "path", dstPath)
			continue
		}

		_, err = io.Copy(dst, src)
		dst.Close()
	}

	utils.SendJSONResponse(w, true, "Files uploaded successfully", nil)
}

// addSftpToZip Helper function for recursively adding files and folders from SFTP to ZIP.
func (h *Handlers) addSftpToZip(client *sftp.Client, zw *zip.Writer, remotePath, baseInZip string) error {
	info, err := client.Stat(remotePath)
	if err != nil {
		return err
	}

	// Creating a title.
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// Set the UTF-8 flag (bit value 0x800).
	header.Flags |= 0x800

	if baseInZip == "" {
		header.Name = info.Name()
	} else {
		header.Name = path.Join(baseInZip, info.Name())
	}

	if info.IsDir() {
		header.Name += "/" // Required for folders.
	} else {
		header.Method = zip.Deflate
	}

	writer, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}

	if info.IsDir() {
		subFiles, err := client.ReadDir(remotePath)
		if err != nil {
			return err
		}
		for _, f := range subFiles {
			// we pass the current path in the archive as a new parent.
			err = h.addSftpToZip(client, zw, path.Join(remotePath, f.Name()), header.Name)
			if err != nil {
				return err
			}
		}
		return nil
	}

	f, err := client.Open(remotePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(writer, f)
	return err
}
