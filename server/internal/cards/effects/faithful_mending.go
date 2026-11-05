package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Faithful Mending — Instant {W}{U}:
//
//	"You gain 2 life, draw two cards, then discard two cards.
//	 Flashback {1}{W}{U}"
//
// Faithless Looting's loot after the life gain, with flashback from
// the graveyard (CR 702.34a).
//
// No simplification. Star Trek Commander (#4974).
func init() {
	Register(Spec{
		OracleID:         "b997a982-09e1-4ae1-a972-3c2e5f101e8c",
		Name:             "Faithful Mending",
		Purpose:          game.Purpose{Draws: 2, Discards: 2},
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{1}{W}{U}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (GainLife{Amount: 2}).Apply(ctx); err != nil {
				return err
			}
			return lootOne(ctx.Game, item, 2)
		},
	})
}
