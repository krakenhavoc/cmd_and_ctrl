package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Verdant Kraken — Creature — Plant Kraken {4}{G}{G}{G}, 6/6:
//
//	"At the beginning of each player's upkeep, you create a 3/3 green
//	 Forest Tentacle land creature token. (It has '{T}: Add {G}.' It's
//	 affected by summoning sickness until your next turn.)"
//
// "You" is Verdant Kraken's controller on every player's upkeep. The
// token is a land creature with the Forest subtype, so it taps for {G}
// like any Forest (the Forest Dryad's shape); made on another player's
// upkeep, it has not been under your control since your turn began, so
// it can neither attack nor tap for mana until your next turn begins
// (CR 302.6, 305.1).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "83c61083-b357-4775-a855-5a9013cc2bdf",
		Name:         "Verdant Kraken",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtEachUpkeep("Verdant Kraken — create a 3/3 Forest Tentacle land creature token",
				Do(CreateToken{Template: TokenCard("3/3 green Forest Tentacle"), N: 1})),
		},
	})
}
