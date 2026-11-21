package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vault 87: Forced Evolution — Enchantment — Saga {3}{G}{U} (Fallout):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Gain control of target non-Mutant creature for as long as you
//	 control this Saga.
//	 II — Put a +1/+1 counter on target creature you control. It becomes a
//	 Mutant in addition to its other types.
//	 III — Draw cards equal to the greatest power among Mutants you control."
//
// I is Extraction Specialist's source-bound duration (CR 611.2b): a Saga
// that has already left, or that I no longer control, never starts the
// theft, and it ends the moment the Saga leaves or changes controller.
// II puts the counter, then adds the Mutant type with no duration (CR
// 611.2a), a layer 4 effect that stays on that object. III reads the
// greatest power among the Mutants I control as it resolves (the creature
// made one in II among them), negative power counted as zero, and draws
// that many.
//
// No simplification.
func init() {
	const name = "Vault 87: Forced Evolution"
	Register(Spec{
		OracleID:     "1eb22223-edd4-42bf-a127-e16191cbd1c1",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, SagaChapterLabel(name, 1, "gain control of target non-Mutant creature for as long as you control this Saga"),
				TargetCreature("target non-Mutant creature", Not(HasSubtype("Mutant"))),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := ctx.ClauseTarget(0)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					d, ok := DurationWhileYouControlSource(ctx, ctx.Source(), ctx.Controller())
					if !ok {
						return nil
					}
					return GainControl{Target: t.ID, Duration: d, Label: name + " — gain control for as long as you control this Saga"}.Apply(ctx)
				}),
			ChapterTriggerTargeting(2, SagaChapterLabel(name, 2, "put a +1/+1 counter on target creature you control; it becomes a Mutant in addition to its other types"),
				TargetCreature("target creature you control", YouControl()),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := ctx.ClauseTarget(0)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
						return err
					}
					return ScopedEffectFor{
						Target:   t.ID,
						Mods:     []game.Mod{game.AddSubtypesMod(ctx.TextSubtypes("Mutant")...)},
						Duration: game.IndefiniteDuration(),
						Label:    name + " — becomes a Mutant in addition to its other types",
					}.Apply(ctx)
				}),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "draw cards equal to the greatest power among Mutants you control"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					best := 0
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsCreature() && c.HasSubtype(ctx.TextSubtype("Mutant")) {
							best = max(best, c.CurrentPower())
						}
					}
					return DrawCards{N: best}.Apply(ctx)
				}),
		},
	})
}
