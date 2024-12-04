package main

import (
	"context"
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
