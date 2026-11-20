package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Battle Brawler — Creature — Orc Warrior {1}{B}, 2/2 (Fate Reforged,
// #5552):
//
//	"As long as you control a red or white permanent, this creature gets
//	 +1/+0 and has first strike."
//
// Ember Weaver's two statics with a two-colour condition
// (frfSelfWhileYouControlEither): +1/+0 in layer 7c (CR 613.4c) and first
// strike in layer 6 (CR 613.1f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d46d8913-7575-4aef-b511-6c0ca9d4b506",
		Name:         "Battle Brawler",
		Completeness: CompletenessFull,
		Static:       battleBrawlerStatics(),
	})
}

// battleBrawlerStatics is "As long as you control a red or white permanent,
// this creature gets +1/+0 and has first strike.": two statics, one
// condition.
func battleBrawlerStatics() []game.StaticAbility {
	pump := game.StaticAbility{
		Label:     "As long as you control a red or white permanent, this creature gets +1/+0.",
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: frfSelfWhileYouControlEither("R", "W"),
		Apply:     pumpBy(1, 0),
	}
	firstStrike := KeywordGrant(frfSelfWhileYouControlEither("R", "W"), "first strike")
	firstStrike.Label = "As long as you control a red or white permanent, this creature has first strike."
	return []game.StaticAbility{pump, firstStrike}
}
