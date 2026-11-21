package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// At Knifepoint — Enchantment {1}{B}{R} (Outlaws of Thunder Junction,
// #4468):
//
//	"During your turn, outlaws you control have first strike.
//	 (Assassins, Mercenaries, Pirates, Rogues, and Warlocks are outlaws.)
//	 Whenever you commit a crime, create a 1/1 red Mercenary creature
//	 token with '{T}: Target creature you control gets +1/+0 until end of
//	 turn. Activate only as a sorcery.' This ability triggers only once
//	 each turn."
//
// The first strike is Hellspur Posse Boss's outlaw grant (CR 700.12),
// your creatures only, held to your turn the way Feisty Spikeling's is:
// it is gone on the other players' turns, when your outlaws block. The
// crime trigger is Marauding Sphinx's once-each-turn crime
// (WheneverYouCommitACrimeOnceEachTurn), and the token is the set's
// Mercenary. The Mercenary is itself an outlaw, so on your turn it has
// first strike too.
//
// No simplification.
func init() {
	outlaws := effects.TribeFilter{
		Tribes:    []string{"Assassin", "Mercenary", "Pirate", "Rogue", "Warlock"},
		YoursOnly: true,
	}
	firstStrike := effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return outlaws.Matches(target, g, source) && effects.IsActivePlayer(g, source.Controller)
	}, "first strike")
	firstStrike.Label = "During your turn, outlaws you control have first strike."
	effects.Register(effects.Spec{
		OracleID:     "bddb2064-afdb-4184-ae4a-633745e64a3c",
		Name:         "At Knifepoint",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{firstStrike},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCommitACrimeOnceEachTurn("At Knifepoint — create a 1/1 red Mercenary", effects.CreateAMercenary()),
		},
	})
}
