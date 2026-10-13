package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Barbarian Bully — Creature — Human Barbarian {2}{R}, 2/2:
//
//	"Discard a card at random: This creature gets +2/+2 until end of
//	 turn unless a player has this creature deal 4 damage to them.
//	 Activate only once each turn."
//
// ADR 0150 §3 (#1903): "a player" is any player, so taking the 4 damage
// is offered to every player in APNAP order from the active player, the
// Bully's controller included, until one takes it. The Bully deals it
// through the ordinary damage path, so it can be prevented, and taking
// it still stops the pump. The +2/+2 lands only on the Bully the
// ability came from (CR 400.7). The random discard is ADR 0109 §7's
// cost, and the once-each-turn limit counts activations (CR 602.1b).
//
// No simplification.
func init() {
	const label = "Discard a card at random: This creature gets +2/+2 until end of turn unless a player has this creature deal 4 damage to them. Activate only once each turn."
	Register(Spec{
		OracleID:     "5a88f616-986d-4e97-9457-26c783489123",
		Name:         "Barbarian Bully",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      DiscardAtRandom(1, "a card at random"),
			Condition: OncePerTurnActivation(label),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return AnyPlayerPays{
					Pay:          DamagePayment(4),
					Question:     "Barbarian Bully — take 4 damage to stop the +2/+2?",
					AcceptLabel:  "Take 4 damage",
					DeclineLabel: "Don't",
					OnNobody: func(ctx *Context) error {
						if !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
							return nil
						}
						return thisGetsUntilEndOfTurn(2, 2, "Barbarian Bully — +2/+2 until end of turn")(ctx.Game, ctx.Item)
					},
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
