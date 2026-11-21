package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Origin of the Hulk — {2}{G} Enchantment — Saga (Marvel Super Heroes
// Commander, #4455):
//
//	"I — Create a 1/1 green and white Citizen creature token.
//	 II — Put two +1/+1 counters on target creature you control.
//	 III — Target creature you control gets +3/+3 and gains trample until
//	       end of turn."
//
// Chapters II and III target as the chapter goes on the stack (CR 603.3d)
// and re-check at resolution (CR 608.2b).
//
// No simplification.
func init() {
	const name = "Origin of the Hulk"
	effects.Register(effects.Spec{
		OracleID:     "7398c8b2-373a-41c7-ac6c-7329f5944462",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "create a 1/1 green and white Citizen creature token"),
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green and white Citizen"), N: 1})),
			effects.ChapterTriggerTargeting(2, effects.SagaChapterLabel(name, 2, "put two +1/+1 counters on target creature you control"),
				effects.TargetCreature("target creature you control", effects.YouControl()), effects.TwoPlusOneCountersOnFirstLegalTarget),
			effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel(name, 3, "target creature you control gets +3/+3 and gains trample until end of turn"),
				effects.TargetCreature("target creature you control", effects.YouControl()),
				effects.Do(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.UntilEndOfTurn(ctx, id, nil, "Origin of the Hulk — +3/+3 and trample",
						game.ModifyPTMod(3, 3), game.AddKeywordsMod("trample"))
				}))),
		},
	})
}
