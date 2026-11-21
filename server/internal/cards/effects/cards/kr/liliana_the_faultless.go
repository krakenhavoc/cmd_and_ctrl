package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Liliana the Faultless — Legendary Creature — Human Cleric {W}, 1/1
// (Reality Fracture):
//
//	"Whenever another creature or planeswalker you control enters, you
//	 gain 1 life.
//	 {1}, {T}, Discard a card: Another target creature or planeswalker
//	 you control gains hexproof until end of turn."
//
// The discard is a cost paid at announce (the activator picks the card
// from hand), so a response that kills the target does not give it back.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "22c2e66d-a2a1-46db-a70a-dff5ad5d6c14",
		Name:         "Liliana the Faultless",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AnotherCreatureOrPlaneswalkerEnteredUnderYourControl,
				"Liliana the Faultless — you gain 1 life",
				effects.Do(effects.GainLife{Amount: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{1}, {T}, Discard a card: Another target creature or planeswalker you control gains hexproof until end of turn.",
			// #4734 slice 7: hexproof keeps the target out of reach of what
			// targets it.
			Purpose: game.Purpose{Answers: game.AnswerProtect, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "hexproof"))},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.DiscardACard()),
			Targets: effects.Another(effects.TargetPermanent("another target creature or planeswalker you control", effects.Or(effects.Creature(), effects.Planeswalker()), effects.YouControl())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"hexproof"}}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
