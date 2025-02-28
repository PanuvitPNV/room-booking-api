const moment = require('moment');

// Helper function to generate random room numbers
function getRandomRoom() {
    return Math.floor(Math.random() * 20) + 101; // Rooms 101-120
}

// Helper function to generate random guest IDs
function getRandomGuest() {
    return Math.floor(Math.random() * 50) + 1; // Guest IDs 1-50
}

// Helper function to generate random future dates
function getRandomFutureDates() {
    const startDate = moment().add(1, 'days');
    const endDate = moment(startDate).add(Math.floor(Math.random() * 5) + 1, 'days');

    return {
        checkIn: startDate.format('YYYY-MM-DD'),
        checkOut: endDate.format('YYYY-MM-DD')
    };
}

// Generate booking data for a new reservation
function generateBookingData(userContext, events, done) {
    const dates = getRandomFutureDates();

    userContext.vars.bookingData = {
        room_num: getRandomRoom(),
        guest_id: getRandomGuest(),
        check_in_date: dates.checkIn,
        check_out_date: dates.checkOut
    };

    return done();
}

// Generate data for availability check
function generateAvailabilityCheckData(userContext, events, done) {
    const dates = getRandomFutureDates();

    userContext.vars.availabilityData = {
        room_num: getRandomRoom(),
        check_in_date: dates.checkIn,
        check_out_date: dates.checkOut
    };

    return done();
}

// Generate data for double booking test
function generateDoubleBookingData(userContext, events, done) {
    const dates = getRandomFutureDates();
    const roomNum = getRandomRoom();

    userContext.vars.booking1Data = {
        room_num: roomNum,
        guest_id: getRandomGuest(),
        check_in_date: dates.checkIn,
        check_out_date: dates.checkOut
    };

    userContext.vars.booking2Data = {
        room_num: roomNum,
        guest_id: getRandomGuest(),
        check_in_date: dates.checkIn,
        check_out_date: dates.checkOut
    };

    userContext.vars.roomNum = roomNum;

    return done();
}

// Generate invalid booking data for transaction rollback testing
function generateInvalidBookingData(userContext, events, done) {
    const dates = getRandomFutureDates();
    const roomNum = getRandomRoom();

    userContext.vars.invalidBookingData = {
        room_num: roomNum,
        guest_id: 999999, // Non-existent guest ID
        check_in_date: dates.checkIn,
        check_out_date: dates.checkOut
    };

    userContext.vars.roomNum = roomNum;

    return done();
}

module.exports = {
    generateBookingData,
    generateAvailabilityCheckData,
    generateDoubleBookingData,
    generateInvalidBookingData
};