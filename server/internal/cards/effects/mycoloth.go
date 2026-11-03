package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mycoloth — Creature — Fungus {3}{G}{G}, 4/4 (Secrets of Strixhaven
// Commander, #4834):
//
//	"Devour 2 (As this creature enters, you may sacrifice any number of
//	 creatures. It enters with twice that many +1/+1 counters on it.)
//	 At the beginning of your upkeep, create a 1/1 green Saproling
//	 creature token for each +1/+1 counter on this creature."
//
// Devour is devour.go's as-enters sacrifice (CR 702.82a). The upkeep
// count is the +1/+1 counters on Mycoloth as the trigger resolves,
// wherever they came from (the ruling), or as it last existed if it has
// left (CR 608.2h), Royal Talon Fighter Jet's read.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d7fd16ce-282d-49cd-b2b4-0d25935e7a72",
		Name:         "Mycoloth",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{Devour("Mycoloth", 2)},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Mycoloth — create a Saproling for each +1/+1 counter on it",
				Do(CreateTokensForEach{
					Template: b11GreenSaprolingToken(),
					Name:     "the +1/+1 counters on it",
					Count:    socB9PlusOneCountersOnSource,
				})),
		},
	})
}
