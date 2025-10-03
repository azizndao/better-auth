package betterauth

import (
	"better-auth/internal/auth"
	"better-auth/internal/config"
	"better-auth/internal/dto"
	"better-auth/pkg/plugins/core"
	"better-auth/pkg/transport"
)

type BetterAuth struct {
	core *auth.AuthCore
}

type Config = config.Config

type User = dto.User

type Session = dto.Session

type UserContext = dto.AuthData

type Plugin = core.Plugin

type Transport = transport.Transport

type SignInPayload = dto.SignInPayload

type SignUpPayload = dto.SignUpPayload
