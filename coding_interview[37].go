package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
)

// Instructions located at the bottom of this file.

// -- Reference Data Structures --

// one of our employees
type Associate struct {
	ID    uint
	Name  string
	Email string
}

// one of our prospective clients
type Contact struct {
	ID    uint
	Email string
	Name  string
}

// one instance of an associate sending an email to a contact
type BlastContact struct {
	ID          uint
	AssociateID uint
	ContactID   uint

	// the earliest date the employee should follow up with the contact
	FollowUpDate time.Time

	Subject string
	Body    string
}

// A historical record of an associate sending an email to a contact
type BlastUpdate struct {
	ID             uint
	BlastContactID uint
	CreatedAt      time.Time
}

// -- Reference Interfaces --

// responsible for sending emails through the email provider
type IMailer interface {
	Send(ctx context.Context, blastContact *BlastContact) error
}

// responsible for interacting with the database
type IRepo interface {
	GetAssociate(ctx context.Context, id uint) (*Associate, error)
	ListAssociates(ctx context.Context) ([]*Associate, error)
	UpdateAssociate(ctx context.Context, associate *Associate) error
	GetContact(ctx context.Context, id uint) (*Contact, error)
	ListContacts(ctx context.Context) ([]*Contact, error)
	UpdateContact(ctx context.Context, contact *Contact) error
	GetBlastContact(ctx context.Context, id uint) (*BlastContact, error)
	ListBlastContacts(ctx context.Context) ([]*BlastContact, error)
	UpdateBlastContact(ctx context.Context, blastContact *BlastContact) error
	GetBlastUpdate(ctx context.Context, id uint) (*BlastUpdate, error)
	ListBlastUpdates(ctx context.Context) ([]*BlastUpdate, error)
	UpdateBlastUpdate(ctx context.Context, blastUpdate *BlastUpdate) error
}

// responsible for enqueuing tasks
type IWorker interface {
	Enqueue(ctx context.Context, task *asynq.Task) (*asynq.TaskInfo, error)
}

// -- TODO: Implement this interface --

type IBlaster interface {
	Process(ctx context.Context, blastContact *BlastContact) error
	Queue(ctx context.Context) error
}

// -- Background --
// Our company has a few employees (called Associates) who send emails to Contacts.
// Assume a BlastContact is one instance of an employee sending an email to a contact, which is unique
// based on the AssociateID, ContactID, and Subject of that email.

// An employee can be emailing one contact multiple times with different subject lines,
// and multiple employees can be emailing the same contact.

// -- Instructions --
// Implement the IBlaster interface.
// - `Queue` should be responsible for queuing up BlastContact instances to be sent.
// - `Process` will be called by a worker processing tasks from the queue, and should handle the business logic of validating
// 	  the BlastContact instance, sending an email and recording that an email was sent.

// This should be a generally high level implementation. I am more interested in your thought process
// rather than specific implementation details. There is no strict time limit - you're welcome to take as long as you'd like,
// but for the sake of the exercise, try to spend no more than 1-2 hours.

// If need be, feel free to make reasonable assumptions and if you think
// another interface/method might be helpful other than the ones provided, feel free to include it
// (with an explanation of why you need it and how it would work).

// -- Limitations/Assumptions --
// - We want to send a max of 100 emails per associate per day.
// - Assume `Queue` is called once per day, at 8am ET.
// - An associate should not email the same contact within a 7 day period.
// - Multiple associates should not be able to email the same contact on the same day.

// -- Example --
// Associate "john@company.com" emails contact "jane@example.com" on 2024-01-01 with subject "Hello".
// Associate "george@company.com" can only email contact "jane@example.com" on or after 2024-01-02.
// Associate "john@company.com" can only email contact "jane@example.com" again on or after 2024-01-08, and the subject could not be "Hello".

type Blaster struct {
	mailer IMailer
	repo   IRepo
	worker IWorker
}

func NewBlaster(repo IRepo, mailer IMailer, worker IWorker) *Blaster {
	return &Blaster{
		repo:   repo,
		mailer: mailer,
		worker: worker,
	}
}

// Queue adds the BlastContact Email Task to the Queue for processing
// Returns a Map of Tasks to perform to listen to or inspect.
func (b *Blaster) Queue(ctx context.Context) error {
	associates, err := b.repo.ListAssociates(ctx) // Get full list of Associates
	if err != nil {
		return fmt.Errorf("failed to list associates: %w", err)
	}

	blastContacts, err := b.repo.ListBlastContacts(ctx) // Get Full list of contacts
	if err != nil {
		return fmt.Errorf("failed to list blast contacts: %w", err)
	}

	const maxEmailsPerDay = 100

	for _, associate := range associates {
		// Filter eligible blast contacts
		eligibleContacts := b.filterEligibleContacts(blastContacts)

		count := 0
		for _, contact := range eligibleContacts {
			if count >= maxEmailsPerDay {
				break
			}

			// Create task for the blast contact
			data := map[string]interface{}{
				"blast_contact_id":   contact.ID,
				"blast_associate_id": associate.ID,
			}

			// Serialize the map into JSON
			payload, err := json.Marshal(data)
			if err != nil {
				log.Fatalf("failed to marshal payload: %v", err)
			}

			// Create the task with serialized payload
			task := asynq.NewTask("process_email", payload)

			_, err = b.worker.Enqueue(ctx, task)
			if err != nil {
				return fmt.Errorf("failed to enqueue task for blast contact ID %d: %w", contact.ID, err)
			}

			count++
		}
	}

	return nil
}

// Process handles the business logic for sending an email and recording it
func (b *Blaster) Process(ctx context.Context, blastContact *BlastContact) error {
	// Validate BlastContact
	if err := b.validateBlastContact(ctx, blastContact); err != nil {
		return fmt.Errorf("blast contact validation failed: %w", err)
	}

	// Send email
	if err := b.mailer.Send(ctx, blastContact); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	// Record the blast update
	blastUpdate := &BlastUpdate{
		BlastContactID: blastContact.ID,
		CreatedAt:      time.Now(),
	}

	if err := b.repo.UpdateBlastUpdate(ctx, blastUpdate); err != nil {
		return fmt.Errorf("failed to record blast update: %w", err)
	}

	return nil
}

/***** UTILITIES *****/

// filterEligibleContacts filters contacts based on rules
func (b *Blaster) filterEligibleContacts(contacts []*BlastContact) []*BlastContact {
	var eligibleContacts []*BlastContact
	for _, contact := range contacts {
		// Skip if the contact was emailed within 7 days
		if time.Since(contact.FollowUpDate) < 7*24*time.Hour {
			continue
		}

		// Add contact to eligible list
		eligibleContacts = append(eligibleContacts, contact)
	}
	return eligibleContacts
}

// validateBlastContact checks if the blast contact follows the rules
func (b *Blaster) validateBlastContact(ctx context.Context, blastContact *BlastContact) error {
	// Check if the associate exists
	associate, err := b.repo.GetAssociate(ctx, blastContact.AssociateID)
	if err != nil || associate == nil {
		return errors.New("invalid associate ID")
	}

	// Check if the contact exists
	contact, err := b.repo.GetContact(ctx, blastContact.ContactID)
	if err != nil || contact == nil {
		return errors.New("invalid contact ID")
	}

	return nil
}
