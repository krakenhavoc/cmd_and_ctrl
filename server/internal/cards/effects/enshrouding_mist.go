package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Enshrouding Mist — Instant {W}:
//
//	"Target creature gets +1/+1 until end of turn. Prevent all damage
//	 that would be dealt to it this turn. If it's renowned, untap it."
//
// #2049: the +1/+1, then ADR 0108 §7's not-one-use shield pinned to the
// same creature (Leap of Faith's), then CR 702.112b's designation read
// as the spell resolves — after the pump and the shield, which do not
// change it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e513c05c-14c0-4d2e-a5b3-e4f29cb4c1a2",
		Name:         "Enshrouding Mist",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 1, Label: "Enshrouding Mist — +1/+1"}).Apply(ctx); err != nil {
					return err
				}
				if err := (PreventDamageFromSource{Protect: ShieldObject(t.ID)}).Apply(ctx); err != nil {
					return err
				}
				if ctx.Game.IsRenowned(t.ID) {
					return UntapTarget{Target: t.ID}.Apply(ctx)
				}
				return nil
			}
			return nil
		},
	})
}
