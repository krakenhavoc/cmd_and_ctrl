package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tainted Strike — Instant {B}:
//
//	"Target creature gets +1/+0 and gains infect until end of turn."
//
// The infect deck's combat trick: cast after blockers, it turns an
// unblocked creature's damage into poison. Infect granted until end of
// turn is an ordinary layer-6 keyword grant, and the damage tail reads
// it off the creature's effective abilities when the damage is dealt
// (ADR 0056, #748) — so the damage is poison this turn and life loss
// the next.
//
// Two turn-scoped statics, as for Ancestors' Aid: the +1/+0 is layer
// 7c and infect is layer 6.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "95a53dd7-76dc-46f1-8833-fafd02ba49c4",
		Name:         "Tainted Strike",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (BoostUntilEOT{
				Target: target,
				Power:  1,
				Label:  "Tainted Strike — +1/+0 until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{
				Target:   target,
				Keywords: []string{"infect"},
				Label:    "Tainted Strike — infect until end of turn",
			}.Apply(ctx)
		},
	})
}
