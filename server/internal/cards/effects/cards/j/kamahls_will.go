package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kamahl's Will — Instant {3}{G} (Commander Legends):
//
//	"Choose one. If you control a commander as you cast this spell, you may
//	 choose both instead.
//	 • Until end of turn, any number of target lands you control become 1/1
//	 Elemental creatures with vigilance, indestructible, and haste. They're
//	 still lands.
//	 • Choose target creature you don't control. Each creature you control
//	 deals damage equal to its power to that creature."
//
// Will of the Mardu's modal frame (choose one, or both with a commander as
// the spell is cast). The first bullet is Nissa's land animation, on any
// number of lands (including none), each a 1/1 Elemental creature with
// vigilance, indestructible and haste until end of turn. The second reads
// each creature the controller controls as it resolves, so the lands
// animated by the first bullet (which resolves first) count, and all the
// damage is dealt at once (CR 120.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a32860dd-e8df-4a3c-bdc2-948dcd07cf9a",
		Name:         "Kamahl's Will",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Until end of turn, any number of target lands you control become 1/1 Elemental creatures with vigilance, indestructible, and haste. They're still lands.",
				effects.TargetPermanent("any number of target lands you control", effects.Land(), effects.YouControl()).WithCount(0, 0)),
			effects.Mode("Choose target creature you don't control. Each creature you control deals damage equal to its power to that creature.",
				effects.TargetCreature("target creature you don't control", effects.OpponentControls())),
		).OrUpToIf(2, effects.YouControlACommander),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if ctx.HasMode(0) {
				for _, t := range effects.OptionTargets(ctx, 0) {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (effects.BecomeCreatureUntilEOT{
						Target:       t.ID,
						Types:        []string{"Creature"},
						Subtypes:     []string{"Elemental"},
						SetPower:     1,
						SetToughness: 1,
						Label:        "Kamahl's Will — becomes a 1/1 Elemental creature",
					}).Apply(ctx.AsGroupMember()); err != nil {
						return err
					}
					if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"vigilance", "indestructible", "haste"},
						Label: "Kamahl's Will — vigilance, indestructible and haste"}).Apply(ctx.AsGroupMember()); err != nil {
						return err
					}
				}
			}
			if ctx.HasMode(1) {
				targets := effects.OptionTargets(ctx, 1)
				if len(targets) == 0 {
					return nil
				}
				victim := targets[0].ID
				ctx.Game.RecomputeLayersIfStaleLocked()
				type hit struct {
					source uuid.UUID
					power  int
				}
				var hits []hit
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if c.Controller == ctx.Controller() && c.IsCreature() && c.CurrentPower() > 0 {
						hits = append(hits, hit{c.InstanceID, c.CurrentPower()})
					}
				}
				return ctx.Game.DamageInstanceForEffect(func() error {
					for _, h := range hits {
						if err := (effects.DealDamage{Source: h.source, Target: victim, Amount: h.power}).Apply(ctx.AsGroupMember()); err != nil {
							return err
						}
					}
					return nil
				})
			}
			return nil
		},
	})
}
