package effects

// Millikin — Artifact Creature — Construct {2}, 0/1 (EDHREC rank
// 1213):
//
//	"{T}, Mill a card: Add {C}. (Activate only as an instant. To mill
//	 a card, put the top card of your library into your graveyard.)"
//
// A mana Myr that feeds the graveyard — the self-mill deck's rock.
//
// "Mill a card" is the real cost component now (ManaAbilityCost.Mill,
// #3845, ADR 0020's 2026-10-15 amendment): validated with the tap
// before anything is paid, so a library of no cards can't pay it
// (CR 701.17b, CR 118.3) and the ability is not offered, and paid last
// (CR 601.2h) through the ordinary mill, so Rest in Peace exiles the
// card instead and a "whenever you mill" payoff sees it. Until #3845
// the cost was built from the two slots a mana ability did have — a
// Condition gating on a non-empty library and a Rider milling after the
// mana landed — which got the observable result but not the order, and
// kept the card off every reader of the cost (the view, the
// enumerator).
//
// A mana ability never uses the stack (CR 605.3b), so the mill and the
// mana are one atomic step. The auto-tapper never plans the row (milling
// a library to pay for a spell is the player's call), and the {T} is on
// a CREATURE, so summoning sickness applies (CR 302.6) — the engine
// enforces that from the type line.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fa4dffda-6f04-4d0b-829d-28a1a5794dee",
		Name:         "Millikin",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, Mill: 1},
			Produced: "{C}",
			Label:    "{T}, Mill a card: Add {C}",
		}},
	})
}
