package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stingerback Terror — Creature — Scorpion Dragon {2}{R}{R}, 7/7
// (Outlaws of Thunder Junction, #4468):
//
//	"Flying, trample
//	 This creature gets -1/-1 for each card in your hand.
//	 Plot {2}{R} (You may pay {2}{R} and exile this card from your hand.
//	 Cast it as a sorcery on a later turn without paying its mana cost.
//	 Plot only as a sorcery.)"
//
// The shrink is a layer 7c modification (CR 613.4c) of the Terror alone,
// counted live off its controller's hand. A hand is not on the
// battlefield, so the static declares DependsOnHandSize (Psychosis
// Crawler's invalidation): every draw, discard or cast resizes it at
// once. With enough cards in hand it dies to state-based actions (CR
// 704.5f). Plot is the ordinary special action (CR 702.170).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5cc6ca73-c38f-4ef8-a2c9-4f0342eebb73",
		Name:            "Stingerback Terror",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		SpecialActions:  []game.SpecialAction{Plot("{2}{R}")},
		Static: []game.StaticAbility{{
			Label:             "Gets -1/-1 for each card in your hand.",
			Layer:             game.Layer7PT,
			SubLayer:          game.SubLayer7C_Modify,
			DependsOnHandSize: true,
			AppliesTo:         selfCreature,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				if p := g.PlayerByIDForEffect(source.Controller); p != nil && p.Hand != nil {
					n = p.Hand.Size()
				}
				c.Power -= n
				c.Toughness -= n
			},
		}},
	})
}
