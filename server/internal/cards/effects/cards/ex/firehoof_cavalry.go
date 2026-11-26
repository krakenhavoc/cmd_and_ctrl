package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Firehoof Cavalry — Creature — Human Berserker {W}, 1/1 (Khans of Tarkir):
//
//	"{3}{R}: This creature gets +2/+0 and gains trample until end of turn."
//
// Battle Cry Goblin's pump and Roofstalker Wight's grant, both pinned to the
// Cavalry and both lasting until the cleanup step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "487de0df-0496-4b5d-8050-27b8e617582b",
		Name:         "Firehoof Cavalry",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{R}: This creature gets +2/+0 and gains trample until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2}},
			Cost:    effects.ManaCost("{3}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.BoostUntilEOT{Target: item.SourceCardID, Power: 2, Label: "Firehoof Cavalry — +2/+0"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"trample"}, Label: "Firehoof Cavalry — trample"}.Apply(ctx)
			},
		}},
	})
}
