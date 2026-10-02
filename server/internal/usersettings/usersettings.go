// Package usersettings is the account half of the client's settings
// (ADR 0110 section 4, migration 0008's user_settings): one JSON object
// per signed-in person, the synced subset of the client's Settings,
// written with a revision check so two tabs or devices never silently
// overwrite each other.
//
// The server does not interpret the body beyond its shape: a JSON
// object, at most MaxBodyBytes, nested at most MaxDepth deep. The
// client's migrate stays the one schema validator, because a server
// copy of the schema would drift from it.
//
// Two Store implementations: SQLStore (over internal/db) and NoStore,
// for a deployment with no database, where no principal carries a
// UserID and so nothing calls the store except defensively.
package usersettings

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	// MaxBodyBytes is the cap on a settings body, ADR 0110 section 4.
	MaxBodyBytes = 32 * 1024
	// MaxDepth is the deepest nesting a body may have. The body itself
	// is depth 1, so {"a":{"b":{"c":{}}}} is depth 4.
	MaxDepth = 4
	// MinVersion and MaxVersion bound the client's settings version.
	MinVersion = 1
	MaxVersion = 1000
)

var (
	// ErrNotFound: the user has no settings row. Callers answer
	// {revision: 0}.
	ErrNotFound = errors.New("usersettings: not found")
	// ErrRevisionConflict: If-Match named a revision that is not the
	// stored one. Put also returns the current copy.
	ErrRevisionConflict = errors.New("usersettings: revision has moved")
	// ErrVersionTooOld: the write's client version is below the stored
	// one. A stale tab must not stamp an older version over a newer
	// client's copy. Put also returns the current copy.
	ErrVersionTooOld = errors.New("usersettings: a newer version saved these settings")
	// ErrInvalid wraps every refusal of a malformed version or body.
	ErrInvalid = errors.New("usersettings: invalid settings")
	// ErrNoStore is what NoStore's Put answers.
	ErrNoStore = errors.New("usersettings: no store configured")
)

// Settings is one user_settings row.
type Settings struct {
	// Version is the client's SETTINGS_VERSION that wrote the body.
	Version int
	// Revision is bumped on every write. 0 means "no row yet" and is
	// what a caller passes as If-Match to create the first copy.
	Revision  int64
	Body      json.RawMessage
	UpdatedAt time.Time
}

// Store reads and writes one person's settings. Safe for concurrent
// use.
type Store interface {
	// Get reads the user's copy. ErrNotFound if there is none.
	Get(ctx context.Context, user uuid.UUID) (Settings, error)
	// Put writes the user's copy if the stored revision equals
	// ifMatch (0 means "there must be no row yet"). It returns the
	// stored copy, with its new revision.
	//
	// Refusals: ErrInvalid (bad version or body, nothing read or
	// written), ErrRevisionConflict, and ErrVersionTooOld. The last two
	// return the current stored copy beside the error (a conflict on a
	// first write that lost a race returns a zero copy). The revision
	// is checked before the version, so a stale revision is always a
	// conflict.
	Put(ctx context.Context, user uuid.UUID, version int, body json.RawMessage, ifMatch int64) (Settings, error)
}

// Validate checks a write's version and body against the caps. It is
// what Put runs first, exported so a route can refuse before it spends
// a rate-limit token on the database.
func Validate(version int, body json.RawMessage) error {
	if version < MinVersion || version > MaxVersion {
		return errors.Join(ErrInvalid, errors.New("version must be an integer from 1 to 1000"))
	}
	if len(body) > MaxBodyBytes {
		return errors.Join(ErrInvalid, errors.New("settings are larger than 32 KiB"))
	}
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 || trimmed[0] != '{' || !json.Valid(trimmed) {
		return errors.Join(ErrInvalid, errors.New("settings must be a JSON object"))
	}
	if jsonDepth(trimmed) > MaxDepth {
		return errors.Join(ErrInvalid, errors.New("settings are nested too deeply"))
	}
	return nil
}

// jsonDepth is the nesting depth of valid JSON: the number of
// containers open at once at the deepest point. Brackets inside
// strings do not count.
func jsonDepth(b []byte) int {
	depth, deepest := 0, 0
	inString, escaped := false, false
	for _, c := range b {
		switch {
		case inString && escaped:
			escaped = false
		case inString && c == '\\':
			escaped = true
		case inString && c == '"':
			inString = false
		case inString:
		case c == '"':
			inString = true
		case c == '{' || c == '[':
			depth++
			if depth > deepest {
				deepest = depth
			}
		case c == '}' || c == ']':
			depth--
		}
	}
	return deepest
}

// NoStore is the Store for a deployment with no database.
type NoStore struct{}

// Get always reports ErrNotFound.
func (NoStore) Get(context.Context, uuid.UUID) (Settings, error) { return Settings{}, ErrNotFound }

// Put always fails: there is nowhere to put the row.
func (NoStore) Put(context.Context, uuid.UUID, int, json.RawMessage, int64) (Settings, error) {
	return Settings{}, ErrNoStore
}

var _ Store = NoStore{}
