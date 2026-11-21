package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oildeep Gearhulk — Artifact Creature — Construct {U}{U}{B}{B}, 4/4
// (Aetherdrift, #4464):
//
//	"Lifelink, ward {1}
//	 When this creature enters, look at target player's hand. You may
//	 choose a card from it. If you do, that player discards that card,
//	 then draws a card."
//
// Salt Vampire's shape with any card and a discard: the hand is LOOKED
// at (ChooseFromRevealedHand.Look), so only the Gearhulk's controller
// sees it, and any player may be the target, you included. The pick is
// optional, and the draw follows only a card actually chosen and
// discarded. It is a discard, so madness and "whenever a player
// discards" see it. Ward {1} is the shared Ward trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "aff0740f-51d5-41c7-ac13-97c298cfd743",
		Name:            "Oildeep Gearhulk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{1}"), "Oildeep Gearhulk — ward {1}"),
			effects.Targeting(
				effects.WhenThisEnters("Oildeep Gearhulk — look at target player's hand; you may have them discard a card, then draw",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						return effects.ChooseFromRevealedHand{
							Player:   effects.TargetedPlayer(ctx),
							Label:    "card",
							Look:     true,
							Optional: true,
							Then:     oildeepGearhulkDraw,
						}.Apply(ctx)
					}),
				effects.TargetPlayer("target player")),
		},
	})
}

// oildeepGearhulkDraw is "If you do, that player discards that card, then
// draws a card": the player whose card was discarded draws one. An
// on-disk key: never renamed, never reused.
var oildeepGearhulkDraw = effects.RevealedPickThen("revealed-pick/oildeep-gearhulk-draw", func(ctx *effects.Context, pick game.RevealedPick) error {
	if len(pick.Chosen) == 0 {
		return nil
	}
	return effects.DrawCards{Player: pick.FromPlayer, N: 1}.Apply(ctx)
})
