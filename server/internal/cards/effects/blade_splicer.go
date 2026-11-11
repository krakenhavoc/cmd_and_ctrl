package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blade Splicer — Creature — Phyrexian Human Artificer {2}{W}, 1/1 (Bloomburrow
// Commander):
//
//	"When this creature enters, create a 3/3 colorless Phyrexian Golem
//	 artifact creature token.
//	 Golems you control have first strike."
//
// The token is the table's Phyrexian Golem row. The grant is a layer 6 lord
// over every Golem its controller controls (the Splicer itself is no Golem
// unless something makes it one), read through the layered view.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "194166ad-0179-42a3-86b9-ba7f322ec576",
		Name:         "Blade Splicer",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Blade Splicer — create a 3/3 Phyrexian Golem",
				Do(CreateToken{Template: TokenCard("3/3 colorless Phyrexian Golem artifact"), N: 1})),
		},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Golem"}, YoursOnly: true}, "first strike"),
		},
	})
}
