package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Attentive Skywarden — Creature — Phyrexian Kor {2}{W}, 2/2 (March of
// the Machine):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player or battle,
//	 transform up to one target Incubator token you control."
//
// The target is chosen as the trigger goes on the stack and is "up to
// one", so it may be declined; it is re-checked as the ability resolves
// (CR 608.2b). Damage to a battle is the same EventDealDamage as damage to
// a player (Bilbo, Unexpected Adventurer).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "896213bb-28a0-45a4-abbd-35af51f1d9aa",
		Name:            "Attentive Skywarden",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventDealDamage, effects.ThisDealtCombatDamageToAPlayerOrBattle,
					"Attentive Skywarden — transform up to one target Incubator token you control",
					func(g *game.Game, item *game.StackItem) error {
						return effects.TransformTargetIncubator(effects.NewContext(g, item))
					}),
				effects.TargetIncubatorYouControl().WithCount(0, 1)),
		},
	})
}
