package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Locust Spray — Instant {B} (Secrets of Strixhaven Mystical Archive,
// #4834):
//
//	"Target creature gets -1/-1 until end of turn.
//	 Cycling {B} ({B}, Discard this card: Draw a card.)"
//
// Tragic Slip's small half with the shared Cycling ability. The target is
// checked again as the spell resolves (CR 608.2b), so a creature that has
// left does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f9dce026-2653-4793-b949-357b08a936ea",
		Name:         "Locust Spray",
		Purpose:      ForTargets(HarmUntilEOT(0, -1, -1)),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return BoostUntilEOT{
				Target:    item.Targets[0].ID,
				Power:     -1,
				Toughness: -1,
				Label:     "Locust Spray — -1/-1",
			}.Apply(ctx)
		},
		Activated: []ActivatedAbility{Cycling("{B}")},
	})
}
