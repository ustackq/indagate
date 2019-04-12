package service

// Action is an enum defining all possible resource operation.
type Action string

const (
	ReadAction      Action = "READ"      //1
	WriteAction     Action = "WRITE"     //2
	READWRITEACTION Action = "READWRITE" //3

)

// ResourceType is an enum defining all resource types that have a permission model in indagate.
type ResourceType string

// Resource is an authorizable resource.
type Resource struct {
	Type ResourceType `json:"type"`
	ID   *ID          `json:"id,omitempty"`
}

// Permission defines an action and resource relactionship.
type Permission struct {
	Action   Action   `json:"action"`
	Resource Resource `json:"resource"`
}

func PermissionAllowed(p Permission, ps []Permission) bool {
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
