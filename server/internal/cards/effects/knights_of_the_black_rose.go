package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Knights of the Black Rose — Creature — Human Knight {3}{W}{B}, 4/4:
//
//	"When this creature enters, you become the monarch.
//	 Whenever an opponent becomes the monarch, if you were the monarch
//	 as the turn began, that player loses 2 life and you gain 2 life."
//
// A punisher for taking your crown (#1722). "An opponent becomes the
// monarch" is EventMonarchChanged with somebody else as the new holder;
// "if you were the monarch as the turn began" is an intervening "if"
// (CR 603.4) read off TurnTally.MonarchAtStart, the stamp
// resetTurnTallyLocked takes as each turn begins. It is checked when
// the ability triggers and again as it resolves; the second check
// cannot change its answer within a turn, and is there for the rule's
// sake.
//
// "That player" is the one who became the monarch, read off the
// triggering event on the item (not captured), so a restored or undone
// table resolves it against the right seat. A player who has since
// left the game loses nothing; you still gain 2 (CR 608.2 — the effect
// does as much as it can).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9dbfa01b-ce7e-4dd1-9257-0c89b831f477",
		Name:         "Knights of the Black Rose",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouBecomeTheMonarch("Knights of the Black Rose"),
			On(game.EventMonarchChanged, AllOf(AnOpponentBecameTheMonarch, youWereTheMonarchAsTheTurnBegan),
				"Knights of the Black Rose — that player loses 2 life and you gain 2 life", knightsOfTheBlackRoseDrain),
		},
	})
}

// youWereTheMonarchAsTheTurnBegan is the intervening "if".
func youWereTheMonarchAsTheTurnBegan(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return g.MonarchAsTurnBegan() == source.Controller
}

// knightsOfTheBlackRoseDrain re-checks the "if" and drains the new
// monarch for 2.
func knightsOfTheBlackRoseDrain(g *game.Game, item *game.StackItem) error {
	if g.MonarchAsTurnBegan() != item.Controller {
		return nil
	}
	ctx := NewContext(g, item)
	victim := item.Trigger.Event.Actor
	if p := g.PlayerByIDForEffect(victim); p != nil && !p.Eliminated {
		if err := g.ChangePlayerLifeForEffect(item.SourceCardID, victim, -2); err != nil {
			return err
		}
	}
	return GainLife{Player: item.Controller, Amount: 2}.Apply(ctx)
}
