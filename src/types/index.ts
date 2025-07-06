export interface Room {
  id: number;
  room_number: string;
  room_type: string;
  price_per_night: number;
  is_available: boolean;
  created_at: Date;
  updated_at: Date;
}

export interface Guest {
  id: number;
  name: string;
  email: string;
  phone: string;
  created_at: Date;
  updated_at: Date;
}

export interface Booking {
  id: number;
  guest_id: number;
  room_id: number;
  check_in_date: Date;
  check_out_date: Date;
  total_amount: number;
  status: 'pending' | 'confirmed' | 'cancelled';
  created_at: Date;
  updated_at: Date;
}

export interface Payment {
  id: number;
  booking_id: number;
  amount: number;
  payment_method: string;
  status: 'pending' | 'completed' | 'failed';
  transaction_id: string;
  created_at: Date;
  updated_at: Date;
}

export interface Receipt {
  id: number;
  booking_id: number;
  payment_id: number;
  receipt_number: string;
  total_amount: number;
  generated_at: Date;
}