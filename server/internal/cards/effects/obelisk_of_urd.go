package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Obelisk of Urd — Artifact {6} (EDHREC rank 4029):
//
//	"Convoke (Your creatures can help cast this spell. Each creature
//	 you tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 As this artifact enters, choose a creature type.
//	 Creatures you control of the chosen type get +2/+2."
//
// The colourless tribal anthem every deck can cast, and the reason it
// is in the batch is that it puts two separate pieces of machinery on
// one card: the S22 convoke tap-cost and the CR 614.12 "as this
// enters, choose" slot. A six-mana artifact in a token deck is
// routinely a two-mana artifact, and the tribe it names is not decided
// until it resolves.
//
// Three details the card file has to get right:
//
//   - Convoke is a cost SPEND, not a cost demand. The printed cost
//     stays {6}; each creature tapped pays for {1} (or one mana of its
//     own colour, which for a generic-only cost is the same thing).
//     The creatures are tapped while casting, so they are already
//     tapped when the Obelisk resolves — a board that convoked itself
//     out gets the anthem on tapped creatures, which is exactly right.
//
//   - The choice is made AS the Obelisk enters, off the stack, with no
//     priority window (CR 614.12). It is not a trigger, so it cannot
//     be responded to and cannot be Stifled — hence AsEnters rather
//     than Triggered.
//
//   - The anthem has no "other": the Obelisk is not a creature, so
//     there is nothing to exclude, and the filter must not carry
//     Others. Every creature its controller has of the named type gets
//     +2/+2, including ones that arrive later and ones a changeling
//     effect turns into the type afterwards — layer 7c reads the
//     post-layer types on every recompute.
//
// A second Obelisk names its own type independently; two Obelisks
// both naming Goblin stack to +4/+4.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e2ef3a24-9e78-47fc-9192-049aa0ddb7a0",
		Name:         "Obelisk of Urd",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		AsEnters:     ChooseCreatureTypeAsEnters("Obelisk of Urd"),
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Chosen: true, YoursOnly: true}, 2, 2),
		},
	})
}
