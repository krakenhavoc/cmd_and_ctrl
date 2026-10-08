package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Confiscation Coup — Sorcery {3}{U}{U}:
//
//	"Choose target artifact or creature. You get {E}{E}{E}{E} (four
//	 energy counters), then you may pay an amount of {E} equal to that
//	 permanent's mana value. If you do, gain control of it."
//
// ADR 0129 §3 (#1995): the amount is the target's mana value as the
// spell resolves, paid through the energy prompt (CR 118.12). Paid, the
// caster gains control of it with no end (CR 611.2a's "no duration").
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c246cdb4-2fd8-487e-944e-3e52fbd1bbaa",
		Name:         "Confiscation Coup",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 4},
		Targets:      TargetPermanent("target artifact or creature", Or(Artifact(), Creature())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if err := (GetEnergy{N: 4}).Apply(ctx); err != nil {
				return err
			}
			return MayPayEnergy{
				N:        firstTargetManaValue(ctx.Game, item),
				Question: "Confiscation Coup — pay {E} equal to its mana value to gain control of it?",
				OnPay: func(ctx *Context) error {
					return GainControl{
						Target:   ctx.Item.Targets[0].ID,
						Duration: game.IndefiniteDuration(),
						Label:    "Confiscation Coup — gain control",
					}.Apply(ctx)
				},
			}.Apply(ctx)
		},
	})
}
