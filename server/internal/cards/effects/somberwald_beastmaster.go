package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Somberwald Beastmaster — Creature — Human Ranger {6}{G}, 1/1 (Midnight
// Hunt Commander, #5497):
//
//	"When this creature enters, create a 2/2 green Wolf creature token, a
//	 3/3 green Beast creature token, and a 4/4 green Beast creature token.
//	 Creature tokens you control have deathtouch."
//
// The three tokens are made by one creation, one group each, so they
// enter together (CR 111.2, 608.2f) and a "one or more tokens" watcher
// sees a single event. The deathtouch is a layer 6 grant over every
// creature token its controller controls (Jaheira's predicate), the new
// tokens and the Beastmaster itself excluded unless it is a token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "df35bf19-4cbf-4624-925d-ceca6bff596e",
		Name:         "Somberwald Beastmaster",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Somberwald Beastmaster — create a 2/2 Wolf, a 3/3 Beast and a 4/4 Beast", somberwaldBeastmasterTokens),
		},
		Static: []game.StaticAbility{
			Labelled(b16GrantKeywords(creatureTokensYouControl, "deathtouch"),
				"Creature tokens you control have deathtouch."),
		},
	})
}

func somberwaldBeastmasterTokens(g *game.Game, item *game.StackItem) error {
	tr := g.TextReaderForEffect(item)
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     item.SourceCardID,
		Groups: []game.TokenGroup{
			{Template: tr.TokenThroughText(TokenCard("2/2 green Wolf")), Count: 1},
			{Template: tr.TokenThroughText(TokenCard("3/3 green Beast")), Count: 1},
			{Template: tr.TokenThroughText(TokenCard("4/4 green Beast")), Count: 1},
		},
	}, game.KeyedTail{})
}
