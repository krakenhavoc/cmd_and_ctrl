package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Somberwald Beastmaster — Creature — Human Ranger {6}{G}, 1/1 (Midnight
// Hunt Commander, #5497):
//
//	"When this creature enters, create a 2/2 green Wolf creature token, a
//	 3/3 green Beast creature token, and a 4/4 green Beast creature token.
//	 Creature tokens you control have deathtouch."
//
// The three tokens are three footprinted CreateToken steps one after
// another inside one resolution, nothing able to act between them, so
// only a watcher of "one or more tokens entering at once" could tell
// them from a single simultaneous creation. The deathtouch is a layer 6 grant over every
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
			WhenThisEnters("Somberwald Beastmaster — create a 2/2 Wolf, a 3/3 Beast and a 4/4 Beast", Do(
				CreateToken{Template: TokenCard("2/2 green Wolf"), N: 1},
				CreateToken{Template: TokenCard("3/3 green Beast"), N: 1},
				CreateToken{Template: TokenCard("4/4 green Beast"), N: 1})),
		},
		Static: []game.StaticAbility{
			Labelled(b16GrantKeywords(creatureTokensYouControl, "deathtouch"),
				"Creature tokens you control have deathtouch."),
		},
	})
}
