package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Doc Ock, Sinister Scientist — Legendary Creature — Human Scientist
// Villain {4}{U}, 4/5 (Marvel Super Heroes Commander):
//
//	"As long as there are eight or more cards in your graveyard, Doc Ock
//	 has base power and toughness 8/8.
//	 As long as you control another Villain, Doc Ock has hexproof."
//
// Two live statics on the source alone. The first is a layer 7b SET, so
// counters and anthems still land on top of the 8/8; the graveyard is
// read on every recompute and a card entering or leaving it is a zone
// move, which drops the layer cache. The second is a layer 6 grant gated
// on another Villain, counted after the layers so a changeling counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dbaa70e6-26de-4ce9-9543-e652cbe387c1",
		Name:         "Doc Ock, Sinister Scientist",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:    "As long as there are eight or more cards in your graveyard, this has base power and toughness 8/8.",
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7B_Set,
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					if !selfOnly(target, g, source) {
						return false
					}
					p := g.PlayerByIDForEffect(source.Controller)
					return p != nil && p.Graveyard != nil && len(p.Graveyard.Cards) >= 8
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power, c.Toughness = 8, 8
				},
			},
			KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return hobA9WhileYouControlAnother(source.TextSubtype("Villain"))(target, g, source)
			}, "hexproof"),
		},
	})
}
