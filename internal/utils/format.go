package utils

import "fmt"

// formatPartialKey generates a partial mapping of the key.
func FormatPartialKey(input interface{}) string {
	switch v := input.(type) {
	case int:
		return fmt.Sprintf("Key ID: %d (******)", v)
	case string:
		if len(v) <= 30 {
			return v // If the key is short, we return it in full
		}
		start := v[:15]
		end := v[len(v)-15:]
		return start + "..." + end
	default:
		return "Unknown key format"
	}
}
