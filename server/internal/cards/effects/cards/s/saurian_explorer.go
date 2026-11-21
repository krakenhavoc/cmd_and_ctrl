package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Saurian Explorer — {3}{G} Creature — Dinosaur Scout 4/3 (Star Trek,
// #4973):
//
//	"Landfall — Whenever a land you control enters, this creature gets
//	 +2/+2 until end of turn.
//	 {6}: Create a tapped Planet land token named New Planet with "{T}:
//	 Add one mana of any color." Activate only as a sorcery."
//
// Landfall is an ability word (CR 207.2c). The New Planet matches the
// printed token (ttrk 27).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "00f4cab8-b135-4118-8e3a-92097178abb6",
		Name:         "Saurian Explorer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Saurian Explorer — +2/+2 until end of turn (landfall)",
				effects.ThisGetsUntilEndOfTurn(2, 2, "Saurian Explorer — +2/+2 until end of turn")),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{6}: Create a tapped Planet land token named New Planet with \"{T}: Add one mana of any color.\" Activate only as a sorcery.",
			Cost:         effects.ManaCost("{6}"),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.TrekB22CreateTappedNewPlanet()),
		}},
	})
}
