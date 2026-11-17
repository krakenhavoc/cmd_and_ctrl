package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mindleech Ghoul — Creature — Zombie {1}{B}, 2/2:
//
//	"Exploit (When this creature enters, you may sacrifice a creature.)
//	 When this creature exploits a creature, each opponent exiles a card
//	 from their hand."
//
// Exploit is the shared constructor (exploit.go); the payoff is an
// untargeted reflexive trigger. Each opponent picks a card from their own
// hand (a floor of one, so it is not optional) and it is exiled face up;
// an opponent with an empty hand exiles nothing (CR 609.3). Nobody may
// play the exiled cards.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e0f47fe2-9e97-400e-98c1-a80ca8be66f3",
		Name:         "Mindleech Ghoul",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Exploit("Mindleech Ghoul", mindleechGhoulExileBody),
		},
	})
}

// mindleechGhoulExileBody is "When this creature exploits a creature,
// each opponent exiles a card from their hand." An on-disk key: never
// renamed, never reused.
var mindleechGhoulExileBody = game.ReflexiveBody("mindleech-ghoul/exile", simpleBody(mindleechGhoulExile), nil)

// mindleechGhoulExile asks each opponent, in turn order, for the card
// they exile. The answer is the shared plain exile (stdB40ExileFromHand).
func mindleechGhoulExile(g *game.Game, item *game.StackItem) error {
	eachOpponent(g, item.Controller, func(opp uuid.UUID) bool {
		p := g.PlayerByIDForEffect(opp)
		if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
			return false
		}
		cards := make([]uuid.UUID, 0, p.Hand.Size())
		for _, c := range p.Hand.Cards {
			cards = append(cards, c.InstanceID)
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    opp,
			FromPlayer: opp,
			Source:     item.SourceCardID,
			Question:   "Mindleech Ghoul — exile a card from your hand",
			Cards:      cards,
			Min:        1,
			Max:        1,
			ThenKey:    stdB40ExileFromHand,
			LiveZone:   game.ZoneHand,
		})
		return false
	})
	return nil
}
