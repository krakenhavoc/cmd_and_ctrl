package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shao Jun — Legendary Creature — Human Assassin {1}{U}{R}, 3/3
// (Assassin's Creed, #5466):
//
//	"Leap Strike — During your turn, Shao Jun has flying and first
//	 strike.
//	 Rope Dart — Tap two untapped artifacts you control: Shao Jun deals
//	 1 damage to each opponent."
//
// "Leap Strike" and "Rope Dart" are ability words. The first is two
// turn-gated layer 6 grants on itself (Shuttle Ace's shape); the second
// is Whirler Rogue's tap-two cost with no tap symbol, so Shao Jun can use
// it the turn it enters.
//
// No simplification.
func init() {
	const tapTwo = "two untapped artifacts you control"
	duringYourTurn := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return effects.ThisPermanent(target, g, source) && effects.IsActivePlayer(g, source.Controller)
	}
	effects.Register(effects.Spec{
		OracleID:     "126aab4d-df6e-4840-8cdc-49e890b1557f",
		Name:         "Shao Jun",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(duringYourTurn, "flying"),
			effects.KeywordGrant(duringYourTurn, "first strike"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Rope Dart — Tap two untapped artifacts you control: Shao Jun deals 1 damage to each opponent.",
			Purpose: game.Purpose{Answers: game.AnswerValue, DamageEachOpponent: 1},
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  2,
				Filter: effects.TargetPermanent(tapTwo, effects.Artifact()),
				Label:  tapTwo,
			}},
			Effect: effects.Do(effects.DamageEachOpponent{N: 1}),
		}},
	})
}
