package testsso

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b13_cards_test.go — Secrets of Strixhaven Commander, thirteenth batch
// (#4834). Each test plays the real catalog entry through the engine and
// asserts what the printed text says.

const (
	socB13ShieldedOracle    = "79085dea-8daa-41c0-a975-61e3e6da2e26"
	socB13SilverquillOracle = "2b65eb80-6fb7-429f-81f2-2fe125eba634"
	socB13SongbirdsOracle   = "2598294c-a9a7-4bef-a562-23f297a80536"
	socB13SpiritOracle      = "ddc85b23-113a-46e9-ab09-fae9312c72fa"
	socB13SqueeOracle       = "0ef760f4-bbec-4735-8d60-fc24d973452a"
	socB13MysticOracle      = "2975590d-6a5b-41e5-8a3b-e9325e71226a"
	socB13SanguinistOracle  = "25877c41-39a4-4cc3-ac4b-8f3dd06d579b"
	socB13SerpentOracle     = "78124cd1-2658-43f3-a339-0755d84154d2"
	socB13ShotcallerOracle  = "10389ff7-2ea4-4413-90cc-0e3ca268c64d"
	socB13StudyHallOracle   = "eb735501-19e7-4910-aa6a-6667fff6f4e5"
)

func TestSocB13EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB13ShieldedOracle:          "Shielded by Faith",
		socB13SilverquillOracle:       "Silverquill Campus",
		socB13SongbirdsOracle:         "Songbirds' Blessing",
		socB13SpiritOracle:            "Spirit of Resilience",
		socB13SqueeOracle:             "Squee, Goblin Nabob",
		socB13MysticOracle:            "Starfield Mystic",
		socB13SanguinistOracle:        "Stensian Sanguinist",
		socB13SanguinistOracle + "#1": "Exsanguinate",
		socB13SerpentOracle:           "Stonecoil Serpent",
		socB13ShotcallerOracle:        "Striding Shotcaller",
		socB13ShotcallerOracle + "#1": "Run the Play",
		socB13StudyHallOracle:         "Study Hall",
	}
	for oracle, name := range want {
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

// --- Silverquill Campus ---------------------------------------------------

func TestSocB13SilverquillCampusEntersTappedMakesWOrBAndScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "Top", "Next")
	id := effectstest.PlayLandFromHand(t, g, "Silverquill Campus", socB13SilverquillOracle)
	effectstest.PassPriorityAroundTable(t, g)
	card, ok := effectstest.BattlefieldCard(g, id)
	if !ok || !card.Tapped {
		t.Fatalf("on battlefield %v, tapped %v; want it to enter tapped", ok, card.Tapped)
	}
	effectstest.UntapForTest(g, id)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err == nil {
		t.Fatal("the land made {R}")
	}
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"B"}}); err != nil {
		t.Fatalf("{T}: Add {B}: %v", err)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.UntapForTest(g, id)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"W"}}); err != nil {
		t.Fatalf("{T}: Add {W}: %v", err)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })

	effectstest.UntapForTest(g, id)
	effectstest.FloatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("scry activated with three mana")
	}
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{4},{T}: scry 1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("scry prompt %v, want one looking at one card", c)
	}
	top := c.ScryCards[0]
	if err := g.ResolveScry(c.ID, me.ID, []uuid.UUID{top}, nil); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	if bottom, err := me.Library.Bottom(); err != nil || bottom.InstanceID != top {
		t.Error("the scried card is not on the bottom")
	}
}

// --- Squee, Goblin Nabob --------------------------------------------------

func socB13Squee(owner *game.Player) uuid.UUID {
	id := uuid.New()
	owner.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: "Squee, Goblin Nabob", TypeLine: "Legendary Creature — Goblin",
		ManaCost: "{2}{R}", OracleID: socB13SqueeOracle, Power: 1, Toughness: 1, Owner: owner.ID, Controller: owner.ID,
	})
	return id
}

func TestSocB13SqueReturnsFromTheGraveyardAtYourUpkeepIfYouChoose(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		squee := socB13Squee(me)
		effectstest.AdvanceToUpkeepOf(t, g, 0)
		if yes {
			// An earlier upkeep may already have passed; make sure this is the
			// upkeep the trigger is raised in.
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
			effectstest.PassPriorityAroundTable(t, g)
			if !me.Hand.Contains(squee) || me.Graveyard.Contains(squee) {
				t.Fatalf("Squee in hand %v / graveyard %v after saying yes", me.Hand.Contains(squee), me.Graveyard.Contains(squee))
			}
			continue
		}
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Hand.Contains(squee) || !me.Graveyard.Contains(squee) {
			t.Fatalf("Squee left the graveyard after saying no")
		}
	}
}

func TestSocB13SqueDoesNotTriggerOnTheBattlefieldOrAtAnOpponentsUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	onBattlefield := effectstest.PushCatalogPermanent(g, me.ID, "Squee, Goblin Nabob", "Legendary Creature — Goblin", socB13SqueeOracle, false)
	squee := socB13Squee(me)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if n := effectstest.TriggersOnStackFrom(g, squee) + effectstest.TriggersOnStackFrom(g, onBattlefield); n != 0 {
		t.Fatalf("%d Squee triggers at the opponent's upkeep", n)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if n := effectstest.TriggersOnStackFrom(g, onBattlefield); n != 0 {
		t.Errorf("the battlefield Squee triggered %d times", n)
	}
	if n := effectstest.TriggersOnStackFrom(g, squee) + len(g.PendingChoices); n == 0 {
		t.Error("the graveyard Squee did not trigger at its owner's upkeep")
	}
}

// --- Starfield Mystic -----------------------------------------------------

func TestSocB13StarfieldMysticDiscountsOnlyYourEnchantmentSpells(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		cost, pool     string
		wantErr        bool
	}{
		{"enchantment is {1} cheaper", "Enchantment", "{1}{W}", "W", false},
		{"enchantment creature is {1} cheaper", "Enchantment Creature — Spirit", "{1}{W}", "W", false},
		{"enchantment at full price", "Enchantment", "{1}{W}", "", true},
		{"a creature is not discounted", "Creature — Bear", "{1}{G}", "G", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.PushCatalogPermanent(g, me.ID, "Starfield Mystic", "Creature — Human Cleric", socB13MysticOracle, false)
			effectstest.AdvanceToMain(t, g)
			effectstest.FloatForTest(g, me, tc.pool)
			id := uuid.New()
			g.WithWriteLock(func() {
				me.Hand.PushTop(game.Card{InstanceID: id, Name: "Test spell", TypeLine: tc.typeLine, ManaCost: tc.cost, Owner: me.ID, Controller: me.ID})
			})
			err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
			if (err != nil) != tc.wantErr {
				t.Fatalf("cast err = %v, want error %v", err, tc.wantErr)
			}
		})
	}
}

func TestSocB13StarfieldMysticDoesNotDiscountAnOpponentsSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Starfield Mystic", "Creature — Human Cleric", socB13MysticOracle, false)
	effectstest.AdvanceToStepInTurn(t, g, game.StepPrecombatMain)
	_ = me
	// Walk to the opponent's main phase and try to cast an enchantment for {W}.
	for i := 0; i < 80 && g.Turn.ActiveSeat != 1; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, opp, "W")
	id := uuid.New()
	g.WithWriteLock(func() {
		opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Test spell", TypeLine: "Enchantment", ManaCost: "{1}{W}", Owner: opp.ID, Controller: opp.ID})
	})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("the opponent's enchantment was discounted by my Mystic")
	}
}

func TestSocB13StarfieldMysticGrowsWhenYourEnchantmentsGoToTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mystic := effectstest.PushCatalogPermanent(g, me.ID, "Starfield Mystic", "Creature — Human Cleric", socB13MysticOracle, false)
	mine := effectstest.B12Permanent(g, me.ID, "My Enchantment", "Enchantment")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	destroy := func(id uuid.UUID) {
		g.WithWriteLock(func() {
			if err := g.DestroyPermanentForEffect(id); err != nil {
				t.Fatalf("destroy: %v", err)
			}
		})
		g.RunStateChecksForTest()
		effectstest.PassPriorityAroundTable(t, g)
	}
	destroy(theirs)
	destroy(rock)
	destroy(bear)
	if got := effectstest.CounterCount(g, mystic, game.CounterPlusOne); got != 0 {
		t.Fatalf("counters %d after an opponent's enchantment, an artifact and a creature died, want 0", got)
	}
	destroy(mine)
	if got := effectstest.CounterCount(g, mystic, game.CounterPlusOne); got != 1 {
		t.Fatalf("counters %d after my enchantment died, want 1", got)
	}
	// An Aura that falls off with its creature is my enchantment going to the graveyard too.
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	aura := effectstest.B12Permanent(g, me.ID, "My Aura", "Enchantment — Aura")
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
	destroy(host)
	if got := effectstest.CounterCount(g, mystic, game.CounterPlusOne); got != 2 {
		t.Errorf("counters %d after the Aura went with its creature, want 2", got)
	}
}

// --- Stonecoil Serpent ----------------------------------------------------

func TestSocB13StonecoilSerpentEntersWithXCountersAndItsKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := socB8Cast(t, g, game.Card{Name: "Stonecoil Serpent", TypeLine: "Artifact Creature — Snake", ManaCost: "{X}", OracleID: socB13SerpentOracle},
		game.CastSpellParams{XValue: 3})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != 3 {
		t.Fatalf("+1/+1 counters %d, want 3", got)
	}
	if p := effectstest.EffectivePowerWithCounters(t, g, id); p != 3 {
		t.Errorf("power %d, want 3", p)
	}
	for _, kw := range []string{"reach", "trample", "protection from multicolored"} {
		if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("no %q", kw)
		}
	}
	_ = me
}

func TestSocB13StonecoilSerpentIsProtectedFromMulticoloredSourcesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	serpent := socB8Cast(t, g, game.Card{Name: "Stonecoil Serpent", TypeLine: "Artifact Creature — Snake", ManaCost: "{X}", OracleID: socB13SerpentOracle},
		game.CastSpellParams{XValue: 4})
	effectstest.PassPriorityAroundTable(t, g)
	gold := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Gold Bear", TypeLine: "Creature — Bear",
		ManaCost: "{G}{W}", Colors: []string{"G", "W"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	green := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear",
		ManaCost: "{1}{G}", Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	damage := func(src uuid.UUID) int {
		before := 0
		if c, ok := effectstest.BattlefieldCard(g, serpent); ok {
			before = c.DamageMarked
		}
		g.WithWriteLock(func() {
			if err := g.DealDamageToCreatureForEffect(src, serpent, 2); err != nil {
				t.Fatalf("damage: %v", err)
			}
		})
		c, _ := effectstest.BattlefieldCard(g, serpent)
		return c.DamageMarked - before
	}
	if got := damage(gold); got != 0 {
		t.Errorf("a multicolored source dealt %d, want 0", got)
	}
	if got := damage(green); got != 2 {
		t.Errorf("a monocolored source dealt %d, want 2", got)
	}
	_ = me
}

// --- Study Hall -----------------------------------------------------------

func socB13Commander(g *game.Game, me *game.Player, casts int) uuid.UUID {
	cmdr := uuid.New()
	g.WithWriteLock(func() {
		me.Command.PushTop(game.Card{InstanceID: cmdr, Name: "Test Commander", TypeLine: "Legendary Creature — Elf",
			ManaCost: "{G}", Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, IsCommander: true})
		me.CommanderCasts[cmdr] = casts
	})
	return cmdr
}

func TestSocB13StudyHallMakesColorlessOrFiltersAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	hall := effectstest.SeedPermanentWithOracle(g, me.ID, "Study Hall", "Land", socB13StudyHallOracle)
	effectstest.ActivateManaFor(t, g, me.ID, hall, 0, game.ManaAbilityParams{})
	if got := effectstest.GmrPoolCount(me, "C"); got != 1 {
		t.Fatalf("pool has %d {C}, want 1", got)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.UntapForTest(g, hall)
	if err := g.ActivateManaAbility(me.ID, hall, 1, game.ManaAbilityParams{Colors: []string{"U"}}); err == nil {
		t.Fatal("the filter ability worked with no {1} to pay")
	}
	effectstest.FloatForTest(g, me, "C")
	for _, color := range []string{"U", "R"} {
		effectstest.UntapForTest(g, hall)
		g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
		effectstest.FloatForTest(g, me, "C")
		effectstest.ActivateManaFor(t, g, me.ID, hall, 1, game.ManaAbilityParams{Colors: []string{color}})
		if got := effectstest.GmrPoolCount(me, color); got != 1 {
			t.Errorf("pool has %d {%s}, want 1", got, color)
		}
	}
}

func TestSocB13StudyHallScriesForTheCommandersCommandZoneCasts(t *testing.T) {
	for _, tc := range []struct {
		name  string
		casts int // before this cast
		want  int
	}{{"first cast", 0, 1}, {"third cast", 2, 3}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
			hall := effectstest.SeedPermanentWithOracle(g, me.ID, "Study Hall", "Land", socB13StudyHallOracle)
			cmdr := socB13Commander(g, me, tc.casts)
			effectstest.FloatForTest(g, me, "C")
			effectstest.ActivateManaFor(t, g, me.ID, hall, 1, game.ManaAbilityParams{Colors: []string{"G"}})
			effectstest.FloatForTest(g, me, strings.Repeat("CC", tc.casts)) // the commander tax: {2} for each earlier cast
			if err := g.CastSpell(me.ID, cmdr, game.CastSpellParams{FromZone: "command", Strict: true}); err != nil {
				t.Fatalf("cast the commander: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g) // the scry trigger resolves
			c := effectstest.ScryChoiceFor(g, me.ID)
			if c == nil || len(c.ScryCards) != tc.want {
				t.Fatalf("scry prompt %+v, want a look at %d cards", c, tc.want)
			}
		})
	}
}

func TestSocB13StudyHallDoesNotScryForAnOrdinarySpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	hall := effectstest.SeedPermanentWithOracle(g, me.ID, "Study Hall", "Land", socB13StudyHallOracle)
	socB13Commander(g, me, 2)
	effectstest.FloatForTest(g, me, "C")
	effectstest.ActivateManaFor(t, g, me.ID, hall, 1, game.ManaAbilityParams{Colors: []string{"G"}})
	effectstest.CastFromHandForTest(t, g, me, "Bear", "Creature — Bear", "{G}", "", game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.ScryChoiceFor(g, me.ID); c != nil {
		t.Errorf("an ordinary spell paid with Study Hall's mana scried: %+v", c)
	}
}
