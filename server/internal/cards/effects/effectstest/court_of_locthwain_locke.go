package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #1729: the one-spell cast permission (CastPermission.CastsLeft) and
// the two cards it unblocks.

const (
	CourtOfLocthwainOracle = "6106ba65-f07b-41ec-8b59-e052bbde5529"
	LockeOracle            = "a800bec4-bacc-43f9-a773-dd795959935b"
)

// seedLibraryTop puts a named card on top of p's library.
func TopOfLibrary(p *game.Player, name, typeLine, manaCost string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine = typeLine
	c.ManaCost = manaCost
	p.Library.PushTop(c)
	return c.InstanceID
}
