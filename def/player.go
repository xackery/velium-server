package def

import (
	"time"

	"github.com/google/uuid"
)

type Player struct {
	Name           string
	UUID           uuid.UUID
	SessionTimeout time.Time
}
