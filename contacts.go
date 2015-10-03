package main

import "net/http"

// Email ...
type Email struct {
	Address *string `json:"address,omitempty"`
	Type    *string `json:"type,omitempty"`
}

// NewEmail creates a new contact email
func NewEmail(address, emailType string) *Email {
	return &Email{Address: &address, Type: &emailType}
}

// Phone ...
type Phone struct {
	Number *string `json:"number,omitempty"`
	Type   *string `json:"type,omitempty"`
}

// NewPhone creates a new contact phone
func NewPhone(number, numberType string) *Phone {
	return &Phone{Number: &number, Type: &numberType}
}

// IMHandle ...
type IMHandle struct {
	Identifier *string `json:"identifier,omitempty"`
	Protocol   *string `json:"protocol,omitempty"`
	Type       *string `json:"type,omitempty"`
}

// NewIMHandle creates a new IM Handle for a contact
func NewIMHandle(identifier, protocol, imType string) *IMHandle {
	return &IMHandle{
		Identifier: &identifier,
		Protocol:   &protocol,
		Type:       &imType,
	}
}

// Organization ...
type Organization struct {
	Company        *string `json:"company,omitempty"db:"company"`
	Type           *string `json:"type,omitempty"db:"type"`
	Title          *string `json:"title,omitempty"db:"title"`
	Department     *string `json:"department,omitempty"db:"department"`
	JobDescription *string `json:"job_description,omitempty"db:"job_description"`
	Symbol         *string `json:"symbol,omitempty"db:"symbol"`
	PhoneticName   *string `json:"phonetic_name,omitempty"db:"phonetic_name"`
	OfficeLocation *string `json:"office_location,omitempty"db:"office_location"`
}

// Relation ...
type Relation struct {
	Name *string `json:"name,omitempty"`
	Type *string `json:"type,omitempty"`
}

// NewRelation creates a contact's Relation
func NewRelation(name, relationType string) *Relation {
	return &Relation{Name: &name, Type: &relationType}
}

// PostalAddress ...
type PostalAddress struct {
	Street       *string `json:"street,omitempty"`
	POBox        *string `json:"po_box,omitempty"`
	Neighborhood *string `json:"neighborhood,omitempty"`
	City         *string `json:"city,omitempty"`
	Region       *string `json:"region,omitempty"`
	PostCode     *string `json:"post_code,omitempty"`
	Country      *string `json:"country,omitempty"`
	Type         *string `json:"type,omitempty"`
}

// NewUSAAddress ...
func NewUSAAddress(street, city, region, postCode, addressType string) *PostalAddress {
	pa := &PostalAddress{
		Street:   &street,
		City:     &city,
		Region:   &region,
		PostCode: &postCode,
		Type:     &addressType,
	}
	usa := "United States of America"
	pa.Country = &usa
	return pa
}

// Website ...
type Website struct {
	Address *string `json:"address,omitempty"`
	Type    *string `json:"type,omitempy"`
}

// NewWebsite ...
func NewWebsite(address, siteType string) *Website {
	return &Website{Address: &address, Type: &siteType}
}

// Event ...
type Event struct {
	StartDate *string `json:"start_date,omitempty"`
	Type      *string `json:"type,omitempty"`
}

// NewEvent ...
func NewEvent(startDate, eventType string) *Event {
	return &Event{StartDate: &startDate, Type: &eventType}
}

// Contact ...
type Contact struct {
	OwnerID         *int64           `json:"owner_id,omitempty"db:"owner_id"`
	Name            *string          `json:"name,omitempty"db:"name"`
	Emails          []*Email         `json:"emails,omitempty"db:"-"`
	Phones          []*Phone         `json:"phones,omitempty"db:"-"`
	IMHandles       []*IMHandle      `json:"im_handles,omitempty"db:"-"`
	Org             *Organization    `json:"organization,omitempty"db:"-"`
	Relations       []*Relation      `json:"relations,omitempty"db:"-"`
	PostalAddresses []*PostalAddress `json:"postal_addresses,omitempty"db:"-"`
	Websites        []*Website       `json:"websites,omitempty"db:"-"`
	Events          []*Event         `json:"events,omitempty"db:"-"`
}

// CreateContactHandler handles POST /contacts
func CreateContactHandler(w http.ResponseWriter, r *http.Request) {

}
