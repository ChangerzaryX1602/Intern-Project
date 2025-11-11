# WebSocket Hooks

This project includes two WebSocket hooks for real-time communication:

## 1. useWebSocket (Native WebSocket)

A hook for native WebSocket connections with automatic reconnection.

### Installation

No additional packages required - uses browser's native WebSocket API.

### Usage

```svelte
<script lang="ts">
  import { useWebSocket } from '$lib/hooks/useWebSocket';

  const { data, status, error, send, connect, disconnect } = useWebSocket(
    'ws://localhost:8080/ws',
    {
      reconnect: true,
      reconnectInterval: 3000,
      reconnectAttempts: 5,
      onOpen: (event) => console.log('Connected'),
      onClose: (event) => console.log('Disconnected'),
      onError: (event) => console.error('Error:', event),
      onMessage: (event) => console.log('Message:', event.data)
    }
  );

  // Send message
  function sendMessage() {
    send({ type: 'message', content: 'Hello!' });
  }

  // Access reactive data
  $: currentData = $data;
  $: connectionStatus = $status; // 'connecting' | 'connected' | 'disconnected' | 'error'
  $: errorMessage = $error;
</script>

<div>
  <p>Status: {$status}</p>
  {#if $error}
    <p>Error: {$error}</p>
  {/if}
  {#if $data}
    <pre>{JSON.stringify($data, null, 2)}</pre>
  {/if}
  <button onclick={sendMessage}>Send Message</button>
</div>
```

### Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `reconnect` | boolean | true | Enable automatic reconnection |
| `reconnectInterval` | number | 3000 | Time between reconnection attempts (ms) |
| `reconnectAttempts` | number | 5 | Maximum number of reconnection attempts |
| `onOpen` | function | - | Callback when connection opens |
| `onClose` | function | - | Callback when connection closes |
| `onError` | function | - | Callback on error |
| `onMessage` | function | - | Callback on message received |

### Return Values

| Property | Type | Description |
|----------|------|-------------|
| `data` | Writable<T \| null> | Latest received data |
| `status` | Writable<WebSocketStatus> | Connection status |
| `error` | Writable<string \| null> | Error message if any |
| `send` | function | Send message to server |
| `connect` | function | Manually connect |
| `disconnect` | function | Manually disconnect |

---

## 2. useSocketIO (Socket.IO)

A hook for Socket.IO connections with event-based communication.

### Installation

First, install Socket.IO client:

```bash
bun add socket.io-client
```

### Usage

```svelte
<script lang="ts">
  import { useSocketIO } from '$lib/hooks/useSocketIO';

  const { data, status, error, emit, on, off, connect, disconnect } = useSocketIO(
    'http://localhost:3000',
    {
      autoConnect: true,
      reconnection: true,
      reconnectionAttempts: 5,
      path: '/socket.io'
    }
  );

  // Listen to custom events
  on('custom_event', (eventData) => {
    console.log('Custom event received:', eventData);
  });

  // Emit events
  function sendCustomEvent() {
    emit('my_event', { message: 'Hello Server!' });
  }

  // Subscribe to camera feed (example from tesa-ui)
  function subscribeCamera(camId: string) {
    emit('subscribe_camera', { cam_id: camId });
  }

  on('object_detection', (detection) => {
    console.log('Detection received:', detection);
  });
</script>

<div>
  <p>Status: {$status}</p>
  <button onclick={sendCustomEvent}>Emit Event</button>
</div>
```

### Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `autoConnect` | boolean | true | Automatically connect on mount |
| `reconnection` | boolean | true | Enable automatic reconnection |
| `reconnectionAttempts` | number | 5 | Maximum reconnection attempts |
| `reconnectionDelay` | number | 3000 | Delay between attempts (ms) |
| `path` | string | '/socket.io' | Server path |
| `transports` | string[] | ['websocket', 'polling'] | Transport methods |
| `auth` | object | - | Authentication data |

### Return Values

| Property | Type | Description |
|----------|------|-------------|
| `data` | Writable<T \| null> | Latest received data |
| `status` | Writable<SocketStatus> | Connection status |
| `error` | Writable<string \| null> | Error message if any |
| `emit` | function | Emit event to server |
| `on` | function | Listen to event |
| `off` | function | Stop listening to event |
| `connect` | function | Manually connect |
| `disconnect` | function | Manually disconnect |

---

## Demo Pages

### WebSocket Demo
Visit `/websocket` to see a live demo of the native WebSocket hook with:
- Real-time connection status
- Message sending and receiving
- Message history
- Reconnection handling

### Configuration

Add to your `.env` file:

```bash
# WebSocket URL
PUBLIC_WS_URL=ws://localhost:8080/ws

# Socket.IO URL (if using Socket.IO)
PUBLIC_SOCKETIO_URL=http://localhost:3000
```

---

## Examples from Your Project

### From tesa-ui (Socket.IO Example)

```typescript
// Subscribe to camera detection events
const { emit, on } = useSocketIO('https://tesa-api.crma.dev');

// Subscribe to specific camera
emit('subscribe_camera', { cam_id: 'CAM001' });

// Listen for object detections
on('object_detection', (data) => {
  console.log('Detection:', data);
});
```

---

## Tips

1. **Native WebSocket** is lighter and faster for simple real-time needs
2. **Socket.IO** provides more features like rooms, namespaces, and automatic reconnection
3. Always disconnect in `onDestroy` to prevent memory leaks (handled automatically by the hooks)
4. Use environment variables for WebSocket URLs for different environments
5. Handle connection status to show loading states to users

## Comparison

| Feature | useWebSocket | useSocketIO |
|---------|--------------|-------------|
| Bundle Size | Smaller | Larger |
| Browser Support | Modern browsers | Wide support |
| Reconnection | Manual | Automatic |
| Event-based | No | Yes |
| Rooms/Namespaces | No | Yes |
| Binary Data | Yes | Yes |
| Best For | Simple real-time | Complex apps |
