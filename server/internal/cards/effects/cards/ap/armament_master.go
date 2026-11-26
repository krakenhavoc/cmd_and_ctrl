package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Armament Master — Creature — Kor Soldier {W}{W}, 2/2 (Zendikar, S82 OLD b496):
//
//	"Other Kor creatures you control get +2/+2 for each Equipment attached to this
//	 creature."
//
// Kellan, the Fae-Blooded's static, over the other Kor creatures its controller
// controls (Kor, a word of the card's text, CR 612.2) and +2/+2 per Equipment
// attached to the Master itself, counted at every recompute so an attach or
// unattach moves the bonus in the same pass. The Master alone gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f1dca9ac-8886-4688-9245-763ef486701c",
		Name:         "Armament Master",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Other Kor creatures you control get +2/+2 for each Equipment attached to this creature.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return effects.OtherCreaturesYouControl(target, nil, source) && target.HasSubtype(source.TextSubtype("Kor"))
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				if g == nil || source == nil {
					return
				}
				for i := range g.Battlefield.Cards {
					att := &g.Battlefield.Cards[i]
					if att.IsAttachedTo(source.InstanceID) && att.HasSubtype("Equipment") {
						c.Power += 2
						c.Toughness += 2
					}
				}
			},
		}},
	})
}
