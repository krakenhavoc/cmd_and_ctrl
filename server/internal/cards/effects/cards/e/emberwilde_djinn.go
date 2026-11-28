package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Emberwilde Djinn — Creature — Djinn {2}{R}{R}, 5/4 (Mirage, S82 OLD batch
// 809, #5633):
//
//	"Flying
//	 At the beginning of each player's upkeep, that player may pay {R}{R} or
//	 2 life. If the player does, they gain control of this creature."
//
// An every-upkeep trigger whose player (TriggeringActor, CR 608.2d) is asked
// once, with either payment on one prompt (S82OldB809ManaOrLifePayment). On
// a payment the Djinn changes hands for good (no stated duration, CR 611.2a)
// unless it is a new object by then (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3e1a7146-75d5-499f-8718-303bc6ea139b",
		Name:            "Emberwilde Djinn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Emberwilde Djinn — that player may pay {R}{R} or 2 life; if they do, they gain control of this creature",
				func(g *game.Game, item *game.StackItem) error {
					player := effects.TriggeringActor(item)
					if player == uuid.Nil {
						return nil
					}
					return effects.MayPay{
						Chooser:  player,
						Pay:      effects.S82OldB809ManaOrLifePayment("{R}{R}", 2),
						Question: "Emberwilde Djinn — pay {R}{R} or 2 life to gain control of it?",
						Then:     emberwildeDjinnPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// emberwildeDjinnPaid is the paid branch: the payer takes the Djinn. An
// on-disk key: never renamed, never reused.
var emberwildeDjinnPaid = effects.PayUnlessBranches("pay-unless/emberwilde-djinn-gain-control",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		return effects.GainControl{
			Target:     ctx.Source(),
			Controller: r.Chooser,
			Duration:   game.IndefiniteDuration(),
			Label:      "Emberwilde Djinn — gain control (no stated duration)",
		}.Apply(ctx)
	}, nil)
