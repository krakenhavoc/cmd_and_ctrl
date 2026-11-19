package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seraph of the Scales — Creature — Angel {2}{W}{B}, 4/3 (Ravnica: Clue
// Edition):
//
//	"Flying
//	 {W}: This creature gains vigilance until end of turn.
//	 {B}: This creature gains deathtouch until end of turn.
//	 Afterlife 2 (When this creature dies, create two 1/1 white and black
//	 Spirit creature tokens with flying.)"
//
// Tithe Taker's afterlife (the dies trigger it reminds, CR 702.135a) with
// two tokens made by Knight of the Last Breath's lazily-read Spirit helper
// (the token row is filed after this file's init runs); the grants are Afterthought Sentry's self grant.
//
// No simplification.
func init() {
	grant := func(kw string) Effect {
		return func(g *game.Game, item *game.StackItem) error {
			return GrantKeywordUntilEOT{
				Target:   item.SourceCardID,
				Keywords: []string{kw},
				Label:    "Seraph of the Scales — " + kw + " until end of turn",
			}.Apply(NewContext(g, item))
		}
	}
	Register(Spec{
		OracleID:        "4a3acbf7-0699-4d7c-b4fa-aa15c7cde189",
		Name:            "Seraph of the Scales",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Tokens: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Seraph of the Scales — afterlife 2: create two 1/1 white and black Spirit creature tokens with flying",
				knightOfTheLastBreathSpirits(2)),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "{W}: This creature gains vigilance until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    ManaCost("{W}"),
				Effect:  grant("vigilance"),
			},
			{
				Label:   "{B}: This creature gains deathtouch until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    ManaCost("{B}"),
				Effect:  grant("deathtouch"),
			},
		},
	})
}
