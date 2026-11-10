package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Mycotyrant — Legendary Creature — Elder Fungus {1}{B}{G} (The Lost
// Caverns of Ixalan, #4470):
//
//	"Trample
//	 The Mycotyrant's power and toughness are each equal to the number of
//	 creatures you control that are Fungi and/or Saprolings.
//	 At the beginning of your end step, create X 1/1 black Fungus creature
//	 tokens with 'This token can't block,' where X is the number of times
//	 you descended this turn."
//
// The size is a characteristic-defining ability (CR 604.3, 613.4a) over the
// layered subtypes, so a Fungus or a Saprolling counts once however many of
// the two it is, and The Mycotyrant itself is a Fungus. X is the number of
// times its controller descended this turn (CR 700.11), one for each
// permanent card put into their graveyard from anywhere, read as the
// trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b69bbb8d-7f8c-45d6-983a-759b5fce6ebd",
		Name:            "The Mycotyrant",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			Label:     "Power and toughness are each equal to the number of creatures you control that are Fungi and/or Saprolings.",
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				for _, p := range g.BattlefieldCardsForEffect() {
					if p.Controller == source.Controller && p.IsCreature() &&
						(p.HasSubtypeWord("Fungus") || p.HasSubtypeWord("Saproling")) {
						n++
					}
				}
				c.Power, c.Toughness = n, n
			},
		}},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("The Mycotyrant — create X 1/1 black Fungus creature tokens with \"This token can't block\"",
				Do(CreateTokensForEach{
					Template: FungusCantBlockToken(),
					Name:     "the number of times you descended this turn",
					Count:    func(ctx *Context) int { return timesDescendedThisTurn(ctx.Game, ctx.Controller()) },
				})),
		},
	})
}
