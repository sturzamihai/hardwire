package server

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type Resource struct {
	ID           uuid.UUID                  `json:"id"`
	Name         string                     `json:"name"`
	Reservations map[uuid.UUID]*Reservation `json:"reservations"`
	lock         *Reservation
	mutex        sync.Mutex
}

type Reservation struct {
	ID     uuid.UUID `json:"id"`
	Client *Client   `json:"client"`
	Start  time.Time `json:"start"`
	End    time.Time `json:"end"`
}

func (r *Resource) isReserved(reservation *Reservation) bool {
	for _, sample := range r.Reservations {
		if sample.ID == reservation.ID {
			continue
		}

		if sample.Start.Before(reservation.End) && sample.End.After(reservation.Start) {
			return true
		}
	}

	return false
}

func (r *Resource) Lock(client *Client) (*Reservation, bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.lock != nil {
		return nil, false
	}

	reservation := &Reservation{ID: uuid.New(), Client: client, Start: time.Now(), End: time.Now().Add(5 * time.Minute)}
	r.lock = reservation

	return reservation, true
}

func (r *Resource) Unlock(client *Client) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.lock == nil || r.lock.Client.Name != client.Name {
		return false
	}

	r.lock = nil
	return true
}

func (r *Resource) Reserve(client *Client, reservationId uuid.UUID, start time.Time, end time.Time) (*Reservation, bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.lock == nil || r.lock.Client.Name != client.Name || r.lock.ID != reservationId {
		return nil, false
	}

	if start.After(end) {
		return nil, false
	}

	reservation := &Reservation{ID: reservationId, Client: client, Start: start, End: end}

	if r.isReserved(reservation) {
		return nil, false
	}

	r.Reservations[reservationId] = reservation

	r.lock = nil

	return reservation, true
}

func (r *Resource) UpdateReservation(client *Client, reservationId uuid.UUID, start time.Time, end time.Time) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	reservation, exists := r.Reservations[reservationId]

	if !exists {
		return false
	}

	if reservation.Client.Name != client.Name {
		return false
	}

	if reservation.Start.After(reservation.End) {
		return false
	}

	if r.isReserved(reservation) {
		return false
	}

	reservation.Start = start
	reservation.End = end

	return true
}

func (r *Resource) CancelReservation(client *Client, reservationId uuid.UUID) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	reservation, exists := r.Reservations[reservationId]

	if !exists {
		return false
	}

	if reservation.Client.Name != client.Name {
		return false
	}

	delete(r.Reservations, reservationId)

	return true
}
