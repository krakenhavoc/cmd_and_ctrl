package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chaos Lord — Creature — Human {4}{R}{R}{R}, 7/7 (Ice Age, S82 OLD batch 851,
// #5636):
//
//	"First strike
//	 At the beginning of your upkeep, target opponent gains control of this
//	 creature if the number of permanents is even.
//	 This creature can attack as though it had haste unless it entered this
//	 turn."
//
// The upkeep trigger targets an opponent when it goes on the stack and counts
// the battlefield as it resolves (the "if" is part of the effect, not an
// intervening one, so it is not rechecked when the trigger would be put on the
// stack): an even count hands the Lord over, indefinitely (CR 611.2a), as
// Bucknard's Everfull Purse hands over its artifact. The haste clause is
// Frenzied Saddlebrute's attack-only waiver (the creature does not gain haste, so
// its {T} abilities stay sick), written on itself and withheld in the turn it
// entered, a thing control changes do not count as (Game.EnteredThisTurn): a Lord
// stolen this turn can attack, one that was cast this turn cannot.
//
// No simplification.
func init() {
	trigger := effects.AtYourUpkeep("Chaos Lord — target opponent gains control of this creature if the number of permanents is even",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			opp := effects.FirstLegalTargetPlayer(ctx)
			if opp == uuid.Nil {
				return nil
			}
			n := len(effects.MatchingBattlefield(ctx, func(*game.Game, uuid.UUID, game.Card) bool { return true }))
			if n%2 != 0 {
				return nil
			}
			return effects.GainControl{
				Target:     item.SourceCardID,
				Controller: opp,
				Duration:   game.IndefiniteDuration(),
				Label:      "Chaos Lord — target opponent gains control of this creature",
			}.Apply(ctx)
		})
	trigger.Targets = effects.TargetPlayer("target opponent", effects.Opponent())
	effects.Register(effects.Spec{
		OracleID:        "16d0aea2-4ca4-4a9d-8674-2863e085f514",
		Name:            "Chaos Lord",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered:       []game.TriggeredAbility{trigger},
		Static:          []game.StaticAbility{effects.CanAttackAsThoughHasteUnlessItEnteredThisTurn()},
	})
}
