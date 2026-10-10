package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Galadriel, Elven-Queen — Legendary Creature — Elf Noble {2}{G}{U},
// 4/5 (EDHREC rank 10901):
//
//	"Will of the council — At the beginning of combat on your turn, if
//	 another Elf entered the battlefield under your control this turn,
//	 starting with you, each player votes for dominion or guidance. If
//	 dominion gets more votes, the Ring tempts you, then you put a +1/+1
//	 counter on your Ring-bearer. If guidance gets more votes or the vote
//	 is tied, draw a card."
//
// The card #2143 was opened for. "If another Elf entered … this turn"
// is an intervening if (CR 603.4), checked as the trigger would go on
// the stack and again as it resolves, against the turn's tally of
// entries by subtype — Éowyn, Shieldmaiden's reading: the Elf as it
// entered, so one that has since died still counts, and "another"
// needs two entries when Galadriel herself entered this turn.
//
// The vote is ADR 0146's (CR 701.38). Dominion tempts Galadriel's
// controller with the Ring (CR 701.54) and then puts a +1/+1 counter on
// the creature they chose as Ring-bearer; one who controls no creature
// is still tempted and nothing gets the counter. A tie goes to
// guidance, as printed.
//
// A bot voting on its own Galadriel votes dominion; an opponent's bot
// votes guidance, the smaller gift.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b97a11c2-7119-4399-9565-bf69520d3988",
		Name:         "Galadriel, Elven-Queen",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventStepBegan},
			AppliesTo: AllOf(
				StepBegan(game.StepBeginCombat, true),
				func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return anotherEnteredWithSubtypeThisTurn(g, source, "Elf")
				},
			),
			Key:    "Galadriel, Elven-Queen — each player votes for dominion or guidance",
			Effect: galadrielElvenQueenVote,
		}},
	})
}

// anotherEnteredWithSubtypeThisTurn is "if another <subtype> entered
// the battlefield under your control this turn": the turn's entry tally
// for the source's controller, needing one more entry when the source
// itself entered this turn (Éowyn, Shieldmaiden; Galadriel,
// Elven-Queen).
func anotherEnteredWithSubtypeThisTurn(g *game.Game, source *game.Card, subtype string) bool {
	need := 1
	if g.EnteredThisTurn(source.InstanceID) {
		need = 2
	}
	return g.EnteredWithSubtypeThisTurn(source.Controller, subtype) >= need
}

func galadrielElvenQueenVote(g *game.Game, item *game.StackItem) error {
	src, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || !anotherEnteredWithSubtypeThisTurn(g, &src, "Elf") {
		// CR 603.4: the intervening if is checked again on resolution.
		return nil
	}
	return Vote{
		Question:      "Galadriel, Elven-Queen — vote for dominion or guidance",
		Words:         []string{"dominion", "guidance"},
		ForController: []int{2, 1},
		ForOpponents:  []int{0, 1},
		Then:          galadrielElvenQueenVoted,
	}.Apply(NewContext(g, item))
}

var galadrielElvenQueenVoted = VoteResultThen("vote/galadriel-elven-queen", func(ctx *Context, r game.VoteResult) error {
	if !r.MoreVotes(0, 1) {
		return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
	}
	return TheRingTemptsYou{Then: func(ctx *Context, ringBearer uuid.UUID) error {
		if ringBearer == uuid.Nil {
			return nil
		}
		return AddCounter{Target: ringBearer, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	}}.Apply(ctx)
})
