package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// For the Ancestors — Instant {2}{G} (Commander Masters, #5948):
//
//	"Choose a creature type. Look at the top six cards of your library. You
//	 may reveal any number of cards of the chosen type from among them and
//	 put the revealed cards into your hand. Put the rest on the bottom of
//	 your library in a random order.
//	 Flashback {3}{G}"
//
// The type is chosen first (CR 700.8 style prompt), then the six cards are
// looked at: the taken cards are revealed, and everything else goes under
// the library in a random order. A changeling matches every type
// (CR 702.73a). Flashback exiles the spell as it leaves the stack (CR 702.34).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "4caa3c55-41b3-48a3-b6a8-4d23aff7d00d",
		Name:             "For the Ancestors",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{3}{G}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			effects.ChooseCreatureTypeTail(ctx.Game, item.Controller, item.SourceCardID, "For the Ancestors — choose a creature type",
				ctx.ItemTail(forTheAncestorsTypeChosen))
			return nil
		},
	})
}

var forTheAncestorsTypeChosen = effects.CreatureTypeChosen("then/for-the-ancestors-type-chosen", func(ctx *effects.Context, t string, _ game.TailOutcome) error {
	player := ctx.Controller()
	return effects.TakeFromLibraryToHand{
		Player:   player,
		Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 6),
		Match:    func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.HasSubtypeWord(t) },
		Optional: true,
		Label:    "For the Ancestors — reveal any number of " + t + " cards and put them into your hand",
		Key:      effects.TakeRevealRestBottomRandom,
	}.Apply(ctx)
})
