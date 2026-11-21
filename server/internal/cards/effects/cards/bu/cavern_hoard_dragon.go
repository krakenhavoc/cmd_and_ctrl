package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cavern-Hoard Dragon — Creature — Dragon {7}{R}{R}, 6/6 (The Hobbit):
//
//	"This spell costs {X} less to cast, where X is the greatest number of
//	 artifacts an opponent controls.
//	 Flying, trample, haste
//	 Whenever this creature deals combat damage to a player, you create a
//	 Treasure token for each artifact that player controls."
//
// The reduction is Ghalta's self cost modifier (CostsLessEach) over the
// largest per-opponent artifact count, a maximum rather than a sum; it eats
// generic mana only, so the spell never costs less than {R}{R}. The trigger
// is Hellkite Tyrant's: the damaged player is read off the damage event and
// the artifacts are counted as the trigger resolves, tokens included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "939e8448-013a-4aef-9a92-7c883a77a1a0",
		Name:            "Cavern-Hoard Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample", "haste"},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.HobB7GreatestArtifactsAnOpponentControls(),
				"This spell costs {X} less to cast, where X is the greatest number of artifacts an opponent controls."),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer(
				"Cavern-Hoard Dragon — create a Treasure for each artifact that player controls",
				func(g *game.Game, item *game.StackItem) error {
					n := len(effects.PermanentsControlledByMatching(g, item.Trigger.Event.Target, effects.Artifact()))
					if n == 0 {
						return nil
					}
					return effects.CreateToken{Template: effects.TreasureToken(), N: n}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
