package service

import "github.com/ustackq/indagate/pkg/utils/errors"

// Action is an enum defining all possible resource operation.
type Action string

const (
	ReadAction      Action = "READ"      //1
	WriteAction     Action = "WRITE"     //2
	READWRITEACTION Action = "READWRITE" //3

)

const (
	// BucketsResourceType gives permissions to one or more buckets.
	BucketsResourceType = ResourceType("buckets") // 1

	// OrgsResourceType gives permissions to one or more orgs.
	OrgsResourceType = ResourceType("orgs") // 3
	// UsersResourceType gives permissions to one or more users.
	UsersResourceType = ResourceType("users") // 7
)

var (
	ErrUnableCreateToken = &errors.Error{
		Msg:  "unable to create token",
		Code: errors.Invalid,
	}
)

// ResourceType is an enum defining all resource types that have a permission model in indagate.
type ResourceType string

// Resource is an authorizable resource.
type Resource struct {
	Type  ResourceType `json:"type"`
	ID    *ID          `json:"id,omitempty"`
	OrgID *ID          `json:"org,omitempty"`
}

// Permission defines an action and resource relactionship.
type Permission struct {
	Action   Action   `json:"action"`
	Resource Resource `json:"resource"`
}

func NewPermissionAtID(id ID, action Action, rt ResourceType, orgID ID) (*Permission, error) {
	p := &Permission{
		Action: action,
		Resource: Resource{
			Type:  rt,
			ID:    &id,
			OrgID: &orgID,
		},
	}
	return p, p.Valid()
}

// NewPermission returns a permission with provided arguments.
func NewPermission(a Action, rt ResourceType, orgID ID) (*Permission, error) {
	p := &Permission{
		Action: a,
		Resource: Resource{
			Type:  rt,
			OrgID: &orgID,
		},
	}

	return p, p.Valid()
}

// NewGlobalPermission constructs a global permission capable of accessing any resource of type rt.
func NewGlobalPermission(a Action, rt ResourceType) (*Permission, error) {
	p := &Permission{
		Action: a,
		Resource: Resource{
			Type: rt,
		},
	}
	return p, p.Valid()
}

func (p *Permission) Valid() error {

	return nil
}

func PermissionAllowed(p Permission, ps []*Permission) bool {
	return true
}

func (p Permission) Mathches(perm Permission) bool {
	if p.Action != perm.Action {
		return false
	}

	if p.Resource.Type != perm.Resource.Type {
		return false
	}

	if p.Resource.ID != nil {
		pID := *p.Resource.ID
		if perm.Resource.ID != nil {
			permID := *perm.Resource.ID
			if pID == permID {
				return true
			}
		}
	}
	return false
}

// Authorizer whicih provider authorize feature component must be impelemented.
type Authorizer interface {
	Allowed(p Permission) bool
	Identifier() ID
	GetUserID() ID
	// Kind metadata for auditing
	Kind() string
}
