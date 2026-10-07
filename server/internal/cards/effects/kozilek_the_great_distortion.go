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
// The counter ability is "Discard a card with mana value X" paired
// with a spell target clause bounded by the same X (#2190, ADR 0113's
// 2026-10-07 amendment). X is announced with the activation (CR
// 602.2b) but is not paid for: it is the mana value the discarded
// card must have (DiscardCardWithManaValueX) and the mana value the
// target must have (WithManaValueEqualsX), so the engine refuses a
// card and a target that disagree at announce, and the target is
// re-checked as the ability resolves (CR 608.2b). A spell on the
// stack counts {X} as the value chosen for it (CR 202.3e) and a card
// in hand counts it as zero, so Kozilek answers a Fireball cast for
// X=3 with a card of mana value four. There is no mana cost: the
// ability is free to activate whenever Kozilek is on the battlefield,
// as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4c1c1537-e519-4e2f-9bc2-d34b289d4487",
		Name:            "Kozilek, the Great Distortion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Activated: []ActivatedAbility{{
			Label:   "Discard a card with mana value X: Counter target spell with mana value X.",
			Cost:    DiscardCardWithManaValueX("a card with mana value X"),
			Targets: TargetSpell("target spell with mana value X").WithManaValueEqualsX(),
			Effect:  counterTheChosenSpell,
		}},
		Triggered: []game.TriggeredAbility{{
			FromStack: true,
			Watches:   []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && b14HandSize(g, ev.Actor) < 7
			},
			Key: "Kozilek, the Great Distortion — draw up to seven cards in hand",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Kozilek, the Great Distortion — draw up to seven cards in hand")
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
