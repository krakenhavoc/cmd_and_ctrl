package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "c4545884-35b7-47cf-939e-3d5d3fe555a4",
		Name:         "Gollum, Patient Plotter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisLeaves("Gollum, Patient Plotter — the Ring tempts you", effects.Do(effects.TheRingTemptsYou{})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{B}, Sacrifice a creature: Return this card from your graveyard to your hand",
			Cost:         effects.Plus(effects.ManaCost("{B}"), effects.SacrificeACreature()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Effect:       effects.ReturnThisCardFromYourGraveyardToHand,
		}},
	})
}
