package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Endless Detour — Instant {G}{W}{U}:
//
//	"The owner of target spell, nonland permanent, or card in a
//	 graveyard puts it on their choice of the top or bottom of their
//	 library."
//
// One target clause over three zones: a spell on the stack, a nonland
// permanent, or a card in any graveyard. The OWNER chooses the lane
// (PutInLibraryInAnyOrder with the owner as chooser), and the move is
// the tuck route from wherever the card is, so a commander's owner is
// offered the command zone (CR 903.9). A spell is moved, not countered:
// one that can't be countered still goes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d65728cf-e9ce-46d8-a890-e1295b18cfbe",
		Name:         "Endless Detour",
		Completeness: CompletenessFull,
		Targets: &game.TargetSpec{
			Mode:  "any",
			Label: "target spell, nonland permanent, or card in a graveyard",
			Zones: []game.ZoneKind{game.ZoneStack, game.ZoneBattlefield, game.ZoneGraveyard},
			CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, zone game.ZoneKind) bool {
				switch zone {
				case game.ZoneStack, game.ZoneGraveyard:
					return true
				case game.ZoneBattlefield:
					return !c.IsLand()
				}
				return false
			},
			Min: 1, Max: 1,
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			ts := ctx.LegalTargets()
			if len(ts) == 0 {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(ts[0].ID)
			if !ok || c.Owner == uuid.Nil {
				return nil
			}
			return PutInLibraryInAnyOrder{
				Chooser:   c.Owner,
				Cards:     []uuid.UUID{ts[0].ID},
				Placement: game.LibraryPlaceTopOrBottom,
				Label:     "Endless Detour — put it on the top or bottom of your library",
			}.Apply(ctx)
		},
	})
}
