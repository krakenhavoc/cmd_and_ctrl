package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pteramander — Creature — Salamander Drake {U}, 1/1 (Outlaws of Thunder
// Junction Commander, #5368):
//
//	"Flying
//	 {7}{U}: Adapt 4. This ability costs {1} less to activate for each
//	 instant and sorcery card in your graveyard. (If this creature has no
//	 +1/+1 counters on it, put four +1/+1 counters on it.)"
//
// The discount is the ability's own cost clause counting the instants and
// sorceries in the activator's graveyard (Balamb Garden's shape, Eddymurk
// Crab's count); it lowers only the generic part, so {U} is always paid.
// Adapt checks for counters as it resolves (CR 701.46a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "00187de2-bc48-4137-97d8-a9a0fafc76c1",
		Name:            "Pteramander",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{7}{U}: Adapt 4. This ability costs {1} less to activate for each instant and sorcery card in your graveyard.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{7}{U}"),
			CostModifiers: []game.CostModifier{
				effects.CostsLessEach(effects.SocB37InstantAndSorceryCardsInYourGraveyard,
					"This ability costs {1} less to activate for each instant and sorcery card in your graveyard."),
			},
			Effect: func(g *game.Game, item *game.StackItem) error { return effects.AdaptN(g, item, 4) },
		}},
	})
}
