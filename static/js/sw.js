self.addEventListener('fetch', (event) => {
    event.respondWith(fetch(event.request));
});

self.addEventListener('push', (event) => {
    if (event.data) {
        let data;
        try {
            data = event.data.json();
        } catch (e) {
            data = { title: 'SSH Manager', body: event.data.text() };
        }

        const options = {
            body: data.body,
            icon: '/static/img/icon-192.png',
            badge: '/static/img/icon-192.png',
            requireInteraction: true,
            data: { url: data.url || '/' }
        };

        // Badge control logic
        if ('setAppBadge' in navigator) {
            event.waitUntil(
                // We get a list of all active notifications for this application.
                self.registration.getNotifications().then(notifications => {
                    // The new notification hasn't been created yet, so we'll take the current ones + 1
                    const currentCount = notifications.length + 1;
                    return navigator.setAppBadge(currentCount);
                }).catch(err => console.error('Badge error:', err))
            );
        }

        event.waitUntil(
            self.registration.showNotification(data.title, options)
        );
    }
});

self.addEventListener('notificationclick', (event) => {
    event.notification.close();

    // When clicking on a notification, it makes sense to either decrease the counter, 
    // or clear it completely if we open the app.
    if ('clearAppBadge' in navigator) {
        navigator.clearAppBadge();
    }

    event.waitUntil(
        clients.openWindow(event.notification.data.url)
    );
});
