self.addEventListener('fetch', (event) => {
    event.respondWith(fetch(event.request));
});

self.addEventListener('push', (event) => {
    if (event.data) {
        let data;
        try {
            data = event.data.json();
        } catch (e) {
            data = { title: 'SSH Manager', body: event.data.text(), badge: 1 };
        }

        if (data.badge !== undefined && 'setAppBadge' in navigator) {
            navigator.setAppBadge(data.badge).catch(err => console.error(err));
        }

        const options = {
            body: data.body,
            icon: '/static/img/icon-192.png',
            badge: '/static/img/icon-192.png',
            requireInteraction: true,
            tag: 'session-timeout-' + Date.now(),
            data: { url: data.url || '/' }
        };

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
