package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Avizoa — Creature — Jellyfish {3}{U}, 2/2 (Weatherlight, S82 older sets batch 778, #5629):
//
//	"Flying
//	 {0}: This creature gets +2/+2 until end of turn. You skip your next untap step. Activate only
//	 once each turn."
//
// Flying rides PrintedKeywords. The ability is free, so it is limited only by its once-per-turn
// condition; resolving it pumps the Jellyfish and records that its controller skips their next untap
// step (CR 614.10, Ivory Gargoyle's one-shot skip of #6144). Each resolution adds a record, so two
// activations over two turns skip two untap steps, oldest first.
//
// No simplification.
func init() {
	const label = "{0}: This creature gets +2/+2 until end of turn. You skip your next untap step. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:        "91520db8-2d50-4c17-9f0e-2f92f63285f7",
		Name:            "Avizoa",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Cost:      effects.ManaCost("{0}"),
			Condition: effects.OncePerTurnActivation(label),
			Purpose:   game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2, Toughness: 2}},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := effects.ThisGetsUntilEndOfTurn(2, 2, "Avizoa — +2/+2")(g, item); err != nil {
					return err
				}
				g.SkipNextStepForEffect(ctx.Controller(), game.StepUntap, ctx.Source(), "Avizoa — you skip your next untap step")
				return nil
			},
		}},
	})
}
