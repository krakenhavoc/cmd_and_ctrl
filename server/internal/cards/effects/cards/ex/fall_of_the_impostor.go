package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fall of the Impostor — Enchantment — Saga {1}{G}{W} (Kaldheim, #5505):
//
//	"I, II — Put a +1/+1 counter on up to one target creature.
//	 III — Exile a creature with the greatest power among creatures target
//	 opponent controls."
//
// Chapters I and II target "up to one" creature, so the chapter goes on the
// stack with no target at all. Chapter III targets the opponent, not a
// creature: the Saga's controller chooses the creature as it resolves
// (CR 608.2c), so hexproof and protection don't stop it. The offer is the
// creatures tied for the greatest power among that opponent's creatures
// (Highcliff Felidar's), read as the chapter resolves; a lone greatest
// creature is exiled unasked and an opponent with no creature loses
// nothing. A target opponent who has left the game makes the chapter do
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	const name = "Fall of the Impostor"
	counter := effects.B36CounterOnChosenAnimal
	upToOne := func() *game.TargetSpec {
		return effects.TargetCreature("up to one target creature").WithCount(0, 1)
	}
	effects.Register(effects.Spec{
		OracleID:     "54deeeef-be90-4975-9213-8355ddf3ce0c",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel(name, 1, "put a +1/+1 counter on up to one target creature"), upToOne(), counter),
			effects.ChapterTriggerTargeting(2, effects.SagaChapterLabel(name, 2, "put a +1/+1 counter on up to one target creature"), upToOne(), counter),
			effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel(name, 3, "exile a creature with the greatest power among creatures target opponent controls"),
				effects.TargetPlayer("target opponent", effects.Opponent()), fallOfTheImpostorExile),
		},
	})
}

// fallOfTheImpostorGreatest is the creatures `of` controls with the
// greatest power among the creatures they control.
func fallOfTheImpostorGreatest(g *game.Game, of uuid.UUID) []uuid.UUID {
	return effects.PermanentsControlledByMatching(g, of, effects.GreatestPowerYouControl())
}

// fallOfTheImpostorExile asks for the creature when the opponent's greatest
// power is tied, and exiles a lone greatest creature unasked.
func fallOfTheImpostorExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var opp uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			opp = t.ID
		}
	}
	if opp == uuid.Nil {
		return nil
	}
	var forced []uuid.UUID
	if ids := fallOfTheImpostorGreatest(g, opp); len(ids) == 1 {
		forced = ids
	}
	return effects.ChoosePermanents{
		Question: "Fall of the Impostor — choose a creature with the greatest power among creatures that player controls; it is exiled",
		Of:       []uuid.UUID{opp},
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			ids := fallOfTheImpostorGreatest(g, of)
			if len(ids) < 2 {
				return nil, 0, 0
			}
			return ids, 1, 1
		},
		Keyed: fallOfTheImpostorExiled,
		Carry: forced,
	}.Apply(ctx)
}

// fallOfTheImpostorExiled exiles the chosen (or unasked) creature. An
// on-disk key.
var fallOfTheImpostorExiled = effects.RegisterPermanentsPick("prompt-run/fall-of-the-impostor-exile", effects.PermanentsDef{
	ThenRun: func(ctx *effects.Context, picked game.PromptedPicks, r game.RunDone) error {
		ids := append([]uuid.UUID(nil), r.Carry...)
		ids = append(ids, picked.Cards()...)
		for _, id := range ids {
			if !effects.OnBattlefield(ctx.Game, id) {
				continue
			}
			if err := (effects.ExileTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	},
})
