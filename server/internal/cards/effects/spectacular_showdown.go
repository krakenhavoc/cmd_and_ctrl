package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spectacular Showdown — Sorcery {1}{R}:
//
//	"Put a double strike counter on target creature, then goad each
//	 creature that had a double strike counter put on it this way. (Until
//	 your next turn, those creatures attack each combat if able and attack
//	 a player other than you if able.)
//	 Overload {4}{R}{R}{R} (You may cast this spell for its overload cost.
//	 If you do, change "target" in its text to "each.")"
//
// The double strike counter is a keyword counter (CR 122.1b): the engine
// reads the creature's double strike off it. Only a creature that really
// had a counter put on it is goaded (the counter can be prevented, or the
// target can be gone). Overloaded, the spell names no target and puts a
// counter on EACH creature on the battlefield, every player's and the
// caster's own included, then goads each one that got it: the printed card
// goads your own creatures too, by you, until your next turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "b6f61b77-d571-4508-a5e8-435b4e097f3a",
		Name:             "Spectacular Showdown",
		Completeness:     CompletenessFull,
		Targets:          TargetCreature("target creature"),
		AlternativeCosts: []game.AlternativeCost{Overload("{4}{R}{R}{R}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			g := ctx.Game
			var ids []uuid.UUID
			if ctx.PaidAltCost("overload") {
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.IsCreature() {
						ids = append(ids, c.InstanceID)
					}
				}
			} else {
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						ids = append(ids, t.ID)
					}
				}
			}
			var goaded []uuid.UUID
			for _, id := range ids {
				if !onBattlefield(g, id) {
					continue
				}
				before := counterCountOn(g, id, game.CounterDoubleStrike)
				if err := (AddCounter{Target: id, Kind: game.CounterDoubleStrike, N: 1}).Apply(ctx.asGroupMember()); err != nil {
					return err
				}
				if counterCountOn(g, id, game.CounterDoubleStrike) > before {
					goaded = append(goaded, id)
				}
			}
			if len(goaded) == 0 {
				return nil
			}
			return goadAndScheduleClear(ctx, goaded)
		},
	})
}
