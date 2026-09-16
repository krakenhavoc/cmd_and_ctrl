package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Heart of Kiran — Legendary Artifact — Vehicle, 4/4, for {2}:
//
//	"Flying, vigilance
//	 Crew 3
//	 You may remove a loyalty counter from a planeswalker you control
//	 rather than pay Heart of Kiran's crew cost."
//
// Two crew abilities, and that is the whole of the "rather than pay"
// (#625). An alternative cost on an ACTIVATED ability has no slot on
// game.AbilityCost — the S22 AlternativeCost machinery is a cast-time
// clause (CR 118.9) that only spells reach — so the alternative is a
// second ability entry with the same effect and its own real cost:
//
//	Crew 3                                  CrewCost(3)
//	Crew — remove a loyalty counter …       RemoveCountersFrom(loyalty, 1, …)
//
// The client already lists abilities separately, so the table sees two
// menu rows and no "choose how to pay" prompt. The reason this was NOT
// done before is the reason it is right now: a second entry with no
// real cost would be activatable with no planeswalker in play, which is
// stronger than printed (#259). The counter is removed at announce from
// a planeswalker the activator controls that actually holds one, and
// the entry cannot be activated otherwise.
//
// Three rules the engine enforces rather than this file:
//
//   - Crew is instant speed, so the counter form is too. It is not a
//     loyalty ability: neither the sorcery-speed window nor the
//     walker's once-per-turn loyalty activation (CR 606.3) is touched,
//     and the walker can still activate its own loyalty ability this
//     turn.
//   - Removing the counter is paying a cost, not an effect, so nothing
//     that modifies counters applies to it.
//   - A walker paid down to 0 loyalty goes to the graveyard through the
//     CR 704.5i state-based action, after the crew ability is on the
//     stack.
func init() {
	Register(Spec{
		OracleID:        "e2ee410f-2467-4f1f-84a0-8a79faedc0b3",
		Name:            "Heart of Kiran",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Activated: []ActivatedAbility{
			{
				Label:  "Crew 3",
				Cost:   CrewCost(3),
				Effect: CrewEffect("Heart of Kiran"),
			},
			{
				Label:  "Crew — remove a loyalty counter from a planeswalker you control",
				Cost:   RemoveCountersFrom(game.CounterLoyalty, 1, "a planeswalker you control", Planeswalker()),
				Effect: CrewEffect("Heart of Kiran"),
			},
		},
	})
}
