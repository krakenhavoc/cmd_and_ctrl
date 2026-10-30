package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Joined Researchers // Secret Rendezvous — Creature — Human Cleric
// Wizard {1}{W}, 2/2 // Sorcery {1}{W}{W} (Secrets of Strixhaven, #4456;
// preparation card, CR 722):
//
//	"First strike
//	 At the beginning of each end step, if an opponent has more cards in
//	 hand than you, this creature becomes prepared. (While it's prepared,
//	 you may cast a copy of its spell. Doing so unprepares it.)"
//
//	Secret Rendezvous — "You and target opponent each draw three cards."
//
// It does not enter prepared. "Each end step" is every player's, so the
// trigger does not read the event's actor (Bloodline Recollector's
// shape). "An opponent has more cards in hand than you" is one opponent
// against you, and it is an intervening if (CR 603.4): asked as the end
// step begins and again as the ability resolves. A creature that is
// already prepared gains nothing (CR 722.3a). The prepare spell is Secret
// Rendezvous word for word, so it shares that card's body
// (secret_rendezvous.go).
//
// No simplification.
const joinedResearchersOracleID = "ec744a2d-4a33-4807-bce0-d95cd5277b1f"

func init() {
	Register(Spec{
		OracleID:        joinedResearchersOracleID,
		Name:            "Joined Researchers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventBeginEndStep && joinedResearchersOpponentHasMore(g, source.Controller)
			}, "Joined Researchers — becomes prepared",
				func(g *game.Game, item *game.StackItem) error {
					if !joinedResearchersOpponentHasMore(g, item.Controller) {
						return nil
					}
					return BecomePrepared{Target: item.SourceCardID}.Apply(NewContext(g, item))
				}),
		},
	})
	Register(Spec{
		OracleID:     joinedResearchersOracleID + "#1",
		Name:         "Secret Rendezvous",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		Purpose:      game.Purpose{Draws: 3, Targets: game.ForTargets(game.TargetPurpose{Slot: 0, Draws: 3})},
		OnResolve:    secretRendezvousOnResolve,
	})
}

// joinedResearchersOpponentHasMore is "if an opponent has more cards in
// hand than you".
func joinedResearchersOpponentHasMore(g *game.Game, you uuid.UUID) bool {
	me := g.PlayerByIDForEffect(you)
	if me == nil || me.Hand == nil {
		return false
	}
	mine := me.Hand.Size()
	return eachOpponent(g, you, func(opp uuid.UUID) bool {
		p := g.PlayerByIDForEffect(opp)
		return p != nil && p.Hand != nil && p.Hand.Size() > mine
	})
}
