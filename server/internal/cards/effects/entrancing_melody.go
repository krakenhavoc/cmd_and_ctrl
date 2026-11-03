package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Entrancing Melody — Sorcery {X}{U}{U} (Secrets of Strixhaven
// Commander, #4834):
//
//	"Gain control of target creature with mana value X."
//
// The target's mana value is bound to the announced X
// (WithManaValueEqualsX), judged at announce (CR 601.2c) and again at
// resolution (CR 608.2b), so a creature whose mana value no longer
// matches is an illegal target and nothing is taken. A permanent with
// {X} in its cost counts that X as 0 (CR 202.3e, the ruling).
//
// The control change has no stated duration, so it lasts indefinitely
// (CR 611.2a, the ruling), the Blatant Thievery shape.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f9f82958-a7c6-4161-b0d8-fdbf7a3dd911",
		Name:         "Entrancing Melody",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature with mana value X").WithManaValueEqualsX(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if err := (GainControl{
					Target:     t.ID,
					Controller: item.Controller,
					Duration:   game.IndefiniteDuration(),
					Label:      "Entrancing Melody — gain control (no stated duration)",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
