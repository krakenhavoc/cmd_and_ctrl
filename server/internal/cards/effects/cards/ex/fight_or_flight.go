package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fight or Flight — Enchantment {3}{W} (Invasion):
//
//	"At the beginning of combat on each opponent's turn, separate all
//	 creatures that player controls into two piles. Only creatures in the
//	 pile of their choice can attack this turn."
//
// Fight or Flight's controller separates the active opponent's creatures
// (an empty pile is legal, CR 700.3d), and that opponent chooses the pile
// that may attack. The restriction is read live (CR 611.2c): "can't attack"
// on every creature outside the chosen pile, so one that arrives later this
// turn can't attack either, and a pile creature that left and came back is
// a new object (CR 400.7). An opponent with no creatures is asked nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2fc42640-983a-49ca-9020-7756135077b4",
		Name:         "Fight or Flight",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventStepBegan, effects.AllOf(effects.StepBegan(game.StepBeginCombat, false), effects.ByAnOpponent),
				"Fight or Flight — separate that player's creatures into two piles; only creatures in the pile of their choice can attack this turn",
				fightOrFlightSeparate),
		},
	})
}

func fightOrFlightSeparate(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := effects.NewContext(g, item)
	opp := item.Trigger.Event.Actor
	creatures := effects.CreaturesControlledByPlayer(g, opp)
	if len(creatures) == 0 {
		return nil
	}
	return effects.PileSplit{
		Splitter:      ctx.Controller(),
		Chooser:       opp,
		Owner:         opp,
		SplitQuestion: "Fight or Flight — separate these creatures into two piles",
		PickQuestion:  "Fight or Flight — choose a pile; only creatures in it can attack this turn",
		Cards:         creatures,
		ThenTail:      ctx.Tail(fightOrFlightChosePile),
	}.Apply(ctx)
}

// fightOrFlightChosePile is the chosen pile ("taken"): only it can attack.
// Keyed (#4253): an on-disk key, never renamed, never reused.
var fightOrFlightChosePile = effects.PileTail("then/fight-or-flight-chosen-pile", func(ctx *effects.Context, taken, _ []uuid.UUID, _ game.TailOutcome) error {
	effects.OldB665OnlyPileCanAttackThisTurn(ctx, taken, "Fight or Flight — only creatures in the chosen pile can attack this turn")
	return nil
})
