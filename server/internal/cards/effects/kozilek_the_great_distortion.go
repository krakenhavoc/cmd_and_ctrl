package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kozilek, the Great Distortion — Legendary Creature — Eldrazi
// {8}{C}{C}, 12/12 (EDHREC rank 1567):
//
//	"When you cast this spell, if you have fewer than seven cards in
//	 hand, draw cards equal to the difference.
//	 Menace
//	 Discard a card with mana value X: Counter target spell with mana
//	 value X."
//
// The Eldrazi that refills. The cast trigger is a FromStack ability
// (cascade's mechanism): it fires when the spell is announced and
// resolves above it, so a countered Kozilek still draws. "If you have
// fewer than seven cards in hand" is an intervening-if (CR 603.4):
// checked as the spell is cast — Kozilek is on the stack by then, not
// in the hand — and again as the trigger resolves, with the
// difference recomputed then (read live off the item's Controller, so
// a card drawn in response shrinks the draw). The {C}{C} in the cost
// wants colorless mana, which the cost engine enforces.
//
// Sandbox simplification, declared — one whole ability omitted, the
// Stoneforge Mystic posture: the counter ability is NOT implemented.
// "Discard a card with mana value X" is a discard-a-card component in
// an activated ability's cost, and AbilityCost has no such component
// (tap, sacrifice, mana, life, loyalty and crew are the whole set).
// Shipping the ability without its cost would be a free Counterspell
// on a stick — stronger than printed, the #259 direction — so the
// ability is left off entirely, which is weaker, and the caveat says
// so.
func init() {
	Register(Spec{
		OracleID:        "4c1c1537-e519-4e2f-9bc2-d34b289d4487",
		Name:            "Kozilek, the Great Distortion",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"The discard-to-counter ability isn't implemented — Kozilek can't counter spells."},
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{{
			FromStack: true,
			Watches:   []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && b14HandSize(g, ev.Actor) < 7
			},
			Key: "Kozilek, the Great Distortion — draw up to seven cards in hand",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Kozilek, the Great Distortion — draw up to seven cards in hand", nil)
				item.Controller, item.Owner = ev.Actor, ev.Actor
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				n := 7 - b14HandSize(g, item.Controller)
				if n <= 0 {
					return nil
				}
				return DrawCards{Player: item.Controller, N: n}.Apply(NewContext(g, item))
			},
		}},
	})
}
