package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oracle's Restoration — Sorcery {G} (Secrets of Strixhaven, #4456):
//
//	"Target creature you control gets +1/+1 until end of turn. You draw a
//	 card and gain 1 life."
//
// The pump is for the target and needs it to be legal; the draw and the life
// are the caster's and happen once the spell resolves. A spell whose only
// target is illegal on resolution does nothing at all (CR 608.2b), so with
// the creature gone neither the draw nor the life happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "00e75a6b-699e-4824-86e5-66d45245cd9b",
		Name:         "Oracle's Restoration",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 1)),
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (BoostUntilEOT{
					Target:    t.ID,
					Power:     1,
					Toughness: 1,
					Label:     "Oracle's Restoration — +1/+1",
				}).Apply(ctx); err != nil {
					return err
				}
				if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
					return err
				}
				return GainLife{Player: item.Controller, Amount: 1}.Apply(ctx)
			}
			return nil
		},
	})
}
