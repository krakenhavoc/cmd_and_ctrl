package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b537_test.go — Shadowmoor, batch 537 (#5591).

const (
	oldB537Safewright  = "003c48d4-0303-4970-8637-47ae060385aa"
	oldB537Folio       = "0440f85f-8128-4399-9a3d-51f39e196c96"
	oldB537Viridescent = "3003e19e-ae60-4e10-b688-d0cac2a63cb2"
	oldB537Salvage     = "a05d478c-a703-4fc4-9496-2894a6acdaef"
	oldB537Aphotic     = "db951ec2-d751-479d-af1f-885d3cf50027"
	oldB537Niveous     = "e4ecd56c-216d-4c3a-9dfa-09ae1be03e4f"
)

func TestOldB537EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB537Safewright: "Safewright Quest", oldB537Folio: "Illuminated Folio",
		oldB537Viridescent: "Viridescent Wisps", oldB537Salvage: "Dream Salvage",
		oldB537Aphotic: "Aphotic Wisps", oldB537Niveous: "Niveous Wisps",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB537Elf(g *game.Game) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Elf", TypeLine: "Creature — Elf", Colors: []string{"G"},
		Power: 1, Toughness: 1, Owner: g.Seats[1].ID, Controller: g.Seats[1].ID,
	})
}

func TestOldB537WispsRecolourAndRiderThenDraw(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, color, keyword string
		power                        int
		tapped                       bool
	}{
		{"Aphotic Wisps", oldB537Aphotic, "B", "fear", 1, false},
		{"Viridescent Wisps", oldB537Viridescent, "G", "", 2, false},
		{"Niveous Wisps", oldB537Niveous, "W", "", 1, true},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		elf := oldB537Elf(g)
		hand := me.Hand.Size()
		effectstest.CastCatalogSpell(t, g, tc.name, "Instant", tc.oracle, effectstest.CardTarget(elf))
		effectstest.PassPriorityAroundTable(t, g)
		colors := effectstest.EffectiveColors(t, g, elf)
		if len(colors) != 1 || colors[0] != tc.color {
			t.Errorf("%s: colours %v, want only %s", tc.name, colors, tc.color)
		}
		if got := effectstest.EffectivePower(t, g, elf); got != tc.power {
			t.Errorf("%s: power %d, want %d", tc.name, got, tc.power)
		}
		if tc.keyword != "" && !effectstest.HasKeywordOnBattlefield(t, g, elf, tc.keyword) {
			t.Errorf("%s: no %s", tc.name, tc.keyword)
		}
		if got := effectstest.IsTapped(g, elf); got != tc.tapped {
			t.Errorf("%s: tapped = %v, want %v", tc.name, got, tc.tapped)
		}
		if me.Hand.Size() != hand+1 {
			t.Errorf("%s: hand %d, want %d", tc.name, me.Hand.Size(), hand+1)
		}
	}
}

func TestOldB537WispsFizzleOnAGoneTargetAndDrawNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elf := oldB537Elf(g)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Aphotic Wisps", "Instant", oldB537Aphotic, effectstest.CardTarget(elf))
	g.WithWriteLock(func() { g.Battlefield.Remove(elf) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("a fizzled Wisps still drew: hand %d → %d", hand, me.Hand.Size())
	}
}

func TestOldB537SafewrightQuestFetchesAForestOrPlainsNotAnIsland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	isl := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	forest := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Wooded Hollow", TypeLine: "Land — Forest"})
	plains := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	effectstest.CastCatalogSpell(t, g, "Safewright Quest", "Sorcery", oldB537Safewright, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	got := 0
	for _, id := range []uuid.UUID{forest, plains} {
		if effectstest.InHandOf(me, id) {
			got++
		}
	}
	if got != 1 {
		t.Errorf("%d Forest or Plains cards fetched, want exactly 1", got)
	}
	if effectstest.InHandOf(me, isl) {
		t.Error("fetched an Island")
	}
}

func TestOldB537DreamSalvageDrawsForTheOpponentsDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(opp.ID, 2) })
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	if n := g.TurnTallyFor(opp.ID).CardsDiscarded; n != 2 {
		t.Fatalf("setup: opponent discarded %d, want 2", n)
	}
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Dream Salvage", "Instant", oldB537Salvage, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	// The spell leaves the hand as it is cast (+1 pushed, -1 cast).
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d, want %d: two cards for the opponent's two discards, none for mine", me.Hand.Size(), hand+2)
	}
}

func TestOldB537DreamSalvageRefusesYourselfAsTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	if err := effectstest.CastCatalogSpellErr(t, g, "Dream Salvage", "Instant", oldB537Salvage, []game.TargetRef{effectstest.PlayerTarget(me.ID)}); err == nil {
		t.Error("targeting yourself was accepted")
	}
}

func TestOldB537IlluminatedFolioNeedsTwoCardsSharingAColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	folio := effectstest.PushCatalogPermanent(g, me.ID, "Illuminated Folio", "Artifact — Book", oldB537Folio, false)
	blue := effectstest.AvB1Hand(me, game.Card{Name: "Blue", TypeLine: "Creature — Bear", Colors: []string{"U"}})
	blue2 := effectstest.AvB1Hand(me, game.Card{Name: "Blue Two", TypeLine: "Creature — Bear", Colors: []string{"U", "R"}})
	red := effectstest.AvB1Hand(me, game.Card{Name: "Red", TypeLine: "Creature — Bear", Colors: []string{"R"}})
	hand := me.Hand.Size()

	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, folio, 0, game.ActivateAbilityParams{Strict: true, RevealIDs: effectstest.Ids(blue, red)}); err == nil {
		t.Fatal("two cards sharing no colour paid")
	}
	if err := g.ActivateCatalogAbility(me.ID, folio, 0, game.ActivateAbilityParams{Strict: true, RevealIDs: effectstest.Ids(blue, blue2)}); err != nil {
		t.Fatalf("activate with a shared colour: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d: revealing moves nothing, the ability draws one", me.Hand.Size(), hand+1)
	}
	if !effectstest.IsTapped(g, folio) {
		t.Error("the Folio did not tap")
	}
}
