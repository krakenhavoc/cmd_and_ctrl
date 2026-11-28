package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Power Armor — Artifact {4} (Invasion, #5616):
//
//	"Domain — {3}, {T}: Target creature gets +1/+1 until end of turn for each
//	 basic land type among lands you control."
//
// Might of Alara's boost as an activated ability: the count is read once, as
// the ability resolves, and the boost ends in the cleanup step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1423b8e6-9165-4a89-a6ed-18085f460bca",
		Name:         "Power Armor",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Domain — {3}, {T}: Target creature gets +1/+1 until end of turn for each basic land type among lands you control.",
			Purpose: effects.WithAnswers(effects.ForTargets(game.TargetPurpose{Slot: 0, UntilEndOfTurn: game.TargetBenefit}), game.AnswerPump),
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ConB1DomainBoostTheTarget(1, "Power Armor — +1/+1 for each basic land type")(item, effects.NewContext(g, item))
			},
		}},
	})
}
