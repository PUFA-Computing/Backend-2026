package services

import (
	"Backend/internal/database/app"
	"Backend/internal/models"
	"time"

	"github.com/google/uuid"
)

type EventService struct {
}

func NewEventService() *EventService {
	return &EventService{}
}

// CreateEvent creates a new event in the database
func (es *EventService) CreateEvent(event *models.Event) error {
	if time.Now().Before(event.StartDate) {
		event.Status = "Upcoming"
	} else if time.Now().After(event.StartDate) && time.Now().Before(event.EndDate) {
		event.Status = "Open"
	} else {
		event.Status = "Ended"
	}

	if err := app.CreateEvent(event); err != nil {
		return err
	}

	return nil
}

// GetEventByID retrieves an event by its ID
func (es *EventService) GetEventByID(eventID int) (*models.Event, error) {
	event, err := app.GetEventByID(eventID)
	if err != nil {
		return nil, err
	}
	return event, nil
}

// GetEventBySlug retrieves an event by its slug
func (es *EventService) GetEventBySlug(slug string) (*models.Event, error) {
	event, err := app.GetEventBySlug(slug)
	if err != nil {
		return nil, err
	}
	return event, nil
}

// EditEvent updates an event in the database
func (es *EventService) EditEvent(eventID int, updatedEvent *models.Event) error {
	if time.Now().Before(updatedEvent.StartDate) {
		updatedEvent.Status = "Upcoming"
	} else if time.Now().After(updatedEvent.StartDate) && time.Now().Before(updatedEvent.EndDate) {
		updatedEvent.Status = "Open"
	} else {
		updatedEvent.Status = "Ended"
	}

	if err := app.UpdateEvent(eventID, updatedEvent); err != nil {
		return err
	}

	return nil
}

// DeleteEvent deletes an event from the database
func (es *EventService) DeleteEvent(eventID int) error {
	if err := app.DeleteEvent(eventID); err != nil {
		return err
	}

	return nil
}

// ListEvents retrieves all events from the database
func (es *EventService) ListEvents(queryParams map[string]string) ([]*models.Event, int, error) {
	events, totalPages, err := app.ListEvents(queryParams)
	if err != nil {
		return nil, totalPages, err
	}

	return events, totalPages, nil
}

// RegisterForEvent registers a user for an event
func (es *EventService) RegisterForEvent(userID uuid.UUID, eventID int, additionalNotes string, filePath string) error {
	if err := app.RegisterForEvent(userID, eventID, additionalNotes, filePath); err != nil {
		return err
	}

	return nil
}

// ListRegisteredUsers retrieves all users registered for an event
func (es *EventService) ListRegisteredUsers(eventID int) ([]*models.User, error) {
	users, err := app.ListRegisteredUsers(eventID)
	if err != nil {
		return nil, err
	}

	// Process file paths for each user
	for _, user := range users {
		if user != nil && user.FilePath != nil && *user.FilePath != "" {
			// Create temporary EventRegistration to process file paths
			tempReg := &models.EventRegistration{
				FilePath: *user.FilePath,
			}
			tempReg.ProcessFilePaths()
			// We could store the processed paths in a separate field if needed
			// For now, we just process them but they'll be processed again in the frontend
		}
	}

	return users, nil
}

// ListEventsRegisteredByUser retrieves all events registered by a user
func (es *EventService) ListEventsRegisteredByUser(userID uuid.UUID) ([]*models.Event, error) {
	events, err := app.ListEventsRegisteredByUser(userID)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (es *EventService) TotalRegisteredUsers(eventID int) (int, error) {
	total, err := app.TotalRegisteredUsers(eventID)
	if err != nil {
		return 0, err
	}

	return total, nil
}
