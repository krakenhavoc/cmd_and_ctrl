package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stand or Fall — Enchantment {3}{R} (Invasion):
//
//	"At the beginning of combat on your turn, for each defending player,
//	 separate all creatures that player controls into two piles and that
//	 player chooses one. Only creatures in the chosen piles can block this
//	 turn."
//
// Attackers are not declared at the beginning of combat, so every
// opponent is a defending player. Stand or Fall's controller separates
// each opponent's creatures in turn order (an empty pile is legal, CR
// 700.3d) and that opponent chooses a pile. The restriction is read live
// (CR 611.2c): each defender's creatures outside their chosen pile can't
// block, including ones that arrive later this turn. An opponent with no
// creatures is asked nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "927d0085-9bd9-4454-8ac5-0e374237ed35",
		Name:         "Stand or Fall",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventStepBegan, effects.StepBegan(game.StepBeginCombat, true),
				"Stand or Fall — for each defending player, separate their creatures into two piles; they choose one; only creatures in the chosen piles can block this turn",
				standOrFallSeparate),
		},
	})
}

func standOrFallSeparate(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, opp := range effects.OpponentsInTurnOrder(g, ctx.Controller()) {
		creatures := effects.CreaturesControlledByPlayer(g, opp)
		if len(creatures) == 0 {
			continue
		}
		tail := ctx.Tail(standOrFallChosePile)
		tail.Carry = []uuid.UUID{opp}
		if err := (effects.PileSplit{
			Splitter:      ctx.Controller(),
			Chooser:       opp,
			Owner:         opp,
			SplitQuestion: "Stand or Fall — separate these creatures into two piles",
			PickQuestion:  "Stand or Fall — choose a pile; only creatures in it can block this turn",
			Cards:         creatures,
			ThenTail:      tail,
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// standOrFallChosePile is one defender's chosen pile ("taken"); the defender
// rides the tail's carry. Keyed (#4253): an on-disk key, never renamed, never
// reused.
var standOrFallChosePile = effects.PileTail("then/stand-or-fall-chosen-pile", func(ctx *effects.Context, taken, _ []uuid.UUID, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	effects.OldB665OnlyPileCanBlockThisTurn(ctx, o.Tail.Carry[0], taken, "Stand or Fall — only creatures in the chosen pile can block this turn")
	return nil
})
