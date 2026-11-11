package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Florian, Voldaren Scion — Legendary Creature — Vampire Noble {1}{B}{R},
// 3/3 (Commander, #5368):
//
//	"First strike
//	 At the beginning of each of your postcombat main phases, look at the
//	 top X cards of your library, where X is the total amount of life your
//	 opponents lost this turn. Exile one of those cards and put the rest
//	 on the bottom of your library in a random order. You may play the
//	 exiled card this turn."
//
// X is the engine's per-turn life-lost tally summed over your opponents
// (opponentLifeLostThisTurn, Rakdos, Lord of Riots'), read as the trigger
// resolves, so life lost in an extra combat counts. The look is private;
// the player picks the card to exile (cmdr2B09FlorianLook), which is
// playable this turn only, a land with a land drop. With X of zero
// nothing happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d74a2cbf-4d13-401d-9b3f-bfbde4942afd",
		Name:            "Florian, Voldaren Scion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			AtYourPostcombatMain("Florian, Voldaren Scion — look at the top X cards, exile one, rest on the bottom", cmdr2B09FlorianLook),
		},
	})
}
