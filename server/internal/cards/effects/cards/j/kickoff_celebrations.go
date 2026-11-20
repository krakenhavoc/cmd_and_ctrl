package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kickoff Celebrations — Enchantment {1}{R}:
//
//	"Start your engines!
//	 When this enchantment enters, you may discard a card. If you do,
//	 draw two cards.
//	 Max speed — Sacrifice this enchantment: Creatures and Vehicles you
//	 control gain haste until end of turn."
//
// ADR 0138 (#2122, #2711). The enters trigger is the shared "may discard,
// if you do draw two" prompt. The haste grant picks its permanents as the
// ability resolves and not later (CR 611.2c): a creature or Vehicle that
// enters afterwards does not get haste.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "94868faf-9e88-4221-8db6-3ad10aa02da5",
		Name:            "Kickoff Celebrations",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{effects.StartYourEngines},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Kickoff Celebrations — you may discard a card. If you do, draw two cards",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayDiscardACardToDrawTwo("Kickoff Celebrations — discard a card to draw two?")(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{
			effects.MaxSpeedActivated(effects.ActivatedAbility{
				Label:   "Max speed — Sacrifice this enchantment: Creatures and Vehicles you control gain haste until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.SacrificeThis(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					you := item.Controller
					return effects.GrantKeywordUntilEOT{
						Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
							return c.Controller == you && (c.IsCreature() || c.HasSubtype("Vehicle"))
						},
						Keywords: []string{"haste"},
						Label:    "Kickoff Celebrations — creatures and Vehicles you control gain haste",
					}.Apply(effects.NewContext(g, item))
				},
			}),
		},
	})
}
