self.addEventListener('fetch', (event) => {
    event.respondWith(fetch(event.request));
});

self.addEventListener('push', (event) => {
    if (!event.data) return;

    const data = event.data.json();
    const newBadge = parseInt(data.badge) || 0;

    const task = caches.open('badge-store').then(async (cache) => {
        const cacheKey = '/last-badge-count';
        const cachedResponse = await cache.match(cacheKey);
        const lastBadge = cachedResponse ? parseInt(await cachedResponse.text()) : 0;

        if (newBadge >= lastBadge || newBadge === 0) {
            if (navigator.setAppBadge) {
                navigator.setAppBadge(newBadge);
                await cache.put(cacheKey, new Response(newBadge.toString()));
            }
        }
    });

    const notification = self.registration.showNotification(data.title, {
        body: data.body,
        icon: '/static/img/icon-192.png',
        badge: '/static/img/icon-192.png',
        data: { url: data.url || '/' }
    });

    event.waitUntil(Promise.all([task, notification]));
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
