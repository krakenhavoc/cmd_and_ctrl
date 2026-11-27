package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gigantomancer — {7}{G} Creature — Human Shaman 1/1 (Rise of the Eldrazi, S82 OLD #5581):
//
//	{1}: Target creature you control has base power and toughness 7/7 until end of turn.
//
// A layer 7b set that lasts until end of turn, so counters and pumps still apply on top (CR 613.4b); a later set wins by timestamp.

func init() {
	effects.Register(effects.Spec{
		OracleID:     "c4b5567a-f195-4638-94b5-0b933dba1850",
		Name:         "Gigantomancer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: Target creature you control has base power and toughness 7/7 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}"),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				t, ok := effects.FirstLegalTarget(ctx)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return effects.UntilEndOfTurn(ctx, t.ID, nil, "Gigantomancer — base power and toughness 7/7 until end of turn",
					game.SetBasePowerMod(7), game.SetBaseToughnessMod(7))
			},
		}},
	})
}
