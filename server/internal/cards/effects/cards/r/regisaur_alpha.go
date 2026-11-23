package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Regisaur Alpha — Creature — Dinosaur {3}{R}{G}, 4/4 (EDHREC rank
// 2501):
//
//	"Other Dinosaurs you control have haste.
//	 When this creature enters, create a 3/3 green Dinosaur creature
//	 token with trample."
//
// Seven power across two bodies for five, and the token swings at
// once: the haste grant is a Layer 6 static over OTHER Dinosaurs the
// controller controls (effective subtypes, so a changeling counts;
// the Alpha itself does not get it, as printed), and the token is
// another Dinosaur, so it enters hasty. The ETB is a real trigger
// with a response window.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0673f4e0-66ff-458c-b4ba-eb067e560cce",
		Name:         "Regisaur Alpha",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Dinosaur"}, Others: true, YoursOnly: true}, "haste"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Regisaur Alpha — create a 3/3 green Dinosaur with trample", effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 green Dinosaur with trample"), N: 1})),
		},
	})
}
