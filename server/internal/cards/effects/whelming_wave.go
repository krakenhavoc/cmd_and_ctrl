package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Whelming Wave — Sorcery {2}{U}{U}:
//
//	"Return all creatures to their owners' hands except for Krakens,
//	 Leviathans, Octopuses, and Serpents."
//
// Evacuation for a sea-monster deck: a one-sided board wipe if your
// whole board is Krakens, and a four-mana symmetric bounce if it
// isn't. The card is only worth its slot in the deck the exclusion
// list was written for, which is exactly what makes it the right
// test case for the exclusion vocabulary.
//
// # "Except for", literally
//
// Except(Creature(), AnySubtype("Kraken", "Leviathan", "Octopus",
// "Serpent")) is the printed sentence, in the same order, with the
// same words. That is the point of the helper: the alternative is
// And(Creature(), Not(Or(Subtype("Kraken"), Subtype("Leviathan"),
// …))), where the reader has to verify a De Morgan transformation
// against the card before believing the list is complete.
//
// The subtypes are EFFECTIVE, so a creature something turned into a
// Kraken is spared. And the exclusion is by TYPE, not by controller
// — an opponent's Leviathan stays on the battlefield too, which is
// the printed text and occasionally a real cost.
func init() {
	Register(Spec{
		OracleID: "e510eaaf-6497-480f-baa8-f4796b5f1086",
		Name:     "Whelming Wave",
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return BounceAllMatching{
				Match: Except(Creature(), AnySubtype("Kraken", "Leviathan", "Octopus", "Serpent")),
			}.Apply(ctx)
		},
	})
}
