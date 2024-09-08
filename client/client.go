package main

import (
	pb "ticket-booking-app/domain"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc"
	"context"
	"log"
	"time"
)

var bookingClient pb.BookingServiceClient
var serverContext context.Context

func main() {

	serverConn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect server: localhost:50051: %v", err)
	}
	defer serverConn.Close()
	bookingClient = pb.NewBookingServiceClient(serverConn)

    var cancel context.CancelFunc
	serverContext, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()

    // List of bookings by section
    log.Printf("\n ***********************************\n")
    retrieveTrainBookingBySection("A")
    retrieveTrainBookingBySection("B")
    log.Printf("***********************************\n")

    //create new train bookings
	createNewTrainBooking("vrushali", "ghadge", "vg@gmail.com")
	createNewTrainBooking("vikram", "ghadge", "vkg@gmail.com")

	// Modify seat for user1
    updateExistingTrainBooking("A", 19, retrieveTrainBookingByUser("vrushali", "ghadge", "vg@gmail.com").GetUser())

	// Remove user2 booking
    cancelExistingTrainBooking(retrieveTrainBookingByUser("vikram", "ghadge", "vkg@gmail.com").GetUser())

    // List of bookings by section
    log.Printf("\n ***********************************\n")
    retrieveTrainBookingBySection("A")
    retrieveTrainBookingBySection("B")
     log.Printf("***********************************\n")
}

func cancelExistingTrainBooking(user *pb.User) {
    cancelBookingRequest := &pb.RemoveBookingByUserRequest{User: user}

    _, cancelBookingErr := bookingClient.RemoveBookingByUser(serverContext, cancelBookingRequest)
    if cancelBookingErr != nil {
    	log.Fatalf("Error while cancelling booking : %v", cancelBookingErr)
    }

    log.Printf("\nBooking for user %s successfully cancelled", user.GetEmail())
}

func updateExistingTrainBooking(sectionTitle string, seatNumber int32, user *pb.User) {
    updateSeatRequest := &pb.SeatModificationRequest{
    		Section: sectionTitle,
    		Seat:    seatNumber,
    		User:    user,
    	}
    updatedSeat, updateSeatErr := bookingClient.ModifySeatByUser(serverContext, updateSeatRequest)
    if updateSeatErr != nil {
    	log.Fatalf("Error in updating the booking : %v", updateSeatErr)
    }

   	log.Printf("\nUpdated booking of user %s successfully with seat %s",  user.GetEmail(), updatedSeat)
}

func retrieveTrainBookingBySection(sectionTitle string) (*pb.BookingListResponse) {
    getBookingsBySectionRequest := &pb.GetBookingsBySectionRequest{
    		Section: sectionTitle,
    	}

    bookingsBySection, bookingsBySectionErr := bookingClient.GetBookingsBySection(serverContext, getBookingsBySectionRequest)
    if bookingsBySectionErr != nil {
    	log.Fatalf("Error in retrieving the bookings : %v", bookingsBySectionErr)
    }

    log.Printf("\nRetrieved booked user for section %s, Details : %s", sectionTitle, bookingsBySection)
    return bookingsBySection
}

func retrieveTrainBookingByUser(fName, lName, email string) (*pb.BookingResponse) {
    getBookingByUserRequest := &pb.GetBookingByUserRequest{
    		User: &pb.User{
    			Firstname: fName,
    			Lastname:  lName,
    			Email:     email,
    		},
    	}

    getBookingByUser, getBookingErr := bookingClient.GetBookingByUser(serverContext, getBookingByUserRequest)
    if getBookingErr != nil {
    	log.Fatalf("Error in retrieving the booking: %v", getBookingErr)
    }

    log.Printf("\nBooked user detail: %s", getBookingByUser)
    return getBookingByUser
}

func createNewTrainBooking(fName, lName, email string) {
    userRequest := createNewTrainBookingRequest(fName, lName, email)
    bookingResponse, bookingErr := bookingClient.CreateBooking(serverContext, userRequest)
    if bookingErr != nil {
        log.Fatalf("Error in creating train booking request : %v , Error: %v", bookingErr, userRequest)
    }

    log.Printf("\nBooking completed Successfully, For user %s", bookingResponse.User.GetEmail())
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
