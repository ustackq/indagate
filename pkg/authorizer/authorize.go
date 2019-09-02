package authorizer

import (
	"context"
	"fmt"
	icontext "github.com/ustackq/indagate/pkg/context"
	"github.com/ustackq/indagate/pkg/service"
	"github.com/ustackq/indagate/pkg/utils/errors"
)

func isAllowed(ctx context.Context, p service.Permission) error {
	auth, err := icontext.GetAuthorizer(ctx)
	if err != nil {
		return err
	}

	if !auth.Allowed(p) {
		return &errors.Error{
			Code: errors.Unauthorized,
			Msg:  fmt.Sprintf("%s is unauthorized", p),
		}
	}
	return nil
}
