package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mindstab Thrull — Creature — Thrull {1}{B}{B}, 2/2 (Fallen Empires, S82 OLD
// b855, #5637):
//
//	"Whenever this creature attacks and isn't blocked, you may sacrifice it. If
//	 you do, defending player discards three cards."
//
// Necrite's shape without a target: the defending player is the one whose
// block declaration fired the trigger, and they discard three of their own
// choosing only if the Thrull really was sacrificed. A player who has left
// the game discards nothing (CR 800.4a).
//
// No simplification.
func init() {
	const label = "Mindstab Thrull — you may sacrifice it. If you do, defending player discards three cards"
	effects.Register(effects.Spec{
		OracleID:     "43abe1f3-7b92-411a-bf93-e158f949685c",
		Name:         "Mindstab Thrull",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Triggered: []game.TriggeredAbility{
			effects.WhenAttacksAndIsNotBlockedEffect(label, func(g *game.Game, item *game.StackItem, _ uuid.UUID) error {
				return effects.MayChoice{
					Question: "Mindstab Thrull — sacrifice it to make defending player discard three cards?",
					YesLabel: "Sacrifice",
					NoLabel:  "Decline",
					Then:     mindstabThrullSacrificed,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

var mindstabThrullDiscard = effects.CardTail("s82-old-b855/mindstab-thrull-discard", func(ctx *effects.Context, o game.TailOutcome) error {
	defender := effects.TriggeringActor(ctx.Item)
	if !effects.OneWentThisWay(o) || defender == uuid.Nil || !effects.DefendingPlayerStillIn(ctx.Game, defender) {
		return nil
	}
	return effects.DiscardCards{Player: defender, N: 3}.Apply(ctx)
})

var mindstabThrullSacrificed = effects.S82OldB367SacrificeAuraAnswer("confirm/s82-old-b855-mindstab-thrull-sacrifice", mindstabThrullDiscard)
