package effects

// Skirge Familiar — Creature — Phyrexian Imp {4}{B}, 3/2:
//
//	"Flying
//	 Discard a card: Add {B}."
//
// The card on #1213's "discard cost on a non-activated-ability
// surface" row, and the whole of what it needed: a discard component
// on effects.ManaAbilityCost, which is the one cost surface that had
// none. `AbilityCost` grew its own with #660 and a cast's additional
// cost has had one since ADR 0021; this is the same game.DiscardCost,
// with the same options walk, the same validator and the same payer,
// declared on a second owner.
//
// A mana ability with no {T} is a loop, and that is the card: with a
// hand full of expensive spells and something to do with the mana
// (Yawgmoth's Will, a big Tendrils), the Imp turns the hand into
// black mana one card at a time. The engine bounds it only the way it
// bounds every other repeatable activation.
//
// Two properties the shared discard door buys without the mana path
// knowing anything about them:
//
//   - the discard is a DISCARD. EventDiscardCard fires once per card,
//     so a discard payoff (Marauding Mako, Glint-Horn Buccaneer) sees
//     it, and the CR 614 window runs over the exit so madness
//     (CR 702.35a) exiles the card instead of binning it.
//   - it does not pause. CR 605.3b resolves a mana ability
//     immediately, with no stack and no priority window, so a cost
//     discard of a commander takes its owner's hand's place in the
//     graveyard rather than opening the CR 903.9 prompt — the same
//     MustSettleNow rule the spell-cost discard already follows.
//
// The AUTO-TAPPER deliberately never plans it: which card to pitch is
// a decision, and the planner makes none. A hand-clicked Skirge
// Familiar is a mana source; an auto-tapped one is not, exactly as
// Springleaf Drum's tap-a-creature cost is skipped for the same
// reason.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ba95f24d-42da-48ce-bcf1-1b7c4b3c45b5",
		Name:            "Skirge Familiar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{DiscardCards: DiscardACard().DiscardCards},
			Produced: "{B}",
			Label:    "Discard a card: Add {B}",
		}},
	})
}
