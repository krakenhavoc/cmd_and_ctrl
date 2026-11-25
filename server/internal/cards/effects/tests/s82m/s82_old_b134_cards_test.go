package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Theros Beyond Death batch 134 (S82 OLD, #5514).

const (
	oldB134Tymaret    = "1d09ebb7-c4c9-4bb7-b6bf-3c8330b954e8"
	oldB134Elspeth    = "40c88392-2aa7-4e12-86a4-8d680219e670"
	oldB134Kiora      = "84d0c616-2fce-47f4-bfdd-783741a63fd2"
	oldB134Medomai    = "e2ea1901-68ac-4afc-adc2-ac6bc7465311"
	oldB134Anax       = "eea6ba94-8755-45c0-94ea-cde68e4a1e7c"
	oldB134Manticore  = "1ecb3349-ccb8-4b54-bd81-8d649d6f96f0"
	oldB134Chimera    = "5cc57770-1070-45f6-9d76-d402c55a0d06"
	oldB134Warden     = "7d9b7fca-9514-4aea-be51-b6c8debb3de4"
	oldB134Wrap       = "b46efd59-eda7-4fe6-ace7-67ca8f6c088d"
	oldB134Purphoros  = "e190f086-ebee-447f-907f-7b37ec8824ac"
	oldB134Saga       = "Enchantment — Saga"
	oldB134Elemental  = "Elemental"
	oldB134Labyrinth  = "9ec5a487-d8ed-459a-8f58-56f6e9a2dfe8"
	oldB134SomeSorcer = "Sorcery"
)

func TestOldB134CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB134Tymaret: "Tymaret Calls the Dead", oldB134Elspeth: "Elspeth Conquers Death", oldB134Kiora: "Kiora Bests the Sea God",
		oldB134Medomai: "Medomai's Prophecy", oldB134Anax: "The Triumph of Anax", oldB134Manticore: "Dreamstalker Manticore",
		oldB134Chimera: "Mischievous Chimera", oldB134Warden: "Skophos Maze-Warden", oldB134Wrap: "Wrap in Flames",
		oldB134Purphoros: "Purphoros's Intervention",
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

func oldB134Zombies(g *game.Game, owner uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.HasSubtype("Zombie") {
			n++
		}
	}
	return n
}

// --- Tymaret Calls the Dead ---------------------------------------------------

func TestOldB134TymaretMillsThenExilesACreatureOrEnchantmentForAZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dead := effectstest.PushGraveyardCreature(g, me.ID, "Dead Bear", "{1}{G}")
	effectstest.CastCatalogSpell(t, g, "Tymaret Calls the Dead", oldB134Saga, oldB134Tymaret, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCards(g, me.ID) == nil {
		t.Fatal("chapter I offered no exile")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.AdZone(g, dead) != game.ZoneExile {
		t.Errorf("the creature card is in zone %v, want exile", effectstest.AdZone(g, dead))
	}
	if n := oldB134Zombies(g, me.ID); n != 1 {
		t.Errorf("%d Zombies, want 1", n)
	}
}

func TestOldB134TymaretDecliningTheExileMakesNoZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushGraveyardCreature(g, me.ID, "Dead Bear", "{1}{G}")
	effectstest.CastCatalogSpell(t, g, "Tymaret Calls the Dead", oldB134Saga, oldB134Tymaret, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB134Zombies(g, me.ID); n != 0 {
		t.Errorf("%d Zombies, want 0", n)
	}
}

func TestOldB134TymaretChapterThreeGainsAndScriesByZombies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, me.ID, "Zombie A", "Creature — Zombie", 2, 2)
	effectstest.B12Creature(g, me.ID, "Zombie B", "Creature — Zombie", 2, 2)
	saga := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Tymaret Calls the Dead", OracleID: oldB134Tymaret, TypeLine: oldB134Saga, Counters: map[string]int{game.CounterLore: 2}})
	before := me.Life
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.AnswerScryIfAsked(t, g, me)
	if got := me.Life - before; got != 2 {
		t.Errorf("gained %d life, want 2", got)
	}
	_ = saga
}

// --- Elspeth Conquers Death ---------------------------------------------------

func TestOldB134ElspethChapterOneExilesAPermanentOfManaValueThreeOrMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Big Thing", TypeLine: "Artifact", ManaCost: "{3}"})
	small := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Small Thing", TypeLine: "Artifact", ManaCost: "{2}"})
	effectstest.CastCatalogSpell(t, g, "Elspeth Conquers Death", oldB134Saga, oldB134Elspeth, nil)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("chapter I asked for no target")
	}
	if effectstest.HasID(p.PickTargetCards, small) || !effectstest.HasID(p.PickTargetCards, big) {
		t.Errorf("chapter I targets %v: mana value 3 or greater only", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, big)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the target was not exiled")
	}
}

func TestOldB134ElspethChapterTwoTaxesOpponentsNoncreatureSpellsUntilYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Elspeth Conquers Death", OracleID: oldB134Elspeth, TypeLine: oldB134Saga, Counters: map[string]int{game.CounterLore: 1}})
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	price := func(typeLine string, who uuid.UUID) string {
		card := game.Card{InstanceID: uuid.New(), Name: "Spell", TypeLine: typeLine, ManaCost: "{1}{R}", Owner: who, Controller: who}
		base, _ := game.ParseCost("{1}{R}")
		out, err := g.ApplyCostModifiers(base, game.CostQuery{Card: card, Controller: who, FromZone: game.ZoneHand})
		if err != nil {
			t.Fatalf("ApplyCostModifiers: %v", err)
		}
		return out.String()
	}
	if got := price("Sorcery", opp.ID); got != "{3}{R}" {
		t.Errorf("an opponent's sorcery costs %q, want {3}{R}", got)
	}
	if got := price("Creature — Bear", opp.ID); got != "{1}{R}" {
		t.Errorf("an opponent's creature spell costs %q, want {1}{R}", got)
	}
	if got := price("Sorcery", me.ID); got != "{1}{R}" {
		t.Errorf("your own sorcery costs %q, want {1}{R}", got)
	}
	// Until your next turn: gone once your next turn begins.
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	if got := price("Sorcery", opp.ID); got != "{1}{R}" {
		t.Errorf("after your next turn begins an opponent's sorcery costs %q, want {1}{R}", got)
	}
}

func TestOldB134ElspethChapterThreeReturnsACreatureWithAPlusOneCounterAndAPlaneswalkerWithLoyalty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushGraveyardCreature(g, me.ID, "Dead Bear", "{1}{G}")
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Elspeth Conquers Death", OracleID: oldB134Elspeth, TypeLine: oldB134Saga, Counters: map[string]int{game.CounterLore: 2}})
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, bear)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var counters map[string]int
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == bear {
			counters = c.Counters
		}
	}
	if effectstest.AdZone(g, bear) != game.ZoneBattlefield || counters[game.CounterPlusOne] != 1 {
		t.Errorf("the creature is in zone %v with counters %v, want the battlefield with one +1/+1 counter", effectstest.AdZone(g, bear), counters)
	}
}

// --- Kiora Bests the Sea God --------------------------------------------------

func TestOldB134KioraMakesAKrakenTapsAndFreezesThenStealsAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Creature", "Creature — Giant", 3, 3)
	land := effectstest.B12Permanent(g, opp.ID, "Their Land", "Land")
	effectstest.CastCatalogSpell(t, g, "Kiora Bests the Sea God", oldB134Saga, oldB134Kiora, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var kraken *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Kraken" {
			kraken = &g.Battlefield.Cards[i]
		}
	}
	if kraken == nil || kraken.Power != 8 || kraken.Toughness != 8 || !game.HasKeyword(kraken, "hexproof") || kraken.Controller != me.ID {
		t.Fatalf("chapter I made %+v, want an 8/8 blue Kraken with hexproof under your control", kraken)
	}
	// Chapter II: tap all nonland permanents target opponent controls.
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatalf("pick the opponent: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == theirs && !c.Tapped {
			t.Error("their creature was not tapped")
		}
		if c.InstanceID == land && c.Tapped {
			t.Error("their land was tapped")
		}
	}
	// Chapter III: gain control of target permanent an opponent controls. Untap it.
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, theirs)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == theirs && c.Controller != me.ID {
			t.Errorf("their creature is controlled by %v, want you", c.Controller)
		}
	}
}

// --- Medomai's Prophecy -------------------------------------------------------

func TestOldB134MedomaisProphecyChoosesANameThenDrawsTwiceOnTheFirstCastWithIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	saga := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Medomai's Prophecy", OracleID: oldB134Medomai, TypeLine: oldB134Saga, Counters: map[string]int{game.CounterLore: 1}})
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ActNameFor(t, g, me.ID, "Lightning Bolt")
	if got := g.ChosenNameOf(saga); got != "Lightning Bolt" {
		t.Fatalf("the Saga chose %q", got)
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	before := effectstest.HandSize(me)
	effectstest.FloatForTest(g, me, "{R}")
	effectstest.CastFromHandForTest(t, g, me, "Lightning Bolt", "Instant", "{R}", "", game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandSize(me) - before; got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
	// Only the first time: a second Lightning Bolt draws nothing.
	before = effectstest.HandSize(me)
	effectstest.FloatForTest(g, me, "{R}")
	effectstest.CastFromHandForTest(t, g, me, "Lightning Bolt", "Instant", "{R}", "", game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandSize(me) - before; got != 0 {
		t.Errorf("a second cast drew %d cards, want 0", got)
	}
}

// --- The Triumph of Anax ------------------------------------------------------

func TestOldB134TriumphOfAnaxPumpsByLoreCountersAndFightsInChapterFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	_ = effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 5)
	effectstest.CastCatalogSpell(t, g, "The Triumph of Anax", oldB134Saga, oldB134Anax, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, mine); p != 3 {
		t.Errorf("chapter I: power %d, want 3 (+1/+0 for one lore counter)", p)
	}
	c := effectstest.GaLayered(t, g, mine)
	if !game.HasKeyword(&c, "trample") {
		t.Error("chapter I: no trample")
	}
}

func TestOldB134TriumphOfAnaxChapterFourFightsUpToOneCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 5)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 5)
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "The Triumph of Anax", OracleID: oldB134Anax, TypeLine: oldB134Saga, Counters: map[string]int{game.CounterLore: 3}})
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, mine)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if v := effectstest.GaLayered(t, g, theirs); v.DamageMarked != 2 {
		t.Errorf("chapter IV: their creature has %d damage, want 2 (Mine's power)", v.DamageMarked)
	}
	if v := effectstest.GaLayered(t, g, mine); v.DamageMarked != 2 {
		t.Errorf("chapter IV: your creature has %d damage, want 2 (Theirs's power)", v.DamageMarked)
	}
}

// --- Dreamstalker Manticore and Mischievous Chimera --------------------------

func oldB134OpponentsTurnCast(t *testing.T, g *game.Game, caster *game.Player) {
	t.Helper()
	// Walk to the next player's turn and cast an instant from the caster's hand.
	effectstest.AdvanceToPrecombatMainOf(t, g, 1)
	effectstest.FloatForTest(g, caster, "{R}")
	effectstest.CastFromHandForTest(t, g, caster, "Shock", "Instant", "{R}", "", game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}})
}

func TestOldB134ManticorePingsOnTheFirstSpellDuringAnOpponentsTurnOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Dreamstalker Manticore", OracleID: oldB134Manticore, TypeLine: "Enchantment Creature — Manticore", ManaCost: "{2}{R}", Power: 4, Toughness: 2})
	effectstest.AdvanceToPrecombatMainOf(t, g, 1)
	effectstest.FloatForTest(g, me, "{R}")
	effectstest.CastFromHandForTest(t, g, me, "Shock", "Instant", "{R}", "", game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if p := effectstest.LatestPickTarget(g, me.ID); p == nil {
		t.Fatal("the first spell on an opponent's turn asked for no target")
	} else if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("pick the player: %v", err)
	}
	life := opp.Life
	effectstest.PassPriorityAroundTable(t, g)
	if life-opp.Life < 1 {
		t.Errorf("the opponent lost %d life, want at least the Manticore's 1", life-opp.Life)
	}
	// A second spell the same turn does not trigger it.
	effectstest.FloatForTest(g, me, "{R}")
	effectstest.CastFromHandForTest(t, g, me, "Shock", "Instant", "{R}", "", game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Error("the second spell triggered the Manticore")
	}
}

func TestOldB134ChimeraDamagesEachOpponentAndScriesOnTheFirstSpellOnAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Mischievous Chimera", OracleID: oldB134Chimera, TypeLine: "Enchantment Creature — Chimera", ManaCost: "{U}{R}", Power: 2, Toughness: 2})
	effectstest.AdvanceToPrecombatMainOf(t, g, 1)
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	effectstest.FloatForTest(g, me, "{R}")
	effectstest.CastFromHandForTest(t, g, me, "Shock", "Instant", "{R}", "", game.CastSpellParams{})
	effectstest.AnswerScryIfAsked(t, g, me)
	for _, s := range g.Seats {
		if s.ID == me.ID {
			continue
		}
		if got := lives[s.ID] - s.Life; got != 1 {
			t.Errorf("seat %v lost %d life, want 1", s.ID, got)
		}
	}
	_ = opp
}

// --- Skophos Maze-Warden ------------------------------------------------------

func TestOldB134MazeWardenPumpsItselfAndFightsATargetedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warden := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Skophos Maze-Warden", OracleID: oldB134Warden, TypeLine: "Creature — Minotaur Warrior", ManaCost: "{3}{R}", Power: 3, Toughness: 4})
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, warden, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate +1/-1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, warden); p != 4 || tt != 3 {
		t.Errorf("the Warden is %d/%d, want 4/3", p, tt)
	}
	// An activated ability of a land named Labyrinth of Skophos that you control
	// targets another creature: you may have the Warden fight it.
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 1, 10)
	lab := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Labyrinth of Skophos", OracleID: oldB134Labyrinth, TypeLine: "Land"})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBecomesTarget, Actor: me.ID, Source: lab, Target: victim, CardID: victim, TargeterKind: game.StackItemActivated})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ResolveTriggerPrompt(effectstest.ConfirmForTriggerPrompt(g, me.ID).ID, me.ID, true); err != nil {
		t.Fatalf("answer the trigger prompt: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if v := effectstest.GaLayered(t, g, victim); v.DamageMarked != 4 {
		t.Errorf("the victim has %d damage, want 4 (the 4/3 Warden's power)", v.DamageMarked)
	}
	// A spell that targets does not trigger it, and neither does an ability of another land.
	other := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Other Land", TypeLine: "Land"})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBecomesTarget, Actor: me.ID, Source: other, Target: victim, CardID: victim, TargeterKind: game.StackItemActivated})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ConfirmForTriggerPrompt(g, me.ID) != nil {
		t.Error("an ability of another land triggered the Warden")
	}
}

// --- Wrap in Flames -----------------------------------------------------------

func TestOldB134WrapInFlamesDamagesEachTargetAndStopsThemBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 3)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 3)
	effectstest.CastCatalogSpell(t, g, "Wrap in Flames", oldB134SomeSorcer, oldB134Wrap, []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		c := effectstest.GaLayered(t, g, id)
		if c.DamageMarked != 1 {
			t.Errorf("a target has %d damage, want 1", c.DamageMarked)
		}
		if effectstest.AuraRestrictions(t, g, id)&game.CantBlock == 0 {
			t.Error("a target can still block")
		}
	}
	_ = me
}

// --- Purphoros's Intervention -------------------------------------------------

// oldB134CastPurphoros casts Purphoros's Intervention with X=2 in the chosen mode.
func oldB134CastPurphoros(t *testing.T, g *game.Game, mode int, targets []game.TargetRef) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.RfdHandCard(me, "Purphoros's Intervention", "Sorcery", "{X}{R}", oldB134Purphoros, 0, 0)
	effectstest.ToMainForCost(t, g)
	effectstest.FloatForTest(g, me, "{R}{R}{R}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{mode}, Targets: targets, XValue: 2}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB134PurphorossInterventionMakesAnXOneElementalThatIsSacrificedAtEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldB134CastPurphoros(t, g, 0, nil)
	var tok *game.Card
	var tokID uuid.UUID
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == oldB134Elemental {
			tok, tokID = &g.Battlefield.Cards[i], g.Battlefield.Cards[i].InstanceID
		}
	}
	if tok == nil {
		t.Fatal("no Elemental token was made")
	}
	if p, tt := effectstest.PtOf(t, g, tokID); p != 2 || tt != 1 || tok.Controller != me.ID {
		t.Errorf("the Elemental is %d/%d, want 2/1 under your control", p, tt)
	}
	c := effectstest.GaLayered(t, g, tokID)
	if !game.HasKeyword(&c, "trample") || !game.HasKeyword(&c, "haste") {
		t.Error("the Elemental lacks trample or haste")
	}
	for i := 0; i < 30 && g.Turn.Step != game.StepEnd; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tokID) {
		t.Error("the Elemental was not sacrificed at the beginning of the end step")
	}
}

func TestOldB134PurphorossInterventionDealsTwiceXToACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Giant", 3, 10)
	oldB134CastPurphoros(t, g, 1, []game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	c := effectstest.GaLayered(t, g, victim)
	if c.DamageMarked != 4 {
		t.Errorf("the creature has %d damage, want 4 (twice X=2)", c.DamageMarked)
	}
}
