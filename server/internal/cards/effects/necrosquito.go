package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Necrosquito — Creature — Phyrexian Insect {3}{B}, 0/0 (Phyrexia: All
// Will Be One):
//
//	"Flying
//	 This creature enters with two oil counters on it.
//	 This creature gets +1/+1 for each oil counter on it.
//	 Whenever another creature or artifact you control is put into a
//	 graveyard from the battlefield, put an oil counter on this creature."
//
// The two counters are an entry replacement (CR 614.1c), so it is never a
// 0/0 on the battlefield. The bonus is read live. The trigger is Vat of
// Rebirth's: one occurrence per permanent, types and controller as it
// last existed (CR 603.10a), tokens included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "41dd7ec7-a238-4723-a244-9a112d4fbfd2",
		Name:            "Necrosquito",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(s82OneOil, 2, "Necrosquito: enters with two oil counters"),
		},
		Static: []game.StaticAbility{
			s82OilPump(1, 1, "This creature gets +1/+1 for each oil counter on it."),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, s82OilOnDeparture, "Necrosquito — put an oil counter on it",
				Do(CounterOnThis{Kind: s82OneOil, N: 1})),
		},
	})
}
