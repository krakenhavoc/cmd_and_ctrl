package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vernal Sovereign — Creature — Elemental Elk {4}{G}{W}, 4/4 (Lorwyn
// Eclipsed Commander, #5368):
//
//	"Whenever this creature enters or attacks, create a green and white
//	 Elemental creature token with 'This token's power and toughness are
//	 each equal to the number of creatures you control.'"
//
// One trigger on two events (WhenThisEntersOrAttacks). The token's size is a
// characteristic-defining ability, so it is live and counts itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bf30b8ae-0125-4a00-bae8-ac1aa7703a08",
		Name:         "Vernal Sovereign",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Vernal Sovereign — create a green and white Elemental token",
				Do(CreateToken{Template: cmdr2S12CreatureCountElementalToken(), N: 1})),
		},
	})
}
