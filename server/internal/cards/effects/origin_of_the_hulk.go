package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "7398c8b2-373a-41c7-ac6c-7329f5944462",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "create a 1/1 green and white Citizen creature token"),
				Do(CreateToken{Template: TokenCard("1/1 green and white Citizen"), N: 1})),
			ChapterTriggerTargeting(2, SagaChapterLabel(name, 2, "put two +1/+1 counters on target creature you control"),
				TargetCreature("target creature you control", YouControl()), twoPlusOneCountersOnFirstLegalTarget),
			ChapterTriggerTargeting(3, SagaChapterLabel(name, 3, "target creature you control gets +3/+3 and gains trample until end of turn"),
				TargetCreature("target creature you control", YouControl()),
				Do(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
					return untilEndOfTurn(ctx, id, nil, "Origin of the Hulk — +3/+3 and trample",
						game.ModifyPTMod(3, 3), game.AddKeywordsMod("trample"))
				}))),
		},
	})
}
