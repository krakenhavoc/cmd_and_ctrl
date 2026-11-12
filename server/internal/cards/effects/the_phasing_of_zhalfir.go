package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Phasing of Zhalfir — Enchantment — Saga {2}{U}{U} (#5664):
//
//	"Read ahead (Choose a chapter and start with that many lore
//	 counters. Add one after your draw step. Skipped chapters don't
//	 trigger. Sacrifice after III.)
//	 I, II — Another target nonland permanent phases out. It can't
//	 phase in for as long as you control this Saga.
//	 III — Destroy all creatures. For each creature destroyed this way,
//	 its controller creates a 2/2 black Phyrexian creature token."
//
// Read ahead is the engine's (CR 702.155, game/read_ahead.go).
//
// Chapters I and II are two sentences. The phase-out is unconditional.
// The second sentence is a prohibition with a duration: while you
// control the Saga the permanent sits out its controller's untap steps,
// and when the Saga is sacrificed, leaves another way or changes
// controller, the permanent phases in during its controller's next
// untap step (CR 502.1), not at once. If the Saga is already gone as
// the chapter resolves, the duration never starts (CR 611.2b) and the
// permanent comes back on the ordinary clock.
//
// Chapter III does not touch the phased-out permanents (CR 702.26b).
// A creature that is not destroyed (indestructible, regenerated) makes
// no token, and the token goes to the player who controlled the
// creature, so a wipe that hits four boards pays four players.
//
// No simplification.
func init() {
	wipe := ChapterTrigger(3, SagaChapterLabel("The Phasing of Zhalfir", 3, "destroy all creatures; a 2/2 Phyrexian for each"),
		phasingOfZhalfirWipe)
	wipe.Purpose = game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}}
	target := func() *game.TargetSpec {
		return Another(TargetPermanent("another target nonland permanent", Nonland()))
	}
	Register(Spec{
		OracleID:        "6b92a5f4-a241-48c8-9aec-c8e9d5b01fa5",
		Name:            "The Phasing of Zhalfir",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordReadAhead},
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, SagaChapterLabel("The Phasing of Zhalfir", 1, "another target nonland permanent phases out"),
				target(), phasingOfZhalfirPhaseOut),
			ChapterTriggerTargeting(2, SagaChapterLabel("The Phasing of Zhalfir", 2, "another target nonland permanent phases out"),
				target(), phasingOfZhalfirPhaseOut),
			wipe,
		},
	})
}

// phasingOfZhalfirPhaseOut is chapters I and II.
func phasingOfZhalfirPhaseOut(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := legalTargetCards(item, g)
	if err := (PhaseOut{Targets: targets}).Apply(ctx); err != nil {
		return err
	}
	return CantPhaseInWhileYouControl{Targets: targets, Source: item.SourceCardID}.Apply(ctx)
}

// phasingOfZhalfirWipe is chapter III.
func phasingOfZhalfirWipe(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return DestroyAllMatching{Match: Creature(), ThenTail: ctx.Tail(phasingOfZhalfirTokens)}.Apply(ctx)
}

// phasingOfZhalfirTokens is "for each creature destroyed this way, its
// controller creates a 2/2 black Phyrexian creature token", keyed. An
// on-disk key: never renamed, never reused.
var phasingOfZhalfirTokens = MassTail("mass/phasing-of-zhalfir-phyrexians", func(ctx *Context, swept []game.Card, _ int) error {
	tally := controllersOf(swept)
	// Seat order, not map order, so the event log reproduces.
	for _, p := range ctx.Game.Seats {
		if p == nil || tally[p.ID] <= 0 {
			continue
		}
		if err := (CreateToken{
			Controller: p.ID,
			Template:   TokenCard("2/2 black Phyrexian"),
			N:          tally[p.ID],
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
