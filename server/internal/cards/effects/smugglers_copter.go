package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Smuggler's Copter — Artifact — Vehicle, 3/3, for {2}:
//
//	"Flying
//	 Whenever this Vehicle attacks or blocks, you may draw a card.
//	 If you do, discard a card.
//	 Crew 1"
//
// One printed ability with two trigger conditions is ONE
// TriggeredAbility watching two kinds (Savvy Hunter's shape, #860):
// the Copter's own attack declaration (EventAttack) or its own block
// declaration (EventBlock). Both are announced at their lock-in —
// EventAttack once per attacker (#859, CR 508.1), EventBlock once per
// (blocker, attacker) pair of the final declaration (#830, CR 509.1)
// — so a Copter re-pointed at another attacker mid-step still loots
// exactly once. The "you may" is Optional (CR 603.5): a yes/no before
// the ability reaches the stack, and a no drops it.
//
// The body is lootOne, the shared draw-then-discard primitive (#797):
// the draw resolves first and the discard is a prompt built from the
// post-draw hand, so the card just drawn is a legal pitch, as in
// paper.
//
// The flying is real and does the work a Vehicle's flying does: the
// Copter is not a creature until it is crewed, so the keyword sits
// inert on an artifact until a crew ability resolves, at which point
// Layer 4 makes it a creature and Layer 6 has already put "flying"
// on its ability list.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "49136bdc-bc50-49a2-999a-1ef9c16ea130",
		Name:            "Smuggler's Copter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:  "Crew 1",
			Cost:   CrewCost(1),
			Effect: CrewEffect("Smuggler's Copter"),
		}},
		Triggered: []game.TriggeredAbility{
			Optional(OnAny([]game.EventKind{game.EventAttack, game.EventBlock},
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclared(ev, source) || b28SelfBlocked(ev, source)
				},
				"Smuggler's Copter — loot",
				func(g *game.Game, item *game.StackItem) error {
					return lootOne(g, item, 1)
				},
			), "Smuggler's Copter — draw a card, then discard a card?"),
		},
	})
}
