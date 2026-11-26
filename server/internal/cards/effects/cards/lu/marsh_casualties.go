package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Marsh Casualties — Sorcery {B}{B} (Duel Decks: Zendikar vs. Eldrazi,
// #5548):
//
//	"Kicker {3}
//	 Creatures target player controls get -1/-1 until end of turn. If this
//	 spell was kicked, those creatures get -2/-2 until end of turn instead."
//
// The creatures are those the player controls as the spell resolves (CR
// 611.2c); a target player who is gone makes it do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "d2bd182c-532f-4c77-adb8-754f2c2fcd82",
		Name:          "Marsh Casualties",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{3}")},
		Targets:       effects.TargetPlayer("target player"),
		Purpose:       game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetPlayer || !ctx.IsTargetLegal(t) {
				return nil
			}
			n := 1
			if ctx.KickedTimes() > 0 {
				n = 2
			}
			return effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.ControlledByPlayer(t.ID)), Power: -n, Toughness: -n,
				Label: "Marsh Casualties — creatures get " + effects.SignedPT(-n, -n),
			}.Apply(ctx)
		},
	})
}
