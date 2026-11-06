package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Federation Resupply — {1}{W} Sorcery (Star Trek, #4973):
//
//	"Search your library for an Aura, Equipment, or Spacecraft card,
//	 reveal it, put it into your hand, then shuffle.
//	 Beam me up {3}{W} (You may cast this card from your graveyard for
//	 {3}{W} if you also return a creature you control to its owner's
//	 hand. Then exile this spell.)"
//
// Open the Armory's tutor with Spacecraft added to the subtypes, read off
// the card's printed type line in the library. Beam me up is
// trekB2BeamMeUp: a graveyard cast for {3}{W} plus the return of a
// creature you control, and the spell is exiled whenever it leaves the
// stack, resolved, countered or fizzled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "4450d4e0-2825-4536-bb0f-f3a36f1f682c",
		Name:             "Federation Resupply",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Tutors: 1},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{trekB2BeamMeUp("{3}{W}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player: ctx.Controller(),
				Predicate: func(c game.Card) bool {
					return c.HasSubtype("Aura") || c.HasSubtype("Equipment") || c.HasSubtype("Spacecraft")
				},
				Dest:    game.ZoneHand,
				Limit:   1,
				Reveal:  true,
				Shuffle: true,
				Reason:  "Federation Resupply — an Aura, Equipment, or Spacecraft card",
			}.Apply(ctx)
		},
	})
}
