// Wire shapes returned by /api/public/*. They are intentionally narrower than
// the backend's CRM models — only fields the public website renders.

export interface SiteConfig {
  brand_name: string;
  phone: string;
  email: string;
  address: string;
  apk_url: string;
  apk_version: string;
  locales: string[];
  chat_enabled: boolean;
}

export interface PublicHouse {
  id: number;
  title: string;
  district: string;
  address: string;
  area: number;
  rooms: number;
  floor: number;
  total_floors: number;
  price_from: number;
  short_desc: string;
  full_desc: string;
  gallery: string[];
  features: string[];
  lat: number;
  lng: number;
  has_tour: boolean;
  tour_slug?: string;
}

export interface PanoramaHotspot {
  id: number;
  panorama_id: number;
  target_panorama_id?: number;
  kind: "link" | "info" | "url";
  label?: string;
  yaw: number;
  pitch: number;
  info_text?: string;
  info_url?: string;
}

export interface Panorama {
  id: number;
  tour_id: number;
  title: string;
  image_url: string;
  image_type: string;
  initial_yaw: number;
  initial_pitch: number;
  initial_zoom: number;
  sort_order: number;
  hotspots?: PanoramaHotspot[];
}

export interface Tour {
  id: number;
  title: string;
  slug: string;
  description?: string;
  house_id?: number;
  cover_image?: string;
  initial_panorama_id?: number;
  panoramas: Panorama[];
}
