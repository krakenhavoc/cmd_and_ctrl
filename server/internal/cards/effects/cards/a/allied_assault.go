package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Allied Assault — Instant {2}{W} (Zendikar Rising, #5508):
//
//	"Up to two target creatures each get +X/+X until end of turn, where X is
//	 the number of creatures in your party. (Your party consists of up to one
//	 each of Cleric, Rogue, Warrior, and Wizard.)"
//
// X is the party (CR 700.8) counted once, as the spell resolves, and both
// creatures get that amount. A target that is no longer legal is skipped
// (CR 608.2b); with no targets nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "31d8c2c4-3309-4eee-a704-f760f381f080",
		Name:         "Allied Assault",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := effects.OldB096PartyOf(ctx.Game, ctx.Controller())
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: x, Toughness: x, Label: "Allied Assault — +X/+X"}).Apply(ctx.AsGroupMember()); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
