package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Anep, Vizier of Hazoret — Legendary Creature — Jackal Warrior {2}{R},
// 4/2:
//
//	"Trample
//	 You may exert Anep as it attacks. When you do, exile the top two
//	 cards of your library. Until the end of your next turn, you may
//	 play those cards. (An exerted creature won't untap during your
//	 next untap step.)"
//
// ADR 0130 §11: exert as it attacks (CR 701.43d) with a linked "when
// you do" (CR 607.2h). The exile and its permission are Reckless
// Impulse's (b19ExileTopTwoUntilEndOfNextTurn): "play", so a land is
// not stranded, and the duration is ADR 0063's "until the end of your
// next turn", which means the same thing from every seat. Exiling
// happens whether or not Anep is still on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0a73289f-0f0b-4316-be21-8edcbb08bd7c",
		Name:            "Anep, Vizier of Hazoret",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		ExertOnAttack:   ExertAsItAttacks(),
		Triggered: []game.TriggeredAbility{
			WhenExerted("Anep, Vizier of Hazoret — exile the top two cards of your library; you may play them until the end of your next turn",
				b19ExileTopTwoUntilEndOfNextTurn),
		},
	})
}
