package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soul of New Phyrexia — Artifact Creature — Phyrexian Avatar {6}, 6/6:
//
//	"Trample
//	 {5}: Permanents you control gain indestructible until end of turn.
//	 {5}, Exile this card from your graveyard: Permanents you control
//	 gain indestructible until end of turn."
//
// Two abilities, the second one working from the graveyard only (CR
// 113.6k, Zones). Both grant to the permanents the controller controls
// as the ability resolves (CR 611.2c: the set is fixed then).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "add500e1-8eeb-45d7-a10d-2fae2cc69233",
		Name:            "Soul of New Phyrexia",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{5}: Permanents you control gain indestructible until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.ManaCost("{5}"),
				Effect:  soulOfNewPhyrexiaIndestructible,
			},
			{
				Label:   "{5}, Exile this card from your graveyard: Permanents you control gain indestructible until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.Plus(effects.ManaCost("{5}"), effects.ExileThis()),
				Zones:   []game.ZoneKind{game.ZoneGraveyard},
				Effect:  soulOfNewPhyrexiaIndestructible,
			},
		},
	})
}

func soulOfNewPhyrexiaIndestructible(g *game.Game, item *game.StackItem) error {
	return effects.GrantKeywordUntilEOT{
		Match:    effects.YouControl(),
		Keywords: []string{"indestructible"},
		Label:    "Soul of New Phyrexia — permanents you control gain indestructible",
	}.Apply(effects.NewContext(g, item))
}
