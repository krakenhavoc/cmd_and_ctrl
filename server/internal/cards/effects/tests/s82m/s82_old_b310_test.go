package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b310_test.go — Battle for Zendikar converge cards (#5547). Every
// test casts the real catalog entry under the strict mana gate, paying with
// exactly the mana named.

const (
	oldB310UnifiedFront = "1f514b02-7827-4087-a6b7-bd6a3f520a64"
	oldB310Roilmage     = "73eaf40a-abc6-486a-958d-374ed6111173"
	oldB310PrismArray   = "8856d849-dad8-41b8-8212-bb839f2b359c"
	oldB310Spectrum     = "88ed66fc-7068-44a7-b00f-3be22fb500cb"
	oldB310Infuse       = "beea5db2-1771-417f-9a31-028b3d31aa66"
	oldB310SkyriderElf  = "c71f19be-aabc-44ef-a62b-57522b3a2c98"
	oldB310ExertInfl    = "feb9cce1-6c3c-4e04-93ee-d646a888bc29"
)

// oldB310Cast casts a card from the active seat's hand paying exactly `mana`
// and passes priority until it resolves.
func oldB310Cast(t *testing.T, g *game.Game, c game.Card, mana string, targets []game.TargetRef, x int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, mana)
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = me.ID, me.ID
	g.WithWriteLock(func() { me.Hand.PushTop(c) })
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true, Targets: targets, XValue: x}); err != nil {
		t.Fatalf("cast %s: %v", c.Name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return c.InstanceID
}

func TestOldB310EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{oldB310UnifiedFront, oldB310Roilmage, oldB310PrismArray, oldB310Spectrum, oldB310Infuse, oldB310SkyriderElf, oldB310ExertInfl} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func TestOldB310UnifiedFrontMakesAKorAllyPerColour(t *testing.T) {
	for _, tc := range []struct {
		mana string
		want int
	}{{"WUBR", 4}, {"WCCC", 1}, {"WWWW", 1}} {
		t.Run(tc.mana, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			oldB310Cast(t, g, game.Card{Name: "Unified Front", TypeLine: "Sorcery", ManaCost: "{3}{W}", OracleID: oldB310UnifiedFront}, tc.mana, nil, 0)
			me := g.Seats[g.Turn.ActiveSeat]
			if got := effectstest.B27CountNamed(g, me.ID, "Kor Ally"); got != tc.want {
				t.Errorf("%d Kor Allies for %s, want %d", got, tc.mana, tc.want)
			}
		})
	}
}

func TestOldB310RoilmagesTrickShrinksOnlyOpponentsAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 3, 3)
	hand := me.Hand.Size()
	oldB310Cast(t, g, game.Card{Name: "Roilmage's Trick", TypeLine: "Instant", ManaCost: "{3}{U}", OracleID: oldB310Roilmage}, "UWBR", nil, 0)
	if got := effectstest.EffectivePower(t, g, theirs); got != -1 {
		t.Errorf("their power %d, want -1 (3 - 4 colours)", got)
	}
	if got := effectstest.EffectivePower(t, g, mine); got != 3 {
		t.Errorf("my power %d, want 3", got)
	}
	if got := me.Hand.Size() - hand; got != 1 { // the spell came and went; the draw added one
		t.Errorf("hand changed by %d, want +1", got)
	}
}

func TestOldB310BrilliantSpectrumDrawsPerColourThenDiscardsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "UWGC")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Brilliant Spectrum", TypeLine: "Sorcery", ManaCost: "{3}{U}", OracleID: oldB310Spectrum, Owner: me.ID, Controller: me.ID})
	})
	hand := me.Hand.Size() - 1 // the spell leaves the hand when cast
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardOwed(g, me.ID) != 2 {
		t.Fatalf("owes %d discards, want 2", effectstest.DiscardOwed(g, me.ID))
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand changed by %d, want +1 (draw 3, discard 2)", got)
	}
}

func TestOldB310InfuseWithTheElementsCountersAndTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	oldB310Cast(t, g, game.Card{Name: "Infuse with the Elements", TypeLine: "Instant", ManaCost: "{3}{G}", OracleID: oldB310Infuse}, "GWUB",
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}}, 0)
	if n := effectstest.CounterCount(g, bear, game.CounterPlusOne); n != 4 {
		t.Errorf("%d +1/+1 counters, want 4", n)
	}
	if !effectstest.P1g2HasKeyword(g, bear, "trample") {
		t.Error("the creature did not gain trample")
	}
}

func TestOldB310SkyriderElfEntersWithACounterPerColour(t *testing.T) {
	for _, tc := range []struct {
		mana string
		x    int
		want int
	}{{"GU", 0, 2}, {"GUWB", 2, 4}, {"GUCC", 2, 2}} {
		t.Run(tc.mana, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := oldB310Cast(t, g, game.Card{Name: "Skyrider Elf", TypeLine: "Creature — Elf Warrior Ally", ManaCost: "{X}{G}{U}", OracleID: oldB310SkyriderElf}, tc.mana, nil, tc.x)
			if n := effectstest.CounterCount(g, id, game.CounterPlusOne); n != tc.want {
				t.Errorf("%d counters for %s, want %d", n, tc.mana, tc.want)
			}
		})
	}
}

func TestOldB310PrismArrayEntersWithCrystalsThatTapCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	array := oldB310Cast(t, g, game.Card{Name: "Prism Array", TypeLine: "Enchantment", ManaCost: "{4}{U}", OracleID: oldB310PrismArray}, "UWBRG", nil, 0)
	if n := effectstest.CounterCount(g, array, effects.S82OldB310CrystalCounter); n != 5 {
		t.Fatalf("%d crystal counters, want 5", n)
	}
	effectstest.B16Activate(t, g, me.ID, array, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, victim) {
		t.Error("the target was not tapped")
	}
	if n := effectstest.CounterCount(g, array, effects.S82OldB310CrystalCounter); n != 4 {
		t.Errorf("%d crystal counters after the activation, want 4", n)
	}
}

func TestOldB310PrismArrayScryNeedsAllFiveColours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pa := effectstest.PushCatalogPermanent(g, me.ID, "Prism Array", "Enchantment", oldB310PrismArray, false)
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.FloatForTest(g, me, "WUBRG")
	effectstest.B16Activate(t, g, me.ID, pa, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.ScryChoiceFor(g, me.ID); c == nil || len(c.ScryCards) != 3 {
		t.Fatalf("activating with WUBRG did not offer a scry 3: %+v", c)
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
}

func TestOldB310ExertInfluenceStealsOnlyAtOrBelowTheColoursSpent(t *testing.T) {
	for _, tc := range []struct {
		mana  string
		power int
		steal bool
	}{{"UWBCC", 3, true}, {"UWBCC", 4, false}, {"UCCCC", 1, true}, {"UCCCC", 2, false}} {
		t.Run(tc.mana, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			target := effectstest.B12Creature(g, opp.ID, "Target", "Creature — Bear", tc.power, 2)
			oldB310Cast(t, g, game.Card{Name: "Exert Influence", TypeLine: "Sorcery", ManaCost: "{4}{U}", OracleID: oldB310ExertInfl}, tc.mana,
				[]game.TargetRef{{Kind: game.TargetCard, ID: target}}, 0)
			c, ok := effectstest.BattlefieldCard(g, target)
			if !ok {
				t.Fatal("the target left the battlefield")
			}
			if got := c.Controller == me.ID; got != tc.steal {
				t.Errorf("stolen = %t, want %t", got, tc.steal)
			}
		})
	}
}
