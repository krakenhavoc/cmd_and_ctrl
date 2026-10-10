package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hanweir, the Writhing Township — Legendary Creature — Eldrazi Ooze,
// 7/4. The combined back face of Hanweir Garrison and Hanweir
// Battlements (CR 712.5b); never a card in a deck, only the melded
// permanent the two become (ADR 0145, #2699):
//
//	"Trample, haste
//	 Whenever Hanweir attacks, create two 3/2 colorless Eldrazi Horror
//	 creature tokens that are tapped and attacking."
//
// Trample and haste are the back face's printed keywords, which the
// melded permanent carries from the deck import. The attack trigger is
// the Garrison's, with bigger tokens
// (whenThisAttacksTokensTappedAndAttacking, meld.go). Its mana value is
// 3, the Garrison's {2}{R} plus the Battlements' none (CR 712.8g).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f4905c40-003e-4992-b8d7-3f07ba09c686",
		Name:         "Hanweir, the Writhing Township",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			whenThisAttacksTokensTappedAndAttacking("Hanweir, the Writhing Township — two tapped and attacking Eldrazi Horrors", "3/2 colorless Eldrazi Horror", 2),
		},
	})
}
