package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ferocity of the Wilds — Enchantment {2}{R} (Throne of Eldraine, #5516):
//
//	"Attacking non-Human creatures you control get +1/+0 and have trample."
//
// Two statics (layer 7c and layer 6) over the same set: your creatures that
// are attacking and are not Humans, re-read on every recompute.
//
// No simplification.
func init() {
	applies := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsCreature() && target.Controller == source.Controller && target.AttackingTarget != uuid.Nil &&
			!target.HasSubtype(source.TextSubtype("Human"))
	}
	pump := game.StaticAbility{
		Label:                    "Attacking non-Human creatures you control get +1/+0.",
		Layer:                    game.Layer7PT,
		SubLayer:                 game.SubLayer7C_Modify,
		DependsOnAttackingStatus: true,
		AppliesTo:                applies,
		Apply:                    effects.PumpBy(1, 0),
	}
	trample := effects.KeywordGrant(applies, "trample")
	trample.Label = "Attacking non-Human creatures you control have trample."
	trample.DependsOnAttackingStatus = true
	effects.Register(effects.Spec{
		OracleID:     "c75bdc5f-22f2-4558-bb8f-5a53ff7ab1ac",
		Name:         "Ferocity of the Wilds",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{pump, trample},
	})
}
