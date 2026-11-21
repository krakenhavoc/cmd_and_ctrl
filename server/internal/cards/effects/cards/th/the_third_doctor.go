package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Third Doctor — Legendary Creature — Time Lord Doctor {2}{G}{U}, 2/2 (Doctor Who):
//
//	"Trample
//	 The Third Doctor gets +1/+1 for each noncreature token you control.
//	 When The Third Doctor enters, create your choice of a Clue token, a Food token,
//	 or a Treasure token."
//
// The bonus is a layer-7c static read through the layers, so a Clue, a Food or a
// Treasure counts and a creature token (or an animated one is a creature) does not.
// The enters trigger asks its controller for one of the three (PickOption); each is
// the engine's own token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2a73db99-5e31-49a8-91e7-b28b3434585d",
		Name:            "The Third Doctor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Purpose:         game.Purpose{Tokens: 1},
		Static: []game.StaticAbility{{
			Label:    "The Third Doctor gets +1/+1 for each noncreature token you control.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				for _, p := range g.BattlefieldCardsForEffect() {
					if p.Controller == source.Controller && p.IsToken() && !p.IsCreature() {
						n++
					}
				}
				c.Power += n
				c.Toughness += n
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("The Third Doctor — create your choice of a Clue token, a Food token, or a Treasure token",
				func(g *game.Game, item *game.StackItem) error {
					return effects.PickOption{
						Question: "The Third Doctor — create a Clue, a Food or a Treasure token?",
						Options:  []game.ChoiceOption{{Label: "Clue"}, {Label: "Food"}, {Label: "Treasure"}},
						ThenKey:  thirdDoctorToken,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var thirdDoctorToken = effects.OptionPickThen("option-pick/the-third-doctor-token", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Option == nil {
		return nil
	}
	var tmpl game.Card
	switch r.Index {
	case 1:
		tmpl = effects.FoodToken()
	case 2:
		tmpl = effects.TreasureToken()
	default:
		tmpl = effects.ClueToken()
	}
	return effects.CreateToken{Controller: uuid.Nil, Template: tmpl, N: 1}.Apply(ctx)
})
