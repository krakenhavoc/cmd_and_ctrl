package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lost Order of Jarkeld — Creature — Human Knight {2}{W}{W}, 1+*/1+* (Ice Age,
// S82 OLD batch 846, #5636):
//
//	"As this creature enters, choose an opponent.
//	 Lost Order of Jarkeld's power and toughness are each equal to 1 plus the
//	 number of creatures the chosen player controls."
//
// The opponent is chosen as it enters (Black Vise) and stored on the
// permanent. The power and toughness are a characteristic-defining ability
// (CR 604.3, 613.4a): layer 7a, read live, one point per creature the chosen
// player controls through the layers, the Order itself counting if it is
// theirs. Before the choice is made it is nobody's count, a 1/1.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75f047dc-4a51-494d-8765-fa7502d1fd06",
		Name:         "Lost Order of Jarkeld",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChoosePlayerAsEnters("Lost Order of Jarkeld", effects.Opponents),
		Static: []game.StaticAbility{{
			Label: "Lost Order of Jarkeld's power and toughness are each equal to 1 plus the number of creatures the chosen player controls.",
			Layer: game.Layer7PT, SubLayer: game.SubLayer7A_CDA,
			AppliesTo: effects.SelfCreature,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 1
				if player := g.ChosenPlayerOf(source.InstanceID); player != uuid.Nil {
					for i := range g.Battlefield.Cards {
						if bc := &g.Battlefield.Cards[i]; bc.Controller == player && bc.IsCreature() {
							n++
						}
					}
				}
				c.Power, c.Toughness = n, n
			},
		}},
	})
}
