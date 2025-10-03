package betterauth

import (
	"better-auth/internal/auth"
	"better-auth/internal/config"
	"better-auth/pkg/plugins/core"
	"better-auth/pkg/transport"
)

type BetterAuth struct {
	core *auth.AuthCore
}

type Config = config.Config

type User = auth.User

type Session = auth.Session

type Organization = auth.Organization

type UserContext = auth.UserContext

type Plugin = core.Plugin

type Transport = transport.Transport

type SignInPayload = auth.SignInRequest

type SignUpPayload = auth.SignUpRequest
