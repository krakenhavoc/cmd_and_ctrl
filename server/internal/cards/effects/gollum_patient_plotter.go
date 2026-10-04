package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gollum, Patient Plotter — Legendary Creature — Halfling Horror
// {1}{B}, 3/1:
//
//	"When Gollum leaves the battlefield, the Ring tempts you.
//	 {B}, Sacrifice a creature: Return this card from your graveyard to
//	 your hand. Activate only as a sorcery."
//
// The leaves trigger looks back (CR 603.10a): Gollum is gone, so it
// can't be chosen. The second ability works only from the graveyard,
// and its sacrifice is of a creature on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c4545884-35b7-47cf-939e-3d5d3fe555a4",
		Name:         "Gollum, Patient Plotter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisLeaves("Gollum, Patient Plotter — the Ring tempts you", Do(TheRingTemptsYou{})),
		},
		Activated: []ActivatedAbility{{
			Label:        "{B}, Sacrifice a creature: Return this card from your graveyard to your hand",
			Cost:         Plus(ManaCost("{B}"), SacrificeACreature()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Effect:       returnThisCardFromYourGraveyardToHand,
		}},
	})
}
