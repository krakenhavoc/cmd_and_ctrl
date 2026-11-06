package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Compassionate Healer — Creature — Human Cleric Ally {1}{W}, 2/2
// (Avatar: The Last Airbender, #4459):
//
//	"Whenever this creature becomes tapped, you gain 1 life and scry 1."
//
// "Becomes tapped" is the tap event and the attack that taps it
// (ThisBecameTapped, #4534), so a vigilance attack is not one; any tap
// counts, a waterbend payment or a convoke tap included. As the ruling
// says, the ability gives no way to tap the Healer. Both halves are
// footprinted Do steps.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6ef464bd-2ed0-4f99-af4a-ebf276f55384",
		Name:         "Compassionate Healer",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
				"Compassionate Healer — you gain 1 life and scry 1",
				Do(GainLife{Amount: 1}, Scry{N: 1})),
		},
	})
}
