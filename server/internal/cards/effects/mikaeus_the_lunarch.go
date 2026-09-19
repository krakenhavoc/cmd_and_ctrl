package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mikaeus, the Lunarch — Legendary Creature — Human Cleric {X}{W},
// 0/0 (EDHREC rank 1959):
//
//	"Mikaeus enters with X +1/+1 counters on it.
//	 {T}: Put a +1/+1 counter on Mikaeus.
//	 {T}, Remove a +1/+1 counter from Mikaeus: Put a +1/+1 counter
//	 on each other creature you control."
//
// The white X-drop that grows itself and then the team. The X body,
// the tap-to-grow and the team pump are live. The pump's cost removes
// a +1/+1 counter from Mikaeus at announce (RemoveCountersFromThis,
// #625), with the tap, so a response cannot see Mikaeus still holding
// a counter he has already spent. Paying with his LAST counter kills
// him, as printed: the engine watched the 0 arrive
// (Card.LostLastCounter), so the toughness check does not mistake the
// 0/0 left behind for the importer's stand-in, and he is put into the
// graveyard (CR 704.5f) with the pump still on the stack, where it
// resolves anyway.
//
// The X counters are the printed CR 614.1c entry clause and ride the
// CR 614 pipeline as one — XCounters, seeded onto the entry event off
// the resolving stack item while the spell is still there (#1002).
// They land on the PERMANENT, after the move and before EventETB, so
// Doubling Season and Hardened Scales apply, the card's own enters
// trigger reads a finished creature, and a "whenever one or more
// counters are put on a permanent you control" payoff sees them —
// which it could not while they went onto a card still on the stack.
//
// Cast for X=0 Mikaeus enters as the printed 0/0 he is and the next
// state-based check puts him into the graveyard (CR 704.5f) before
// the tap can grow him, as in paper. CR 601.2b allows the
// announcement; it simply does not survive it. He used to survive
// here and carried a second caveat saying so — the one way this card
// played STRONGER than printed; #691 took both the gap and the caveat
// away.
func init() {
	Register(Spec{
		OracleID:                   "82f3faa8-39fa-450b-843f-d60a4c36d8f7",
		Name:                       "Mikaeus, the Lunarch",
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		XMatters:                   true,
		Completeness:               CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:  "{T}: Put a +1/+1 counter on Mikaeus.",
				Cost:   TapCost(),
				Effect: putCounterOnSelf,
			},
			{
				Label:  "{T}, Remove a +1/+1 counter from Mikaeus: Put a +1/+1 counter on each other creature you control.",
				Cost:   Plus(TapCost(), RemoveCountersFromThis(game.CounterPlusOne, 1)),
				Effect: putCounterOnEachOtherCreatureYouControl,
			},
		},
	})
}
