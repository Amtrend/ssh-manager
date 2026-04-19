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

    const unreadCount = parseInt(data.badge);
    const promises = [];

    if (navigator.setAppBadge) {
        if (unreadCount && unreadCount > 0) {
            promises.push(navigator.setAppBadge(unreadCount));
        } else {
            promises.push(navigator.clearAppBadge());
        }
    }

    const options = {
        body: data.body,
        icon: '/static/img/icon-192.png',
        badge: '/static/img/icon-192.png',
        data: { url: data.url || '/' }
    };
    promises.push(self.registration.showNotification(data.title, options));

    event.waitUntil(Promise.all(promises));
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
