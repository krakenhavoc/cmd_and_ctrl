package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lavabrink Venturer — Creature — Human Soldier {2}{W}, 3/3 (Ikoria: Lair of
// Behemoths, S82 OLD batch 120, #5512):
//
//	"As this creature enters, choose odd or even. (Zero is even.)
//	 This creature has protection from each mana value of the chosen
//	 quality."
//
// The choice is Gollum, Riddle Master's CR 614.12 one (ChooseOptionAsEnters),
// stored on the permanent. Each answer gates one static, the Siege shape:
// "odd" gives this creature protection from odd mana values and "even" from
// even mana values (CR 702.16, over the parity qualities of #4307). The
// source's mana value is read as it is (a token or a land is 0, so even), and
// before the choice is answered the Venturer has neither.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea66ffa3-d084-434f-91a6-fdff49806285",
		Name:         "Lavabrink Venturer",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseOptionAsEnters("Lavabrink Venturer", "odd", "even"),
		Static: []game.StaticAbility{
			effects.StaticWhenChosen("odd", effects.Labelled(effects.KeywordGrant(effects.SelfOnly, "protection from odd mana values"),
				"This creature has protection from odd mana values.")),
			effects.StaticWhenChosen("even", effects.Labelled(effects.KeywordGrant(effects.SelfOnly, "protection from even mana values"),
				"This creature has protection from even mana values.")),
		},
	})
}
