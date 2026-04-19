self.addEventListener('fetch', (event) => {
    event.respondWith(fetch(event.request));
});

self.addEventListener('push', (event) => {
    if (!event.data) return;

    let data;
    try {
        data = event.data.json();
    } catch (e) {
        data = { title: 'SSH Manager', body: event.data.text(), badge: 1 };
    }

    const unreadCount = parseInt(data.badge) || 0;

    event.waitUntil((async () => {
        if (navigator.setAppBadge) {
            try {
                if (unreadCount > 0) {
                    await navigator.setAppBadge(unreadCount);
                } else {
                    await navigator.clearAppBadge();
                }
            } catch (err) {
            }
        }

        const options = {
            body: data.body,
            icon: '/static/img/icon-192.png',
            badge: '/static/img/icon-192.png',
            data: { url: data.url || '/' }
        };
        
        await self.registration.showNotification(data.title, options);
    })());
});

self.addEventListener('notificationclick', (event) => {
    event.notification.close();

    // When clicking on a notification, it makes sense to either decrease the counter, 
    // or clear it completely if we open the app.
    if (navigator.clearAppBadge) {
        event.waitUntil(navigator.clearAppBadge());
    }

    event.waitUntil(
        clients.openWindow(event.notification.data.url || '/')
    );
});
