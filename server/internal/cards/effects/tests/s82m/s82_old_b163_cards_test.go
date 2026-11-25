package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB163AriaOracle   = "406fad02-e1a6-4c24-b46d-976c548e1c78"
	oldB163KnightOracle = "673ddb72-d800-49c3-9ea9-ff74a0e20a97"
)

func TestOldB163AriaOfFlameGivesOpponentsTenLifeWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	effectstest.CastCatalogSpell(t, g, "Aria of Flame", "Enchantment", oldB163AriaOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := before[p.ID]
		if p.ID != me.ID {
			want += 10
		}
		if p.Life != want {
			t.Errorf("life %d, want %d", p.Life, want)
		}
	}
}

func TestOldB163AriaOfFlameCountsVersesAndDamagesTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	aria := effectstest.PushCatalogPermanent(g, me.ID, "Aria of Flame", "Enchantment", oldB163AriaOracle, false)
	effectstest.AdvanceToMain(t, g)
	for i, want := range []int{1, 2} {
		life := opp.Life
		cast := effectstest.HandCardFull(me, "Spell", "Instant", "{R}", "", []string{"R"})
		effectstest.FillPoolColored(me, "R", 1)
		if err := g.CastSpell(me.ID, cast, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("cast %d: %v", i, err)
		}
		for j := 0; j < 8; j++ {
			if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
				if err := g.ResolvePickTarget(p.ID, me.ID, effectstest.PbPlayer(opp.ID)); err != nil {
					t.Fatalf("ResolvePickTarget: %v", err)
				}
				continue
			}
			if effectstest.StackFullyEmpty(g) {
				break
			}
			if err := g.PassPriority(); err != nil {
				t.Fatal(err)
			}
		}
		if got := effectstest.CountersOn(g, aria, "verse"); got != want {
			t.Errorf("verse counters %d, want %d", got, want)
		}
		if opp.Life != life-want {
			t.Errorf("opponent life %d, want %d (cast %d)", opp.Life, life-want, i+1)
		}
	}
}

func TestOldB163KnightOfOldBenaliaPumpsTheOthersWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	knight := effectstest.CastCatalogSpell(t, g, "Knight of Old Benalia", "Creature — Human Knight", oldB163KnightOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, mine), effectstest.EffectiveToughness(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("your other creature is %d/%d, want 3/3", p, tough)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 {
		t.Errorf("their creature has power %d, want 2", p)
	}
	// The test card carries no printed power, so a +1/+1 on the Knight itself would show as 1.
	if p := effectstest.EffectivePower(t, g, knight); p != 0 {
		t.Errorf("the Knight itself has power %d, want it unpumped", p)
	}
}

func TestOldB163KnightOfOldBenaliaCanBeSuspended(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	card := effectstest.HandCardFull(me, "Knight of Old Benalia", "Creature — Human Knight", "{3}{W}{W}", oldB163KnightOracle, []string{"W"})
	effectstest.FillPoolColored(me, "W", 1)
	if err := g.PerformSpecialAction(me.ID, card, game.SpecialActionSuspend, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("suspend: %v", err)
	}
	if !g.Exile.Contains(card) {
		t.Fatal("the suspended Knight is not in exile")
	}
	if got := effectstest.CounterCountInExile(g, card, game.CounterTime); got != 5 {
		t.Errorf("%d time counters, want 5", got)
	}
}
