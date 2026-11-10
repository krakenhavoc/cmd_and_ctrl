package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zoyowa Lava-Tongue — Legendary Creature — Goblin Warlock {B}{R}, 2/2
// (The Lost Caverns of Ixalan, #4470):
//
//	"Deathtouch
//	 At the beginning of your end step, if you descended this turn, each
//	 opponent may discard a card or sacrifice a permanent of their choice.
//	 Zoyowa deals 3 damage to each opponent who didn't."
//
// The intervening "if" (CR 603.4) is descend's: a permanent card was put
// into your graveyard this turn, read as the end step begins and as the
// trigger resolves. Each opponent is asked on their own, one question
// naming both payments (a branch they can't pay isn't offered, CR 118.3),
// and a "no", or an opponent with no card in hand and no permanent, takes
// 3 damage from Zoyowa as it last existed (CR 608.2h). An opponent who
// pays takes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e90f298f-c54b-4fe8-82db-5b9153a57712",
		Name:            "Zoyowa Lava-Tongue",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			AtYourEndStepIfDescended("Zoyowa Lava-Tongue — each opponent may discard a card or sacrifice a permanent; Zoyowa deals 3 damage to each who didn't",
				zoyowaOffer),
		},
	})
}

var zoyowaPayment = EitherPayment(DiscardPayment(1), SacrificePayment(1, "permanent", "permanents", game.PermanentQuery{}))

func zoyowaOffer(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, opp := range ctx.Opponents() {
		cost, action := zoyowaPayment.prompt()
		if err := (UpkeepPayUnless{
			Chooser:  opp,
			Cost:     cost,
			Action:   action,
			Question: "Zoyowa Lava-Tongue — discard a card or sacrifice a permanent, or take 3 damage?",
			Else:     zoyowaDamage,
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// zoyowaDamage is the "didn't": Zoyowa deals 3 damage to the opponent who
// was asked.
var zoyowaDamage = PayUnlessBranches("pay-unless/zoyowa-lava-tongue-damage", nil,
	func(ctx *Context, r game.PayUnlessAnswer) error {
		return DealDamage{Source: ctx.Source(), Target: r.Chooser, Amount: 3}.Apply(ctx)
	})
