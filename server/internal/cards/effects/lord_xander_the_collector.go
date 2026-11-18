package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lord Xander, the Collector — Legendary Creature — Vampire Demon Noble
// {4}{U}{B}{R}, 6/6 (Streets of New Capenna, #5491):
//
//	"When Lord Xander enters, target opponent discards half the cards in their
//	 hand, rounded down.
//	 Whenever Lord Xander attacks, defending player mills half their library,
//	 rounded down.
//	 When Lord Xander dies, target opponent sacrifices half the nonland
//	 permanents they control of their choice, rounded down."
//
// Three printed abilities, three triggers. Every half is worked out as the
// ability resolves, not as it triggered:
//
//   - the discard is the opponent's own pick of half their hand, rounded down
//     (DiscardPrompt.HalfHand, Pox Plague's), so a one-card hand discards none;
//   - the mill reads the defending player off the attack declaration (CR 506.4,
//     not a target) and mills half their library, rounded down;
//   - the sacrifice counts the nonland permanents the opponent controls then and
//     they choose which (CR 701.21a); the chosen permanents leave together.
//
// A target that has left the game or become illegal makes its ability do
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	enters := WhenThisEnters("Lord Xander — target opponent discards half the cards in their hand, rounded down",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			opp := TargetedPlayer(ctx)
			if opp == uuid.Nil || g.PlayerByIDForEffect(opp) == nil {
				return nil
			}
			g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
				Player:   opp,
				Source:   item.SourceCardID,
				HalfHand: true,
				Question: "Lord Xander — discard half the cards in your hand, rounded down",
			})
			return nil
		})
	enters.Targets = TargetPlayer("target opponent", Opponent())

	const attackLabel = "Lord Xander — defending player mills half their library, rounded down"
	attacks := game.TriggeredAbility{
		Watches: []game.EventKind{game.EventAttack},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return attackDeclared(ev, source)
		},
		Key: attackLabel,
		// "Defending player" is a fact about the attack declaration, read once
		// as the trigger is built (CR 506.4).
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, attackLabel)
			item.Params.Player = b17DefendingPlayer(g, ev)
			return item
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			defender := item.Params.Player
			if defender == uuid.Nil || g.PlayerByIDForEffect(defender) == nil {
				return nil
			}
			return b22MillHalf(NewContext(g, item), defender)
		},
	}

	dies := WhenThisDies("Lord Xander — target opponent sacrifices half the nonland permanents they control of their choice, rounded down",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			victim := TargetedPlayer(ctx)
			if victim == uuid.Nil {
				return nil
			}
			if p := g.PlayerByIDForEffect(victim); p == nil || p.Eliminated {
				return nil
			}
			owed := len(permanentsControlledByMatching(g, victim, Nonland())) / 2
			if owed == 0 {
				return nil
			}
			asks := make([]uuid.UUID, owed)
			for i := range asks {
				asks[i] = victim
			}
			g.PlayersSacrificeForEffect(item.SourceCardID, asks,
				sacrificeSpec("a nonland permanent", Nonland()),
				"Lord Xander — sacrifice half the nonland permanents you control")
			return nil
		})
	dies.Targets = TargetPlayer("target opponent", Opponent())

	Register(Spec{
		OracleID:     "d17bfa94-5046-46ae-80af-be5d1d35af13",
		Name:         "Lord Xander, the Collector",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters, attacks, dies},
	})
}
