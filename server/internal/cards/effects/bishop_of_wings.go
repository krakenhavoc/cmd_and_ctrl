package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bishop of Wings — Creature — Human Cleric {W}{W}, 1/4 (EDHREC rank
// 3824):
//
//	"Whenever an Angel you control enters, you gain 4 life.
//	 Whenever an Angel you control dies, create a 1/1 white Spirit
//	 creature token with flying."
//
// The Angel deck's glue. Two triggers, both on Angels the controller
// controls — effective subtypes, so a changeling counts. The Bishop
// is a Human Cleric and never matches either; the dies trigger reads
// the card in the graveyard, which still says Angel.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "af4684cf-f109-44be-adfb-7c551a36635e",
		Name:         "Bishop of Wings",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventETB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b21AngelYouControlEntered(ev, source, g)
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Bishop of Wings — you gain 4 life", b36GainLife(4))
				},
			},
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b36AngelYouControlDied(ev, source, g)
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Bishop of Wings — create a 1/1 white Spirit with flying",
						b34CreateTokens(b28WhiteSpiritFlyingToken, 1))
				},
			},
		},
	})
}
