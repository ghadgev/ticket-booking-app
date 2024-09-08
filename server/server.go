package main

import (
	pb "ticket-booking-app/domain"
    "ticket-booking-app/server/api"
    "google.golang.org/grpc"
    _ "github.com/mattn/go-sqlite3"
	"database/sql"
	"net"
	"log"
)

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Error in listening on port 50051: %v", err)
	}

    db, err := sql.Open("sqlite3", "./ticket_booking.db")
    if err != nil {
        log.Fatalf("Failed to open database: %v", err)
    }
    defer db.Close()
    createDatabaseTables(db)

    // Start server and register the all APIs
	server := grpc.NewServer()

	bookingService := api.NewBookingService(db)
	pb.RegisterBookingServiceServer(server, bookingService)

	log.Printf("Server started successfully. Listening %v", listener.Addr())
	if err := server.Serve(listener); err != nil {
		log.Fatalf("Server :: Error : %v", err)
	}
}

func createDatabaseTables(db *sql.DB) {
     _, err := db.Exec(`CREATE TABLE IF NOT EXISTS tickets (
            t_id TEXT PRIMARY KEY,
            t_from TEXT,
            t_to TEXT,
            t_price INTEGER,
            t_seat INTEGER,
            t_section TEXT,
            t_user_id TEXT
        )`)
     if err != nil {
        log.Fatalf("Failed to create TICKET table: %v", err)
     }

      _, dbErr := db.Exec(`CREATE TABLE IF NOT EXISTS users (
                 u_id TEXT PRIMARY KEY,
                 u_user_fname TEXT,
                 u_user_lname TEXT,
                 u_user_email TEXT
             )`)
      if dbErr != nil {
          log.Fatalf("Failed to create USER table: %v", dbErr)
      }
}

