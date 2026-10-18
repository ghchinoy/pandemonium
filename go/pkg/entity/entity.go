// Package entity provides the base actor implementations for Pandemonium.
package entity

import (
	"github.com/ghchinoy/pandemonium/pkg/types"
)

// ActionSpec specifies what the entity is being asked to do.
type ActionSpec = types.ActionSpec

// Entity is the base interface for all actors.
type Entity = types.Entity
