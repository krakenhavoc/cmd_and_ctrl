package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Celestial Convergence — Enchantment {2}{W}{W} (Prophecy, S82 OLD batch 675,
// #5617):
//
//	"This enchantment enters with seven omen counters on it.
//	 At the beginning of your upkeep, remove an omen counter from this
//	 enchantment. If there are no omen counters on this enchantment, the player
//	 with the highest life total wins the game. If two or more players are tied
//	 for highest life total, the game is a draw."
//
// Midnight Oil's enters-with-counters replacement, and an upkeep trigger that
// removes one counter and then looks at what is left: the same permanent,
// still on the battlefield (a copy that left and came back is a new object,
// CR 400.7), with none left. The highest life total among the players still in
// the game wins (an effect win, ADR 0057, so a "can't win" gate stops it); two
// or more tied for the highest make the game a draw (CR 104.4c).
//
// No simplification.
func init() {
	const omen = "omen"
	effects.Register(effects.Spec{
		OracleID:     "92f0677f-64d8-4c83-abad-ffa2a37f88c1",
		Name:         "Celestial Convergence",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(omen, 7, "Celestial Convergence — enters with seven omen counters"),
		},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Celestial Convergence — remove an omen counter; if there are none, the player with the highest life total wins the game",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					g := ctx.Game
					if !effects.OnBattlefield(g, ctx.Source()) || effects.SourceIsNewObject(g, ctx.Item) {
						return nil
					}
					c, ok := g.LookupCardForEffect(ctx.Source())
					if !ok {
						return nil
					}
					if c.Counters[omen] > 0 {
						if err := (effects.AddCounter{Target: ctx.Source(), Kind: omen, N: -1}).Apply(ctx); err != nil {
							return err
						}
					}
					// "If there are no omen counters": read after the removal.
					if c, ok = g.LookupCardForEffect(ctx.Source()); !ok || c.Counters[omen] > 0 {
						return nil
					}
					return effects.HighestLifeTotalWins{}.Apply(ctx)
				}))),
		},
	})
}
