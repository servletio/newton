package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// StructuredName ...
type StructuredName struct {
	DisplayName        *string `json:"display_name,omitempty"db:"display_name"`
	Prefix             *string `json:"prefix,omitempty"`
	GivenName          *string `json:"given_name,omitempty"db:"given_name"`
	MiddleName         *string `json:"middle_name,omitempty"db:"middle_name"`
	FamilyName         *string `json:"family_name,omitempty"db:"family_name"`
	Suffix             *string `json:"suffix,omitempty"`
	PhoneticGivenName  *string `json:"phonetic_given_name,omitempty"db:"phonetic_given_name"`
	PhoneticMiddleName *string `json:"phonetic_middle_name,omitempty"db:"phonetic_middle_name"`
	PhoneticFamilyName *string `json:"phonetic_family_name,omitempty"db:"phonetic_family_name"`
}

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
	Company        *string `json:"company,omitempty"`
	Type           *string `json:"type,omitempty"`
	Title          *string `json:"title,omitempty"`
	Department     *string `json:"department,omitempty"`
	JobDescription *string `json:"job_description,omitempty"db:"job_description"`
	Symbol         *string `json:"symbol,omitempty"`
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
	POBox        *string `json:"po_box,omitempty"db:"po_box"`
	Neighborhood *string `json:"neighborhood,omitempty"`
	City         *string `json:"city,omitempty"`
	Region       *string `json:"region,omitempty"`
	PostCode     *string `json:"post_code,omitempty"db:"post_code"`
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
	StartDate *string `json:"start_date,omitempty"db:"start_date"`
	Type      *string `json:"type,omitempty"`
}

// NewEvent ...
func NewEvent(startDate, eventType string) *Event {
	return &Event{StartDate: &startDate, Type: &eventType}
}

// Contact ...
type Contact struct {
	ID              *int64           `json:"id,omitempty"`
	Name            *StructuredName  `json:"name,omitempty"`
	Emails          []*Email         `json:"emails,omitempty"db:"-"`
	Phones          []*Phone         `json:"phones,omitempty"db:"-"`
	IMHandles       []*IMHandle      `json:"im_handles,omitempty"db:"-"`
	Org             *Organization    `json:"organization,omitempty"db:"-"`
	Relations       []*Relation      `json:"relations,omitempty"db:"-"`
	PostalAddresses []*PostalAddress `json:"postal_addresses,omitempty"db:"-"`
	Websites        []*Website       `json:"websites,omitempty"db:"-"`
	Events          []*Event         `json:"events,omitempty"db:"-"`
	OwnerID         *int64           `json:"owner_id,omitempty"db:"owner_id"`
}

func parseContactID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	vars := mux.Vars(r)
	idStr := vars["contact_id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		sendBadReq(w, "invalid contact id")
		return 0, false
	}
	exists, err := db().ContactExists(id)
	if err != nil {
		sendInternalErr(w, err)
		return 0, false
	}
	if !exists {
		sendNotFound(w, fmt.Sprintf("contact %d not found", id))
		return 0, false
	}

	return id, true
}

// CreateContactHandler handles POST /contacts
func CreateContactHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticate(w, r)
	if !ok {
		return
	}

	dec := json.NewDecoder(r.Body)
	contact := &Contact{}
	err := dec.Decode(contact)
	if err != nil {
		sendBadReq(w, "unable to decode the request json")
		return
	}
	// we shouldn't be using any id received from the POST body
	contact.ID = nil

	contact.OwnerID = &userID
	var contactID int64
	contactID, err = db().CreateContact(contact)
	if err != nil {
		sendInternalErr(w, err)
		return
	}
	contact.ID = &contactID
	sendSuccess(w, contact)
}

// GetContactsHandler handles GET /contacts
func GetContactsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticate(w, r)
	if !ok {
		return
	}

	contacts, err := db().Contacts(userID)
	if err != nil {
		sendInternalErr(w, err)
		return
	}

	sendSuccess(w, contacts)
}

// GetContactHandler handles GET /contacts/{contact_id}
func GetContactHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticate(w, r)
	if !ok {
		return
	}

	contactID, ok := parseContactID(w, r)
	if !ok {
		return
	}

	contact, err := db().Contact(contactID, userID)
	if err != nil {
		sendInternalErr(w, err)
		return
	}

	if *contact.OwnerID != userID {
		sendNotFound(w, "contact not found")
		return
	}

	sendSuccess(w, contact)
}

// DeleteContactHandler handles DELETE /contacts/{contact_id}
func DeleteContactHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticate(w, r)
	if !ok {
		return
	}

	contactID, ok := parseContactID(w, r)
	if !ok {
		return
	}

	err := db().DeleteContact(contactID, userID)
	if err != nil {
		sendInternalErr(w, err)
		return
	}

	sendSuccess(w, nil)
}

// GetContactPhotoHandler handles GET /contacts/{contact_id}
func GetContactPhotoHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticate(w, r)
	if !ok {
		return
	}

	contactID, ok := parseContactID(w, r)
	if !ok {
		return
	}

	ownerID, err := db().ContactOwner(contactID)
	if err != nil {
		sendInternalErr(w, err)
		return
	}
	if ownerID != userID {
		sendNotFound(w, "contact not found")
		return
	}

	imgData, err := db().ContactPhoto(contactID)
	if err != nil {
		sendInternalErr(w, err)
		return
	}

	w.Header().Set("Content-Type", "image/*")
	w.Write(imgData)
}

// DeleteContactPhotoHandler handles DELETE /contacts/{contact_id}
func DeleteContactPhotoHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticate(w, r)
	if !ok {
		return
	}

	contactID, ok := parseContactID(w, r)
	if !ok {
		return
	}

	ownerID, err := db().ContactOwner(contactID)
	if err != nil {
		sendInternalErr(w, err)
		return
	}
	if ownerID != userID {
		sendNotFound(w, "contact not found")
		return
	}

	err = db().SetContactPhoto(contactID, nil)
	if err != nil {
		sendInternalErr(w, err)
		return
	}

	sendSuccess(w, nil)
}
