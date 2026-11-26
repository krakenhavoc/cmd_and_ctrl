package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gruul Charm — Instant {R}{G} (Gatecrash, S82 OLD #5565):
//
//	"Choose one —
//	 • Creatures without flying can't block this turn.
//	 • Gain control of all permanents you own.
//	 • Gruul Charm deals 3 damage to each creature with flying."
//
// Mode 1 is Falter's restriction on a set read live (faltering), mode 2 is
// Brand's indefinite control change of every permanent the caster owns as the
// charm resolves, and mode 3 is a Squall Line-style damage to each creature
// with flying (read as it resolves) as one damage event.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e552cab2-cee0-4420-9419-4bada4b2f241",
		Name:         "Gruul Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Creatures without flying can't block this turn.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.Faltering("Gruul Charm")(item, ctx)
				}),
			effects.ModeDoing("Gain control of all permanents you own.", nil,
				func(item *game.StackItem, spell *effects.Context, _ int) error {
					ctx := spell.AsGroupMember()
					var owned []uuid.UUID
					for _, c := range ctx.Game.BattlefieldCardsForEffect() {
						if c.Owner == item.Controller {
							owned = append(owned, c.InstanceID)
						}
					}
					for _, id := range owned {
						if err := (effects.GainControl{
							Target:   id,
							Duration: game.IndefiniteDuration(),
							Label:    "Gruul Charm — gain control of all permanents you own",
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
			effects.ModeWithPurpose(effects.ModeDoing("Gruul Charm deals 3 damage to each creature with flying.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return ctx.Game.DamageInstanceForEffect(func() error {
						return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.HasKeyword("flying")), 3)
					})
				}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3, Partial: true}}),
		),
	})
}
