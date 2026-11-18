package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Darling of the Masses — Creature — Elf Citizen {2}{G}{W}, 2/4 (Streets of
// New Capenna, #5491):
//
//	"Other Citizens you control get +1/+0.
//	 Whenever this creature attacks, create a 1/1 green and white Citizen
//	 creature token."
//
// The lord is Canopy Tactician's TribalAnthem (others, yours only) at +1/+0.
// The attack trigger makes the Citizen token from the token catalog, as
// Boss's Chauffeur's death trigger does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "96818698-c104-48f4-b73f-c8722d1d1be4",
		Name:         "Darling of the Masses",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Citizen"}, Others: true, YoursOnly: true}, 1, 0),
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Darling of the Masses — create a 1/1 green and white Citizen creature token",
				Do(CreateToken{Template: TokenCard("1/1 green and white Citizen"), N: 1})),
		},
	})
}
