package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rakdos Pit Dragon — Creature — Dragon {2}{R}{R}, 3/3 (Ravnica Remastered,
// #5948):
//
//	"{R}{R}: This creature gains flying until end of turn.
//	 {R}: This creature gets +1/+0 until end of turn.
//	 Hellbent — This creature has double strike as long as you have no cards
//	 in hand."
//
// The two abilities are Endling's grant and Cinder Shade's pump. Hellbent
// (CR 207.2c, an ability word with no rules meaning of its own) is a layer 6
// grant to this creature while its controller's hand is empty; the static
// declares that it depends on hand sizes, so a draw, a discard or a card put
// into the hand re-reads it.
//
// No simplification.
func init() {
	hellbent := KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return selfCreature(target, g, source) && stickerHandSize(g, source.Controller) == 0
	}, "double strike")
	hellbent.Label = "Hellbent — Has double strike as long as you have no cards in hand."
	hellbent.DependsOnHandSize = true
	Register(Spec{
		OracleID:     "63dd611e-1c36-4e5d-ba45-87c7f629cd54",
		Name:         "Rakdos Pit Dragon",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{hellbent},
		Activated: []ActivatedAbility{
			{
				Label:   "{R}{R}: This creature gains flying until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    ManaCost("{R}{R}"),
				Effect:  thisGainsKeywordUntilEndOfTurn("flying", "Rakdos Pit Dragon — flying until end of turn"),
			},
			thisGetsRow("{R}", 1, 0, "Rakdos Pit Dragon"),
		},
	})
}
