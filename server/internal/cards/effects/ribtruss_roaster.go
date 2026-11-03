package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ribtruss Roaster — Creature — Troll Druid {4}{G}, 3/3 (Secrets of
// Strixhaven Commander, #4834):
//
//	"Devour 1 (As this creature enters, you may sacrifice any number of
//	 creatures. This creature enters with that many +1/+1 counters on it.)
//	 At the beginning of your end step, create a number of 1/1 black and
//	 green Pest creature tokens equal to the number of +1/+1 counters on
//	 this creature. They have "When this token dies, you gain 1 life.""
//
// Devour is the as-enters sacrifice (Devour, CR 702.82a): the counters ride
// the entry, so a creature entering at the same time is never a candidate
// (the ruling, CR 614.13a). The end step trigger is Mycoloth's count of
// +1/+1 counters on the source, read as the source last existed when it has
// left the battlefield by the time the trigger resolves (CR 608.2h, the
// ruling). The Pest is the predefined token (PestToken), whose printed
// "When this token dies, you gain 1 life." is exactly the ability the card
// gives the tokens, so each one is its own token's trigger and its
// controller gains the life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "86c55fe5-98b5-4e53-8828-a7f04ee3e985",
		Name:         "Ribtruss Roaster",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{Devour("Ribtruss Roaster", 1)},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Ribtruss Roaster — create a Pest for each +1/+1 counter on it",
				Do(CreateTokensForEach{
					Template: PestToken(),
					Name:     "the +1/+1 counters on it",
					Count:    socB9PlusOneCountersOnSource,
				})),
		},
	})
}
