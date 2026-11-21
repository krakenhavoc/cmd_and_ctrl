package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Trial of a Time Lord — Enchantment — Saga {1}{W}{W}:
//
//	"(As this Saga enters and after your draw step, add a lore
//	 counter. Sacrifice after IV.)
//	 I, II, III — Exile target nontoken creature an opponent controls
//	 until this Saga leaves the battlefield.
//	 IV — Starting with you, each player votes for innocent or guilty.
//	 If guilty gets more votes, the owner of each card exiled with this
//	 Saga puts that card on the bottom of their library."
//
// A Saga chapter whose resolution is ADR 0146's will of the council (CR
// 701.38), so nothing new in the vote: chapters I to III are the "until"
// exile of CR 610.3 (ExileUntil, the record that brings the card back
// the moment the Saga leaves), and chapter IV asks the vote.
//
// "Cards exiled with this Saga" (CR 607.2a) is read where Skyclave
// Apparition reads it: the cards the Saga's own chapter abilities
// exiled, in this incarnation of the Saga (b27ExiledWith, once per
// chapter, since each chapter is its own ability). A card that has left
// exile by then (a flicker effect, a cast from exile) is not one.
//
// Innocent, or a tie, does nothing: the Saga is sacrificed once the
// chapter has resolved and every exiled creature comes back. Guilty
// takes them out of exile first, so the Saga leaving returns nothing.
// Each owner with several of the cards orders their own on the bottom
// (CR 401.4).
//
// A bot casting it votes guilty; an opponent's bot votes innocent.
//
// No simplification.
func init() {
	const name = "Trial of a Time Lord"
	chapters := make([]game.TriggeredAbility, 0, 4)
	for n := 1; n <= 3; n++ {
		label := trialOfATimeLordExileLabel(n)
		chapters = append(chapters, effects.ChapterTriggerTargeting(n, label,
			effects.TargetCreature("target nontoken creature an opponent controls", effects.Not(effects.IsTokenPredicate()), effects.OpponentControls()),
			effects.ExileChosenTargetUntilThisLeaves(name+" — returns when the Saga leaves the battlefield")))
	}
	chapters = append(chapters, effects.ChapterTrigger(4, effects.SagaChapterLabel(name, 4, "each player votes for innocent or guilty"),
		func(g *game.Game, item *game.StackItem) error {
			return effects.Vote{
				Question:      name + " — vote for innocent or guilty",
				Words:         []string{"innocent", "guilty"},
				ForController: []int{0, 1},
				ForOpponents:  []int{1, 0},
				Then:          trialOfATimeLordVoted,
			}.Apply(effects.NewContext(g, item))
		}))
	effects.Register(effects.Spec{
		OracleID:     "a5425d4f-cd6c-4240-89a4-dd13e0e1a95f",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered:    chapters,
	})
}

// trialOfATimeLordExileLabel is chapter n's ability label, which is
// also the event-log key the exile is read back under.
func trialOfATimeLordExileLabel(n int) string {
	return effects.SagaChapterLabel("Trial of a Time Lord", n, "exile target nontoken creature an opponent controls until this Saga leaves the battlefield")
}

var trialOfATimeLordVoted = effects.VoteResultThen("vote/trial-of-a-time-lord", func(ctx *effects.Context, r game.VoteResult) error {
	if !r.MoreVotes(1, 0) {
		return nil
	}
	var ids []uuid.UUID
	for n := 1; n <= 3; n++ {
		ids = append(ids, effects.B27ExiledWith(ctx.Game, ctx.Source(), trialOfATimeLordExileLabel(n))...)
	}
	if len(ids) == 0 {
		return nil
	}
	return effects.TuckCardsToBottomByOwner(ctx, ctx.Item, ids, "Trial of a Time Lord — put these on the bottom of your library in any order")
})
