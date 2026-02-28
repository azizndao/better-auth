// Package core provides core plugin interfaces for better-auth
package core

import "github.com/azizndao/grouter"

type Handler interface {
	RegisterRoutes(group grouter.RouteGroup) error
}
