package api

import (
	"errors"
	"math/rand"
)

const MaxSeatsPerSection = 20

var MaxSeatsLimitReached = errors.New("Max seat limit reached")
var SeatNotAvailable = errors.New("Seat is already booked")

type SeatAllocator struct {
	occupiedSeatsSectionA map[int32]bool
	occupiedSeatsSectionB map[int32]bool
	sections              [2]string
}

// helps to create new seat allocator
func NewSeatAllocator() *SeatAllocator {
	return &SeatAllocator{
		occupiedSeatsSectionA: make(map[int32]bool),
		occupiedSeatsSectionB: make(map[int32]bool),
		sections:              [2]string{"A", "B"},
	}
}

func (s *SeatAllocator) AllocateSeat() (int32, string, error) {

	section, err := s.findSection()
	if err != nil {
		return 0, "", err
	}

	seatNumber := rand.Int31n(MaxSeatsPerSection)
	if section == "A" {
		if _, ok := s.occupiedSeatsSectionA[seatNumber]; ok {
			return s.AllocateSeat()
		}

		s.occupiedSeatsSectionA[seatNumber] = true
	} else if section == "B" {
		if _, ok := s.occupiedSeatsSectionB[seatNumber]; ok {
			return s.AllocateSeat()
		}

		s.occupiedSeatsSectionB[seatNumber] = true
	}
	return seatNumber, section, nil
}

func (s *SeatAllocator) DeallocateSeat(seatNumber int32, section string) {
	if section == "A" {
		delete(s.occupiedSeatsSectionA, seatNumber)
	} else if section == "B" {
		delete(s.occupiedSeatsSectionB, seatNumber)
	}
}

func (s *SeatAllocator) AllocateSpecificSeat(seatNumber int32, section string) error {
	if !s.isSeatAvailable(seatNumber, section) {
		return SeatNotAvailable
	}
	if section == "A" {
		s.occupiedSeatsSectionA[seatNumber] = true
	} else if section == "B" {
		s.occupiedSeatsSectionB[seatNumber] = true
	}
	return nil
}

func (s *SeatAllocator) isSeatAvailable(seatNumber int32, section string) bool {
    isSeatAvailable := false
	if section == "A" && len(s.occupiedSeatsSectionA) < MaxSeatsPerSection {
		if _, ok := s.occupiedSeatsSectionA[seatNumber]; !ok {
			isSeatAvailable = true
		}
	} else if section == "B" && len(s.occupiedSeatsSectionB) < MaxSeatsPerSection {
		if _, ok := s.occupiedSeatsSectionB[seatNumber]; !ok {
			isSeatAvailable = true
		}
	}

	return isSeatAvailable
}

func (s *SeatAllocator) findSection() (string, error) {
	if len(s.occupiedSeatsSectionA) < MaxSeatsPerSection && len(s.occupiedSeatsSectionB) < MaxSeatsPerSection {
		return s.sections[rand.Intn(2)], nil
	} else if len(s.occupiedSeatsSectionA) < MaxSeatsPerSection {
		return "A", nil
	} else if len(s.occupiedSeatsSectionB) < MaxSeatsPerSection {
		return "B", nil
	} else {
		return "", MaxSeatsLimitReached
	}
}
