package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jeskai Baller — Creature — Human Athlete {2}{W}, 2/3:
//
//	"When you cast this spell, create a 1/1 white Athlete creature token.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// The cast trigger is the printed one (WhenYouCastThisSpell, so it
// resolves before the spell does). Rebound is printed as the keyword and
// is the engine's (game/rebound.go): CR 702.88a replaces "putting it into
// your graveyard as it resolves", which a permanent spell never does, so
// on a creature it has nothing to replace — the Baller enters the
// battlefield as it resolves and rebound never exiles it.
func init() {
	Register(Spec{
		OracleID:        "e789e7bf-d2f9-41e8-976f-24321e8da928",
		Name:            "Jeskai Baller",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		Triggered: []game.TriggeredAbility{
			WhenYouCastThisSpell("Jeskai Baller — create a 1/1 white Athlete creature token",
				Do(CreateToken{Template: TokenCard("1/1 white Athlete"), N: 1})),
		},
	})
}
