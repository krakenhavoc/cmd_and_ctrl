package effects

// Cadaverous Bloom — Enchantment {3}{B}{G}:
//
//	"Exile a card from your hand: Add {B}{B} or {G}{G}."
//
// #1283's card, and the whole of what it needed: an EXILE-A-CARD
// component on a battlefield mana ability. #1228 gave ManaAbilityCost
// an exile-THIS-card cost for the Spirit Guides, and that is the wrong
// clause here — it exiles the source, asks nothing, and is refused at
// boot on an ability that functions from the battlefield. This one
// exiles a card the activator PICKS out of their hand, which makes it
// Skirge Familiar's discard one keyword action over:
//
//   - the same shape (a count, a printed label, a predicate), the same
//     candidate walk, validator and payer pattern, the same wire triple
//     under its own names (`exile_cost_n` / `exile_cost_label` /
//     `exile_cost_options`, answered with `exile_ids`) and the same
//     client picker;
//   - a different EXIT. An exiled card is not discarded: no
//     EventDiscardCard, no CR 614 discard window, nothing for madness
//     to see (game.ExileCost). A Marauding Mako beside the Bloom does
//     not grow.
//
// "{B}{B} or {G}{G}" is one activation, one pick, two mana — the
// Gilded Lotus grammar with two options, "{B2|G2}". The pick is the
// ordinary mana-colour prompt.
//
// The AUTO-TAPPER never plans the Bloom: which card leaves the hand is
// a decision, and the planner makes none — the Skirge Familiar
// posture. It is a hand-clicked source, which is how the card is
// played: pitch the spell you were never going to cast, float four,
// cast the one you were.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fbb0f73b-5e30-4632-99c1-e49582e41f8d",
		Name:         "Cadaverous Bloom",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{ExileCards: ExileACardFromHand()},
			Produced: "{B2|G2}",
			Label:    "Exile a card from your hand: Add {B}{B} or {G}{G}",
		}},
	})
}
