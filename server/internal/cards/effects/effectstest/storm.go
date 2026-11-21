package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// storm_test.go — CR 702.40 end to end, on the five cards that print
// it. The engine half (TurnTally.Casts, the count, the announcement)
// is pinned in game/storm_test.go; this file is about what a player
// sees.
//
// Storm is the keyword where the obvious test is the wrong one. "N
// copies were created" is easy to assert and tells you almost
// nothing: the mistakes that matter are all about WHICH spells the N
// counted (an opponent's, a countered one, the storm spell itself, a
// copy) and about each copy being its own object with its own
// target and its own resolution. So every test below is either about
// the count's membership or about the copies being separate.

const (
	GrapeshotOracle       = "ebd2d760-5ad8-4027-b124-171822f3edfe"
	BrainFreezeOracle     = "464c0150-3dbc-403b-9ada-fef25ab1f29d"
	EmptyTheWarrensOracle = "3a59c882-8bb8-49ba-862f-125020dd5bec"
	FlusterstormOracle    = "86bf58f2-7f25-4e10-b797-25e0e8e67769"
	TendrilsOracle        = "78db3190-59bc-4033-95bc-32d8ed872b6a"
)

// stormFiller casts a plain non-catalog instant, which is what a
// storm count is made of: a spell that was cast and did nothing else
// worth asserting.
func StormFiller(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := CastCatalogSpell(t, g, "Filler", "Instant", "", nil)
	PassPriorityAroundTable(t, g)
	return id
}

// castFromSeat puts a card in `seat`'s hand and casts it from there,
// whoever's turn it is. castCatalogSpell always casts from the ACTIVE
// seat, and half of storm is about the spells somebody ELSE cast.
func CastFromSeat(t *testing.T, g *game.Game, seat *game.Player, name, typeLine, oracle string, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	id := uuid.New()
	seat.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: seat.ID, Controller: seat.ID,
	})
	if err := g.CastSpell(seat.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("cast %s from seat %s: %v", name, seat.ID, err)
	}
	return id
}

// answerCopyPrompts settles a storm resolution: pass priority until a
// re-target prompt opens, answer it with `target`, repeat, and stop
// when the stack is empty. `wantPrompts` is asserted, because "how
// many separate choices did the player get" IS the per-copy shape
// (ADR 0086 Decision 4) and a batched prompt would pass every
// life-total assertion in this file.
func AnswerCopyPrompts(t *testing.T, g *game.Game, chooser, target uuid.UUID, wantPrompts int) {
	t.Helper()
	seen := 0
	for i := 0; i < 64; i++ {
		if StackFullyEmpty(g) && LatestPickTarget(g, chooser) == nil {
			break
		}
		if p := LatestPickTarget(g, chooser); p != nil {
			seen++
			if err := g.ResolvePickTarget(p.ID, chooser,
				game.TargetRef{Kind: game.TargetPlayer, ID: target}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if seen != wantPrompts {
		t.Errorf("the copies opened %d re-target prompts, want %d (one per copy, CR 707.10c)", seen, wantPrompts)
	}
	if !StackFullyEmpty(g) {
		t.Fatalf("the stack never emptied: %d cards, %d items, %d pending triggers",
			g.Stack.Size(), len(g.StackMeta), len(g.PendingTriggers))
	}
}
