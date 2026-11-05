package effects

import (
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b27_cards_test.go — Star Trek, batch 27 (#4973). Each test plays the
// real catalog entry through the engine and asserts what its text says.

const (
	trekB27SpecialistOracle = "9fc49a69-f80c-478f-b237-66f155c0c09a"
	trekB27DS9Oracle        = "14442f5b-0bad-4c58-b268-d775787571d8"
	trekB27IsolinearOracle  = "ef5d7b38-b782-4fb0-b6cd-3da36995ab39"
	trekB27ReplicatorOracle = "a91b8eb4-d4f7-4e24-ae49-1f51e610c7ca"
	trekB27SickbayOracle    = "1d4163d9-ca02-4f25-9f53-01f6d7cc4b70"
	trekB27TeaOracle        = "6f2e8083-29af-4cdb-a2cb-216b2f3f5616"
	trekB27TricorderOracle  = "1385679f-bdcd-4d7e-8266-1df0e36cc2da"
	trekB27PhaserOracle     = "9d4d3682-e904-412d-b8ae-734962865dfb"
	trekB27CerritosOracle   = "472eac8f-2b82-4fa9-a40c-af11ec40f19e"
)

func TestTrekB27EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB27SpecialistOracle: "Cybernetic Specialist",
		trekB27DS9Oracle:        "Deep Space 9",
		trekB27IsolinearOracle:  "Isolinear Processor",
		trekB27ReplicatorOracle: "Replicator",
		trekB27SickbayOracle:    "Sickbay",
		trekB27TeaOracle:        "Tea, Earl Grey, Hot",
		trekB27TricorderOracle:  "Tricorder",
		trekB27PhaserOracle:     "Type-2 Phaser",
		trekB27CerritosOracle:   "U.S.S. Cerritos, Following Up",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- shared test helpers ------------------------------------------------------

// trekB27Cast puts a card with a mana cost in p's hand, floats that cost,
// casts it strictly and resolves the spell.
func trekB27Cast(t *testing.T, g *game.Game, p *game.Player, name, typeLine, cost, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle,
			Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID,
		})
	})
	floatMana(t, g, p, trekB27Colorless(cost))
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	return id
}

// trekB27Colorless turns a generic cost such as "{3}" into that many {C}
// for floatMana.
func trekB27Colorless(cost string) string {
	n, err := strconv.Atoi(strings.Trim(cost, "{}"))
	if err != nil {
		return cost
	}
	return strings.Repeat("{C}", n)
}

// trekB27Activate activates ability `idx` of a catalog permanent strictly.
func trekB27Activate(g *game.Game, p *game.Player, src uuid.UUID, idx int, params game.ActivateAbilityParams) error {
	params.Strict = true
	return g.ActivateCatalogAbility(p.ID, src, idx, params)
}

// trekB27Surveil answers the open surveil prompt for p, binning the card
// when bin is set, and returns the card looked at.
func trekB27Surveil(t *testing.T, g *game.Game, p *game.Player, bin bool) uuid.UUID {
	t.Helper()
	c := surveilChoiceFor(g, p.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if len(c.ScryCards) != 1 {
		t.Fatalf("looked at %d cards, want 1 (surveil 1)", len(c.ScryCards))
	}
	looked := c.ScryCards[0]
	var err error
	if bin {
		err = g.ResolveSurveil(c.ID, p.ID, c.ScryCards, nil)
	} else {
		err = g.ResolveSurveil(c.ID, p.ID, nil, c.ScryCards)
	}
	if err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	return looked
}

func trekB27InZone(g *game.Game, id uuid.UUID, kind game.ZoneKind) bool {
	var ok bool
	g.ReadSnapshot(func() {
		z := g.FindCardZoneForEffect(id)
		ok = z != nil && z.Kind == kind
	})
	return ok
}

func trekB27TokensNamed(g *game.Game, controller uuid.UUID, name string) []game.Card {
	var out []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == controller && c.Name == name && c.IsToken() {
				out = append(out, c)
			}
		}
	})
	return out
}

// trekB27AddCharge puts n charge counters on a permanent through the
// effect API, as station would.
func trekB27AddCharge(t *testing.T, g *game.Game, id uuid.UUID, n int) {
	t.Helper()
	var err error
	g.WithWriteLock(func() { err = g.AddCounterForEffect(id, game.CounterCharge, n) })
	if err != nil {
		t.Fatalf("AddCounterForEffect: %v", err)
	}
}

// --- Cybernetic Specialist ---------------------------------------------------

func TestTrekB27SpecialistSurveilsOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Fodder")
	trekB27Cast(t, g, me, "Cybernetic Specialist", "Artifact Creature — Human Scientist", "{2}", trekB27SpecialistOracle, 2, 1)
	looked := trekB27Surveil(t, g, me, true)
	if !trekB27InZone(g, looked, game.ZoneGraveyard) {
		t.Error("the surveilled card was not binned")
	}
}

func TestTrekB27SpecialistTucksACardFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spec := pushCatalogPermanent(g, me.ID, "Cybernetic Specialist", "Artifact Creature — Human Scientist", trekB27SpecialistOracle, false)
	mine, theirs := uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: mine, Name: "Mine", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Theirs", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	})
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}")
	if err := trekB27Activate(g, me, spec, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err == nil {
		t.Fatal("a card in an opponent's graveyard was a legal target")
	}
	if err := trekB27Activate(g, me, spec, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if bottom, _ := me.Library.Bottom(); bottom.InstanceID != mine {
		t.Errorf("library bottom is %q, want the tucked card", bottom.Name)
	}
	if !layeredCard(t, g, spec).Tapped {
		t.Error("the {T} cost did not tap the Specialist")
	}
}

func TestTrekB27SpecialistCantTapTheTurnItArrives(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spec := pushCatalogPermanent(g, me.ID, "Cybernetic Specialist", "Artifact Creature — Human Scientist", trekB27SpecialistOracle, true)
	mine := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: mine, Name: "Mine", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	})
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}")
	if err := trekB27Activate(g, me, spec, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}},
	}); err == nil {
		t.Fatal("a summoning-sick Specialist paid {T}")
	}
}

// --- Deep Space 9 ---------------------------------------------------------------

func TestTrekB27DS9PutsThePermanentCardsOntoTheBattlefieldOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		// Pushed top-down in reverse: the Fifth ends up under the four.
		for _, c := range []game.Card{
			{InstanceID: uuid.New(), Name: "Fifth", TypeLine: "Creature — Bear", Power: 1, Toughness: 1},
			{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact"},
			{InstanceID: uuid.New(), Name: "Spell", TypeLine: "Sorcery"},
			{InstanceID: uuid.New(), Name: "Wastes", TypeLine: "Basic Land"},
			{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
		} {
			c.Owner, c.Controller = me.ID, me.ID
			me.Library.PushTop(c)
		}
	})
	trekB27Cast(t, g, me, "Deep Space 9", "Legendary Artifact — Spacecraft", "{9}", trekB27DS9Oracle, 0, 0)
	passPriorityAroundTable(t, g)
	// Each card is a new object after the exile (CR 400.7), so look for it
	// by name.
	onBF := map[string]game.Card{}
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			onBF[c.Name] = c
		}
	})
	for _, name := range []string{"Bear", "Wastes", "Rock"} {
		c, ok := onBF[name]
		if !ok {
			t.Errorf("%s is not on the battlefield", name)
		} else if c.Controller != me.ID {
			t.Errorf("%s entered under the wrong player", name)
		}
	}
	if _, ok := onBF["Spell"]; ok {
		t.Error("the sorcery reached the battlefield")
	}
	if _, ok := onBF["Fifth"]; ok {
		t.Error("a fifth card was taken")
	}
	exiled := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Exile.Cards {
			if c.Name == "Spell" {
				exiled++
			}
		}
	})
	if exiled != 1 {
		t.Error("the sorcery is not left in exile")
	}
}

func TestTrekB27DS9TriggersAtYourUpkeepOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Deep Space 9", "Legendary Artifact — Spacecraft", trekB27DS9Oracle, false)
	before := me.Library.Size()
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != before {
		t.Fatalf("library %d -> %d on an opponent's upkeep", before, got)
	}
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != before-4 {
		t.Errorf("library %d -> %d at your upkeep, want four exiled", before, got)
	}
}

func TestTrekB27DS9PumpsYourCreaturesAtNine(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ds9 := pushCatalogPermanent(g, me.ID, "Deep Space 9", "Legendary Artifact — Spacecraft", trekB27DS9Oracle, false)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	trekB27AddCharge(t, g, ds9, 8)
	if p := effectivePower(t, g, mine); p != 2 {
		t.Fatalf("at 8 charge counters your creature has power %d, want 2", p)
	}
	trekB27AddCharge(t, g, ds9, 1)
	if p, tg := effectivePower(t, g, mine), effectiveToughness(t, g, mine); p != 5 || tg != 5 {
		t.Errorf("at 9 your creature is %d/%d, want 5/5", p, tg)
	}
	if p := effectivePower(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature has power %d, want 2", p)
	}
	if layeredCard(t, g, ds9).IsCreature() {
		t.Error("Deep Space 9 became a creature")
	}
}

func TestTrekB27DS9Stations(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ds9 := pushCatalogPermanent(g, me.ID, "Deep Space 9", "Legendary Artifact — Spacecraft", trekB27DS9Oracle, false)
	crew := b12Creature(g, me.ID, "Crew", "Creature — Giant", 9, 9)
	advanceToMain(t, g)
	station(t, g, me, ds9, crew)
	passPriorityAroundTable(t, g)
	if got := counterOn(g, ds9, game.CounterCharge); got != 9 {
		t.Fatalf("charge counters = %d, want 9", got)
	}
	if p := effectivePower(t, g, crew); p != 12 {
		t.Errorf("the crew has power %d at 9+, want 12", p)
	}
}

// --- Isolinear Processor ---------------------------------------------------------

func TestTrekB27IsolinearManaPaysArtifactsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	proc := pushCatalogPermanent(g, me.ID, "Isolinear Processor", "Artifact", trekB27IsolinearOracle, false)
	advanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, proc, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Processor: %v", err)
	}
	cost, _ := game.ParseCost("{1}")
	rock := game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{1}"}
	bear := game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}"}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(rock)) {
		t.Error("refused an artifact spell")
	}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(rock)) {
		t.Error("refused an ability of an artifact source")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(bear)) {
		t.Error("paid for a creature spell")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(bear)) {
		t.Error("paid for a creature's ability")
	}
	// A real cast: the restricted {C} pays for an artifact spell.
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Rock", TypeLine: "Artifact", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
	})
	bearID := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: bearID, Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, bearID, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("the Processor's mana cast a creature spell")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the Processor's mana could not cast an artifact: %v", err)
	}
}

func TestTrekB27IsolinearLoots(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	proc := pushCatalogPermanent(g, me.ID, "Isolinear Processor", "Artifact", trekB27IsolinearOracle, false)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}")
	if err := trekB27Activate(g, me, proc, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the loot was paid with one mana")
	}
	floatMana(t, g, me, "{C}")
	hand := me.Hand.Size()
	if err := trekB27Activate(g, me, proc, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("loot: %v", err)
	}
	if discards := trekB9Settle(t, g, me.ID); discards != 1 {
		t.Errorf("answered %d discard prompts, want 1", discards)
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d -> %d, want the same size after draw then discard", hand, got)
	}
	if me.Graveyard.Size() != 1 {
		t.Errorf("graveyard has %d cards, want 1", me.Graveyard.Size())
	}
}

// --- Replicator ---------------------------------------------------------------

func TestTrekB27ReplicatorMakesFoodAndCopiesItTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rep := trekB27Cast(t, g, me, "Replicator", "Artifact", "{2}", trekB27ReplicatorOracle, 0, 0)
	passPriorityAroundTable(t, g)
	foods := trekB27TokensNamed(g, me.ID, "Food")
	if len(foods) != 1 {
		t.Fatalf("%d Food tokens, want 1", len(foods))
	}
	nontoken := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	floatMana(t, g, me, "{C}{C}")
	if err := trekB27Activate(g, me, rep, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: nontoken}},
	}); err == nil {
		t.Fatal("a nontoken creature was a legal target")
	}
	if err := trekB27Activate(g, me, rep, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: foods[0].InstanceID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	foods = trekB27TokensNamed(g, me.ID, "Food")
	if len(foods) != 2 {
		t.Fatalf("%d Food tokens, want 2", len(foods))
	}
	tapped := 0
	for _, f := range foods {
		if f.Tapped {
			tapped++
		}
	}
	if tapped != 1 {
		t.Errorf("%d tapped Food tokens, want just the copy", tapped)
	}
}

func TestTrekB27ReplicatorCantCopyAnOpponentsToken(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rep := pushCatalogPermanent(g, me.ID, "Replicator", "Artifact", trekB27ReplicatorOracle, false)
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Officer", TypeLine: "Token Creature — Officer", Power: 1, Toughness: 1,
		Owner: opp.ID, Controller: opp.ID,
	})
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}")
	if err := trekB27Activate(g, me, rep, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err == nil {
		t.Fatal("an opponent's token was a legal target")
	}
}

// --- Sickbay ----------------------------------------------------------------------

func TestTrekB27SickbayGainsOnEntryAndCountsTheLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Sickbay", "Artifact", trekB27SickbayOracle, false)
	advanceToMain(t, g)
	life := me.Life
	theirs := uuid.New()
	var err error
	g.WithWriteLock(func() {
		opp.Hand.PushTop(game.Card{InstanceID: theirs, Name: "Their Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
		_, err = g.PutFromHandOntoBattlefieldForEffect(theirs, game.HandEntryOptions{})
	})
	if err != nil {
		t.Fatalf("put their bear: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Fatalf("life %d -> %d on an opponent's creature", life, me.Life)
	}
	bear := trekB27Cast(t, g, me, "Bear", "Creature — Bear", "{1}", "", 2, 2)
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Fatalf("life %d -> %d, want one gained", life, me.Life)
	}
	advanceTo(t, g, game.StepBeginCombat)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if got := counterOn(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2 (one plus one life gained)", got)
	}
}

func TestTrekB27SickbayPutsOneCounterWithNoLifeGained(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Sickbay", "Artifact", trekB27SickbayOracle, false)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
		t.Fatal("an opponent's creature was a legal target")
	}
	answerPickTarget(t, g, mine)
	passPriorityAroundTable(t, g)
	if got := counterOn(g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

// --- Tea, Earl Grey, Hot ---------------------------------------------------------

func TestTrekB27TeaDrawsAndBurnsAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hand := me.Hand.Size()
	tea := trekB27Cast(t, g, me, "Tea, Earl Grey, Hot", "Artifact — Food", "{2}", trekB27TeaOracle, 0, 0)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("hand %d -> %d, want one card drawn", hand, got)
	}
	floatMana(t, g, me, "{C}{C}")
	if err := trekB27Activate(g, me, tea, 0, game.ActivateAbilityParams{
		Modes: []int{0}, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	}); err == nil {
		t.Fatal("the damage mode targeted its controller")
	}
	life := opp.Life
	if err := trekB27Activate(g, me, tea, 0, game.ActivateAbilityParams{
		Modes: []int{0}, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !trekB27InZone(g, tea, game.ZoneGraveyard) {
		t.Error("the sacrifice cost was not paid")
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent life %d -> %d, want 2 damage", life, opp.Life)
	}
}

func TestTrekB27TeaGainsThreeLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tea := pushCatalogPermanent(g, me.ID, "Tea, Earl Grey, Hot", "Artifact — Food", trekB27TeaOracle, false)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}")
	if err := trekB27Activate(g, me, tea, 0, game.ActivateAbilityParams{Modes: []int{1}}); err == nil {
		t.Fatal("the ability was paid with one mana")
	}
	floatMana(t, g, me, "{C}")
	life := me.Life
	if err := trekB27Activate(g, me, tea, 0, game.ActivateAbilityParams{Modes: []int{1}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d -> %d, want 3 gained", life, me.Life)
	}
}

// --- Tricorder ---------------------------------------------------------------------

func TestTrekB27TricorderPumpsAndSurveilsOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tri := pushCatalogPermanent(g, me.ID, "Tricorder", "Artifact — Equipment", trekB27TricorderOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}")
	equipTo(t, g, me.ID, tri, bear)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tg != 3 {
		t.Fatalf("equipped bear is %d/%d, want 3/3", p, tg)
	}
	seedLibrary(me, "Fodder")
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	looked := trekB27Surveil(t, g, me, true)
	if !trekB27InZone(g, looked, game.ZoneGraveyard) {
		t.Error("the surveilled card was not binned")
	}
}

func TestTrekB27TricorderIgnoresAnUnequippedAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Tricorder", "Artifact — Equipment", trekB27TricorderOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if surveilChoiceFor(g, me.ID) != nil {
		t.Error("an unequipped attacker surveilled")
	}
}

// --- Type-2 Phaser -----------------------------------------------------------------

func TestTrekB27PhaserStunsOnEntryAndPumpsPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	phaser := trekB27Cast(t, g, me, "Type-2 Phaser", "Artifact — Equipment", "{2}", trekB27PhaserOracle, 0, 0)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
		t.Fatal("a creature the caster controls was a legal target")
	}
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if c := layeredCard(t, g, theirs); !c.Tapped || c.Counters[game.CounterStun] != 1 {
		t.Errorf("their creature: tapped %v, stun %d", c.Tapped, c.Counters[game.CounterStun])
	}
	floatMana(t, g, me, "{C}{C}")
	if err := trekB27Activate(g, me, phaser, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}},
	}); err == nil {
		t.Fatal("equip {3} was paid with two mana")
	}
	floatMana(t, g, me, "{C}")
	equipTo(t, g, me.ID, phaser, mine)
	if p, tg := effectivePower(t, g, mine), effectiveToughness(t, g, mine); p != 5 || tg != 2 {
		t.Errorf("equipped creature is %d/%d, want 5/2", p, tg)
	}
}

// --- U.S.S. Cerritos, Following Up ----------------------------------------------------

func TestTrekB27CerritosMakesTwoOfficersAndFliesAtSix(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ship := trekB27Cast(t, g, me, "U.S.S. Cerritos, Following Up", "Legendary Artifact — Spacecraft", "{5}", trekB27CerritosOracle, 4, 4)
	passPriorityAroundTable(t, g)
	officers := trekB27TokensNamed(g, me.ID, "Officer")
	if len(officers) != 2 {
		t.Fatalf("%d Officer tokens, want 2", len(officers))
	}
	for _, o := range officers {
		if o.Power != 1 || o.Toughness != 1 || len(o.Colors) != 1 || o.Colors[0] != "R" {
			t.Errorf("Officer %d/%d %v, want a 1/1 red", o.Power, o.Toughness, o.Colors)
		}
	}
	if layeredCard(t, g, ship).IsCreature() {
		t.Fatal("unstationed, the Spacecraft is a creature")
	}
	trekB27AddCharge(t, g, ship, 5)
	if layeredCard(t, g, ship).IsCreature() {
		t.Fatal("at 5 charge counters the Spacecraft is a creature")
	}
	crew := b12Creature(g, me.ID, "Crew", "Creature — Human", 1, 1)
	station(t, g, me, ship, crew)
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, ship)
	if !c.IsCreature() || c.CurrentPower() != 4 || c.CurrentToughness() != 4 || !game.HasKeyword(&c, "flying") {
		t.Errorf("at 6+: creature %v %d/%d flying %v; want a 4/4 flier",
			c.IsCreature(), c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(&c, "flying"))
	}
}
