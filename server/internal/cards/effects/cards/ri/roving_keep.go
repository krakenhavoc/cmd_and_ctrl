package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Roving Keep — Artifact Creature — Wall {7}, 5/7 (Jumpstart, #5510):
//
//	"Defender
//	 {7}: This creature gets +2/+0 and gains trample until end of turn. It
//	 can attack this turn as though it didn't have defender."
//
// Three until-end-of-turn effects on the source, the last Hightide Hermit's
// defender waiver (the keyword stays). Nothing happens if the Keep has left
// the battlefield: a new object is not "this creature" (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6a88784c-d3be-4076-83eb-895cc1b50947",
		Name:            "Roving Keep",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{7}: This creature gets +2/+0 and gains trample until end of turn. It can attack this turn as though it didn't have defender.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{7}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				id := item.SourceCardID
				if err := (effects.BoostUntilEOT{Target: id, Power: 2, Label: "Roving Keep — +2/+0"}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Roving Keep — trample"}).Apply(ctx); err != nil {
					return err
				}
				return effects.WaiveDefenderUntilEOT(ctx, id, "Roving Keep — can attack as though it didn't have defender")
			},
		}},
	})
}
