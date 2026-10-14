package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Samut, Tyrant of Naktamun — Legendary Creature — Human Wizard 2/1
// {1}{U}:
//
//	"Instant and sorcery spells you control have split second. (As long
//	 as a spell with split second is on the stack, players can't cast
//	 spells or activate abilities that aren't mana abilities.)"
//
// #3502. The grant is Spec.SplitSecondGrants, read LIVE (CR 611.3a): a
// spell has split second while Samut is on the battlefield under its
// controller with her abilities, not because she was there when it was
// cast. A spell Samut's controller casts has it at once, an opponent's
// spells never do, and a Samut who leaves while one is on the stack (a
// mana-ability sacrifice, which split second does not stop) takes it
// off. The stack chip and the table's split-second gate read the same
// answer.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "a275f7ac-7dae-4422-875a-f34f24b6f28a",
		Name:              "Samut, Tyrant of Naktamun",
		Completeness:      CompletenessFull,
		SplitSecondGrants: []game.SplitSecondGrant{InstantAndSorcerySpellsYouControlHaveSplitSecond()},
	})
}
