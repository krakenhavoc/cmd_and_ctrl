package effects

import (
	"fmt"

	"github.com/google/uuid"
)

// manifest_library.go — "manifest the top card of that player's library"
// (#5429), the card side of game/manifest_library.go.
//
// Manifest (manifest.go) is the top card of your own library.
// ManifestTopOfLibrary splits the two players the way CloakTopOfLibrary
// (cloak_action.go) does: the resolving item's controller manifests the
// card and controls the face-down 2/2 (CR 701.40a, CR 110.2a), and
// LibraryOwner is whose library it comes from. That player still owns the
// card, and may not look at it (CR 708.5).

// ManifestTopOfLibrary manifests the top card of LibraryOwner's library
// under Controller's control. Controller zero is the resolving item's
// controller; LibraryOwner zero is the controller, which is Manifest{N: 1}.
type ManifestTopOfLibrary struct {
	Controller   uuid.UUID
	LibraryOwner uuid.UUID
}

// Do manifests the card and returns the permanent that entered (uuid.Nil
// on an empty library).
func (m ManifestTopOfLibrary) Do(ctx *Context) (uuid.UUID, error) {
	controller := m.Controller
	if controller == uuid.Nil {
		controller = ctx.Controller()
	}
	owner := m.LibraryOwner
	if owner == uuid.Nil {
		owner = controller
	}
	id, err := ctx.Game.ManifestTopOfLibraryForEffect(controller, owner)
	if err != nil {
		return uuid.Nil, fmt.Errorf("manifest: %w", err)
	}
	return id, nil
}

// Apply manifests the card and ignores which permanent it became.
func (m ManifestTopOfLibrary) Apply(ctx *Context) error {
	_, err := m.Do(ctx)
	return err
}
