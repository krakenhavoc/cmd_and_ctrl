package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:        "087d2e42-c69f-4cd2-b2e3-e0b64ac2893b",
		Name:            "Momentum Breaker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{StartYourEngines},
		Triggered: []game.TriggeredAbility{
			TriggerEdict(WhenThisEnters("Momentum Breaker — each opponent sacrifices a creature or Vehicle; each who can't discards a card",
				func(g *game.Game, item *game.StackItem) error {
					return EachPlayerSacrifices{
						ExceptController: true,
						Match:            Or(Creature(), OfSubtype("Vehicle")),
						Label:            "a creature or Vehicle",
						Then: func(ctx *Context, sacrificed game.PromptedSacrifices) error {
							var cantSacrifice []uuid.UUID
							for _, opp := range ctx.Opponents() {
								if !sacrificed.Sacrificed(opp) {
									cantSacrifice = append(cantSacrifice, opp)
								}
							}
							if len(cantSacrifice) == 0 {
								return nil
							}
							return ctx.Game.PlayersDiscardThenForEffect(cantSacrifice, game.DiscardPrompt{
								Source:   ctx.Source(),
								N:        1,
								Question: "Momentum Breaker — discard a card",
							}, nil)
						},
					}.Apply(NewContext(g, item))
				})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, Sacrifice this enchantment: You gain life equal to your speed.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}"), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GainLife{Amount: YourSpeed(g, item.Controller)}.Apply(NewContext(g, item))
			},
		}},
	})
}
