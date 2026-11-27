package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thunderscape Apprentice — Creature — Human Wizard {R}, 1/1 (Invasion, #5616):
//
//	"{B}, {T}: Target player loses 1 life.
//	 {G}, {T}: Target creature gets +1/+1 until end of turn."
//
// Two printed activated abilities, each needing the {T}, so the Apprentice
// can't use either the turn it arrives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "34e325f0-b66f-44b9-a25c-da88357dbac4",
		Name:         "Thunderscape Apprentice",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.S82OldB659DrainOne(),
			{
				Label:   "{G}, {T}: Target creature gets +1/+1 until end of turn.",
				Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1)), game.AnswerPump),
				Cost:    effects.Plus(effects.ManaCost("{G}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					target := effects.FirstLegalBattlefieldTarget(ctx)
					if target == uuid.Nil {
						return nil
					}
					return effects.BoostUntilEOT{Target: target, Power: 1, Toughness: 1, Label: "Thunderscape Apprentice — +1/+1 until end of turn"}.Apply(ctx)
				},
			},
		},
	})
}
