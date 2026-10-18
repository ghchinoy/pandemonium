package store

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"github.com/ghchinoy/pandemonium/pkg/types"
)

// FirestoreStore implements StateStore using Google Cloud Firestore.
type FirestoreStore struct {
	client *firestore.Client
}

// NewFirestoreStore initializes a new Firestore client.
func NewFirestoreStore(ctx context.Context, projectID string) (*FirestoreStore, error) {
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("initializing firestore: %w", err)
	}
	return &FirestoreStore{client: client}, nil
}

// SaveScenario persists a scenario template to Firestore.
func (s *FirestoreStore) SaveScenario(ctx context.Context, sc types.Scenario) error {
	if _, err := s.client.Collection("scenarios").Doc(sc.Name).Set(ctx, sc); err != nil {
		return fmt.Errorf("saving scenario %q to firestore: %w", sc.Name, err)
	}
	return nil
}

// GetScenario retrieves a scenario template by name.
func (s *FirestoreStore) GetScenario(ctx context.Context, name string) (types.Scenario, error) {
	doc, err := s.client.Collection("scenarios").Doc(name).Get(ctx)
	if err != nil {
		return types.Scenario{}, fmt.Errorf("fetching scenario %q from firestore: %w", name, err)
	}
	var sc types.Scenario
	if err := doc.DataTo(&sc); err != nil {
		return types.Scenario{}, fmt.Errorf("decoding scenario %q from firestore: %w", name, err)
	}
	return sc, nil
}

// ListScenarios returns the names of all persisted scenarios.
func (s *FirestoreStore) ListScenarios(ctx context.Context) ([]string, error) {
	iter := s.client.Collection("scenarios").Documents(ctx)
	var names []string
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("listing scenarios from firestore: %w", err)
		}
		names = append(names, doc.Ref.ID)
	}
	return names, nil
}

// SaveEvent adds a simulation event to a session's event sub-collection.
func (s *FirestoreStore) SaveEvent(ctx context.Context, sessionID string, event types.Event) error {
	if _, _, err := s.client.Collection("sessions").Doc(sessionID).Collection("events").Add(ctx, event); err != nil {
		return fmt.Errorf("saving event for session %q to firestore: %w", sessionID, err)
	}
	return nil
}

// GetSessionEvents retrieves all events associated with a specific session ID.
func (s *FirestoreStore) GetSessionEvents(ctx context.Context, sessionID string) ([]types.Event, error) {
	iter := s.client.Collection("sessions").Doc(sessionID).Collection("events").Documents(ctx)
	var events []types.Event
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("listing events for session %q from firestore: %w", sessionID, err)
		}
		var e types.Event
		if err := doc.DataTo(&e); err != nil {
			return nil, fmt.Errorf("decoding event for session %q from firestore: %w", sessionID, err)
		}
		events = append(events, e)
	}
	return events, nil
}
