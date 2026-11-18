package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "add500e1-8eeb-45d7-a10d-2fae2cc69233",
		Name:            "Soul of New Phyrexia",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []ActivatedAbility{
			{
				Label:   "{5}: Permanents you control gain indestructible until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    ManaCost("{5}"),
				Effect:  soulOfNewPhyrexiaIndestructible,
			},
			{
				Label:   "{5}, Exile this card from your graveyard: Permanents you control gain indestructible until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    Plus(ManaCost("{5}"), ExileThis()),
				Zones:   []game.ZoneKind{game.ZoneGraveyard},
				Effect:  soulOfNewPhyrexiaIndestructible,
			},
		},
	})
}

func soulOfNewPhyrexiaIndestructible(g *game.Game, item *game.StackItem) error {
	return GrantKeywordUntilEOT{
		Match:    YouControl(),
		Keywords: []string{"indestructible"},
		Label:    "Soul of New Phyrexia — permanents you control gain indestructible",
	}.Apply(NewContext(g, item))
}
