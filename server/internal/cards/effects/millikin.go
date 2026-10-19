package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Millikin — Artifact Creature — Construct {2}, 0/1 (EDHREC rank
// 1213):
//
//	"{T}, Mill a card: Add {C}. (Activate only as an instant. To mill
//	 a card, put the top card of your library into your graveyard.)"
//
// A mana Myr that feeds the graveyard — the self-mill deck's rock.
//
// An ordinary activated ability, not a mana ability (#3971): CR 605.1a
// says an ability whose cost or effect moves a card to or from a library
// is not one, and "Mill a card" puts the top card of the library into
// the graveyard. So it uses the stack (CR 602.2), an opponent may
// respond and the ability can be countered, and the mana is added on
// resolution (CR 106.4, via AddMana). CR 605.3a does not apply, so it is
// not available in the middle of paying a cost, and the auto-tapper
// never plans it: milling a library to pay for a spell is the player's
// call.
//
// "Mill a card" is the real cost component (AbilityCost.Mill, #3845,
// ADR 0020's 2026-10-15 amendment): validated with the tap before
// anything is paid, so a library of no cards can't pay it (CR 701.17b,
// CR 118.3) and the ability is not offered, and paid last (CR 601.2h)
// through the ordinary mill, so Rest in Peace exiles the card instead and
// a "whenever you mill" payoff sees it. The {T} is on a CREATURE, so
// summoning sickness applies (CR 302.6) — the engine enforces that from
// the type line.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fa4dffda-6f04-4d0b-829d-28a1a5794dee",
		Name:         "Millikin",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Mill a card: Add {C}",
			Cost:    Plus(TapCost(), MillCardsCost(1)),
			Purpose: game.Purpose{Answers: game.AnswerValue}, // ADR 0142: mana answers nothing
			Effect:  Do(AddMana{Produced: "{C}"}),
		}},
	})
}
