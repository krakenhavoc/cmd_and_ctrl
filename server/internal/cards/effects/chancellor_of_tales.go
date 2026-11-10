package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chancellor of Tales — Creature — Faerie Advisor {3}{U}, 3/3 (Wilds of
// Eldraine, #4471):
//
//	"Flying
//	 Whenever you cast an Adventure spell, you may copy it. You may choose
//	 new targets for the copy."
//
// An Adventure spell is the Adventure half cast from a card (its type line
// carries the Adventure subtype on the stack); the creature half of the same
// card is not one. The "you may" is asked as the trigger resolves, and a
// yes is Owlin Spiralmancer's CR 707.10 copy with the new-targets offer
// (CR 707.10c), made from last-known information if the spell has been
// countered (CR 608.2h). The copy is created, not cast, and is not a card,
// so it is not exiled on resolution (CR 715.3d speaks of the card, and a
// copy ceases to exist).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "529f259a-df66-4283-bc28-de3933c42e67",
		Name:            "Chancellor of Tales",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventCast, YouCast(Subtype("Adventure")), "Chancellor of Tales — you may copy the Adventure spell you cast", owlinSpiralmancerCopy),
				"Chancellor of Tales — copy that spell? You may choose new targets for the copy."),
		},
	})
}
