package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Momentum Breaker — Enchantment {1}{B}:
//
//	"Start your engines!
//	 When this enchantment enters, each opponent sacrifices a creature or
//	 Vehicle of their choice. Each opponent who can't discards a card.
//	 {2}, Sacrifice this enchantment: You gain life equal to your speed."
//
// ADR 0138 (#2122, #2711). The edict is EachPlayerSacrifices, whose
// continuation runs once every opponent has answered: an opponent who was
// never asked (no creature or Vehicle) is exactly "an opponent who can't",
// and gets a discard prompt of their own to pick the card (CR 701.9a). An
// opponent who sacrificed does not discard. The life gain reads the
// speed as the ability resolves (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "087d2e42-c69f-4cd2-b2e3-e0b64ac2893b",
		Name:            "Momentum Breaker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{effects.StartYourEngines},
		Triggered: []game.TriggeredAbility{
			effects.TriggerEdict(effects.WhenThisEnters("Momentum Breaker — each opponent sacrifices a creature or Vehicle; each who can't discards a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.EachPlayerSacrifices{
						ExceptController: true,
						Match:            effects.Or(effects.Creature(), effects.OfSubtype("Vehicle")),
						Label:            "a creature or Vehicle",
						Keyed:            momentumBreakerEdicted,
					}.Apply(effects.NewContext(g, item))
				})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Sacrifice this enchantment: You gain life equal to your speed.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GainLife{Amount: effects.YourSpeed(g, item.Controller)}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

// momentumBreakerEdicted is the discard for each opponent who could not
// sacrifice, keyed (#4290).
var momentumBreakerEdicted = effects.RegisterEdictThen("prompt-run/momentum-breaker-discard", func(ctx *effects.Context, sacrificed game.PromptedSacrifices) error {
	var cantSacrifice []uuid.UUID
	for _, opp := range ctx.Opponents() {
		if !sacrificed.Sacrificed(opp) {
			cantSacrifice = append(cantSacrifice, opp)
		}
	}
	if len(cantSacrifice) == 0 {
		return nil
	}
	ctx.Game.PlayersDiscardForEffect(cantSacrifice, game.DiscardPrompt{
		Source:   ctx.Source(),
		N:        1,
		Question: "Momentum Breaker — discard a card",
	})
	return nil
})
