/**
 * Hotel Booking System - Main Application Logic
 * Using Alpine.js for reactivity
 */
function hotelBookingApp() {
    return {
        // UI State
        activeTab: 'rooms',
        isLoading: false,
        notification: {
            show: false,
            message: '',
            type: 'alert-info'
        },

        // Data
        rooms: [],
        roomTypes: [],
        facilities: {
            1: "Wi-Fi",
            2: "Air Conditioning",
            3: "TV",
            4: "Mini Bar",
            5: "Coffee Machine",
            6: "Balcony",
            7: "Bathtub",
            8: "Kitchen"
        },
        selectedRoomType: null,
        selectedRoom: null,
        availableRooms: [],
        currentBooking: null,
        currentReceipt: null,
        confirmedBooking: null,

        // Forms
        bookingForm: {
            booking_name: '',
            room_num: '',
            check_in_date: '',
            check_out_date: ''
        },
        lookupForm: {
            booking_id: ''
        },
        paymentForm: {
            booking_id: '',
            payment_method: '',
            amount: 0
        },
        modifyForm: {
            check_in_date: '',
            check_out_date: ''
        },

        // Lifecycle
        init() {
            // Fetch data on initialization
            this.fetchRooms();
            this.fetchRoomTypes();

            // Check if room ID is in URL params and select it
            const urlParams = new URLSearchParams(window.location.search);
            const roomId = urlParams.get('roomId');
            if (roomId) {
                this.activeTab = 'booking';
                this.bookingForm.room_num = parseInt(roomId);
            }
        },

        // UI Methods
        setActiveTab(tab) {
            this.activeTab = tab;
        },

        toggleTheme() {
            const html = document.querySelector('html');
            html.dataset.theme = html.dataset.theme === 'dark' ? 'light' : 'dark';
        },

        showNotification(message, type = 'alert-info') {
            this.notification = {
                show: true,
                message,
                type
            };

            // Auto-hide after 4 seconds
            setTimeout(() => {
                this.notification.show = false;
            }, 4000);
        },

        // Room Methods
        fetchRooms() {
            this.isLoading = true;
            fetch('/api/v1/rooms')
                .then(response => {
                    if (!response.ok) throw new Error('Failed to fetch rooms');
                    return response.json();
                })
                .then(data => {
                    this.rooms = data;
                    this.isLoading = false;
                })
                .catch(error => {
                    console.error('Error fetching rooms:', error);
                    this.showNotification('Failed to load rooms. Please try again.', 'alert-error');
                    this.isLoading = false;
                });
        },

        fetchRoomTypes() {
            fetch('/api/v1/rooms/types')
                .then(response => {
                    if (!response.ok) throw new Error('Failed to fetch room types');
                    return response.json();
                })
                .then(data => {
                    this.roomTypes = data;
                })
                .catch(error => {
                    console.error('Error fetching room types:', error);
                    this.showNotification('Failed to load room types. Please try again.', 'alert-error');
                });
        },

        viewRoomDetails(room) {
            this.selectedRoom = room;
            document.getElementById('roomDetailsModal').showModal();
        },

        selectRoomForBooking(room) {
            this.bookingForm.room_num = room.room_num;
            this.setActiveTab('booking');
            document.getElementById('roomDetailsModal').close();
        },

        getRoomImage(roomType) {
            // Map room types to image URLs
            // In a real app, these would be URLs to actual room images
            const images = {
                'Standard': '/static/img/standard-room.png',
                'Deluxe': '/static/img/deluxe-room.png',
                'Suite': '/static/img/suite-room.png',
                'Family Room': '/static/img/family-room.png'
            };

            // Return the specific image or a default
            return images[roomType] || '/static/img/default-room.png';
        },

        get filteredRooms() {
            if (this.selectedRoomType === null) {
                return this.rooms;
            }
            return this.rooms.filter(room => room.type_id === this.selectedRoomType);
        },

        getFacilityName(facilityId) {
            return this.facilities[facilityId] || 'Facility';
        },

        // Booking Methods
        getTodayDate() {
            return new Date().toISOString().split('T')[0];
        },

        getMinCheckoutDate() {
            if (!this.bookingForm.check_in_date) {
                return this.getTodayDate();
            }

            const checkInDate = new Date(this.bookingForm.check_in_date);
            const nextDay = new Date(checkInDate);
            nextDay.setDate(nextDay.getDate() + 1);
            return nextDay.toISOString().split('T')[0];
        },

        getMinModifyCheckoutDate() {
            if (!this.modifyForm.check_in_date) {
                return this.getTodayDate();
            }

            const checkInDate = new Date(this.modifyForm.check_in_date);
            const nextDay = new Date(checkInDate);
            nextDay.setDate(nextDay.getDate() + 1);
            return nextDay.toISOString().split('T')[0];
        },

        checkAvailability() {
            if (!this.bookingForm.check_in_date || !this.bookingForm.check_out_date) {
                return;
            }
            this.isLoading = true;

            // Format dates for API request
            const requestData = {
                check_in_date: new Date(this.bookingForm.check_in_date).toISOString(),
                check_out_date: new Date(this.bookingForm.check_out_date).toISOString()
            };

            fetch('/api/v1/rooms/available', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(requestData)
            })
                .then(response => {
                    if (!response.ok) throw new Error('Failed to fetch available rooms');
                    return response.json();
                })
                .then(data => {
                    this.availableRooms = data;
                    this.isLoading = false;

                    // Clear the room selection if previously selected room is no longer available
                    if (this.bookingForm.room_num) {
                        const stillAvailable = this.availableRooms.some(room => room.room_num === parseInt(this.bookingForm.room_num));
                        if (!stillAvailable) {
                            this.bookingForm.room_num = '';
                        }
                    }
                })
                .catch(error => {
                    console.error('Error checking availability:', error);
                    this.showNotification('Failed to check room availability. Please try again.', 'alert-error');
                    this.isLoading = false;
                });
        },

        getSelectedRoom() {
            if (!this.bookingForm.room_num) return null;
            const roomNum = parseInt(this.bookingForm.room_num);
            return this.availableRooms.find(room => room.room_num === roomNum);
        },

        calculateNights() {
            if (!this.bookingForm.check_in_date || !this.bookingForm.check_out_date) {
                return 0;
            }

            const checkIn = new Date(this.bookingForm.check_in_date);
            const checkOut = new Date(this.bookingForm.check_out_date);
            const diffTime = Math.abs(checkOut - checkIn);
            return Math.ceil(diffTime / (1000 * 60 * 60 * 24));
        },

        calculateTotalPrice() {
            const selectedRoom = this.getSelectedRoom();
            if (!selectedRoom) return 0;

            const nights = this.calculateNights();
            return selectedRoom.room_type.price_per_night * nights;
        },

        isBookingFormValid() {
            return this.bookingForm.booking_name &&
                this.bookingForm.room_num &&
                this.bookingForm.check_in_date &&
                this.bookingForm.check_out_date &&
                this.calculateNights() > 0;
        },

        createBooking() {
            if (!this.isBookingFormValid()) return;

            this.isLoading = true;

            // Format dates for API request
            const bookingData = {
                booking_name: this.bookingForm.booking_name,
                room_num: parseInt(this.bookingForm.room_num),
                check_in_date: new Date(this.bookingForm.check_in_date).toISOString(),
                check_out_date: new Date(this.bookingForm.check_out_date).toISOString()
            };

            fetch('/api/v1/bookings', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(bookingData)
            })
                .then(response => {
                    if (!response.ok) throw new Error('Failed to create booking');
                    return response.json();
                })
                .then(data => {
                    this.isLoading = false;
                    this.confirmedBooking = data;

                    // Show confirmation modal
                    document.getElementById('bookingConfirmModal').showModal();

                    // Reset form
                    this.bookingForm = {
                        booking_name: '',
                        room_num: '',
                        check_in_date: '',
                        check_out_date: ''
                    };

                    this.showNotification('Booking created successfully!', 'alert-success');
                })
                .catch(error => {
                    console.error('Error creating booking:', error);
                    this.showNotification('Failed to create booking. Please try again.', 'alert-error');
                    this.isLoading = false;
                });
        },

        goToPayment() {
            if (!this.confirmedBooking) return;

            // Close the confirmation modal
            document.getElementById('bookingConfirmModal').close();

            // Switch to manage tab and set up payment form
            this.setActiveTab('manage');
            this.lookupForm.booking_id = this.confirmedBooking.booking_id;
            this.lookupBooking();
        },

        // Manage Booking Methods
        lookupBooking() {
            if (!this.lookupForm.booking_id) return;

            this.isLoading = true;

            fetch(`/api/v1/bookings/${this.lookupForm.booking_id}`)
                .then(response => {
                    if (!response.ok) throw new Error('Booking not found');
                    return response.json();
                })
                .then(data => {
                    this.currentBooking = data;

                    // Try to get receipt if it exists
                    this.fetchReceipt(data.booking_id);

                    // Pre-fill modification form
                    this.modifyForm = {
                        check_in_date: new Date(data.check_in_date).toISOString().split('T')[0],
                        check_out_date: new Date(data.check_out_date).toISOString().split('T')[0]
                    };

                    // Set up payment form
                    this.paymentForm = {
                        booking_id: data.booking_id,
                        payment_method: '',
                        amount: data.total_price
                    };

                    this.isLoading = false;
                })
                .catch(error => {
                    console.error('Error looking up booking:', error);
                    this.showNotification('Booking not found. Please check your booking ID.', 'alert-error');
                    this.isLoading = false;
                });
        },

        fetchReceipt(bookingId) {
            fetch(`/api/v1/receipts/booking/${bookingId}`)
                .then(response => {
                    if (!response.ok) {
                        if (response.status === 404) {
                            // No receipt found, which is OK
                            this.currentReceipt = null;
                            return;
                        }
                        throw new Error('Failed to fetch receipt');
                    }
                    return response.json();
                })
                .then(data => {
                    if (data) {
                        this.currentReceipt = data;
                    }
                })
                .catch(error => {
                    console.error('Error fetching receipt:', error);
                    // Don't show notification as this is expected for unpaid bookings
                });
        },

        processPayment() {
            if (!this.paymentForm.payment_method || !this.currentBooking) return;

            this.isLoading = true;

            const paymentData = {
                booking_id: this.currentBooking.booking_id,
                payment_method: this.paymentForm.payment_method,
                amount: this.currentBooking.total_price
            };

            fetch('/api/v1/receipts', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(paymentData)
            })
                .then(response => {
                    if (!response.ok) throw new Error('Failed to process payment');
                    return response.json();
                })
                .then(data => {
                    this.currentReceipt = data;
                    this.isLoading = false;
                    this.showNotification('Payment processed successfully!', 'alert-success');
                })
                .catch(error => {
                    console.error('Error processing payment:', error);
                    this.showNotification('Failed to process payment. Please try again.', 'alert-error');
                    this.isLoading = false;
                });
        },

        confirmCancelBooking() {
            if (!this.currentBooking) return;
            document.getElementById('cancelConfirmModal').showModal();
        },

        cancelBooking() {
            if (!this.currentBooking) return;

            this.isLoading = true;

            fetch(`/api/v1/bookings/${this.currentBooking.booking_id}`, {
                method: 'DELETE'
            })
                .then(response => {
                    if (!response.ok) throw new Error('Failed to cancel booking');
                    return response.json();
                })
                .then(() => {
                    document.getElementById('cancelConfirmModal').close();
                    this.isLoading = false;
                    this.showNotification('Booking cancelled successfully.', 'alert-success');
                    this.currentBooking = null;
                    this.currentReceipt = null;
                })
                .catch(error => {
                    console.error('Error cancelling booking:', error);
                    this.showNotification('Failed to cancel booking. Please try again.', 'alert-error');
                    this.isLoading = false;
                });
        },

        confirmRefund() {
            if (!this.currentBooking || !this.currentReceipt) return;
            document.getElementById('refundConfirmModal').showModal();
        },

        processRefund() {
            if (!this.currentBooking || !this.currentReceipt) return;

            this.isLoading = true;

            const refundData = {
                booking_id: this.currentBooking.booking_id
            };

            fetch('/api/v1/receipts/refund', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(refundData)
            })
                .then(response => {
                    if (!response.ok) throw new Error('Failed to process refund');
                    return response.json();
                })
                .then(() => {
                    document.getElementById('refundConfirmModal').close();
                    this.isLoading = false;
                    this.showNotification('Refund processed successfully.', 'alert-success');
                    this.currentBooking = null;
                    this.currentReceipt = null;
                })
                .catch(error => {
                    console.error('Error processing refund:', error);
                    this.showNotification('Failed to process refund. Please try again.', 'alert-error');
                    this.isLoading = false;
                });
        },

        modifyBooking() {
            if (!this.currentBooking) return;
            document.getElementById('modifyBookingModal').showModal();
        },

        updateBookingDates() {
            if (!this.currentBooking || !this.modifyForm.check_in_date || !this.modifyForm.check_out_date) return;

            this.isLoading = true;

            const updateData = {
                check_in_date: new Date(this.modifyForm.check_in_date).toISOString(),
                check_out_date: new Date(this.modifyForm.check_out_date).toISOString()
            };

            fetch(`/api/v1/bookings/${this.currentBooking.booking_id}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(updateData)
            })
                .then(response => {
                    if (!response.ok) throw new Error('Failed to update booking');
                    return response.json();
                })
                .then(() => {
                    document.getElementById('modifyBookingModal').close();
                    this.isLoading = false;
                    this.showNotification('Booking updated successfully.', 'alert-success');

                    // Refresh booking details
                    this.lookupBooking();
                })
                .catch(error => {
                    console.error('Error updating booking:', error);
                    this.showNotification('Failed to update booking. Please try again.', 'alert-error');
                    this.isLoading = false;
                });
        },

        // Helper Methods
        formatDate(dateString) {
            if (!dateString) return '';
            const options = { year: 'numeric', month: 'long', day: 'numeric' };
            return new Date(dateString).toLocaleDateString(undefined, options);
        },

        formatDateTime(dateString) {
            if (!dateString) return '';
            const options = {
                year: 'numeric',
                month: 'long',
                day: 'numeric',
                hour: '2-digit',
                minute: '2-digit'
            };
            return new Date(dateString).toLocaleDateString(undefined, options);
        }
    };
}
