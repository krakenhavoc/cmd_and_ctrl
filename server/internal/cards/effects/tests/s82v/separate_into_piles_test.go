package testss82v

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// separate_into_piles_test.go — #6248, SeparateIntoPiles and
// SacrificeARandomPileEach: "Starting with you, each player separates all
// permanents they control into three piles. Then each player chooses one
// of their piles at random and sacrifices those permanents." (Whims of the
// Fates). CR 700.3 (piles), CR 700.3d (a pile can be empty), CR 101.4
// (turn order), CR 701.21a (sacrifice).

func startPiles(t *testing.T, g *game.Game, then effects.PilesThen) {
	t.Helper()
	me := g.Seats[0]
	src := game.NewCard("Whims", me.ID)
	item := &game.StackItem{Kind: game.StackItemSpell, Controller: me.ID, Owner: me.ID, SourceCardID: src.InstanceID}
	var err error
	g.WithWriteLock(func() { err = effects.SeparateIntoPiles{Piles: 3, Then: then}.Apply(effects.NewContext(g, item)) })
	if err != nil {
		t.Fatalf("SeparateIntoPiles: %v", err)
	}
}

func openPileChoice(g *game.Game) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		if c := g.PendingChoices[i]; c != nil && c.Kind == game.PendingChoiceRevealPick {
			return c
		}
	}
	return nil
}

func makePile(t *testing.T, g *game.Game, who *game.Player, wantOffered []uuid.UUID, pick ...uuid.UUID) {
	t.Helper()
	c := openPileChoice(g)
	if c == nil || c.Chooser != who.ID {
		t.Fatalf("want %s asked for a pile, got %+v", who.Name, c)
	}
	if !sameIDs(c.ChooseCards, wantOffered) {
		t.Fatalf("%s offered %v, want %v", who.Name, c.ChooseCards, wantOffered)
	}
	if err := g.ResolveRevealPick(c.ID, who.ID, pick); err != nil {
		t.Fatalf("%s's pile: %v", who.Name, err)
	}
}

func sameIDs(a, b []uuid.UUID) bool {
	x, y := slices.Clone(a), slices.Clone(b)
	slices.SortFunc(x, func(p, q uuid.UUID) int { return compareUUID(p, q) })
	slices.SortFunc(y, func(p, q uuid.UUID) int { return compareUUID(p, q) })
	return slices.Equal(x, y)
}

func compareUUID(p, q uuid.UUID) int {
	for i := range p {
		if p[i] != q[i] {
			return int(p[i]) - int(q[i])
		}
	}
	return 0
}

func threePermanents(g *game.Game, p *game.Player) []uuid.UUID {
	return []uuid.UUID{
		effectstest.PushArtifactToBattlefieldForTest(g, p.ID, "Rock A", "Artifact", ""),
		effectstest.PushArtifactToBattlefieldForTest(g, p.ID, "Rock B", "Artifact", ""),
		effectstest.PushArtifactToBattlefieldForTest(g, p.ID, "Rock C", "Artifact", ""),
	}
}

var pilesSeen []effects.PlayerPiles

var recordPiles = effects.RegisterPilesThen("number/test-separate-into-piles-record", func(ctx *effects.Context, piles []effects.PlayerPiles) error {
	pilesSeen = piles
	return nil
})

// Starting with you and in turn order, each player makes two piles of
// their own permanents and the rest is the third; a pile may be empty and
// a later pick offers only what is left.
func TestSeparateIntoPilesStartsWithYouInTurnOrder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	pilesSeen = nil
	perms := make([][]uuid.UUID, len(g.Seats))
	for i, p := range g.Seats {
		perms[i] = threePermanents(g, p)
	}
	startPiles(t, g, recordPiles)
	for i, p := range g.Seats {
		x := perms[i]
		if i == 1 {
			// An empty first pile, then two of three.
			makePile(t, g, p, x)
			makePile(t, g, p, x, x[0], x[1])
			continue
		}
		makePile(t, g, p, x, x[0])
		makePile(t, g, p, x[1:], x[1])
	}
	if openPileChoice(g) != nil {
		t.Fatal("a pile prompt is still open after every player split")
	}
	if len(pilesSeen) != len(g.Seats) {
		t.Fatalf("piles for %d players, want %d", len(pilesSeen), len(g.Seats))
	}
	for i, pp := range pilesSeen {
		x := perms[i]
		if pp.Player != g.Seats[i].ID {
			t.Errorf("player %d = %v, want %s (turn order from you)", i, pp.Player, g.Seats[i].Name)
		}
		want := [][]uuid.UUID{{x[0]}, {x[1]}, {x[2]}}
		if i == 1 {
			want = [][]uuid.UUID{nil, {x[0], x[1]}, {x[2]}}
		}
		for k := range want {
			if !sameIDs(pp.Piles[k], want[k]) {
				t.Errorf("%s pile %d = %v, want %v", g.Seats[i].Name, k+1, pp.Piles[k], want[k])
			}
		}
	}
}

// The acceptance: each player splits three permanents 1/1/1; one pile per
// player is chosen at random from the seeded stream and sacrificed, so
// each player loses exactly one permanent, and every other permanent
// stays.
func TestSacrificeARandomPileEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	perms := make([][]uuid.UUID, len(g.Seats))
	for i, p := range g.Seats {
		perms[i] = threePermanents(g, p)
	}
	startPiles(t, g, effects.SacrificeARandomPileEach)
	for i, p := range g.Seats {
		x := perms[i]
		makePile(t, g, p, x, x[0])
		makePile(t, g, p, x[1:], x[1])
	}
	for i, p := range g.Seats {
		left := 0
		for _, id := range perms[i] {
			if g.Battlefield.Contains(id) {
				left++
			}
		}
		if left != 2 {
			t.Errorf("%s has %d of their three permanents, want 2", p.Name, left)
		}
	}
}

// A player with no permanents is asked nothing; a table waiting on a split
// is a restore point.
func TestSeparateIntoPilesSkipsAnEmptyBoardAndSurvivesASnapshot(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	pilesSeen = nil
	a := threePermanents(g, g.Seats[0])
	c := threePermanents(g, g.Seats[2])
	startPiles(t, g, recordPiles)
	makePile(t, g, g.Seats[0], a, a[0])
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var snap game.GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	r, err := snap.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	makePile(t, r, r.Seats[0], a[1:], a[1])
	makePile(t, r, r.Seats[2], c, c[0], c[1], c[2])
	if openPileChoice(r) != nil {
		t.Fatal("a pile prompt is still open: a player with nothing left is asked nothing")
	}
	if len(pilesSeen) != len(r.Seats) {
		t.Fatalf("piles for %d players, want %d", len(pilesSeen), len(r.Seats))
	}
	if !sameIDs(pilesSeen[2].Piles[0], c) || len(pilesSeen[1].Piles[0])+len(pilesSeen[1].Piles[2]) != 0 {
		t.Errorf("piles = %+v", pilesSeen)
	}
}
