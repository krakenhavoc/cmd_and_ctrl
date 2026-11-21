package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sink into Stupor // Soporific Springs — the FRONT face, Instant
// {1}{U}{U}:
//
//	"Return target spell or nonland permanent an opponent controls to
//	 its owner's hand."
//
// The land back (pay 3 life or enter tapped; {T}: Add {U}) is the
// mdfc_lands.go row under "<oracle>#1"; this is face 0, which keeps
// the bare oracle ID (game.CatalogKey).
//
// One pick across two zones — TargetSpellOrPermanent, Venser's clause
// — and "an opponent controls" governs both halves, the way "target
// artifact or creature an opponent controls" does: a spell an opponent
// cast (a spell's controller is the player who cast it, CR 112.2), or
// a nonland permanent an opponent controls. Your own spells and
// permanents are never offered.
//
// A spell is returned, not countered (ReturnSpellToHand): a spell
// that can't be countered still goes, and a returned copy of a spell
// ceases to exist (CR 707.10a). A permanent is bounced; a commander
// goes where its owner chooses (CR 903.9b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bcc6eece-75ea-494c-b33a-d4477d504e0b",
		Name:         "Sink into Stupor",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetSpellOrPermanent("target spell or nonland permanent an opponent controls",
			effects.OpponentControls(), effects.And(effects.Nonland(), effects.OpponentControls())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ReturnTargetSpellOrPermanentToHand(ctx.Game, item)
		},
	})
}
