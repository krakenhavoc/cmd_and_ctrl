package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jace's Sanctum — Enchantment {3}{U} (EDHREC rank 4185):
//
//	"Instant and sorcery spells you cast cost {1} less to cast.
//	 Whenever you cast an instant or sorcery spell, scry 1."
//
// The spellslinger's four-mana enchantment: a Baral discount that
// cannot be killed by a creature removal spell, plus a scry on every
// spell that smooths the next draw. It is in this batch because both
// halves are now ordinary declarations — the cost half needed the
// S28 cost-modification pipeline (#93), which is one of the five
// mechanics the 2026-09-18 re-triage freed.
//
// The discount is a CostModifiers entry rather than a Static for the
// reason spec.go gives at length: a cost reduction changes what
// someone PAYS for an object that is not on the battlefield and has
// no characteristic, so there is no CR 613 layer for it to sit in.
// The engine asks the battlefield for every modifier at cast time and
// prices the spell through them in CR 601.2f order, which means two
// Sanctums stack and mana value is untouched (CR 202.3) — a
// Counterspell under two Sanctums still has mana value 2 for
// Chalice of the Void.
//
// "Spells YOU cast" is an explicit predicate (YourSpell), not a
// default: a CostModifier with no who-cast-it clause prices EVERY
// player's spell, which is Sphere of Resistance's shape and not this
// card's.
//
// Generic-only reduction: {1} less never eats a coloured pip, so a
// Counterspell still costs {U}{U} and a Ponder still costs {U}. That
// is CR 601.2f and the pipeline enforces it; nothing here has to.
//
// The scry is a real CR 603 trigger on CAST, so it goes on the stack
// ABOVE the spell and resolves first — you scry before the spell
// resolves, and you scry even if the spell is countered. That is the
// printed card and it is what makes the trigger worth having in a
// counterspell deck.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "22d7048e-d538-4fd3-9fb3-2e99e74875fe",
		Name:         "Jace's Sanctum",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Jace's Sanctum: your instants and sorceries cost {1} less", YourSpell(), InstantOrSorcerySpell()),
		},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Instant(), Sorcery()), "Jace's Sanctum — scry 1", Do(Scry{N: 1})),
		},
	})
}
