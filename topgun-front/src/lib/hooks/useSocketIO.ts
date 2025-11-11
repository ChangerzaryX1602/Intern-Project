import { onMount, onDestroy } from 'svelte';
import { writable, type Writable } from 'svelte/store';
import { io, type Socket } from 'socket.io-client';

export type SocketStatus = 'connecting' | 'connected' | 'disconnected' | 'error';

export interface SocketIOOptions {
	autoConnect?: boolean;
	reconnection?: boolean;
	reconnectionAttempts?: number;
	reconnectionDelay?: number;
	path?: string;
	transports?: string[];
	auth?: Record<string, any>;
}

export interface UseSocketIOReturn<T> {
	data: Writable<T | null>;
	status: Writable<SocketStatus>;
	error: Writable<string | null>;
	emit: (event: string, data?: any) => void;
	on: (event: string, callback: (data: any) => void) => void;
	off: (event: string, callback?: (data: any) => void) => void;
	connect: () => void;
	disconnect: () => void;
}

/**
 * Custom hook for Socket.IO connection
 * @param url - Socket.IO server URL
 * @param options - Socket.IO configuration options
 * @returns Socket.IO utilities and state
 */
export function useSocketIO<T = any>(
	url: string,
	options: SocketIOOptions = {}
): UseSocketIOReturn<T> {
	const {
		autoConnect = true,
		reconnection = true,
		reconnectionAttempts = 5,
		reconnectionDelay = 3000,
		path = '/socket.io',
		transports = ['websocket', 'polling'],
		auth
	} = options;

	let socket: Socket | null = null;

	const data = writable<T | null>(null);
	const status = writable<SocketStatus>('disconnected');
	const error = writable<string | null>(null);

	function connect() {
		if (socket?.connected) {
			return;
		}

		try {
			status.set('connecting');
			error.set(null);

			socket = io(url, {
				autoConnect,
				reconnection,
				reconnectionAttempts,
				reconnectionDelay,
				path,
				transports,
				auth
			});

			socket.on('connect', () => {
				status.set('connected');
				console.log('Socket.IO connected:', socket?.id);
			});

			socket.on('disconnect', (reason: string) => {
				status.set('disconnected');
				console.log('Socket.IO disconnected:', reason);
			});

			socket.on('connect_error', (err: Error) => {
				status.set('error');
				error.set(err.message);
				console.error('Socket.IO connection error:', err);
			});

			if (autoConnect) {
				socket.connect();
			}
		} catch (e) {
			status.set('error');
			error.set(e instanceof Error ? e.message : 'Failed to connect');
			console.error('Failed to create Socket.IO connection:', e);
		}
	}

	function disconnect() {
		socket?.disconnect();
		socket = null;
		status.set('disconnected');
	}

	function emit(event: string, eventData?: any) {
		if (socket?.connected) {
			socket.emit(event, eventData);
		} else {
			console.warn('Socket.IO is not connected. Event not emitted.');
		}
	}

	function on(event: string, callback: (eventData: any) => void) {
		socket?.on(event, callback);
	}

	function off(event: string, callback?: (eventData: any) => void) {
		if (callback) {
			socket?.off(event, callback);
		} else {
			socket?.off(event);
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
		emit,
		on,
		off,
		connect,
		disconnect
	};
}
