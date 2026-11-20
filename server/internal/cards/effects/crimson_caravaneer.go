package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crimson Caravaneer — Creature — Human Scout {2}{R}, 1/2 (Fallout,
// #5948):
//
//	"Double strike, trample
//	 Whenever this creature deals combat damage to a player, create a
//	 Junk token. (It's an artifact with "{T}, Sacrifice this token: Exile
//	 the top card of your library. You may play that card this turn.
//	 Activate only as a sorcery.")"
//
// Double strike deals damage in both combat damage steps (CR 702.4b), so a
// Caravaneer that connects both times makes two Junk, one per damage event
// (CR 603.2c). The token is the printed Junk (JunkToken).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ab3360e5-6a60-4806-bbc2-b7b26c119856",
		Name:            "Crimson Caravaneer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike", "trample"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Crimson Caravaneer — create a Junk token",
				Do(CreateToken{Template: JunkToken(), N: 1})),
		},
	})
}
