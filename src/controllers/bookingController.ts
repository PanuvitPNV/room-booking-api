import { Request, Response } from 'express';
import { BookingService } from '../services/bookingService';
import { logger } from '../utils/logger';

const bookingService = new BookingService();

export const createBooking = async (req: Request, res: Response) => {
  try {
    const result = await bookingService.createBooking(req.body);
    res.status(201).json({
      success: true,
      data: result,
      message: 'Booking created successfully'
    });
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    logger.error('Failed to create booking', { error: errorMessage });
    res.status(400).json({
      success: false,
      message: errorMessage
    });
  }
};

export const getBooking = async (req: Request, res: Response) => {
  try {
    const bookingId = parseInt(req.params.id);
    const booking = await bookingService.getBookingDetails(bookingId);
    
    if (!booking) {
      return res.status(404).json({
        success: false,
        message: 'Booking not found'
      });
    }

    res.json({
      success: true,
      data: booking
    });
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    logger.error('Failed to get booking', { error: errorMessage });
    res.status(500).json({
      success: false,
      message: errorMessage
    });
  }
};

export const cancelBooking = async (req: Request, res: Response) => {
  try {
    const bookingId = parseInt(req.params.id);
    await bookingService.cancelBooking(bookingId);
    
    res.json({
      success: true,
      message: 'Booking cancelled successfully'
    });
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    logger.error('Failed to cancel booking', { error: errorMessage });
    res.status(400).json({
      success: false,
      message: errorMessage
    });
  }
};

export const setRowLocking = async (req: Request, res: Response) => {
  try {
    const { enabled } = req.body;
    bookingService.setRowLocking(enabled);
    
    res.json({
      success: true,
      message: `Row locking ${enabled ? 'enabled' : 'disabled'}`
    });
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    logger.error('Failed to set row locking', { error: errorMessage });
    res.status(500).json({
      success: false,
      message: errorMessage
    });
  }
};