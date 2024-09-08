package api

import (
	pb "ticket-booking-app/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	_ "github.com/mattn/go-sqlite3"
	"github.com/google/uuid"
	"database/sql"
	"context"
	"log"
)

type BookingService struct {
	seatAllocator *SeatAllocator
	db            *sql.DB
}

func NewBookingService(dbInstance *sql.DB) *BookingService {
	return &BookingService{
		seatAllocator: NewSeatAllocator(),
		db: dbInstance,
	}
}

func (b *BookingService) CreateBooking(ctx context.Context, req *pb.BookingRequest) (*pb.BookingResponse, error) {
	if req.GetFrom() == "" || req.GetTo() == "" || req.GetPrice() == 0 || req.GetUser() == nil {
	    log.Fatalf("Invalid create request : %v", req)
		return nil, status.Errorf(codes.InvalidArgument, "Invalid create booking request")
	}

    //check if user exists before inserting new record
    var userId string
    dbUser, isUserExists := retrieveUserIfExists(b, req.GetUser().GetFirstname(), req.GetUser().GetLastname(), req.GetUser().GetEmail())
    if isUserExists {
        userId = dbUser.GetId()
    } else {
        userId = uuid.NewString()
        _, userDbErr := b.db.Exec("INSERT INTO users (u_id, u_user_fname, u_user_lname, u_user_email) VALUES (?, ?, ?, ?)",
                    userId, req.GetUser().GetFirstname(), req.GetUser().GetLastname(), req.GetUser().GetEmail())
        if userDbErr != nil {
            return nil, userDbErr
        }
        log.Printf("Added new user with email %v \n", req.GetUser().GetEmail())
    }

    //check if booking already exists for user with requested location details
    dbBooking, isBookingExists := retrieveBookingIfExists(b, userId, req.GetFrom(), req.GetTo())
    if isBookingExists {
        log.Printf("Ticket already exsits from %s to %s for user %s with seat number %d\n", dbBooking.GetFrom(),
            dbBooking.GetTo(), userId, dbBooking.GetSeat())
        return nil, nil
    }
    seat, section, err := b.seatAllocator.AllocateSeat()
    if err != nil {
    	return nil, status.Errorf(codes.Internal, "Error while allocating seat: %v", err)
    }

    ticketId := uuid.NewString()
    _, dbErr := b.db.Exec("INSERT INTO tickets (t_id, t_from, t_to, t_price, t_seat, t_section, t_user_id) VALUES (?, ?, ?, ?, ?, ?, ?)",
           ticketId, req.GetFrom(), req.GetTo(), req.GetPrice(), seat, section, userId)
    if dbErr != nil {
        return nil, dbErr
    }
    log.Printf("Booked new ticket from %s to %s for user %s with seat number %v\n", req.From,
         req.To, req.GetUser().GetEmail(), seat)
    return nil, nil
}

func (b *BookingService) GetBookingByUser(ctx context.Context, req *pb.GetBookingByUserRequest) (*pb.BookingResponse, error) {
	booking, _ := b.getBookingFromDataByUser(ctx, req.GetUser())
	if booking == nil {
		return nil, status.Errorf(codes.NotFound, "No booking exists with email %s", req.GetUser().GetEmail())
	}
	return transformAsBookingResponse(booking), nil
}

func (b *BookingService) GetBookingsBySection(ctx context.Context, req *pb.GetBookingsBySectionRequest) (*pb.BookingListResponse, error) {
    var bookings []*pb.BookingResponse
    rows, err := b.db.Query("SELECT * FROM tickets WHERE t_section = ?", req.GetSection())
     if err != nil {
        log.Fatalf("List :: ticket error : %v", err)
         return nil, err
     }
    defer rows.Close()

    for rows.Next() {
         var response pb.BookingDbResponse
         //log.Printf("List :: indv row : %v", rows)
         if err := rows.Scan(&response.Id, &response.From, &response.To, &response.Price, &response.Seat, &response.Section, &response.Userid); err != nil {
             log.Fatalf("List :: error : %v", err)
             return nil, err
         }

         var dbUser pb.User
         dbRow := b.db.QueryRow("SELECT * FROM users WHERE u_id = ?", response.GetUserid())
         if dbErr := dbRow.Scan(&dbUser.Id , &dbUser.Firstname, &dbUser.Lastname, &dbUser.Email); dbErr != nil {
             log.Fatalf("List :: Error in retrieving user of id : %s, Error : %v", dbUser.Id, dbErr)
             return nil, dbErr
         }
         bookings = append(bookings, transformDbResponseToBookingResponse(response, dbUser))
    }

	return &pb.BookingListResponse{Bookings: bookings}, nil
}

func (b *BookingService) RemoveBookingByUser(ctx context.Context, req *pb.RemoveBookingByUserRequest) (*pb.RemoveBookingResponse, error) {
	booking, _ := b.getBookingFromDataByUser(ctx, req.GetUser())
	if booking == nil {
		return nil, status.Errorf(codes.NotFound, "No booking exists with email %s", req.GetUser().GetEmail())
	}

	b.seatAllocator.DeallocateSeat(booking.GetSeat(), booking.GetSection())

	_, err := b.db.Exec("DELETE FROM tickets WHERE t_id = ?", booking.GetId())
    if err != nil {
        return nil, err
    }
    log.Printf("Cancelled booking for user %v \n", req.GetUser().GetEmail())

	return &pb.RemoveBookingResponse{}, nil
}

func (b *BookingService) ModifySeatByUser(ctx context.Context, req *pb.SeatModificationRequest) (*pb.SeatModificationResponse, error) {
    log.Printf("Received booking modification request for user %v \n", req.GetUser().GetEmail())

	booking, _ := b.getBookingFromDataByUser(ctx, req.GetUser())
	if booking == nil {
		return nil, status.Errorf(codes.NotFound, "No booking exists with email %s", req.GetUser().GetEmail())
	}

	if booking.GetSection() == req.GetSection() && booking.GetSeat() == req.GetSeat() {
		return nil, status.Errorf(codes.InvalidArgument, "Old and new seats can't be same")
	}

	allocationErr := b.seatAllocator.AllocateSpecificSeat(req.GetSeat(), req.GetSection())
	if allocationErr != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Seat number %d in section %s is not available", req.GetSeat(), req.GetSection())
	}

	booking.Seat = req.GetSeat()
	booking.Section = req.GetSection()
	return &pb.SeatModificationResponse{Seat: booking.GetSeat(), Section: booking.GetSection(), User: booking.GetUser()}, nil
}

func (b *BookingService) getBookingFromDataByUser(ctx context.Context, user *pb.User) (*pb.Booking, int) {
    var dbUser pb.User
    dbRow := b.db.QueryRow("SELECT * FROM users WHERE u_user_email = ?", user.GetEmail())
    if dbErr := dbRow.Scan(&dbUser.Id , &dbUser.Firstname, &dbUser.Lastname, &dbUser.Email); dbErr != nil {
        log.Fatalf("Error in retrieving user : %s, Error : %v", dbUser.Id, dbErr)
        return nil, -1
    }

    //log.Printf("User response : %v , %s, %s, %s \n", dbUser.Id, dbUser.Firstname, dbUser.Lastname, dbUser.Email)
    var response pb.BookingDbResponse
    row := b.db.QueryRow("SELECT * FROM tickets WHERE t_user_id = ?", dbUser.GetId())
    if err := row.Scan(&response.Id, &response.From, &response.To, &response.Price, &response.Seat, &response.Section, &response.Userid); err != nil {
        log.Fatalf("Error in retrieving ticket for user of id : %s, Error : %v", dbUser.Id, err)
        return nil, -1
    }

    return transformDbResponseToBooking(response, dbUser), -1
}

func retrieveUserIfExists(b *BookingService, firstName, lastName, userEmail string) (*pb.User, bool) {
    var dbUser pb.User
    dbRow := b.db.QueryRow("SELECT * FROM users WHERE u_user_fname = ? and u_user_lname = ? and u_user_email = ?", firstName, lastName, userEmail)
    if dbErr := dbRow.Scan(&dbUser.Id , &dbUser.Firstname, &dbUser.Lastname, &dbUser.Email); dbErr != nil {
       return nil, false
    }

    return &dbUser, true
}

func retrieveBookingIfExists(b *BookingService, userId, fromLocation, toLocation string) (*pb.BookingDbResponse, bool){
    var response pb.BookingDbResponse
    row := b.db.QueryRow("SELECT * FROM tickets WHERE t_user_id = ? and t_from = ? and t_to = ?", userId, fromLocation, toLocation)
    if err := row.Scan(&response.Id, &response.From, &response.To, &response.Price, &response.Seat, &response.Section, &response.Userid); err != nil {
       return nil, false
    }

    return &response, true
}

func transformAsBookingResponse(booking *pb.Booking) *pb.BookingResponse {
	return &pb.BookingResponse{
		Id:      booking.GetId(),
		From:    booking.GetFrom(),
		To:      booking.GetTo(),
		Price:   booking.GetPrice(),
		Seat:    booking.GetSeat(),
		Section: booking.GetSection(),
		User:    booking.GetUser(),
	}
}

func transformDbResponseToBooking(bookingDbResp pb.BookingDbResponse, userDbResp pb.User) *pb.Booking {
	return &pb.Booking{
		Id:      bookingDbResp.GetId(),
		From:    bookingDbResp.GetFrom(),
		To:      bookingDbResp.GetTo(),
		Price:   bookingDbResp.GetPrice(),
		Seat:    bookingDbResp.GetSeat(),
		Section: bookingDbResp.GetSection(),
        User: &pb.User{
        	Firstname: userDbResp.GetFirstname(),
        	Lastname:  userDbResp.GetLastname(),
        	Email:     userDbResp.GetEmail(),
        	},
    }
}

func transformDbResponseToBookingResponse(bookingDbResp pb.BookingDbResponse, userDbResp pb.User) *pb.BookingResponse {
	return &pb.BookingResponse{
		Id:      bookingDbResp.GetId(),
		From:    bookingDbResp.GetFrom(),
		To:      bookingDbResp.GetTo(),
		Price:   bookingDbResp.GetPrice(),
		Seat:    bookingDbResp.GetSeat(),
		Section: bookingDbResp.GetSection(),
        User: &pb.User{
        	Firstname: userDbResp.GetFirstname(),
        	Lastname:  userDbResp.GetLastname(),
        	Email:     userDbResp.GetEmail(),
        	},
    }
}