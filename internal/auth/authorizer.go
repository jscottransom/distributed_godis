package auth

import (
	"fmt"
	"log"

	"github.com/casbin/casbin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func New(model, policy string) *Authorizer {
	enforcer := casbin.NewEnforcer("/Users/jimislick/dev/distributed_godis/internal/auth/model.conf", "/Users/jimislick/dev/distributed_godis/internal/auth/policy.csv")

	return &Authorizer{
		enforcer: enforcer,
	}
}

type Authorizer struct {
	enforcer *casbin.Enforcer
}

func (a *Authorizer) Authorize(subject, object, action string) error {
	
	log.Printf("Attempting to Authorize")
	
	if !a.enforcer.Enforce(subject, object, action) {
		msg := fmt.Sprintf(
			"%s not permitted to %s to %s",
			subject,
			action,
			object,
		)
		log.Printf("Something went wrong authorizing")
		st := status.New(codes.PermissionDenied, msg)
		return st.Err()
	}
	log.Printf("Nothing went wrong authorizing")
	return nil
}