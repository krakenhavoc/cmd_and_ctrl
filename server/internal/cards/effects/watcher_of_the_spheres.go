package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Watcher of the Spheres — 2/2 Creature — Bird Wizard for {W}{U}
// (EDHREC rank 4402):
//
//	"Flying
//	 Creature spells with flying you cast cost {1} less to cast.
//	 Whenever another creature you control with flying enters, this
//	 creature gets +1/+1 until end of turn."
//
// A two-mana flier that makes every other flier cheaper — the payoff
// slot in a Bird or Angel deck, where half the curve has the keyword
// already. Roadmap batch 42 (#449) filed it under cost modification;
// #93 shipped that in S28 and the card is now three declarations.
//
// The discount is a CostModifier and deliberately NOT a Static: the
// CR 613 layer engine models continuous effects on a characteristic
// of an OBJECT, and a spell being priced is in no zone the layers
// reach. The engine asks the battlefield for every modifier on every
// cast and prices the spell through them in CR 601.2f order, so two
// Watchers stack and one of them dying takes exactly its own
// reduction away.
//
// Two rules the engine owns rather than this file:
//
//   - The reduction spends against GENERIC mana only and stops at
//     zero (CR 601.2f), so a {2}{W} Angel becomes {1}{W} and a {W}{W}
//     one is untouched.
//   - Mana value is not changed (CR 202.3), so a discounted Angel is
//     still the mana value its card says.
//
// "With flying" is read through game.HasKeyword rather than off a
// type line, because the spell is not on the battlefield and
// HasKeyword is the accessor that falls back to the catalog's printed
// keywords for exactly that case. A creature that would only GAIN
// flying once it is on the battlefield is not a creature spell with
// flying, which is correct.
//
// The pump is "ANOTHER creature you control with flying" — the
// Watcher does not pump itself on arrival — and it is a Layer 7c
// turn-scoped static rather than a counter, so it wears off at
// cleanup.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d43463d2-4395-4616-acc3-c4f80ebe3242",
		Name:            "Watcher of the Spheres",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Creature spells with flying you cast cost {1} less to cast.",
				YourSpell(), b42CreatureSpellWithFlying()),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB,
				func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					if !AnotherCreatureEnteredUnderYourControl(ev, source, lki, g) {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && game.HasKeyword(&c, "flying")
				},
				"Watcher of the Spheres — +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{
						Target: item.SourceCardID,
						Power:  1, Toughness: 1,
						Label: "Watcher of the Spheres — +1/+1 until end of turn",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
