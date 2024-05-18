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
	ID       uuid.UUID `json:"id"`
	Client   *Client   `json:"client"`
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
	Resolved bool      `json:"resolved"`
}

func (r *Resource) isReserved(start time.Time, end time.Time) bool {
	for _, reservation := range r.Reservations {
		if reservation.Resolved {
			continue
		}

		if reservation.Start.Before(end) && reservation.End.After(start) {
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

	reservation := &Reservation{ID: uuid.New(), Client: client, Start: time.Now(), End: time.Now().Add(5 * time.Minute), Resolved: false}
	r.lock = reservation

	return reservation, true
}

func (r *Resource) Unlock(client *Client, reservation *Reservation) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.lock == nil || r.lock.Client.Name != client.Name || r.lock.ID != reservation.ID {
		return false
	}

	r.lock = nil
	return true
}

func (r *Resource) Reserve(client *Client, reservation *Reservation) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.isReserved(reservation.Start, reservation.End) {
		return false
	}

	if r.lock == nil || r.lock.Client.Name != client.Name || r.lock.ID != reservation.ID {
		return false
	}

	reservation.Client = client
	r.Reservations[reservation.ID] = reservation
	r.lock = nil

	return true
}
