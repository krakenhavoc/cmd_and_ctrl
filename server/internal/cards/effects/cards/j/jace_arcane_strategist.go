package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jace, Arcane Strategist — Legendary Planeswalker {4}{U}{U}, loyalty 4
// (Jumpstart 2022, #5948):
//
//	"Whenever you draw your second card each turn, put a +1/+1 counter on
//	 target creature you control.
//	 +1: Draw a card.
//	 −7: Creatures you control can't be blocked this turn."
//
// Alandra's second-draw tally trigger (read after the draw, so it fires on
// that draw alone) with a target chosen as the trigger goes on the stack;
// the −7 is Glaring Spotlight's unblockable grant, whose creature set is
// the one you control as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4126bab7-6a2a-43b2-ba65-77299f9bb8cf",
		Name:            "Jace, Arcane Strategist",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.B41DrewYourSecondCardThisTurn(ev, source, g)
				}, "Jace, Arcane Strategist — put a +1/+1 counter on target creature you control",
					effects.PlusOneCounterOnFirstLegalTarget),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Draw a card.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{Draws: 1},
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			},
			{
				Label: "−7: Creatures you control can't be blocked this turn.",
				Cost:  effects.LoyaltyCost(-7),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.RestrictUntilEOT{
						Scope:        game.ScopeYourCreatures,
						Restrictions: game.CantBeBlocked,
						Label:        "Jace, Arcane Strategist — can't be blocked",
					}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
