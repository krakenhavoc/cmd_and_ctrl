package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goblin Grenadiers — Creature — Goblin {3}{R}, 2/2 (Weatherlight, S82 older sets batch 779, #5629):
//
//	"Whenever this creature attacks and isn't blocked, you may sacrifice it. If you do, destroy
//	 target creature and target land."
//
// The trigger puts target creature and target land on the stack as it triggers (CR 603.3d); the "may" is asked as it
// resolves (CR 603.5). Only if the Grenadiers really is sacrificed are the targets destroyed, each still legal as it
// resolves (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Goblin Grenadiers — you may sacrifice it; if you do, destroy target creature and target land"
	effects.Register(effects.Spec{
		OracleID:     "be83a518-cfab-46b1-9fe9-7e3d54311c37",
		Name:         "Goblin Grenadiers",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenAttacksAndIsNotBlockedEffect(label, func(g *game.Game, item *game.StackItem, _ uuid.UUID) error {
				return effects.MayChoice{
					Question: "Goblin Grenadiers — sacrifice it to destroy target creature and target land?",
					YesLabel: "Sacrifice",
					NoLabel:  "Decline",
					Then:     goblinGrenadiersSacrificed,
				}.Apply(effects.NewContext(g, item))
			}), effects.Clauses(
				effects.TargetCreature("target creature"),
				effects.TargetPermanent("target land", effects.Land()),
			)),
		},
	})
}

// goblinGrenadiersSacrificed is the "you may sacrifice it" answered: on a yes the Grenadiers is sacrificed and the two
// targets are destroyed only if it really left. An on-disk key.
var goblinGrenadiersSacrificed = effects.S82OldB367SacrificeAuraAnswer("confirm/s82-old-b779-goblin-grenadiers-sacrifice", effects.S82OldB779DestroyTargetsTail)
