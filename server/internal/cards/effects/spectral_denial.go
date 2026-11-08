package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spectral Denial — Instant {X}{U} (Tarkir: Dragonstorm, #4463):
//
//	"This spell costs {1} less to cast for each creature you control with
//	 power 4 or greater.
//	 Counter target spell unless its controller pays {X}."
//
// The reduction is Salt Road Packbeast's CostsLessEach over the creatures
// the caster controls, counted by their current power, so anthems and
// counters count. It spends against generic mana only (CR 601.2f), which
// for this spell is the {X}: it lowers what you pay, never the X itself, so
// the counter still asks for the X you chose. The counter is Syncopate's
// "unless its controller pays {X}" without the exile: X read off the
// resolving spell, and X of 0 asks for {0}, which is paid for nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f8536726-993b-4100-918c-e3c7177d4c65",
		Name:         "Spectral Denial",
		Completeness: CompletenessFull,
		XMatters:     true,
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(PermanentsYouControl(And(Creature(), PowerGE(4))),
				"This spell costs {1} less to cast for each creature you control with power 4 or greater."),
		},
		Targets: TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			cost := "{" + strconv.Itoa(ctx.X()) + "}"
			return CounterUnlessPaid{
				StackID:  item.Targets[0].ID,
				Cost:     cost,
				Question: "Spectral Denial — pay " + cost + " or your spell is countered",
			}.Apply(ctx)
		},
	})
}
