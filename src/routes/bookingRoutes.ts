import { Router } from 'express';
import { createBooking, getBooking, cancelBooking, setRowLocking } from '../controllers/bookingController';

const router = Router();

router.post('/bookings', createBooking);
router.get('/bookings/:id', getBooking);
router.delete('/bookings/:id', cancelBooking);
router.post('/settings/row-locking', setRowLocking);

export default router;