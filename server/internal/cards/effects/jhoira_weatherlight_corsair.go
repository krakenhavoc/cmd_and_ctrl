package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jhoira, Weatherlight Corsair — Legendary Creature — Human Pirate
// {4}{B}{B}, 4/5:
//
//	"Whenever Jhoira enters or attacks, target opponent reveals cards
//	 from the top of their library until they reveal a historic permanent
//	 card. You put that card onto the battlefield under your control and
//	 lose life equal to that permanent's mana value. That player puts the
//	 rest of the revealed cards on the bottom of their library in a
//	 random order. (Artifacts, legendaries, and Sagas are historic.)"
//
// The reveal is the shared reveal-until walk (revealUntil), run against
// the OPPONENT's library; a token in a library is revealed but never
// stops the run. The hit enters under the controller's control through
// the library-entry door, so entry replacements ask their questions, and
// the life loss and the bottoming wait for the entry to finish. A library
// with no historic permanent card is revealed whole and put back on the
// bottom in a random order, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cf9b0bb5-d545-4efd-9255-39b10c2be0ee",
		Name:         "Jhoira, Weatherlight Corsair",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEntersOrAttacks("Jhoira, Weatherlight Corsair — target opponent reveals until a historic permanent card; you put it onto the battlefield", rfCreatureCJhoiraEffect),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}
