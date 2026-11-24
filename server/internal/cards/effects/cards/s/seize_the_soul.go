package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Seize the Soul — Instant {2}{B}{B} (Guildpact):
//
//	"Destroy target nonwhite, nonblack creature. Create a 1/1 white Spirit
//	 creature token with flying.
//	 Haunt (When this spell card is put into a graveyard after resolving,
//	 exile it haunting target creature.)
//	 When the creature this card haunts dies, destroy target nonwhite,
//	 nonblack creature. Create a 1/1 white Spirit creature token with
//	 flying."
//
// Haunt on a spell is CR 702.55a's second sentence (#6083): the card
// haunts only if it resolved, so a Seize the Soul whose target became
// illegal (CR 608.2b) neither makes the Spirit nor haunts. From exile its
// owner controls the second trigger (CR 108.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "91557f08-15b1-4e3e-90fd-cec35ed0868a",
		Name:         "Seize the Soul",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetNonwhiteNonblackCreature(),
		OnResolve: effects.ResolveSteps(effects.Instruction(func(ctx *effects.Context) error {
			return effects.DestroyTargetAndMakeAWhiteSpirit(ctx.Game, ctx.Item)
		})),
		Triggered: []game.TriggeredAbility{
			effects.HauntSpell("Seize the Soul"),
			effects.Targeting(
				effects.WhenTheCreatureThisCardHauntsDies("Seize the Soul — destroy target nonwhite, nonblack creature and create a 1/1 white Spirit with flying",
					effects.DestroyTargetAndMakeAWhiteSpirit),
				effects.TargetNonwhiteNonblackCreature(),
			),
		},
	})
}
