package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ghostly Prison — Enchantment {2}{W}:
//
//	"Creatures can't attack you unless their controller pays {2} for
//	 each creature they control that's attacking you."
//
// Propaganda's sentence in white, word for word, and the second card
// #1063 was opened for. Everything Propaganda's file says applies
// here; read that one for the reasoning.
//
// The pair is also the proof that two taxes STACK. A table with both
// on one player's board charges {2} twice per attacking creature —
// the pricer asks every tax every permanent of the defending player
// contributes and concatenates all of them, so one attacker costs
// "{2}{2}" and three cost "{2}{2}{2}{2}{2}{2}". Neither enchantment
// knows the other exists, which is the property a per-permanent static
// is for.
func init() {
	Register(Spec{
		OracleID:     "e828b189-0e8f-43b8-b909-4c23e742e028",
		Name:         "Ghostly Prison",
		Completeness: CompletenessFull,
		AttackTaxes: []game.AttackTax{
			AttackTax("{2}", "Creatures can't attack you unless their controller pays {2} for each creature they control that's attacking you."),
		},
	})
}
