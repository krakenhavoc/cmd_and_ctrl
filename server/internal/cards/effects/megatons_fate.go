package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Megaton's Fate — Sorcery {5}{R}:
//
//	"Choose one —
//	 • Disarm — Destroy target artifact. Create four Treasure tokens.
//	 • Detonate — Megaton's Fate deals 8 damage to each creature. Each
//	   player gets four rad counters."
//
// #2042. Disarm's Treasures are the caster's, and they are made only if
// the spell resolves (an artifact gone in response leaves it with no
// legal target, CR 608.2b). Detonate's damage is one simultaneous
// instance; every player still in the game, the caster included, gets
// the counters.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "09a688c1-21f6-4c7d-a6bd-aedc0d76a671",
		Name:         "Megaton's Fate",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Disarm — Destroy target artifact. Create four Treasure tokens.",
				TargetPermanent("target artifact", Artifact()),
				func(item *game.StackItem, ctx *Context, occ int) error {
					if err := DestroyTheModesTarget(item, ctx, occ); err != nil {
						return err
					}
					return CreateToken{Controller: ctx.Controller(), Template: TreasureToken(), N: 4}.Apply(ctx)
				}),
			ModeWithPurpose(ModeDoing("Detonate — Megaton's Fate deals 8 damage to each creature. Each player gets four rad counters.",
				nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					if err := damageEachMatching(ctx, Creature(), 8); err != nil {
						return err
					}
					return eachPlayerGetsRadCounters(ctx.Game, ctx.Controller(), 4)
				}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 8}}),
		),
	})
}
