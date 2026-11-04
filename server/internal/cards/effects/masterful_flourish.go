package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Masterful Flourish — Instant {B}, Secrets of Strixhaven (#4456):
//
//	"Target creature you control gets +1/+0 and gains indestructible until
//	 end of turn."
//
// Two turn-scoped statics on the one target, as for Tainted Strike: the
// +1/+0 is layer 7c and indestructible is a layer 6 grant. If the creature
// is gone or no longer one you control as the spell resolves, the whole
// spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1cb419cc-d8e4-4676-987f-089cebccc374",
		Name:         "Masterful Flourish",
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 0, "indestructible")),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (BoostUntilEOT{
				Target: target,
				Power:  1,
				Label:  "Masterful Flourish — +1/+0 until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{
				Target:   target,
				Keywords: []string{"indestructible"},
				Label:    "Masterful Flourish — indestructible until end of turn",
			}.Apply(ctx)
		},
	})
}
