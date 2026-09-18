package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Annoyed Altisaur — 6/5 Creature — Dinosaur for {5}{G}{G} (EDHREC
// rank 3967):
//
//	"Reach, trample
//	 Cascade (When you cast this spell, exile cards from the top of
//	 your library until you exile a nonland card that costs less. You
//	 may cast it without paying its mana cost. Put the exiled cards on
//	 the bottom in a random order.)"
//
// The biggest cascade hit in the format: seven mana means the trigger
// digs for anything with mana value six or less, which in a green
// deck is most of the deck. It is in the batch as the high end of the
// cascade range — Bloodbraid Elf caps at three — and every card that
// exercises a keyword at a different number is a free check that the
// number is read off the spell rather than hard-coded.
//
// Important: the "costs less" comparison is against the ALTISAUR's
// mana value, seven, even though the Altisaur itself is green and the
// hit need not be. Cascade is colour-blind.
//
// No simplification: reach and trample are printed keywords, and
// cascade is the engine keyword whole — the exile loop, the free
// cast offer, and the random bottoming of the rest.
func init() {
	Register(Spec{
		OracleID:        "a8135179-51ef-454e-98ff-69137440339f",
		Name:            "Annoyed Altisaur",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
		Triggered:       []game.TriggeredAbility{Cascade()},
	})
}
