package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vigor — Creature — Elemental Incarnation {3}{G}{G}{G}, 6/6:
//
//	"Trample
//	 If damage would be dealt to another creature you control, prevent
//	 that damage. Put a +1/+1 counter on that creature for each 1 damage
//	 prevented this way.
//	 When Vigor is put into a graveyard from anywhere, shuffle it into
//	 its owner's library."
//
// The prevention is an ADR 0108 §8 static (#1906), once per creature
// the damage would have been dealt to (Test of Faith's counters body).
// Vigor itself is not "another" creature, so damage to it is dealt.
// Damage that can't be prevented is dealt and puts no counter on (CR
// 615.12, "for each 1 damage prevented this way").
//
// The last line is the graveyard trigger (#2038,
// WhenThisIsPutIntoAGraveyardFromAnywhereShuffleItIn).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "70787d8b-0a40-43ea-837a-9617dc13e7e5",
		Name:            "Vigor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements: []game.ReplacementEffect{
			PreventDamageDealtTo(PreventionStatic{
				To:    ToAnotherCreatureYouControl,
				Then:  countersOnItPerPreventedBody,
				Label: "Vigor — prevent damage to another creature you control and put that many +1/+1 counters on it",
			}),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisIsPutIntoAGraveyardFromAnywhereShuffleItIn("Vigor — shuffle it into its owner's library"),
		},
	})
}
