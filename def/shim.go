package def

import (
	"time"

	"github.com/google/uuid"
)

// Shim represents a connected mqvelium client.
type Shim struct {
	Name           string
	UUID           uuid.UUID
	SessionTimeout time.Time
}
