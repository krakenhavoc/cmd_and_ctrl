package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Winds of Abandon — Sorcery {1}{W} (EDHREC rank 1996):
//
//	"Exile target creature you don't control. For each creature
//	 exiled this way, its controller searches their library for a
//	 basic land card. Those players put those cards onto the
//	 battlefield tapped, then shuffle.
//	 Overload {4}{W}{W} (You may cast this spell for its overload
//	 cost. If you do, change "target" in its text to "each.")"
//
// Path to Exile at sorcery speed, and a one-sided Plague Wind for
// six. The two modes share one body, b18ExileCreaturesThenControllersFetch:
// the set is the single target or every creature you don't control,
// and the predicate is the same one the target clause is built from,
// so "you don't control" is real in both modes.
//
// The search is mandatory (no "may"), one per player rather than one
// per creature — a player who lost three creatures runs one
// three-card search, so no two prompts are ever open over the same
// library. Lands enter tapped through the search's own tapped
// clause.
//
// No simplification.
func init() {
	creatureYouDontControl := effects.And(effects.Creature(), effects.OpponentControls())
	effects.Register(effects.Spec{
		OracleID:     "499a4715-6776-461b-bee4-22ed172d455e",
		Name:         "Winds of Abandon",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		AlternativeCosts: []game.AlternativeCost{
			effects.Overload("{4}{W}{W}"),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			const reason = "Winds of Abandon — a basic land for each creature you lost, onto the battlefield tapped"
			if ctx.PaidAltCost("overload") {
				return effects.B18ExileCreaturesThenControllersFetch(ctx, effects.MatchingBattlefield(ctx, creatureYouDontControl), reason)
			}
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(item.Targets[0].ID)
			if !ok {
				return nil
			}
			return effects.B18ExileCreaturesThenControllersFetch(ctx, []game.Card{c}, reason)
		},
	})
}
