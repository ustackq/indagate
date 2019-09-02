package http

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"net/http"

	icontext "github.com/ustackq/indagate/pkg/context"
	"github.com/ustackq/indagate/pkg/service"
	"github.com/ustackq/indagate/pkg/utils/errors"

	"github.com/julienschmidt/httprouter"
)

// AuthorizationBackend is all services and associated parameters required to construct
// the AuthorizationHandler.
type AuthorizationBackend struct {
	Logger *zap.Logger

	AuthorizationService service.AuthorizationService
	OrganizationService  service.OrganizationService
	UserService          service.UserService
	LookupService        service.LookupService
}

func NewAuthorizationBackend(ab *APIBackend) *AuthorizationBackend {
	return &AuthorizationBackend{
		Logger: ab.Logger.With(zap.String("handler", "authorization")),

		AuthorizationService: ab.AuthenticationService,
		OrganizationService:  ab.OrganizationService,
		UserService:          ab.UserService,
		LookupService:        ab.LookupService,
	}
}

type AuthorizationHandler struct {
	*httprouter.Router
	Logger *zap.Logger

	OrganizationService  service.OrganizationService
	UserService          service.UserService
	AuthorizationService service.AuthorizationService
	LookupService        service.LookupService
}

const (
	authsPath  = "/api/v1/authorizations"
	authIDPath = "/api/v1/authorizations/:id"
)

func NewAuthorizationHandler(ab *AuthorizationBackend) *AuthorizationHandler {
	ah := &AuthorizationHandler{
		Router: NewRouter(),
		Logger: ab.Logger,

		AuthorizationService: ab.AuthorizationService,
		OrganizationService:  ab.OrganizationService,
		UserService:          ab.UserService,
		LookupService:        ab.LookupService,
	}

	ah.POST(authsPath, ah.handlePostAuthorization)
	ah.GET(authsPath, ah.handleGetAuthorizations)
	ah.GET(authIDPath, ah.handleGetAuthorization)
	ah.PATCH(authIDPath, ah.handleUpdateAuthorization)
	ah.DELETE(authIDPath, ah.handleDeleteAuthorization)

	return ah
}

type postAuthorizationRequest struct {
	Status      service.Status        `json:"status"`
	OrgID       service.ID            `json:"orgID"`
	UserID      service.ID            `json:"userID"`
	Description string                `json:"description"`
	Permissions []*service.Permission `json:"permissions"`
}

func (pa *postAuthorizationRequest) toPlatform(userID service.ID) *service.Authorization {
	return &service.Authorization{
		OrgID:       pa.OrgID,
		Status:      pa.Status,
		Description: pa.Description,
		Permissions: pa.Permissions,
		UserID:      userID,
	}
}

func (pa *postAuthorizationRequest) SetDefaults() {
	if pa.Status == "" {
		pa.Status = service.Active
	}
}

func (pa *postAuthorizationRequest) Validate() error {
	if len(pa.Permissions) == 0 {
		return &errors.Error{
			Code: errors.Invalid,
			Msg:  "authorization must include permissions",
		}
	}

	for _, p := range pa.Permissions {
		if err := p.Valid(); err != nil {
			return &errors.Error{
				Err: err,
			}
		}
	}

	if !pa.OrgID.Valid() {
		return &errors.Error{
			Code: errors.Invalid,
			Msg:  "org id required",
		}
	}

	if pa.Status == "" {
		pa.SetDefaults()
	}

	if err := pa.Status.Valid(); err != nil {
		return err
	}

	return nil
}

func descodePostAuthorizationRequest(r *http.Request) (*postAuthorizationRequest, error) {
	auth := &postAuthorizationRequest{}
	if err := json.NewDecoder(r.Body).Decode(auth); err != nil {
		return nil, &errors.Error{
			Code: errors.Invalid,
			Msg:  "invalid json structure",
			Err:  err,
		}
	}

	auth.SetDefaults()
	return auth, auth.Validate()
}

func getAuthorizeUser(r *http.Request, svc service.UserService) (*service.User, error) {
	ctx := r.Context()

	a, err := icontext.GetAuthorizer(ctx)
	if err != nil {
		return nil, err
	}

	return svc.FindUserByID(ctx, a.GetUserID())
}

type permissionResponse struct {
	Action   service.Action   `json:"action"`
	Resource resourceResponse `json:"resource"`
}

type resourceResponse struct {
	service.Resource
	Name         string `json:"name,omitempty"`
	Organization string `json:"org,omitempty"`
}

func newPermissionsResponse(ctx context.Context, ps []*service.Permission, ls service.LookupService) ([]permissionResponse, error) {
	res := make([]permissionResponse, len(ps))
	for index, p := range ps {
		res[index] = permissionResponse{
			Action: p.Action,
			Resource: resourceResponse{
				Resource: p.Resource,
			},
		}

		if p.Resource.ID != nil {
			name, err := ls.Name(ctx, p.Resource.Type, *p.Resource.ID)
			if errors.ErrorCode(err) == errors.NotFound {
				continue
			}
			if err != nil {
				return nil, err
			}
			res[index].Resource.Name = name
		}

		if p.Resource.OrgID != nil {
			name, err := ls.Name(ctx, service.OrgsResourceType, *p.Resource.OrgID)
			if errors.ErrorCode(err) == errors.NotFound {
				continue
			}
			if err != nil {
				return nil, err
			}
			res[index].Resource.Organization = name
		}
	}

	return res, nil
}

func (ah *AuthorizationHandler) handlePostAuthorization(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	req, err := descodePostAuthorizationRequest(r)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	user, err := getAuthorizeUser(r, ah.UserService)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	auth := req.toPlatform(user.ID)

	org, err := ah.OrganizationService.FindOrganizationByID(ctx, auth.OrgID)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	if err = ah.AuthorizationService.CreateAuthorization(ctx, auth); err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	perms, err := newPermissionsResponse(ctx, auth.Permissions, ah.LookupService)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	if err := encodeResponse(ctx, rw, http.StatusCreated, newAuthResponse(auth, org, user, perms)); err != nil {
		EncodeError(ctx, err, rw)
		return
	}

}

type getAuthorizationRequest struct {
	filter service.AuthorizationFilter
}

func decodeGetAuthorizationsRequest(ctx context.Context, r *http.Request) (*getAuthorizationRequest, error) {
	query := r.URL.Query()
	req := &getAuthorizationRequest{}
	userID := query.Get("userID")

	if userID != "" {
		id, err := service.IDFromString(userID)
		if err != nil {
			return nil, err
		}
		req.filter.UserID = id
	}

	user := query.Get("user")
	if user != "" {
		req.filter.User = &user
	}

	orgID := query.Get("orgID")
	if orgID != "" {
		orgID, err := service.IDFromString(orgID)
		if err != nil {
			return nil, err
		}
		req.filter.OrgID = orgID
	}

	org := query.Get("org")
	if user != "" {
		req.filter.Org = &org
	}

	authID := query.Get("id")
	if authID != "" {
		id, err := service.IDFromString(authID)
		if err != nil {
			return nil, err
		}
		req.filter.ID = id
	}

	return req, nil
}

type authResponse struct {
	ID          service.ID           `json:"id"`
	Token       string               `json:"token"`
	Status      service.Status       `json:"status"`
	Description string               `json:"description"`
	OrgID       service.ID           `json:"orgID"`
	Org         string               `json:"org"`
	UserID      service.ID           `json:"userID"`
	User        string               `json:"user"`
	Permissions []permissionResponse `json:"permissions"`
	Links       map[string]string    `json:"links"`
}

func newAuthResponse(a *service.Authorization, org *service.Organization, user *service.User, ps []permissionResponse) *authResponse {
	res := &authResponse{
		ID:          a.ID,
		Token:       a.Token,
		Status:      a.Status,
		Description: a.Description,
		OrgID:       a.OrgID,
		UserID:      a.UserID,
		User:        user.Name,
		Org:         org.Name,
		Permissions: ps,
		Links: map[string]string{
			"self": fmt.Sprintf("/api/v1/authorizations/%s", a.ID),
			"user": fmt.Sprintf("/api/v1/users/%s", a.UserID),
		},
	}

	return res
}

type authsResponse struct {
	Links map[string]string `json:"links"`
	Auths []*authResponse   `json:"authorizations"`
}

func newAuthsResponse(as []*authResponse) *authsResponse {
	return &authsResponse{
		Links: map[string]string{
			"self": "/api/v1/authorizations",
		},
		Auths: as,
	}
}

func (ah *AuthorizationHandler) handleGetAuthorizations(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	req, err := decodeGetAuthorizationsRequest(ctx, r)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	opts := service.FindOptions{}
	as, _, err := ah.AuthorizationService.FindAuthorization(ctx, req.filter, opts)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	auths := make([]*authResponse, 0, len(as))
	for _, a := range as {
		org, err := ah.OrganizationService.FindOrganizationByID(ctx, a.ID)
		if err != nil {
			continue
		}

		user, err := ah.UserService.FindUserByID(ctx, a.UserID)
		if err != nil {
			continue
		}

		pes, err := newPermissionsResponse(ctx, a.Permissions, ah.LookupService)
		if err != nil {
			EncodeError(ctx, err, rw)
			return
		}

		auths = append(auths, newAuthResponse(a, org, user, pes))
	}

	if err := encodeResponse(ctx, rw, http.StatusOK, newAuthsResponse(auths)); err != nil {
		EncodeError(ctx, err, rw)
		return
	}
}

type requestFilter struct {
	ID service.ID
}

func decodeRequest(r *http.Request, ps httprouter.Params) (*requestFilter, error) {
	id := ps.ByName("id")
	if id == "" {
		return nil, &errors.Error{
			Code: errors.Invalid,
			Msg:  "url missing id",
		}
	}

	var i service.ID
	if err := i.DecodeFromString(id); err != nil {
		return nil, err
	}

	return &requestFilter{
		ID: i,
	}, nil
}

func (ah *AuthorizationHandler) handleGetAuthorization(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	req, err := decodeRequest(r, ps)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	a, err := ah.AuthorizationService.FindAuthorizationByID(ctx, req.ID)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	org, err := ah.OrganizationService.FindOrganizationByID(ctx, a.OrgID)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	user, err := ah.UserService.FindUserByID(ctx, a.UserID)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	pes, err := newPermissionsResponse(ctx, a.Permissions, ah.LookupService)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	if err := encodeResponse(ctx, rw, http.StatusOK, newAuthResponse(a, org, user, pes)); err != nil {
		EncodeError(ctx, err, rw)
		return
	}
}

type updateAuthorizationRequest struct {
	ID service.ID
	*service.AuthorizationUpdate
}

func decodeUpdateAuthorizationRequest(r *http.Request, ps httprouter.Params) (*updateAuthorizationRequest, error) {
	id := ps.ByName("id")
	if id == "" {
		return nil, &errors.Error{
			Code: errors.Invalid,
			Msg:  "url missing id",
		}
	}

	var i service.ID
	if err := i.DecodeFromString(id); err != nil {
		return nil, err
	}

	update := &service.AuthorizationUpdate{}
	if err := json.NewDecoder(r.Body).Decode(update); err != nil {
		return nil, err
	}

	return &updateAuthorizationRequest{
		ID:                  i,
		AuthorizationUpdate: update,
	}, nil
}

func (ah *AuthorizationHandler) handleUpdateAuthorization(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	req, err := decodeUpdateAuthorizationRequest(r, ps)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	a, err := ah.AuthorizationService.FindAuthorizationByID(ctx, req.ID)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	a, err = ah.AuthorizationService.UpdateAuthorization(ctx, a.ID, req.AuthorizationUpdate)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	org, err := ah.OrganizationService.FindOrganizationByID(ctx, a.OrgID)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	user, err := ah.UserService.FindUserByID(ctx, a.UserID)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	pes, err := newPermissionsResponse(ctx, a.Permissions, ah.LookupService)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	if err := encodeResponse(ctx, rw, http.StatusOK, newAuthResponse(a, org, user, pes)); err != nil {
		EncodeError(ctx, err, rw)
		return
	}
}

func (ah *AuthorizationHandler) handleDeleteAuthorization(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	req, err := decodeRequest(r, ps)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	if err := ah.AuthorizationService.DeleteAuthorization(ctx, req.ID); err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}
