package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crater Elemental — Creature — Elemental {2}{R}, 0/6 (Dragons of Tarkir, S82
// older sets):
//
//	"{R}, {T}, Sacrifice this creature: It deals 4 damage to target creature.
//	 Formidable — {2}{R}: This creature has base power 8 until end of turn.
//	 Activate only if creatures you control have total power 8 or greater."
//
// The first ability pays the sacrifice as a cost (CR 118.12) and the Elemental
// deals the damage as it last existed (CR 608.2h). The second is a layer 7b set of
// base power (CR 613.4b) pinned to the Elemental until end of turn, so counters
// and pumps still add to the 8; it needs the formidable total (CR 602.5b), read as
// the ability is announced. A new object (the Elemental flickered in response) is
// not changed (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8ea607cd-c14b-4bb1-aba1-c5cdc397a825",
		Name:         "Crater Elemental",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{R}, {T}, Sacrifice this creature: It deals 4 damage to target creature.",
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 4)),
				Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost(), effects.SacrificeThis()),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DamageToFirstTargetStep(4)(item, effects.NewContext(g, item))
				},
			},
			effects.OldB324FormidableRow(effects.ActivatedAbility{
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{2}{R}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !effects.SourceIsStillThisPermanent(g, item) {
						return nil
					}
					return effects.UntilEndOfTurn(effects.NewContext(g, item), item.SourceCardID, nil,
						"Crater Elemental — base power 8 until end of turn", game.SetBasePowerMod(8))
				},
			}, "{2}{R}: This creature has base power 8 until end of turn."),
		},
	})
}
