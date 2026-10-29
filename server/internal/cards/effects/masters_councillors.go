package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Master's Councillors — Creature — Human Advisor {1}{U}, 1/3:
//
//	"Vigilance
//	 This creature gets +2/+0 for each graveyard with seven or more
//	 cards in it.
//	 Whenever you draw your second card each turn, target player
//	 mills three cards."
//
// The bonus counts EVERY graveyard at the table, the controller's and
// each opponent's, and is re-read on every layer pass, so a graveyard
// that fills past six or is exiled below seven moves it at once. A
// player who has left the game has no graveyard to count.
//
// The mill is Pelargir Survivor's TargetPlayerMills on Knights of Dol
// Amroth's "second card" condition.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b9079564-5c86-4765-988a-f2334e016fa6",
		Name:            "Master's Councillors",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static: []game.StaticAbility{{
			Label:     "This creature gets +2/+0 for each graveyard with seven or more cards in it.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				c.Power += 2 * mastersCouncillorsFullGraveyards(g)
			},
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Master's Councillors — target player mills three cards", TargetPlayerMills(3)),
				TargetPlayer("target player")),
		},
	})
}

// mastersCouncillorsFullGraveyards counts the graveyards with seven or
// more cards in them.
func mastersCouncillorsFullGraveyards(g *game.Game) int {
	n := 0
	for _, p := range g.Seats {
		if p != nil && p.Graveyard != nil && len(p.Graveyard.Cards) >= 7 {
			n++
		}
	}
	return n
}
