// Type definitions for Defensive Dashboard

export interface Camera {
	id: string;
	name: string;
	location: string;
	token: string;
	institute: string;
	created_at: string;
	updated_at: string;
}

export interface Detection {
	id: number;
	camera_id: string;
	detected_at?: string;
	timestamp?: string;
	path: string;
	detected_objects?: DetectedObject[];
	objects?: DetectedObject[];  // New MQTT format
	image_base64?: string;
	mime_type?: string;
	camera?: Camera;
}

// Detected objects can come in different shapes depending on the stream
// - older streams: { class_name, confidence }
// - drone stream: { obj_id, type, lat, lng, objective, size, details }
export interface DetectedObject {
	// generic/classifier fields
	class_name?: string;
	confidence?: number;

	// drone/object tracking fields
	obj_id?: string;
	type?: string;
	// lat/lng may be strings from backend, allow both
	lat?: string | number;
	lng?: string | number;
	objective?: string;
	size?: string;
	details?: any;

	// MQTT detection format (new)
	h?: number;
	w?: number;
	x?: number;
	y?: number;
	alt?: number;
	lon?: number;
	track_id?: number;
	timestamp?: number;
}

export interface Pagination {
	page: number;
	limit: number;
	total: number;
	totalPages: number;
}

export interface CameraListResponse {
	success: boolean;
	data: {
		cameras: Camera[];
		pagination: {
			page: number;
			limit: number;
			total: number;
			total_pages: number;
		};
	};
}
