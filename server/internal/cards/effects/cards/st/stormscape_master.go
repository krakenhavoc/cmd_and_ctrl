package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stormscape Master — Creature — Human Wizard {2}{U}{U}, 2/2 (Invasion,
// #5616):
//
//	"{W}{W}, {T}: Target creature gains protection from the color of your
//	 choice until end of turn.
//	 {B}{B}, {T}: Target player loses 2 life and you gain 2 life."
//
// The protection is Mother of Runes' colour prompt (the colour named is the
// one the creature is protected FROM), on any creature; the drain is Syphon
// Life's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bfcd6c5d-7189-4d84-b0df-0a5ce5f43528",
		Name:         "Stormscape Master",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{W}{W}, {T}: Target creature gains protection from the color of your choice until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerProtect, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "protection"))},
				Cost:    effects.Plus(effects.ManaCost("{W}{W}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					target := effects.FirstLegalBattlefieldTarget(ctx)
					if target == uuid.Nil {
						return nil
					}
					effects.ChooseProtectionColorFor(ctx, item.Controller, target, "Stormscape Master", "Stormscape Master — choose a color")
					return nil
				},
			},
			effects.S82OldB659DrainTwoGainTwo(),
		},
	})
}
