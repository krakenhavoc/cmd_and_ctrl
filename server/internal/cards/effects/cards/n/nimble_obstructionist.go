package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nimble Obstructionist — {2}{U} Creature — Bird Wizard 3/1 (Commander 2020):
//
//	"Flash
//	 Flying
//	 Cycling {2}{U}
//	 When you cycle this card, counter target activated or triggered ability
//	 you don't control."
//
// Solar Blast's cycling trigger (from the graveyard, CR 702.29c) with Green
// Slime's ability target; the clause names only abilities an opponent
// controls (S82OldB123AnAbilityYouDontControl). It counters it as it resolves
// if the ability is still on the stack (CR 608.2b).
func init() {
	trig := effects.Targeting(
		effects.InGraveyard(effects.On(game.EventCycle, effects.Self, "Nimble Obstructionist — counter target activated or triggered ability you don't control",
			func(g *game.Game, item *game.StackItem) error {
				for _, t := range effects.NewContext(g, item).LegalTargets() {
					return effects.CounterTarget{StackID: t.ID}.Apply(effects.NewContext(g, item))
				}
				return nil
			})),
		effects.AbilityOnStack("target activated or triggered ability you don't control", effects.S82OldB123AnAbilityYouDontControl()))
	effects.Register(effects.Spec{
		OracleID:        "750d846e-2774-4296-8d9f-5207477d842d",
		Name:            "Nimble Obstructionist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Activated:       []effects.ActivatedAbility{effects.Cycling("{2}{U}")},
		Triggered:       []game.TriggeredAbility{trig},
	})
}
