package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

const hazelOracle = "845b40b1-dc09-44fa-9910-22544c1424ba"

func hazelPush(g *game.Game, owner uuid.UUID) uuid.UUID {
	return blbB1Push(g, owner, "Hazel of the Rootbloom", "Legendary Creature — Squirrel Druid", hazelOracle, 3, 5)
}

// hazelSquirrels makes n Squirrel tokens for owner and returns their ids.
func hazelSquirrels(t *testing.T, g *game.Game, owner uuid.UUID, n int) []uuid.UUID {
	t.Helper()
	before := map[uuid.UUID]bool{}
	for _, c := range g.Battlefield.Cards {
		before[c.InstanceID] = true
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(owner, TokenCard("1/1 green Squirrel"), n) })
	var ids []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if !before[c.InstanceID] && IsToken(c) && c.Controller == owner {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) != n {
		t.Fatalf("made %d tokens, want %d", len(ids), n)
	}
	return ids
}

func hazelPoolTotal(p *game.Player) int {
	total := 0
	for _, tok := range p.ManaPool {
		total += tok.N()
	}
	return total
}

// X is the number of tokens tapped: two tokens are two mana, each its own
// colour pick, for 2 life and Hazel's own tap.
func TestHazelAddsOneManaPerTokenTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hazel := hazelPush(g, me.ID)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	tokens := blbBC2Tokens(t, g, me.ID, 3)
	g.WithWriteLock(func() { me.ManaPool = nil })
	life := me.Life

	if err := g.ActivateManaAbility(me.ID, hazel, 0, game.ManaAbilityParams{TapIDs: []uuid.UUID{tokens[0], bear}}); err == nil {
		t.Fatal("a nontoken creature paid the token cost")
	}
	if err := g.ActivateManaAbility(me.ID, hazel, 0, game.ManaAbilityParams{TapIDs: []uuid.UUID{tokens[0], tokens[0]}}); err == nil {
		t.Fatal("one token paid twice")
	}
	assertTapped(t, g, false, hazel, bear, tokens[0])
	if me.Life != life {
		t.Fatalf("a refused activation cost life (%d → %d)", life, me.Life)
	}

	activateManaFor(t, g, me.ID, hazel, 0, game.ManaAbilityParams{TapIDs: tokens[:2]})
	if got := b10ResolveAllManaPicks(t, g, me.ID, "U"); got != 2 {
		t.Fatalf("%d colour picks, want 2 (one per token tapped)", got)
	}
	assertTapped(t, g, true, hazel, tokens[0], tokens[1])
	assertTapped(t, g, false, tokens[2], bear)
	if got := hazelPoolTotal(me); got != 2 {
		t.Errorf("mana produced %d (%v), want 2", got, me.ManaPool)
	}
	if me.Life != life-2 {
		t.Errorf("life %d → %d, want 2 paid", life, me.Life)
	}
}

// A token that arrived this turn pays (CR 302.6), a tapped one does not,
// and tapping none is a legal activation that adds nothing.
func TestHazelTokenCostIsNotTheTapSymbol(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hazel := hazelPush(g, me.ID)
	tokens := blbBC2Tokens(t, g, me.ID, 2)
	g.WithWriteLock(func() {
		me.ManaPool = nil
		findBattlefieldCardForTest(g, tokens[0]).SummonedThisTurn = true
		findBattlefieldCardForTest(g, tokens[1]).Tapped = true
	})
	if err := g.ActivateManaAbility(me.ID, hazel, 0, game.ManaAbilityParams{TapIDs: tokens}); err == nil {
		t.Fatal("an already tapped token paid the cost")
	}
	activateManaFor(t, g, me.ID, hazel, 0, game.ManaAbilityParams{TapIDs: tokens[:1]})
	if got := b10ResolveAllManaPicks(t, g, me.ID, "R"); got != 1 {
		t.Fatalf("%d colour picks, want 1", got)
	}
	if got := hazelPoolTotal(me); got != 1 {
		t.Errorf("mana produced %d, want 1", got)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	hazel2 := hazelPush(g2, me2.ID)
	g2.WithWriteLock(func() { me2.ManaPool = nil })
	life := me2.Life
	activateManaFor(t, g2, me2.ID, hazel2, 0, game.ManaAbilityParams{})
	if got := hazelPoolTotal(me2); got != 0 || me2.Life != life-2 {
		t.Errorf("X=0: mana %d (want 0), life %d → %d (want 2 paid)", got, life, me2.Life)
	}
}

// The enumerator offers one payment per count of tokens, never the
// no-mana X=0, and nothing at all with no token to tap.
func TestHazelManaMovesNameTheTokensTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hazel := hazelPush(g, me.ID)
	count := func() map[int]int {
		sizes := map[int]int{}
		for _, m := range legal.EnumerateFor(g, me.ID) {
			if m.Source != hazel || m.Type != legal.TypeActivateManaAbility {
				continue
			}
			var p struct {
				TapIDs []string `json:"tap_ids"`
			}
			if err := json.Unmarshal(m.Params, &p); err != nil {
				t.Fatalf("decode %s: %v", m.Params, err)
			}
			sizes[len(p.TapIDs)]++
		}
		return sizes
	}
	if got := count(); len(got) != 0 {
		t.Fatalf("mana moves offered with no token to tap: %v", got)
	}
	blbBC2Tokens(t, g, me.ID, 2)
	got := count()
	if got[0] != 0 || got[1] != 1 || got[2] != 1 {
		t.Fatalf("payments by token count = %v, want one each for 1 and 2 and none for 0", got)
	}
}

// The end step trigger copies the target token once, or twice if it is a
// Squirrel.
func TestHazelCopiesATokenTwiceIfItIsASquirrel(t *testing.T) {
	for _, squirrel := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		hazelPush(g, me.ID)
		var token uuid.UUID
		name, want := "Rabbit", 2
		if squirrel {
			token, name, want = hazelSquirrels(t, g, me.ID, 1)[0], "Squirrel", 3
		} else {
			token = blbBC2Tokens(t, g, me.ID, 1)[0]
		}
		advanceToEndStepOf(t, g, 0)
		ch := pickTargetPrompt(t, g)
		if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: token}); err != nil {
			t.Fatalf("squirrel %v: ResolvePickTarget: %v", squirrel, err)
		}
		passPriorityAroundTable(t, g)
		if n, _ := countTokensNamed(g, me.ID, name); n != want {
			t.Errorf("squirrel %v: %d %s tokens, want %d", squirrel, n, name, want)
		}
	}
}

// A nontoken permanent and an opponent's token are not targets.
func TestHazelTargetsOnlyATokenYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hazel := hazelPush(g, me.ID)
	blbBC2Tokens(t, g, me.ID, 1)
	theirs := blbBC2Tokens(t, g, opp.ID, 1)[0]
	advanceToEndStepOf(t, g, 0)
	ch := pickTargetPrompt(t, g)
	for _, bad := range []uuid.UUID{hazel, theirs} {
		if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatalf("%s was accepted as the target", bad)
		}
	}
}
