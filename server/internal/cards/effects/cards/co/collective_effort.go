package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "a97e760a-99c8-47ea-a875-451aca913ee7",
		Name:         "Collective Effort",
		Completeness: effects.CompletenessFull,
		Modes: effects.Escalating(effects.ChooseOneOrMore(
			effects.ModeDoing("Destroy target creature with power 4 or greater.",
				effects.TargetCreature("target creature with power 4 or greater", effects.PowerGE(4)),
				effects.DestroyTheModesTarget),
			effects.ModeDoing("Destroy target enchantment.",
				effects.TargetPermanent("target enchantment", effects.Enchantment()),
				effects.DestroyTheModesTarget),
			effects.ModeDoing("Put a +1/+1 counter on each creature target player controls.",
				effects.TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
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
						if err := (effects.AddCounter{Target: c.InstanceID, Kind: "+1/+1", N: 1}).Apply(ctx.AsGroupMember()); err != nil {
							return err
						}
					}
					return nil
				}),
		), effects.EscalateTapCreature()),
	})
}
