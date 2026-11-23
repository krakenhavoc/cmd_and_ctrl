package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sycorax Commander — Creature — Alien Soldier {2}{B}{R}, 4/2:
//
//	"First strike, haste
//	 Sanctified Rules of Combat — When this creature enters, each
//	 opponent faces a villainous choice — That opponent discards all
//	 the cards in their hand, then draws that many cards minus one, or
//	 this creature deals damage to that player equal to the number of
//	 cards in their hand."
//
// A villainous choice (CR 701.55, #3140) asked of each opponent in APNAP
// order, one at a time (CR 701.55d). Both numbers are read when the
// opponent answers, from the hand they have then: the first option
// discards the whole hand (nothing to choose, so the random-order
// discard primitive is exact) and draws one fewer than the hand held,
// which is nothing for an empty hand; the second deals that many damage
// with the Commander as the source. An opponent with an empty hand may
// pick the damage and take none, as printed.
//
// No simplifications.
var sycoraxCommanderChoice = effects.NewVillainousChoice("sycorax-commander",
	"Sycorax Commander — discard your hand and draw that many minus one, or take damage equal to your hand?",
	effects.VillainousOption{
		Label:    "Discard all the cards in your hand, then draw that many cards minus one",
		Harm:     1,
		Possible: func(g *game.Game, facing uuid.UUID) bool { return effects.VillainousHandSize(g, facing) > 0 },
		Body: func(ctx *effects.Context, f effects.Facing) error {
			return sycoraxCommanderWheel.RunAt(ctx, effects.SeatIndexOf(ctx.Game, f.Player))
		},
	},
	effects.VillainousOption{
		Label:    "This creature deals damage to you equal to the number of cards in your hand",
		Harm:     1,
		Possible: func(g *game.Game, facing uuid.UUID) bool { return effects.VillainousHandSize(g, facing) > 0 },
		Body: func(ctx *effects.Context, f effects.Facing) error {
			n := effects.VillainousHandSize(ctx.Game, f.Player)
			if n <= 0 {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: f.Player, Amount: n}.Apply(ctx)
		},
	})

func init() {
	effects.Register(effects.Spec{
		OracleID:        "78ed3e01-6d22-429a-a9db-baff38d7bc03",
		Name:            "Sycorax Commander",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sycorax Commander — each opponent faces a villainous choice",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.FaceVillainousChoice{Players: ctx.Opponents(), Choice: sycoraxCommanderChoice}.Apply(ctx)
				}),
		},
	})
}

// sycoraxCommanderWheel is "discard all the cards in your hand, then draw
// that many cards minus one" for the player facing the choice. Each
// opponent faces it at once, so the steps find their player from the seat
// RunAt was given and file the count under that player. The draw waits for a
// prompt a discard asks (a replacement offer such as Library of Leng's; CR 608.2c, #4155).
var sycoraxCommanderWheel = effects.RegisterThenSteps("then-steps/sycorax-commander-wheel",
	effects.Instruction(func(ctx *effects.Context) error {
		who := effects.SeatAtOccurrence(ctx)
		n, err := effects.DiscardWholeHand(ctx.Game, who)
		ctx.RecordStep(sycoraxCommanderRecord(who), nil, n)
		return err
	}),
	effects.Instruction(func(ctx *effects.Context) error {
		who := effects.SeatAtOccurrence(ctx)
		n := ctx.StepRecord(sycoraxCommanderRecord(who)).N()
		if n <= 1 {
			return nil
		}
		return effects.DrawCards{Player: who, N: n - 1}.Apply(ctx)
	}),
)

// sycoraxCommanderRecord is the number of cards a player discarded.
func sycoraxCommanderRecord(who uuid.UUID) string {
	return "sycorax-commander/discarded/" + who.String()
}
