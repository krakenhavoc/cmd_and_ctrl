package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wicked Reward — Instant {1}{B} (Multiverse Gift Box, #5632):
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Target creature gets +4/+2 until end of turn."
//
// Bone Splinters' additional cost with Giant Growth's pump. The sacrifice
// is paid while casting (CR 601.2h), after the target is chosen, so a
// creature sacrificed to pay for it is gone before the spell resolves and
// can't usefully be its target. With no creature to sacrifice the cast is
// refused.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "7aad0a80-cf7b-4224-8e72-d611e0be3383",
		Name:           "Wicked Reward",
		Completeness:   CompletenessFull,
		Purpose:        ForTargets(BenefitUntilEOT(0, 4, 2)),
		AdditionalCost: SacrificeCost("a creature", Creature()),
		Targets:        TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return BoostUntilEOT{
				Target:    item.Targets[0].ID,
				Power:     4,
				Toughness: 2,
				Label:     "Wicked Reward — +4/+2",
			}.Apply(ctx)
		},
	})
}
