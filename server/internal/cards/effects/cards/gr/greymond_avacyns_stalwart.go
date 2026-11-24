package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Greymond, Avacyn's Stalwart — Legendary Creature — Human Soldier {2}{W}{W},
// 3/4 (Secret Lair Drop):
//
//	"As Greymond, Avacyn's Stalwart enters, choose two abilities from among
//	 first strike, vigilance, and lifelink.
//	 Humans you control have each of the chosen abilities.
//	 As long as you control four or more Humans, Humans you control get +2/+2."
//
// The as-enters choice is Frostcliff Siege's anchor-word choice, with the three
// pairs as the anchor words (choosing two of three is choosing the one left out,
// so there are three answers), stored on Greymond (Card.ChosenOption, per object,
// CR 614.12). Each ability is one grant to Humans you control, gated on the pairs
// that include it, so until the choice is made nothing is granted. "Human" is a
// word of Greymond's text (CR 612.2), read on each creature's types through the
// layers, Greymond included. The anthem is a layer 7c modification of the same
// Humans, counted live: four or more Humans you control (Greymond among them,
// changelings included).
//
// No simplification.
func init() {
	const (
		fsVig  = "first strike and vigilance"
		fsLink = "first strike and lifelink"
		vigLnk = "vigilance and lifelink"
	)
	humans := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsCreature() && target.Controller == source.Controller && target.HasSubtype(source.TextSubtype("Human"))
	}
	grant := func(keyword string, options ...string) []game.StaticAbility {
		var out []game.StaticAbility
		for _, o := range options {
			ab := effects.StaticWhenChosen(o, effects.KeywordGrant(humans, keyword))
			ab.Label = "Humans you control have " + keyword + " (chosen: " + o + ")."
			out = append(out, ab)
		}
		return out
	}
	var static []game.StaticAbility
	static = append(static, grant("first strike", fsVig, fsLink)...)
	static = append(static, grant("vigilance", fsVig, vigLnk)...)
	static = append(static, grant("lifelink", fsLink, vigLnk)...)
	static = append(static, game.StaticAbility{
		Label:    "As long as you control four or more Humans, Humans you control get +2/+2.",
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			if !humans(target, g, source) {
				return false
			}
			n := 0
			for _, c := range g.BattlefieldCardsForEffect() {
				if humans(&c, g, source) {
					n++
				}
			}
			return n >= 4
		},
		Apply: effects.PumpBy(2, 2),
	})
	effects.Register(effects.Spec{
		OracleID:     "739a5220-50cd-4148-be58-1bbdeef2954e",
		Name:         "Greymond, Avacyn's Stalwart",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseOptionAsEnters("Greymond, Avacyn's Stalwart", fsVig, fsLink, vigLnk),
		Static:       static,
	})
}
