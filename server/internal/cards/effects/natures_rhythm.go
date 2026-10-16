package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nature's Rhythm — Sorcery {X}{G}{G} (#3931):
//
//	"Search your library for a creature card with mana value X or less,
//	 put it onto the battlefield, then shuffle.
//	 Harmonize {X}{G}{G}{G}{G} (You may cast this card from your
//	 graveyard for its harmonize cost. You may tap a creature you control
//	 to reduce that cost by an amount of generic mana equal to its power.
//	 Then exile this spell.)"
//
// Green Sun's Zenith's search without the colour, and harmonize (CR
// 702.180) from the graveyard: an alternative cost that includes {X}, so
// X is announced as usual (CR 107.3a), and the tapped creature's power
// comes off the {X} once there is no printed generic left (CR 601.2f), so
// a 5-power creature buys five more X for the same mana.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "0259f063-3e3f-4e07-9066-f748d4819829",
		Name:             "Nature's Rhythm",
		Completeness:     CompletenessFull,
		XMatters:         true,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Harmonize("{X}{G}{G}{G}{G}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := ctx.X()
			src := ctx.Source()
			owner := item.Owner
			return SearchLibrary{
				Player: ctx.Controller(),
				Predicate: func(c game.Card) bool {
					return c.IsCreature() && c.ManaValue() <= x
				},
				Dest:   game.ZoneBattlefield,
				Limit:  1,
				Reason: "Nature's Rhythm — a creature card with mana value X or less",
				Source: src,
				Then: func(g *game.Game, _ []uuid.UUID) error {
					return g.ShuffleLibraryForEffect(owner)
				},
			}.Apply(ctx)
		},
	})
}
