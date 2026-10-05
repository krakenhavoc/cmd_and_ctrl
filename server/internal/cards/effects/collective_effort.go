package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Collective Effort — Sorcery {1}{W}{W}:
//
//	"Escalate—Tap an untapped creature you control. (Pay this cost
//	 for each mode chosen beyond the first.)
//	 Choose one or more —
//	 • Destroy target creature with power 4 or greater.
//	 • Destroy target enchantment.
//	 • Put a +1/+1 counter on each creature target player controls."
//
// Escalate (CR 702.120a, #2126): one untapped creature tapped per mode
// beyond the first, any power, paid with the spell on the stack. It is
// not the {T} symbol, so a creature with summoning sickness may pay it
// (CR 302.6). The third bullet snapshots the target player's creatures
// as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a97e760a-99c8-47ea-a875-451aca913ee7",
		Name:         "Collective Effort",
		Completeness: CompletenessFull,
		Modes: Escalating(ChooseOneOrMore(
			ModeDoing("Destroy target creature with power 4 or greater.",
				TargetCreature("target creature with power 4 or greater", PowerGE(4)),
				DestroyTheModesTarget),
			ModeDoing("Destroy target enchantment.",
				TargetPermanent("target enchantment", Enchantment()),
				DestroyTheModesTarget),
			ModeDoing("Put a +1/+1 counter on each creature target player controls.",
				TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					var ids []game.Card
					for _, c := range ctx.Game.BattlefieldCardsForEffect() {
						if c.Controller == t.ID && c.IsCreature() {
							ids = append(ids, c)
						}
					}
					for _, c := range ids {
						if z := ctx.Game.FindCardZoneForEffect(c.InstanceID); z == nil || z.Kind != game.ZoneBattlefield {
							continue
						}
						if err := (AddCounter{Target: c.InstanceID, Kind: "+1/+1", N: 1}).Apply(ctx.asGroupMember()); err != nil {
							return err
						}
					}
					return nil
				}),
		), EscalateTapCreature()),
	})
}
