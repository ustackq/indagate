package http

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"github.com/julienschmidt/httprouter"

	"github.com/ustackq/indagate/pkg/service"
)

// RegisterInstall ...
// when install yunus platform, we split the step into four parts
// a. config and check healthz
// b. preflight
// c. postflight
// d. complete check
// we not only provide to install a new cluster, but a new node can be
// join to a current cluster.

const (
	setupPath = "/api/v1/setup"
)

// SetupBackend define installer service
type SetupBackend struct {
	Logger       *zap.Logger
	SetupService service.SetupService
}

// NewSetupBackend return a instance of SetupBackend
func NewSetupBackend(ab *APIBackend) *SetupBackend {
	return &SetupBackend{
		Logger:       ab.Logger,
		SetupService: ab.SetupService,
	}
}

// SetupHandler define an HTTP API handler for setup.
type SetupHandler struct {
	*httprouter.Router
	Logger       *zap.Logger
	SetupService service.SetupService
}

// NewSetupHandler return a instance of SetupHandler
func NewSetupHandler(stb *SetupBackend) *SetupHandler {
	sh := &SetupHandler{
		Router:       NewRouter(),
		Logger:       stb.Logger,
		SetupService: stb.SetupService,
	}
	sh.POST(setupPath, sh.Setup)
	sh.GET(setupPath, sh.isInstalling)

	return sh
}

// isInstalling handler check wether install
func (sh *SetupHandler) isInstalling(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()
	result, err := sh.SetupService.IsInstalling(ctx)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}
	if err := encodeResponse(ctx, rw, http.StatusOK, service.IsInstallingResponse{result}); err != nil {
		LogEncodeError(sh.Logger, r, err)
		return
	}
}

type setupResponse struct {
	User         *userResponse `json:"user"`
	Organization *orgResponse  `json:"org"`
	Auth         *authResponse `json:"auth"`
}

func newSetupResponse(results *service.SetupResult) *setupResponse {
	ps := make([]permissionResponse, len(results.Auth.Permissions))
	for i, p := range results.Auth.Permissions {
		ps[i] = permissionResponse{
			Action: p.Action,
			Resource: resourceResponse{
				Resource: p.Resource,
			},
		}
	}

	return &setupResponse{
		User:         newUserResponse(results.User),
		Organization: newOrgResponse(results.Org),
		Auth:         newAuthResponse(results.Auth, results.Org, results.User, ps),
	}
}

// Install handle indagate install
func (sh *SetupHandler) Setup(rw http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	ctx := r.Context()
	req, err := decodeSetupRequest(ctx, r)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}
	results, err := sh.SetupService.Setup(ctx, req)
	if err != nil {
		EncodeError(ctx, err, rw)
		return
	}
	if err := encodeResponse(ctx, rw, http.StatusCreated, newSetupResponse(results)); err != nil {
		LogEncodeError(sh.Logger, r, err)
		return
	}
}

func decodeSetupRequest(ctx context.Context, r *http.Request) (*service.SetupRequest, error) {
	req := &service.SetupRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return nil, err
	}
	return nil, nil
}

type setupService struct {
	Config string
}

// implement SetupService
func (ss *setupService) IsInstalling(ctx context.Context) (bool, error) {
	// build connection
	// query database and checkout table
	// check table paremetes.Installed == true
	// hold connection
	return true, nil
}

func (ss *setupService) Setup(ctx context.Context, req *service.SetupRequest) (*service.SetupResult, error) {
	// check connection,if not, checkout config and build connection
	// checkout table paremethes.Installed == true
	// if not, initial database and others
	return nil, nil
}
