package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Midsummer Revel — Enchantment {3}{G}{G} (Urza's Saga, S82 OLD b724):
//
//	At the beginning of your upkeep, you may put a verse counter on this enchantment.
//	{G}, Sacrifice this enchantment: Create X 3/3 green Beast creature tokens, where X is the number of verse counters on this enchantment.
//
// The upkeep counter is a "you may" trigger (Barrin's Codex's shape). The cost is paid with the enchantment sacrificed, so X is the verse counters it had as it last existed on the battlefield (CR 608.2h). The tokens are the shared 3/3 green Beast.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d3579282-5045-4b37-9998-9a282e9932b9",
		Name:         "Midsummer Revel",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.S82OldB724VerseUpkeep("Midsummer Revel")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, Sacrifice this enchantment: Create X 3/3 green Beast creature tokens, where X is the number of verse counters on this enchantment.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.S82OldB724SacrificeCost("{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.CreateToken{Template: effects.TokenCard("3/3 green Beast"), N: effects.S82OldB724X(ctx)}.Apply(ctx)
			},
		}},
	})
}
