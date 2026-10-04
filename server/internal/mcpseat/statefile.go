package mcpseat

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// savedSession is the one file per seat §8 allows: enough to reattach to
// the seat after the MCP client restarts the binary. The invite token is
// NOT in it; it is the owner's to paste.
type savedSession struct {
	Origin    string    `json:"origin"`
	GameID    uuid.UUID `json:"game_id"`
	PlayerID  uuid.UUID `json:"player_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// stateStore keeps saved sessions under
// $XDG_STATE_HOME/cmdctrl-mcpseat/<server host>/<game id>.json.
type stateStore struct {
	root string
}

// DefaultStateDir is $XDG_STATE_HOME/cmdctrl-mcpseat, falling back to
// ~/.local/state/cmdctrl-mcpseat.
func DefaultStateDir() (string, error) {
	if x := os.Getenv("XDG_STATE_HOME"); x != "" && filepath.IsAbs(x) {
		return filepath.Join(x, "cmdctrl-mcpseat"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("no state directory: %w", err)
	}
	return filepath.Join(home, ".local", "state", "cmdctrl-mcpseat"), nil
}

// errInsecureState is the ssh-style refusal: a token file or directory
// anyone but the owner can read is not used.
var errInsecureState = errors.New("insecure permissions")

func (s stateStore) dirFor(origin string) string {
	host := strings.TrimPrefix(strings.TrimPrefix(origin, "https://"), "http://")
	host = strings.NewReplacer(":", "_", "/", "_", "\\", "_").Replace(host)
	return filepath.Join(s.root, host)
}

func (s stateStore) pathFor(origin string, gameID uuid.UUID) string {
	return filepath.Join(s.dirFor(origin), gameID.String()+".json")
}

// checkPrivate refuses a path that is group- or world-accessible.
func checkPrivate(path string, fi fs.FileInfo) error {
	if fi.Mode().Perm()&0o077 != 0 {
		return fmt.Errorf("%s is mode %#o, readable by others: %w (want 0700 for the directory, 0600 for the file)",
			path, fi.Mode().Perm(), errInsecureState)
	}
	return nil
}

// load reads a saved session. A missing file is (nil, nil).
func (s stateStore) load(origin string, gameID uuid.UUID) (*savedSession, error) {
	for _, dir := range []string{s.root, s.dirFor(origin)} {
		fi, err := os.Stat(dir)
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		if err := checkPrivate(dir, fi); err != nil {
			return nil, err
		}
	}
	path := s.pathFor(origin, gameID)
	fi, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := checkPrivate(path, fi); err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(path) // #nosec G304 -- our own state path
	if err != nil {
		return nil, err
	}
	var ss savedSession
	if err := json.Unmarshal(raw, &ss); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if ss.Origin != origin || ss.GameID != gameID {
		return nil, fmt.Errorf("%s names another table", path)
	}
	return &ss, nil
}

// save writes the session 0600 through a temp file and a rename, in a
// directory created 0700.
func (s stateStore) save(ss *savedSession) error {
	dir := s.dirFor(ss.Origin)
	for _, d := range []string{s.root, dir} {
		if err := os.MkdirAll(d, 0o700); err != nil {
			return err
		}
		fi, err := os.Stat(d)
		if err != nil {
			return err
		}
		if err := checkPrivate(d, fi); err != nil {
			return err
		}
	}
	raw, err := json.Marshal(ss)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".session-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(raw); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, s.pathFor(ss.Origin, ss.GameID))
}

// remove deletes a saved session; a missing file is not an error.
func (s stateStore) remove(origin string, gameID uuid.UUID) error {
	err := os.Remove(s.pathFor(origin, gameID))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}
