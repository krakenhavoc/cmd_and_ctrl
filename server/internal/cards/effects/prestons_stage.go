package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Preston's Stage — Plane — Las Vegas (Unknown Event, #4870):
//
//	"Whenever a player performs a magic trick, they create a Hat token.
//	 (You perform a magic trick when you cast an instant or spell with
//	 flash, or turn a face-down creature face-up. Hat tokens are artifact
//	 tokens with "{1}, {T}, Sacrifice this artifact: Create a 1/1 white
//	 Rabbit creature token.")
//	 Whenever Chaos ensues, until your next turn, creature cards in your
//	 hand have flash."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6). A
// magic trick is anyone's cast of an instant or of a spell with flash (read
// on the spell as it sits on the stack), or anyone's turning of a face-down
// permanent face up (the special action or an effect); the trigger is one
// ability watching both, and the player who performed it creates the Hat. The
// chaos ability stores a flash statement on the planar controller for creature
// spells cast from their hand, lasting until their next turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "804c913e-7dcd-4e83-a8d2-47b26e796c2e",
		Name:         "Preston's Stage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventCast, game.EventTurnedFaceUp}, aMagicTrickWasPerformed,
				"Preston's Stage — that player creates a Hat token",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					who := item.Trigger.Event.Actor
					if who == uuid.Nil {
						if c, ok := g.LookupCardForEffect(item.Trigger.Event.CardID); ok {
							who = c.Controller
						}
					}
					if who == uuid.Nil {
						return nil
					}
					return CreateToken{Controller: who, Template: planesB13HatToken(), N: 1}.Apply(NewContext(g, item))
				}),
			WheneverChaosEnsues("Preston's Stage — until your next turn, creature cards in your hand have flash",
				Do(GrantCastTiming{
					Filter:            game.PermissionFilter{CreatureOnly: true},
					FromZone:          game.ZoneHand,
					UntilYourNextTurn: true,
					Label:             "Preston's Stage — creature cards in your hand have flash until your next turn",
				})),
		},
	})
}

// aMagicTrickWasPerformed is "a player performs a magic trick": a cast of an
// instant or of a spell with flash, or a face-down permanent turned face up.
func aMagicTrickWasPerformed(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	switch ev.Kind {
	case game.EventTurnedFaceUp:
		return true
	case game.EventCast:
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && (c.IsInstant() || game.HasKeyword(&c, "flash"))
	}
	return false
}
