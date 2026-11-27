package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aurochs Herd — Creature — Aurochs {5}{G}, 4/4 (Coldsnap):
//
//	"Trample
//	 When this creature enters, you may search your library for an Aurochs
//	 card, reveal it, put it into your hand, then shuffle.
//	 Whenever this creature attacks, it gets +1/+0 until end of turn for each
//	 other attacking Aurochs."
//
// The search is "you may" and a stated quality, so the player may find nothing
// (CR 701.23b). The attack bonus counts the other attacking creatures that are
// Aurochs as the trigger resolves, itself excluded.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b07bd1cd-4000-4e6b-b16b-283c76593377",
		Name:            "Aurochs Herd",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Aurochs Herd — you may search your library for an Aurochs card, reveal it, put it into your hand, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					sub := ctx.TextSubtype("Aurochs")
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.HasSubtype(sub) },
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Optional:  true,
						Reason:    "Aurochs Herd — an Aurochs card",
					}.Apply(ctx)
				}),
			effects.WheneverThisAttacks("Aurochs Herd — +1/+0 until end of turn for each other attacking Aurochs",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					sub := ctx.TextSubtype("Aurochs")
					n := 0
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.InstanceID != item.SourceCardID && c.AttackingTarget != uuid.Nil && c.HasSubtype(sub) {
							n++
						}
					}
					if n == 0 {
						return nil
					}
					return effects.BoostUntilEOT{Target: item.SourceCardID, Power: n, Label: "Aurochs Herd — +1/+0 for each other attacking Aurochs"}.Apply(ctx)
				}),
		},
	})
}
