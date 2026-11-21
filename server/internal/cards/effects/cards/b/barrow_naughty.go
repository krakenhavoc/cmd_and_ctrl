package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Barrow Naughty — Creature — Faerie {1}{B}, 1/3 (Wilds of Eldraine,
// #4471):
//
//	"Flying
//	 This creature has lifelink as long as you control another Faerie.
//	 {2}{B}: This creature gets +1/+0 until end of turn."
//
// The lifelink is Dáin's Company's: a self-only layer 6 keyword grant
// gated on some OTHER Faerie its controller controls, read live, so a
// changeling counts and the grant ends the moment the last other Faerie
// leaves. The pump is Burnout Bashtronaut's self boost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "95664c29-9801-4dfc-9dbd-465abf6bb4ee",
		Name:            "Barrow Naughty",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			effects.KeywordGrant(effects.HobA9WhileYouControlAnother("Faerie"), "lifelink"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{2}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Barrow Naughty — +1/+0"}.Apply(ctx)
			},
		}},
	})
}
