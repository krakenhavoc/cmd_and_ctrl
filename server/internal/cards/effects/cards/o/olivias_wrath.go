package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Olivia's Wrath — Sorcery {4}{B} (Commander, #5368):
//
//	"Each non-Vampire creature gets -X/-X until end of turn, where X is
//	 the number of Vampires you control."
//
// X is counted once, as the spell resolves (CR 608.2h), through the
// layers: a changeling counts as a Vampire (it has every creature type,
// CR 702.73a) and so is both counted and spared. The creatures that get
// -X/-X are those that are not Vampires at that moment, fixed as the
// effect begins (CR 611.2c); the shrink is a layer 7c effect until end
// of turn, so it dies to state-based actions where toughness reaches 0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9aebd891-c8c5-4309-a07b-6cab1b204172",
		Name:         "Olivia's Wrath",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, AmountIsX: true, Partial: true}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			x := 0
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == item.Controller && c.IsCreature() && c.HasSubtypeWord("Vampire") {
					x++
				}
			}
			if x == 0 {
				return nil
			}
			return effects.BoostUntilEOT{
				Match:     effects.And(effects.Creature(), effects.Not(effects.HasSubtype("Vampire"))),
				Power:     -x,
				Toughness: -x,
				Label:     "Olivia's Wrath — -X/-X",
			}.Apply(ctx)
		},
	})
}
