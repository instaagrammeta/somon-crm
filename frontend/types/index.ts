/**
 * Domain types — match Go backend models.
 */
export interface User {
  id: number;
  full_name: string;
  login: string;
  age?: number;
  personal_phones?: string[];
  work_phones?: string[];
  photo?: string;
  category?: string;
  role: "admin" | "employee" | "manager";
  telegram_username?: string;
  has_telegram?: boolean;
  is_active: boolean;
  created_at: string;
}

export interface ApiError {
  success: false;
  error: string;
  code?: string;
}

export interface Task {
  id: number;
  title: string;
  description?: string;
  author_id?: number;
  executor_id?: number;
  author_name?: string;
  executor_name?: string;
  photo?: string;
  status: "new" | "in_progress" | "done" | "cancelled";
  created_at: string;
}

export interface RequestsBoard {
  id: number;
  title: string;
  color: string;
  is_public: boolean;
  author_id?: number;
}

export interface RequestsColumn {
  id: number;
  board_id: number;
  title: string;
  color: string;
  order_index: number;
}

export interface RequestItem {
  id: number;
  column_id: number;
  board_id: number;
  property_type?: string;
  address?: string;
  area?: number;
  rooms?: number;
  windows?: number;
  floor?: number;
  total_floors?: number;
  total_price?: number;
  price_per_m2?: number;
  phone?: string;
  client_name?: string;
  comment?: string;
  author_name?: string;
  executor_name?: string;
  executor_id?: number;
  files?: any[];
  order_index: number;
  created_at: string;
}

export interface KanbanBoard {
  id: number;
  title: string;
  color: string;
  is_public: boolean;
  is_archived: boolean;
}

export interface KanbanColumn {
  id: number;
  board_id: number;
  title: string;
  color: string;
  order_index: number;
}

export interface KanbanLead {
  id: number;
  board_id: number;
  column_id: number;
  client_name: string;
  phone?: string;
  topic?: string;
  comment?: string;
  source?: string;
  mortgage?: boolean;
  box?: boolean;
  author_name?: string;
  order_index: number;
  created_at: string;
}

export interface House {
  id: number;
  title: string;
  construction_type?: string;
  district?: string;
  address?: string;
  area?: number;
  rooms?: number;
  windows?: number;
  floor?: number;
  total_floors?: number;
  price_per_m2?: number;
  total_price?: number;
  developer?: string;
  contact_phone?: string;
  files?: string[];
}

export interface RealtyObject {
  id: number;
  name: string;
  description?: string;
  address?: string;
  lat: number;
  lng: number;
  district?: string;
  total_floors?: number;
  price_per_m2?: number;
  total_price?: number;
  developer?: string;
}

export interface ObjektProject {
  id: number;
  name: string;
  address?: string;
  developer?: string;
  description?: string;
}

export interface ObjektBlock {
  id: number;
  project_id: number;
  name: string;
  floor_from: number;
  floor_to: number;
  default_area: number;
  default_rooms: number;
  default_windows: number;
  default_price_per_m2: number;
  default_balcony?: boolean;
  default_bathroom_type?: string;
  default_bathroom_count?: number;
  default_plan_image?: string;
  description?: string;
  order_index: number;
}

export interface ObjektApartment {
  id: number;
  block_id: number;
  block_name?: string;
  floor: number;
  area: number;
  rooms: number;
  windows: number;
  price_per_m2: number;
  total_price: number;
  status: "free" | "reserved" | "sold";
  balcony?: boolean;
  bathroom_type?: string;
  bathroom_count?: number;
  plan_image?: string;
  description?: string;
  client_name?: string;
  client_phone?: string;
}

export interface SimCard {
  id: number;
  phone_number: string;
  operator: string;
  assigned_to?: number;
  assigned_name?: string;
  phone_id?: number;
  phone_model?: string;
  description?: string;
  status: "active" | "blocked";
  created_at: string;
}

export interface CompanyPhone {
  id: number;
  model: string;
  phone_id: string;
  assigned_to?: number;
  assigned_name?: string;
  description?: string;
  status: string;
}

export interface SimTariff {
  id: number;
  sim_id: number;
  minutes: number;
  gb: number;
  sms: number;
  cost: number;
  start_date?: string;
  end_date?: string;
  status: string;
}

export interface Post {
  id: number;
  title: string;
  description?: string;
  category?: string;
  project?: string;
  user_id?: number;
  user_name?: string;
  media_path?: string;
  media_type?: string;
  likes: number;
  comments: number;
  shares: number;
  views: number;
  reach: number;
  is_published: boolean;
  created_at: string;
}

export interface Message {
  id: number;
  user_id?: number;
  user_name?: string;
  message: string;
  file_path?: string;
  file_name?: string;
  file_type?: string;
  created_at: string;
}

export interface Folder {
  id: number;
  name: string;
  parent_id: number;
  author_name?: string;
}

export interface FolderFile {
  id: number;
  folder_id: number;
  filename: string;
  original_name: string;
  filepath: string;
  filetype: string;
  filesize: number;
  author_name?: string;
}

export interface Bank {
  id: number;
  name: string;
  slug: string;
  logo?: string;
  description?: string;
  is_active: boolean;
  order_index: number;
}

export interface MortgageCondition {
  id: number;
  bank_id: number;
  currency: string;
  interest_yearly: number;
  interest_monthly: number;
  min_amount: number;
  max_amount: number;
  min_months: number;
  max_months: number;
  down_payment_percent: number;
  guarantor_required: boolean;
  collateral_required: boolean;
  extra_conditions?: string;
}

export interface InstallmentObject {
  id: number;
  name: string;
  slug: string;
  image?: string;
  description?: string;
  address?: string;
  developer?: string;
  is_active: boolean;
  order_index: number;
}

export interface DashboardStats {
  users: number;
  requests: number;
  requests_open: number;
  leads: number;
  tasks: number;
  my_tasks: number;
  sim_cards: number;
  apartments: number;
  apartments_free: number;
  apartments_sold: number;
  leads_7d: { day: string; count: number }[];
  requests_7d: { day: string; count: number }[];
}
