package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vizkopa Confessor — Creature — Human Cleric {3}{W}{B}, 1/3 (Gatecrash, S82
// OLD b408, #5565):
//
//	"Extort (Whenever you cast a spell, you may pay {W/B}. If you do, each
//	 opponent loses 1 life and you gain that much life.)
//	 When this creature enters, pay any amount of life. Target opponent reveals
//	 that many cards from their hand. You choose one of them and exile it."
//
// Extort is the keyword. The enters trigger names its opponent as it goes on the
// stack (CR 603.3d); the payment is a number prompt capped at the controller's
// life (CR 119.4, PayLifeAmount), paid on resolution, and the opponent then
// chooses that many cards from their hand to reveal (all of it if it is smaller),
// the controller chooses one of them (Acquisitions Expert's reveal-and-pick) and
// it is exiled. Paying 0 reveals nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9db7cb7f-1309-4f1d-9c9d-423c91c0896c",
		Name:         "Vizkopa Confessor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Extort("Vizkopa Confessor"),
			effects.Targeting(effects.WhenThisEnters("Vizkopa Confessor — pay any amount of life; target opponent reveals that many cards from their hand; you choose one and exile it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					victim := effects.TargetedPlayer(ctx)
					if victim == uuid.Nil {
						return nil
					}
					return effects.PayLifeAmount{
						Question: "Vizkopa Confessor — pay any amount of life",
						Unit:     "life",
						Then:     vizkopaConfessorPaid,
						Carry:    []uuid.UUID{victim},
					}.Apply(ctx)
				}), effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}

// vizkopaConfessorReveal is the reveal-and-exile: the opponent reveals the
// paid number of cards, and the controller exiles one of them.
var vizkopaConfessorReveal = effects.NewRevealFromHandPick("Vizkopa Confessor", "exiles",
	"choose-cards/s82-old-b408-vizkopa-confessor-reveal", "choose-cards/s82-old-b408-vizkopa-confessor-exile",
	func(ctx *effects.Context, _, card uuid.UUID) error {
		return ctx.Game.ExileCardsKeyedForEffect([]uuid.UUID{card}, game.KeyedTail{})
	})

// vizkopaConfessorPaid runs once the life is paid, with the number paid. An
// on-disk key: never renamed, never reused.
var vizkopaConfessorPaid = effects.NumberThen("number/s82-old-b408-vizkopa-confessor-paid", func(ctx *effects.Context, r game.NumberAnswer) error {
	if r.Number <= 0 || len(r.Carry) == 0 {
		return nil
	}
	return vizkopaConfessorReveal.Ask(ctx, r.Carry[0], r.Number)
})
