package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unnerving Assault — Instant {2}{U/R}:
//
//	"Creatures your opponents control get -1/-0 until end of turn if {U} was
//	 spent to cast this spell, and creatures you control get +1/+0 until end of
//	 turn if {R} was spent to cast this spell. (Do both if {U}{R} was spent.)"
//
// River's Grasp's mana-spent reading: each half asks whether any mana of its
// colour was spent on the cast (ReadsOwnPayment, so a waived payment asks what
// was paid and a copy does neither). WantsColors names both so an auto-paid
// cast reaches for both. Each half pumps the creatures it names as the spell
// resolves (CR 611.2c).
func init() {
	effects.Register(effects.Spec{
		OracleID:       "023f9e20-51fc-493d-a7aa-c9fcbdc7cd04",
		Name:           "Unnerving Assault",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"U", "R"},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if ctx.ManaSpentOfColor("U") > 0 {
				if err := (effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.OpponentControls()), Power: -1,
					Label: "Unnerving Assault — -1/-0 until end of turn"}).Apply(ctx); err != nil {
					return err
				}
			}
			if ctx.ManaSpentOfColor("R") > 0 {
				return effects.CreaturesYouControlGetUntilEOT(1, 0, "Unnerving Assault — +1/+0 until end of turn")(nil, ctx)
			}
			return nil
		},
	})
}
