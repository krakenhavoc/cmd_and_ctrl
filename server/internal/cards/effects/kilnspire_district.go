package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kilnspire District — Plane — Ravnica (Planechase Anthology, #4870):
//
//	"When you planeswalk to Kilnspire District and at the beginning of your
//	 first main phase, put a charge counter on Kilnspire District, then add
//	 {R} for each charge counter on it.
//	 Whenever chaos ensues, you may pay {X}. If you do, Kilnspire District
//	 deals X damage to any target."
//
// A plane (ADR 0216 §6). The counters sit on the face-up card in the command
// zone (PutCountersOnFaceUpPlanarCardForEffect); a planeswalk away makes it a
// new object with none (CR 901.7a), and every counter counts, whoever put
// it there. The mana is added as the ability resolves, to the planar
// controller. The chaos ability targets as it goes on the stack, and X is
// chosen and paid as it resolves (the ruling): MayPayX asks then, and the
// damage goes to the target the ability chose.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fc650404-991c-4639-b311-c35df81a254a",
		Name:         "Kilnspire District",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			planesR6WhenYouPlaneswalkToThisAndAtYourFirstMain("Kilnspire District — put a charge counter on it, then add {R} for each charge counter on it",
				func(g *game.Game, item *game.StackItem) error {
					n, ok := g.PutCountersOnFaceUpPlanarCardForEffect(item.SourceCardID, game.CounterCharge, 1)
					if !ok || n <= 0 {
						return nil
					}
					return AddMana{Amount: ManaN(n, "R")}.Apply(NewContext(g, item))
				}),
			Targeting(WheneverChaosEnsues("Kilnspire District — you may pay {X}; it deals X damage to any target",
				func(g *game.Game, item *game.StackItem) error {
					return MayPayX{
						Cost:  "{X}",
						Label: "Kilnspire District",
						Buys:  "deal X damage to the target",
						Goal:  XAsHighAsYouCan,
						Unit:  game.PayAmountDamage,
						Then:  kilnspireDistrictPaid,
					}.Apply(NewContext(g, item))
				}), TargetAny()),
		},
	})
}

// kilnspireDistrictPaid is the paid half, keyed: X rides the prompt and
// the target is the asking ability's.
var kilnspireDistrictPaid = PayUnlessBranches("pay-unless/kilnspire-district-damage",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if r.Number <= 0 || r.Item == nil || len(r.Item.Targets) == 0 {
			return nil
		}
		return DealDamage{
			Source: r.Item.SourceCardID,
			Target: r.Item.Targets[0].ID,
			Amount: r.Number,
		}.Apply(ctx)
	}, nil)
