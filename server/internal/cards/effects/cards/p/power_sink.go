package cardsp

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Power Sink — Instant {X}{U} (Urza's Saga, S82 OLD batch 732, #5624):
//
//	"Counter target spell unless its controller pays {X}. If that player
//	 doesn't, they tap all lands with mana abilities they control and lose all
//	 unspent mana."
//
// Broken Ambitions' tax: the spell's controller is asked to pay {X} as the
// spell resolves, with the spell still on the stack. A decline, or a "pay" they
// can't fund, counters it (if it is still there) and then that player taps
// every untapped land they control that has a mana ability, without adding any
// mana (CR 701.26 — it is not activating the abilities), and loses all mana
// in their pool, the mana they just spent included only if they paid, which
// they didn't.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "39412e6d-2837-4729-abf9-e64a5ba87e40",
		Name:         "Power Sink",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			cost := "{" + strconv.Itoa(ctx.X()) + "}"
			return effects.CounterUnlessPaid{
				StackID:  targets[0].ID,
				Cost:     cost,
				Question: "Power Sink — pay " + cost + " or your spell is countered, your lands tapped and your mana lost",
				Then:     powerSinkAnswer,
			}.Apply(ctx)
		},
	})
}

// powerSinkAnswer is Power Sink's two branches, keyed: a payment lets the spell
// resolve; a decline counters it, taps the payer's lands with mana abilities
// and empties their pool. An on-disk key.
var powerSinkAnswer = effects.PayUnlessBranches("pay-unless/power-sink-tap-lands",
	nil,
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		if err := ctx.Game.CounterIfStillOnStackForEffect(r.Guards); err != nil {
			return err
		}
		payer := r.Chooser
		if payer == uuid.Nil {
			return nil
		}
		for _, id := range ctx.Game.LandsControlledForEffect(payer) {
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok || c.Tapped || len(game.ManaAbilitiesOf(&c)) == 0 {
				continue
			}
			if err := (effects.TapTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
		}
		ctx.Game.LoseAllUnspentManaForEffect(payer)
		return nil
	})
