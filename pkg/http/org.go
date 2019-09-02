package http

import (
	"context"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/ustackq/indagate/pkg/service"
	"github.com/ustackq/indagate/pkg/utils/errors"
	"net/http"

	"go.uber.org/zap"
)

type OrgBackend struct {
	Logger *zap.Logger

	OrganizationService        service.OrganizationService
	UserResourceMappingService service.UserResourceMappingService
	UserService                service.UserService
}

func NewOrgBackend(ab *APIBackend) *OrgBackend {
	return &OrgBackend{
		Logger: ab.Logger.With(zap.String("handler", "org")),

		OrganizationService:        ab.OrganizationService,
		UserResourceMappingService: ab.UserResourceMappingService,
		UserService:                ab.UserService,
	}
}

type OrgHandler struct {
	*httprouter.Router

	Logger *zap.Logger

	OrganizationService        service.OrganizationService
	UserResourceMappingService service.UserResourceMappingService
	UserService                service.UserService
}

const (
	orgPath = "/api/v1/orgs"

	orgIDPath          = "/api/v1/orgs:id"
	orgIDMembersPath   = "/api/v1/orgs/:id/members"
	orgIDMembersIDPath = "/api/v1/orgs/:id/members/:userID"
	orgIDOwnersPath    = "/api/v1/orgs/:id/owners"
	orgIDOwnersIDPath  = "/api/v1/orgs/:id/owners/:userID"
)

func NewOrgHandler(ab *OrgBackend) *OrgHandler {
	oh := &OrgHandler{
		Router: NewRouter(),
		Logger: zap.NewNop(),

		OrganizationService:        ab.OrganizationService,
		UserResourceMappingService: ab.UserResourceMappingService,
		UserService:                ab.UserService,
	}

	oh.POST(orgPath, oh.handlePostOrg)
	oh.GET(orgPath, oh.handleGetOrgs)
	oh.GET(orgIDPath, oh.handleGetOrg)
	oh.PATCH(orgIDPath, oh.handlePatchOrg)
	oh.DELETE(orgIDPath, oh.handleDeleteOrg)

	memberBackend := MemberBackend{
		Logger:                     ab.Logger.With(zap.String("handler", "member")),
		ResourceType:               service.OrgsResourceType,
		UserType:                   service.Member,
		UserResourceMappingService: ab.UserResourceMappingService,
		UserService:                ab.UserService,
	}

	oh.POST(orgIDMembersPath, newPostMemberHandler(memberBackend))
	oh.GET(orgIDMembersIDPath, newGetMembersHandler(memberBackend))
	oh.DELETE(orgIDMembersIDPath, newDeleteMemberHandler(memberBackend))

	ownerBackend := MemberBackend{
		Logger:                     ab.Logger.With(zap.String("handler", "member")),
		ResourceType:               service.OrgsResourceType,
		UserType:                   service.Owner,
		UserResourceMappingService: ab.UserResourceMappingService,
		UserService:                ab.UserService,
	}

	oh.POST(orgIDOwnersPath, newPostMemberHandler(ownerBackend))
	oh.GET(orgIDOwnersPath, newGetMembersHandler(ownerBackend))
	oh.DELETE(orgIDOwnersIDPath, newDeleteMemberHandler(ownerBackend))

	return oh
}

type orgResponse struct {
	Links map[string]string `json:"links"`
	service.Organization
}

func newOrgResponse(org *service.Organization) *orgResponse {
	return &orgResponse{
		Links: map[string]string{
			"self":    fmt.Sprintf("/api/v1/orgs/%s", org.ID),
			"members": fmt.Sprintf("/api/v1/orgs/%s/members", org.ID),
			"owners":  fmt.Sprintf("/api/v1/orgs/%s/owners", org.ID),
		},
		Organization: *org,
	}
}

func (oh *OrgHandler) handlePostOrg(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	req, err := decodePostOrgRequest(ctx, r)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	if err := oh.OrganizationService.CreateOrganization(ctx, req.Org); err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	if err := encodeResponse(ctx, rw, http.StatusCreated, newOrgResponse(req.Org)); err != nil {
		LogEncodeError(oh.Logger, r, err)
		return
	}
}

type getOrgRequest struct {
	filter service.OrganizationFilter
}

func decodeGetOrgRequest(ctx context.Context, r *http.Request) (req *getOrgRequest, err error) {
	query := r.URL.Query()
	if OrgID := query.Get(OrgID); OrgID != "" {
		id, err := service.IDFromString(OrgID)
		if err != nil {
			return nil, err
		}

		req.filter.ID = id
	}

	if name := query.Get(OrgName); name != "" {
		req.filter.Name = &name
	}

	return
}

type orgsResponse struct {
	Links         map[string]string `json:"links"`
	Organizations []*orgResponse    `json:"orgs"`
}

func newOrgsResponse(orgs []*service.Organization) *orgsResponse {
	res := orgsResponse{
		Links: map[string]string{
			"self": "/api/v1/orgs",
		},
		Organizations: []*orgResponse{},
	}

	for _, org := range orgs {
		res.Organizations = append(res.Organizations, newOrgResponse(org))
	}

	return &res
}

func (oh *OrgHandler) handleGetOrgs(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	req, err := decodeGetOrgRequest(ctx, r)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	orgs, _, err := oh.OrganizationService.FindOrganizations(ctx, req.filter)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	if err := encodeResponse(ctx, rw, http.StatusOK, newOrgsResponse(orgs)); err != nil {
		LogEncodeError(oh.Logger, r, err)
		return
	}
}

func (oh *OrgHandler) handleGetOrg(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()
	var id service.ID
	if err := id.DecodeFromString(ps.ByName("id")); err != nil {
		EncodeError(ctx, &errors.Error{
			Code: errors.Invalid,
			Msg:  "invalid org id",
			Err:  err,
		}, rw)
		return
	}
	org, err := oh.OrganizationService.FindOrganizationByID(ctx, id)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	if err := encodeResponse(ctx, rw, http.StatusOK, newOrgResponse(org)); err != nil {
		LogEncodeError(oh.Logger, r, err)
		return
	}
}

func (oh *OrgHandler) handleDeleteOrg(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	var id service.ID

	if err := id.DecodeFromString(ps.ByName("id")); err != nil {
		EncodeError(ctx, &errors.Error{
			Code: errors.Invalid,
			Msg:  "invalid org id",
			Err:  err,
		}, rw)
		return
	}

	err := oh.OrganizationService.DeleteOrganization(ctx, id)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	rw.WriteHeader(http.StatusNoContent)
}

type patchOrgRequest struct {
	Update service.OrganizationUpdate
	OrgID  service.ID
}

func decodePatchOrgRequest(r *http.Request, id string) (*patchOrgRequest, error) {
	var i service.ID

	if err := i.DecodeFromString(id); err != nil {
		return nil, err
	}

	var update service.OrganizationUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		return nil, err
	}

	return &patchOrgRequest{
		Update: update,
		OrgID:  i,
	}, nil

}

func (oh *OrgHandler) handlePatchOrg(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()

	req, err := decodePatchOrgRequest(r, ps.ByName("id"))
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	org, err := oh.OrganizationService.UpdateOrganization(ctx, req.OrgID, req.Update)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}

	if err := encodeResponse(ctx, rw, http.StatusOK, newOrgResponse(org)); err != nil {
		LogEncodeError(oh.Logger, r, err)
		return
	}
}
