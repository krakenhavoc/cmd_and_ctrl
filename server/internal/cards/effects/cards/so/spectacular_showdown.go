package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:         "b6f61b77-d571-4508-a5e8-435b4e097f3a",
		Name:             "Spectacular Showdown",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetCreature("target creature"),
		AlternativeCosts: []game.AlternativeCost{effects.Overload("{4}{R}{R}{R}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
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
				if !effects.OnBattlefield(g, id) {
					continue
				}
				before := effects.CounterCountOn(g, id, game.CounterDoubleStrike)
				if err := (effects.AddCounter{Target: id, Kind: game.CounterDoubleStrike, N: 1}).Apply(ctx.AsGroupMember()); err != nil {
					return err
				}
				if effects.CounterCountOn(g, id, game.CounterDoubleStrike) > before {
					goaded = append(goaded, id)
				}
			}
			if len(goaded) == 0 {
				return nil
			}
			return effects.GoadAndScheduleClear(ctx, goaded)
		},
	})
}
