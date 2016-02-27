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

// EmailType values describe different address types (e.g. home, work, etc.)
type EmailType int

// Types to describe a contact's email address
const (
	EmailTypeCustom EmailType = 0
	EmailTypeHome             = 1
	EmailTypeWork             = 2
	EmailTypeOther            = 3
	EmailTypeMobile           = 4
)

// Email ...
type Email struct {
	Address string    `json:"address,omitempty"`
	Type    EmailType `json:"type,omitempty"`
	Label   *string   `json:"label,omitempty"`
}

// PhoneType values describe different phone number types (e.g. home, work, etc.)
type PhoneType int

// Enumeration of different phone types
const (
	PhoneTypeCustom      PhoneType = 0
	PhoneTypeHome                  = 1
	PhoneTypeMobile                = 2
	PhoneTypeWork                  = 3
	PhoneTypeFaxWork               = 4
	PhoneTypeFaxHome               = 5
	PhoneTypePager                 = 6
	PhoneTypeOther                 = 7
	PhoneTypeCallback              = 8
	PhoneTypeCar                   = 9
	PhoneTypeCompanyMain           = 10
	PhoneTypeISDN                  = 11
	PhoneTypeMain                  = 12
	PhoneTypeOtherFax              = 13
	PhoneTypeRadio                 = 14
	PhoneTypeTelex                 = 15
	PhoneTypeTTYTDD                = 16
	PhoneTypeWorkMobile            = 17
	PhoneTypeWorkPager             = 18
)

// Phone ...
type Phone struct {
	Number string    `json:"number,omitempty"`
	Type   PhoneType `json:"type,omitempty"`
	Label  *string   `json:"label,omitempty"`
}

// IMType ...
type IMType int

// Enumeration of IMTypes
const (
	IMTypeCustom IMType = 0
	IMTypeHome          = 1
	IMTypeWork          = 2
	IMTypeOther         = 3
)

// IMProtocol ...
type IMProtocol int

// Enumeration of IMProtocols
const (
	IMProtocolCustom   IMProtocol = 0
	IMProtocolAIM                 = 1
	IMProtocolMSN                 = 2
	IMProtocolYahoo               = 3
	IMProtocolSkype               = 4
	IMProtocolQQ                  = 5
	IMProtocolHangouts            = 6
	IMProtocolICQ                 = 7
	IMProtocolXMPP                = 8
)

// IMAccount ...
type IMAccount struct {
	Handle         string     `json:"handle,omitempty"`
	Type           IMType     `json:"type,omitempty"`
	Label          *string    `json:"label,omitempty"`
	Protocol       IMProtocol `json:"protocol,omitempty"`
	CustomProtocol *string    `json:"custom_protocol,omitempty"db:"custom_protocol"`
}

// Organization ...
type Organization struct {
	Company *string `json:"company,omitempty"`
	Title   *string `json:"title,omitempty"`
}

// RelationType ...
type RelationType int

// Enumeration of RelationTypes
const (
	RelationTypeCustom          RelationType = 0
	RelationTypeAssistant                    = 1
	RelationTypeBrother                      = 2
	RelationTypeChild                        = 3
	RelationTypeDomesticPartner              = 4
	RelationTypeFather                       = 5
	RelationTypeFriend                       = 6
	RelationTypeManager                      = 7
	RelationTypeMother                       = 8
	RelationTypeParent                       = 9
	RelationTypePartner                      = 10
	RelationTypeReferredBy                   = 11
	RelationTypeRelative                     = 12
	RelationTypeSister                       = 13
	RelationTypeSpouse                       = 14
)

// Relation ...
type Relation struct {
	Name  string       `json:"name,omitempty"`
	Type  RelationType `json:"type,omitempty"`
	Label *string      `json:"label,omitempty"`
}

// PostalAddressType ...
type PostalAddressType int

// Enumeration of PostalAddresssTypes
const (
	PostalAddressTypeCustom PostalAddressType = 0
	PostalAddressTypeHome                     = 1
	PostalAddressTypeWork                     = 2
	PostalAddresssTypeOther                   = 3
)

// PostalAddress ...
type PostalAddress struct {
	Street       *string           `json:"street,omitempty"`
	POBox        *string           `json:"po_box,omitempty"db:"po_box"`
	Neighborhood *string           `json:"neighborhood,omitempty"`
	City         *string           `json:"city,omitempty"`
	Region       *string           `json:"region,omitempty"`
	PostCode     *string           `json:"post_code,omitempty"db:"post_code"`
	Country      *string           `json:"country,omitempty"`
	Type         PostalAddressType `json:"type,omitempty"`
	Label        *string           `json:"label,omitempty"`
}

// NewUSAAddress ...
func NewUSAAddress(street, city, region, postCode string, addressType PostalAddressType) *PostalAddress {
	pa := &PostalAddress{
		Street:   &street,
		City:     &city,
		Region:   &region,
		PostCode: &postCode,
		Type:     addressType,
	}
	usa := "United States of America"
	pa.Country = &usa
	return pa
}

// EventType ...
type EventType int

// Enumeration of EventTypes
const (
	EventTypeCustom      EventType = 0
	EventTypeAnniversary           = 1
	EventTypeOther                 = 2
	EventTypeBirthday              = 3
)

// Event ...
type Event struct {
	StartDate string    `json:"start_date,omitempty"db:"start_date"`
	Type      EventType `json:"type,omitempty"`
	Label     *string   `json:"label,omitempty"`
}

// Contact ...
type Contact struct {
	ID              *int64           `json:"id,omitempty"`
	Name            *StructuredName  `json:"name,omitempty"`
	Nickname        *string          `json:"nickname,omitempty"`
	Emails          []*Email         `json:"emails,omitempty"db:"-"`
	Phones          []*Phone         `json:"phones,omitempty"db:"-"`
	IMAccounts      []*IMAccount     `json:"im_accounts,omitempty"db:"-"`
	Org             *Organization    `json:"organization,omitempty"db:"-"`
	Relations       []*Relation      `json:"relations,omitempty"db:"-"`
	PostalAddresses []*PostalAddress `json:"postal_addresses,omitempty"db:"-"`
	Websites        []string         `json:"websites,omitempty"db:"-"`
	Events          []*Event         `json:"events,omitempty"db:"-"`
	Note            *string          `json:"note,omitempty"`
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
