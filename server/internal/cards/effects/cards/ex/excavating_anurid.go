package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Excavating Anurid — Creature — Frog Beast {4}{G}, 4/4 (Modern Horizons):
//
//	"When this creature enters, you may sacrifice a land. If you do, draw a
//	 card.
//	 Threshold — As long as there are seven or more cards in your graveyard,
//	 this creature gets +1/+1 and has vigilance."
//
// The enters trigger is Yuma's "may sacrifice a land, if you do draw": the
// draw is the sacrifice's keyed continuation, paid only for a land that really
// left. Threshold is an ability word with no rules of its own (CR 207.2c): a
// layer 7c self-modifier (CR 613.4c) and a layer 6 keyword, both read live
// against the graveyard on every recompute, like Anurid Barkripper and
// Dreadwing Scavenger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0239b661-80be-4b52-b3b7-fca229192b58",
		Name:         "Excavating Anurid",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.FraThresholdSelfPT(1, 1),
			effects.FraThresholdSelfKeywords("vigilance"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Excavating Anurid — you may sacrifice a land; if you do, draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "Excavating Anurid — sacrifice a land to draw a card?",
						effects.Land(), false, excavatingAnuridSacrificed)
				}),
		},
	})
}

// excavatingAnuridSacrificed is "if you do, draw a card": keyed. An on-disk
// key: never renamed, never reused.
var excavatingAnuridSacrificed = effects.RegisterSacrificedOne("prompt-run/excavating-anurid-sacrifice", func(ctx *effects.Context, _ uuid.UUID) error {
	return effects.DrawCards{N: 1}.Apply(ctx)
})
