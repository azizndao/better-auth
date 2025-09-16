// Package core provides core plugin interfaces for better-auth
package core

import "better-auth/pkg/router"

type Handler interface {
	RegisterRoutes(group router.RouteGroup) error
}
