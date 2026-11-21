package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "96818698-c104-48f4-b73f-c8722d1d1be4",
		Name:         "Darling of the Masses",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Citizen"}, Others: true, YoursOnly: true}, 1, 0),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Darling of the Masses — create a 1/1 green and white Citizen creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green and white Citizen"), N: 1})),
		},
	})
}
