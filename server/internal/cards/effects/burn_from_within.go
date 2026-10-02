package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Burn from Within — Sorcery {X}{R}:
//
//	"Burn from Within deals X damage to any target. If a creature is
//	 dealt damage this way, it loses indestructible until end of turn. If
//	 that creature would die this turn, exile it instead."
//
// "That creature" is the creature dealt damage this way, so both riders
// come from the damage's continuation (ADR 0108 §1 decision 3): a
// creature whose damage was all prevented keeps indestructible and is
// not marked. The creature already has the damage marked when it loses
// indestructible, which is the point of the card: the state-based check
// after resolution destroys it (CR 702.12b, 704.5g), and it is exiled.
// A player or a noncreature permanent is only dealt the damage.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "8a1ccbdb-3d89-42fb-a731-6db4241acf24",
		Name:         "Burn from Within",
		XMatters:     true,
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				return DealDamageThen(ctx, t.ID, ctx.X(), AllThen(
					g2LosesIndestructibleIfDealtDamage(item),
					ExileIfDealtDamageWouldDie(item, false),
				))
			}
			return nil
		},
	})
}
