package api

import (
	pb "ticket-booking-app/domain"
    "ticket-booking-app/server/api"
	"github.com/stretchr/testify/assert"
	"testing"
	"context"
)

const(
    FIRST_NAME = "testFirstName"
    LAST_NAME = "testLastName"
    EMAIL = "testEmail@test.com"
)

func TestShouldCreateTrainBooking(t *testing.T) {
    request := createNewTrainBookingRequest(FIRST_NAME, LAST_NAME, EMAIL)
	response := createTrainBookingResponse(FIRST_NAME, LAST_NAME, EMAIL)
	booking := createMockTrainBooking(request)

    assert.Equal(t, response.GetUser(), booking.GetUser(), "Booking creation is not working as expected")
}

func TestShouldReturnBookingByUser(t *testing.T) {
	request := createNewTrainBookingRequest(FIRST_NAME, LAST_NAME, EMAIL)

    bookingService := api.NewBookingService()
    _, err := bookingService.CreateBooking(context.TODO(), request)
    if err != nil {
       	t.Errorf("Error in creating booking %v ", err)
        return
    }

	reqGetBookingByUser := &pb.GetBookingByUserRequest{User: request.GetUser()}
	response := createTrainBookingResponse(FIRST_NAME, LAST_NAME, EMAIL)

	got, err := bookingService.GetBookingByUser(context.TODO(), reqGetBookingByUser)
	if err != nil {
		t.Errorf("Error in retrieving booking %v ", err)
		return
	}

    assert.Equal(t, response.GetUser(), got.GetUser(), "Can't retrieve existing booking")
}

func createMockTrainBooking(request *pb.BookingRequest) (*pb.BookingResponse) {
    bookingService := api.NewBookingService()
    booking, err := bookingService.CreateBooking(context.TODO(), request)
    if err != nil {
    	return nil
    }

    return booking
}

func createTrainBookingResponse(fName, lName, email string) (*pb.BookingResponse) {
    return &pb.BookingResponse{
    		From:  "London",
    		To:    "France",
    		Price: 20,
    		User: &pb.User{
    			Firstname: fName,
    			Lastname:  lName,
    			Email:     email,
    		},
    	}
}

func createNewTrainBookingRequest(fName, lName, email string) (*pb.BookingRequest) {
    return &pb.BookingRequest{
           		From:  "London",
           		To:    "France",
           		Price: 20,
           		User: &pb.User{
           			Firstname: fName,
           			Lastname:  lName,
           			Email:     email,
           		},
           	};
}
