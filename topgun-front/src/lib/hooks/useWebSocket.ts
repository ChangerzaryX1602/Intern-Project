import { onMount, onDestroy } from 'svelte';
import { writable, type Writable } from 'svelte/store';

export type WebSocketStatus = 'connecting' | 'connected' | 'disconnected' | 'error';

export interface WebSocketOptions {
	reconnect?: boolean;
	reconnectInterval?: number;
	reconnectAttempts?: number;
	onOpen?: (event: Event) => void;
	onClose?: (event: CloseEvent) => void;
	onError?: (event: Event) => void;
	onMessage?: (event: MessageEvent) => void;
}

export interface UseWebSocketReturn<T> {
	data: Writable<T | null>;
	status: Writable<WebSocketStatus>;
	error: Writable<string | null>;
	send: (data: string | object) => void;
	connect: () => void;
	disconnect: () => void;
}

/**
 * Custom hook for WebSocket connection
 * @param url - WebSocket URL
 * @param options - Configuration options
 * @returns WebSocket utilities and state
 */
export function useWebSocket<T = any>(
	url: string,
	options: WebSocketOptions = {}
): UseWebSocketReturn<T> {
	const {
		reconnect = true,
		reconnectInterval = 3000,
		reconnectAttempts = 5,
		onOpen,
		onClose,
		onError,
		onMessage
	} = options;

	let ws: WebSocket | null = null;
	let reconnectCount = 0;
	let reconnectTimeout: ReturnType<typeof setTimeout> | null = null;

	const data = writable<T | null>(null);
	const status = writable<WebSocketStatus>('disconnected');
	const error = writable<string | null>(null);

	function connect() {
		if (ws?.readyState === WebSocket.OPEN || ws?.readyState === WebSocket.CONNECTING) {
			return;
		}

		try {
			status.set('connecting');
			error.set(null);

			ws = new WebSocket(url);

			ws.onopen = (event) => {
				status.set('connected');
				reconnectCount = 0;
				console.log('WebSocket connected');
				onOpen?.(event);
			};

			ws.onmessage = (event) => {
				try {
					const parsedData = JSON.parse(event.data);
					data.set(parsedData);
					onMessage?.(event);
				} catch (e) {
					console.error('Failed to parse WebSocket message:', e);
					data.set(event.data as T);
				}
			};

			ws.onerror = (event) => {
				status.set('error');
				error.set('WebSocket error occurred');
				console.error('WebSocket error:', event);
				onError?.(event);
			};

			ws.onclose = (event) => {
				status.set('disconnected');
				console.log('WebSocket disconnected');
				onClose?.(event);

				// Attempt to reconnect
				if (reconnect && reconnectCount < reconnectAttempts) {
					reconnectCount++;
					console.log(
						`Attempting to reconnect... (${reconnectCount}/${reconnectAttempts})`
					);
					reconnectTimeout = setTimeout(() => {
						connect();
					}, reconnectInterval);
				}
			};
		} catch (e) {
			status.set('error');
			error.set(e instanceof Error ? e.message : 'Failed to connect');
			console.error('Failed to create WebSocket connection:', e);
		}
	}

	function disconnect() {
		if (reconnectTimeout) {
			clearTimeout(reconnectTimeout);
			reconnectTimeout = null;
		}
		reconnectCount = reconnectAttempts; // Prevent reconnection
		ws?.close();
		ws = null;
		status.set('disconnected');
	}

	function send(message: string | object) {
		if (ws?.readyState === WebSocket.OPEN) {
			const dataToSend = typeof message === 'string' ? message : JSON.stringify(message);
			ws.send(dataToSend);
		} else {
			console.warn('WebSocket is not connected. Message not sent.');
		}
	}

	onMount(() => {
		connect();
	});

	onDestroy(() => {
		disconnect();
	});

	return {
		data,
		status,
		error,
		send,
		connect,
		disconnect
	};
}
