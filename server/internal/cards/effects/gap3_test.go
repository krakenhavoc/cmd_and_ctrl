package effects

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gap3_test.go — card-level coverage for the play-rate slice cut from
// tracker #293's two gap lists after the #2342 / #2358 / #2359 slices.

const (
	g3MangaraOracle         = "cbcb6d9a-6ae5-4bcd-8013-2b657553764a"
	g3IcetillOracle         = "109cdefd-e8cc-4ac7-b6ba-2cfdef8d780f"
	g3PuresteelOracle       = "74a62c7b-4753-4af2-b7a1-9a4ae8988801"
	g3MirarisWakeOracle     = "852657c0-18a4-4b28-b9ae-7728acdb5044"
	g3LifeFromLoamOracle    = "ac8fba34-512c-4f24-999a-ab72f1ce4acb"
	g3FlareOracle           = "4ff07b7c-e97a-4b57-bcf1-2f22c37a8bd6"
	g3FloodcallerOracle     = "4879c8f0-8832-4290-bc71-9838940f75cd"
	g3NivParunOracle        = "33666a98-812f-4892-9f8d-33e0cbecc340"
	g3RangerCaptainOracle   = "cada3481-cc2b-4412-b9b5-0436af53aad2"
	g3CoruscationOracle     = "88bb91b5-2ccd-4ce9-8cd4-e54d63c12abf"
	g3CankerbloomOracle     = "d5b80895-621a-40df-bf48-6c7295658f21"
	g3IrrigatedFarmlandOrcl = "406eabe2-df62-49e2-bb39-c0227509d875"
	g3EscapeTunnelOracle    = "0056fc91-4398-471c-b561-7ff99750ac8a"
	g3SunscorchedOracle     = "8d2b2675-19df-4f40-9e8e-196ec097b91c"
	g3TocasiasOracle        = "25c983e0-a8c9-4784-91a4-8fe04c6df882"
)

func TestGap3CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		g3MangaraOracle:         "Mangara, the Diplomat",
		g3IcetillOracle:         "Icetill Explorer",
		g3PuresteelOracle:       "Puresteel Paladin",
		g3MirarisWakeOracle:     "Mirari's Wake",
		g3LifeFromLoamOracle:    "Life from the Loam",
		g3FlareOracle:           "Flare of Fortitude",
		g3FloodcallerOracle:     "Valley Floodcaller",
		g3NivParunOracle:        "Niv-Mizzet, Parun",
		g3RangerCaptainOracle:   "Ranger-Captain of Eos",
		g3CoruscationOracle:     "Coruscation Mage",
		g3CankerbloomOracle:     "Cankerbloom",
		g3IrrigatedFarmlandOrcl: "Irrigated Farmland",
		g3EscapeTunnelOracle:    "Escape Tunnel",
		g3SunscorchedOracle:     "Sunscorched Divide",
		g3TocasiasOracle:        "Tocasia's Welcome",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
		} else if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
}

func g3Spec(t *testing.T, oracle string) Spec {
	t.Helper()
	spec, ok := Lookup(oracle)
	if !ok {
		t.Fatalf("%s is not registered", oracle)
	}
	return spec
}

// --- Mangara, the Diplomat -----------------------------------------

func TestGap3MangaraLifelinkAndSecondSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[0]
	mangara := seedEnchantment(g, g3MangaraOracle, "Mangara, the Diplomat", "Legendary Creature — Human Cleric", me.ID)
	if !hasEffectiveKeyword(t, g, mangara, "lifelink") {
		t.Error("Mangara has lifelink")
	}
	e7Main(t, g)
	before := me.Hand.Size()

	if err := e7CastInstant(g, opp); err != nil {
		t.Fatalf("first spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 0 {
		t.Fatalf("the first spell drew %d cards, want 0", got)
	}
	if err := e7CastInstant(g, opp); err != nil {
		t.Fatalf("second spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Fatalf("the second spell drew %d cards, want 1", got)
	}
	if err := e7CastInstant(g, opp); err != nil {
		t.Fatalf("third spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("the third spell drew too: total %d, want 1", got)
	}
}

func TestGap3MangaraDrawsOnceForATwoCreatureAttackAtYou(t *testing.T) {
	for _, tc := range []struct {
		name      string
		attackers int
		atMe      bool
		want      int
	}{
		{"one attacker draws nothing", 1, true, 0},
		{"three attackers draw exactly one", 3, true, 1},
		{"two attackers at somebody else draw nothing", 2, false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp, other := g.Seats[1], g.Seats[0], g.Seats[2]
			seedEnchantment(g, g3MangaraOracle, "Mangara, the Diplomat", "Legendary Creature — Human Cleric", me.ID)
			var ids []uuid.UUID
			for i := 0; i < tc.attackers; i++ {
				ids = append(ids, pushVanillaCreature(g, opp.ID, "Bear", 2, 2))
			}
			defender := other.ID
			if tc.atMe {
				defender = me.ID
			}
			before := me.Hand.Size()
			declareAttack(t, g, defender, ids...)
			passPriorityAroundTable(t, g)
			if got := me.Hand.Size() - before; got != tc.want {
				t.Errorf("drew %d cards, want %d", got, tc.want)
			}
		})
	}
}

// --- Icetill Explorer ----------------------------------------------

func TestGap3IcetillExplorerExtraLandGraveyardPlayAndMill(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Icetill Explorer", "Creature — Insect Scout", g3IcetillOracle, false)
	advanceTo(t, g, game.StepPrecombatMain)

	// The helper that plays lands raises the base allowance for its own
	// fixtures, so the drops are counted by hand here.
	land := func() (uuid.UUID, error) {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
		return id, g.CastSpell(me.ID, id, game.CastSpellParams{})
	}
	libBefore := me.Library.Size()
	if _, err := land(); err != nil {
		t.Fatalf("first land: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := libBefore - me.Library.Size(); got != 1 {
		t.Errorf("landfall milled %d, want 1", got)
	}
	if _, err := land(); err != nil {
		t.Fatalf("the extra land drop was refused: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, err := land(); err == nil {
		t.Error("a third land drop was allowed")
	}
	// A graveyard land is playable, but a graveyard spell is not.
	if !grantedPermissionOn(g, me.ID, pushGraveyardCardTyped(me, "Dead Forest", "Basic Land — Forest"), game.ZoneGraveyard).Granted() {
		t.Error("lands in the graveyard are playable")
	}
	if grantedPermissionOn(g, me.ID, pushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear"), game.ZoneGraveyard).Granted() {
		t.Error("creature cards in the graveyard are not castable")
	}
}

// --- Puresteel Paladin ---------------------------------------------

func TestGap3PuresteelPaladinDrawsOnEquipmentOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Puresteel Paladin", "Creature — Human Knight", g3PuresteelOracle, false)
	advanceTo(t, g, game.StepPrecombatMain)

	before := me.Hand.Size()
	castAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 0 {
		t.Fatalf("a creature drew %d cards", got)
	}

	castAndResolveCreature(t, g, "Sword", "Artifact — Equipment", "")
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("the Equipment drew %d cards, want 1", got)
	}

	castAndResolveCreature(t, g, "Sword Two", "Artifact — Equipment", "")
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("declining still drew: total %d, want 1", got)
	}
}

// --- Mirari's Wake -------------------------------------------------

func TestGap3MirarisWakeAnthemAndDoubledMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Mirari's Wake", "Enchantment", g3MirarisWakeOracle, false)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	if got := effectivePower(t, g, mine); got != 3 {
		t.Errorf("my creature power = %d, want 3", got)
	}
	if got := effectivePower(t, g, theirs); got != 2 {
		t.Errorf("their creature power = %d, want 2", got)
	}
	forest := seedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	tapForMana(t, g, me.ID, forest)
	if got := sortedPool(me); !reflect.DeepEqual(got, []string{"G", "G"}) {
		t.Errorf("pool = %v, want {G}{G}", got)
	}
}

// --- Life from the Loam --------------------------------------------

func TestGap3LifeFromTheLoamReturnsUpToThreeLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := pushGraveyardCardTyped(me, "Forest", "Basic Land — Forest")
	b := pushGraveyardCardTyped(me, "Island", "Basic Land — Island")
	c := pushGraveyardCardTyped(me, "Swamp", "Basic Land — Swamp")
	d := pushGraveyardCardTyped(me, "Mountain", "Basic Land — Mountain")
	bear := pushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	refs := func(ids ...uuid.UUID) []game.TargetRef {
		var out []game.TargetRef
		for _, id := range ids {
			out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
		}
		return out
	}

	if err := castCatalogSpellErr(t, g, "Life from the Loam", "Sorcery", g3LifeFromLoamOracle, refs(bear)); err == nil {
		t.Fatal("a creature card is not a land card")
	}
	if err := castCatalogSpellErr(t, g, "Life from the Loam", "Sorcery", g3LifeFromLoamOracle, refs(a, b, c, d)); err == nil {
		t.Fatal("four targets are over the limit")
	}
	castCatalogSpell(t, g, "Life from the Loam", "Sorcery", g3LifeFromLoamOracle, refs(a, b, c))
	passPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"Forest": a, "Island": b, "Swamp": c} {
		if !me.Hand.Contains(id) {
			t.Errorf("%s did not return to hand", name)
		}
	}
	if !me.Graveyard.Contains(d) || !me.Graveyard.Contains(bear) {
		t.Error("an untargeted card left the graveyard")
	}
	if spec := g3Spec(t, g3LifeFromLoamOracle); spec.Completeness != CompletenessFull {
		t.Errorf("dredge landed in #2127, so the card is %v, want full", spec.Completeness)
	}
}

// --- Flare of Fortitude --------------------------------------------

func TestGap3FlareOfFortitudeSacrificeCostAndEffects(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	flare := handCardAtMain(t, g, "Flare of Fortitude", "Instant", g3FlareOracle)
	green := seedColoredCreature(g, me.ID, "Elf", []string{"G"}, 0)
	token := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Soldier",
		TypeLine: "Token Creature — Soldier", Colors: []string{"W"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	white := seedColoredCreature(g, me.ID, "Cleric", []string{"W"}, 0)
	theirs := seedColoredCreature(g, opp.ID, "Their Cleric", []string{"W"}, 0)

	for name, id := range map[string]uuid.UUID{"a green creature": green, "a white token": token, "an opposing creature": theirs} {
		if err := castForAltCost(g, flare, "", "sacrifice", []uuid.UUID{id}, nil); err == nil {
			t.Fatalf("%s paid \"sacrifice a nontoken white creature\"", name)
		}
	}
	if err := castForAltCost(g, flare, "", "sacrifice", []uuid.UUID{white}, nil); err != nil {
		t.Fatalf("Flare for a nontoken white creature: %v", err)
	}
	if g.Battlefield.Contains(white) {
		t.Error("the white creature was not sacrificed")
	}
	passPriorityAroundTable(t, g)

	for _, kw := range []string{"hexproof", "indestructible"} {
		if !hasEffectiveKeyword(t, g, green, kw) {
			t.Errorf("my creature lacks %s", kw)
		}
		if hasEffectiveKeyword(t, g, theirs, kw) {
			t.Errorf("their creature has %s", kw)
		}
	}
	life := me.Life
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -3) })
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2) })
	if me.Life != life {
		t.Errorf("life changed %d -> %d under Flare of Fortitude", life, me.Life)
	}
}

// --- Valley Floodcaller --------------------------------------------

func TestGap3ValleyFloodcallerPumpsAndUntapsItsTribes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Valley Floodcaller", "Creature — Otter Wizard", g3FloodcallerOracle, false)
	otter := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Otter",
		TypeLine: "Creature — Otter", Power: 1, Toughness: 1, Tapped: true, Owner: me.ID, Controller: me.ID})
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bear",
		TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Tapped: true, Owner: me.ID, Controller: me.ID})
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Rat",
		TypeLine: "Creature — Rat", Power: 1, Toughness: 1, Tapped: true, Owner: opp.ID, Controller: opp.ID})

	castAndResolveCreature(t, g, "Bear Cub", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, otter); got != 1 {
		t.Fatalf("a creature spell pumped the Otter: power %d", got)
	}

	if err := e7CastInstant(g, me); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, otter); got != 2 {
		t.Errorf("Otter power = %d, want 2", got)
	}
	if tapped, _ := battlefieldCardTapped(g, otter); tapped {
		t.Error("the Otter stayed tapped")
	}
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("a Bear was pumped: power %d", got)
	}
	if tapped, _ := battlefieldCardTapped(g, bear); !tapped {
		t.Error("a Bear was untapped")
	}
	if got := effectivePower(t, g, theirs); got != 1 {
		t.Errorf("an opposing Rat was pumped: power %d", got)
	}
}

func TestGap3ValleyFloodcallerLetsYouCastANoncreatureSpellAtInstantSpeed(t *testing.T) {
	cast := func(floodcaller bool) error {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		if floodcaller {
			pushCatalogPermanent(g, me.ID, "Valley Floodcaller", "Creature — Otter Wizard", g3FloodcallerOracle, false)
		}
		// The draw step is not a main phase, so a sorcery is untimely.
		sorcery := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: sorcery, Name: "Test Sorcery", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		return g.CastSpell(me.ID, sorcery, game.CastSpellParams{})
	}
	if err := cast(false); err == nil {
		t.Fatal("a sorcery was cast outside a main phase with no Floodcaller")
	}
	if err := cast(true); err != nil {
		t.Errorf("with Floodcaller the sorcery was refused: %v", err)
	}
}

// --- Niv-Mizzet, Parun ---------------------------------------------

func TestGap3NivMizzetParunDrawsOffAnyInstantAndPingsPerDraw(t *testing.T) {
	if !g3Spec(t, g3NivParunOracle).CantBeCountered {
		t.Error("Niv-Mizzet can't be countered")
	}
	g := newCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[0]
	niv := b21Push(g, me.ID, "Niv-Mizzet, Parun", "Legendary Creature — Dragon Wizard", g3NivParunOracle, 5, 5, "U", "R")
	if !hasEffectiveKeyword(t, g, niv, "flying") {
		t.Error("Niv-Mizzet flies")
	}
	e7Main(t, g)

	// A creature spell by anyone draws nothing.
	hand := me.Hand.Size()
	castAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatalf("a creature spell drew me a card")
	}

	// An opponent's instant draws me a card, and that draw pings.
	if err := e7CastInstant(g, opp); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards off the instant, want 1", got)
	}
	if opp.Life != 39 {
		t.Errorf("the drawn card pinged for %d, want 1", 40-opp.Life)
	}
}

// --- Ranger-Captain of Eos -----------------------------------------

func TestGap3RangerCaptainTutorsAOneDropAndBansNoncreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	seedSearchLibrary(me,
		game.Card{Name: "Big Guy", TypeLine: "Creature — Giant", ManaCost: "{3}{G}", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID},
		game.Card{Name: "Cheap Elf", TypeLine: "Creature — Elf", ManaCost: "{G}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID},
		game.Card{Name: "Cheap Spell", TypeLine: "Instant", ManaCost: "{G}", Owner: me.ID, Controller: me.ID},
	)
	captain := castAndResolveCreature(t, g, "Ranger-Captain of Eos", "Creature — Human Soldier Ranger", g3RangerCaptainOracle)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Big Guy") != uuid.Nil || searchOptionNamed(g, c, "Cheap Spell") != uuid.Nil {
		t.Error("only a creature card with mana value 1 or less is a legal find")
	}
	answerSearchNamed(t, g, me.ID, "Cheap Elf")
	passPriorityAroundTable(t, g)
	if !handHasNamed(me, "Cheap Elf") {
		t.Error("the Elf did not reach the hand")
	}

	spell := instantCardFor(t, g)
	creature := game.NewCard("Test Creature", g.Seats[0].ID)
	creature.TypeLine = "Creature — Bear"
	if err := castGateFor(g, opp.ID, spell, game.ZoneHand); err != nil {
		t.Fatalf("an opponent's instant was refused before the sacrifice: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, captain, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("sacrifice ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(captain) {
		t.Error("the Captain was not sacrificed")
	}
	var banErr *game.CantCastError
	if err := castGateFor(g, opp.ID, spell, game.ZoneHand); !errors.As(err, &banErr) {
		t.Errorf("an opponent's noncreature spell: got %v, want a cast refusal", err)
	}
	if err := castGateFor(g, opp.ID, creature, game.ZoneHand); err != nil {
		t.Errorf("an opponent's creature spell was refused: %v", err)
	}
	if err := castGateFor(g, me.ID, spell, game.ZoneHand); err != nil {
		t.Errorf("the controller's own spell was refused: %v", err)
	}
}

func handHasNamed(p *game.Player, name string) bool {
	for _, c := range p.Hand.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}

// --- Coruscation Mage ----------------------------------------------

func TestGap3CoruscationMageOffspringAndPing(t *testing.T) {
	for _, tc := range []struct {
		name     string
		optional []int
		want     int
	}{
		{"unpaid", nil, 1},
		{"offspring paid", []int{0}, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			if _, err := castWithOptionalCosts(t, g, "Coruscation Mage", "Creature — Otter Wizard", g3CoruscationOracle, nil, tc.optional, nil); err != nil {
				t.Fatalf("cast: %v", err)
			}
			passPriorityAroundTable(t, g)
			n := 0
			for _, c := range g.Battlefield.Cards {
				if c.Name == "Coruscation Mage" {
					n++
					if c.IsToken() && (c.Power != 1 || c.Toughness != 1) {
						t.Errorf("the token is %d/%d, want 1/1", c.Power, c.Toughness)
					}
				}
			}
			if n != tc.want {
				t.Errorf("Mages on the battlefield = %d, want %d", n, tc.want)
			}
		})
	}

	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Coruscation Mage", "Creature — Otter Wizard", g3CoruscationOracle, false)
	before := b29Lives(g)
	castAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Life != before[i] {
			t.Errorf("a creature spell changed seat %d's life", i)
		}
	}
	if err := e7CastInstant(g, me); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		want := before[i] - 1
		if p == me {
			want = before[i]
		}
		if p.Life != want {
			t.Errorf("seat %d life = %d, want %d", i, p.Life, want)
		}
	}
}

// --- Cankerbloom ---------------------------------------------------

func TestGap3CankerbloomModes(t *testing.T) {
	setup := func(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID, uuid.UUID) {
		g := newCatalogGame(t)
		advanceToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		bloom := pushCatalogPermanent(g, me.ID, "Cankerbloom", "Creature — Phyrexian Fungus", g3CankerbloomOracle, false)
		rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
		aura := b12Permanent(g, opp.ID, "Their Anthem", "Enchantment")
		if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
			t.Fatalf("AddManaForEffect: %v", err)
		}
		return g, me, bloom, rock, aura
	}
	t.Run("destroy an artifact", func(t *testing.T) {
		g, me, bloom, rock, aura := setup(t)
		b16Activate(t, g, me.ID, bloom, 0, game.ActivateAbilityParams{Modes: []int{0},
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}}})
		if g.Battlefield.Contains(rock) || g.Battlefield.Contains(bloom) {
			t.Error("the artifact and the sacrificed Cankerbloom should both be gone")
		}
		if !g.Battlefield.Contains(aura) {
			t.Error("the enchantment was destroyed too")
		}
	})
	t.Run("destroy an enchantment", func(t *testing.T) {
		g, me, bloom, rock, aura := setup(t)
		b16Activate(t, g, me.ID, bloom, 0, game.ActivateAbilityParams{Modes: []int{1},
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: aura}}})
		if g.Battlefield.Contains(aura) {
			t.Error("the enchantment survived")
		}
		if !g.Battlefield.Contains(rock) {
			t.Error("the artifact was destroyed too")
		}
	})
	t.Run("an artifact is not a legal enchantment target", func(t *testing.T) {
		g, me, bloom, rock, _ := setup(t)
		err := g.ActivateCatalogAbility(me.ID, bloom, 0, game.ActivateAbilityParams{Modes: []int{1},
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}}})
		if err == nil {
			t.Fatal("an artifact was accepted for the enchantment bullet")
		}
		if !g.Battlefield.Contains(bloom) {
			t.Error("a refused activation sacrificed the Cankerbloom")
		}
	})
	t.Run("proliferate", func(t *testing.T) {
		g, me, bloom, _, _ := setup(t)
		mine := seedCreatureWithCounters(g, me.ID, map[string]int{"+1/+1": 1})
		b16Activate(t, g, me.ID, bloom, 0, game.ActivateAbilityParams{Modes: []int{2}})
		if got := countersOn(g, mine, "+1/+1"); got != 2 {
			t.Errorf("+1/+1 counters = %d, want 2", got)
		}
		if g.Battlefield.Contains(bloom) {
			t.Error("the Cankerbloom was not sacrificed")
		}
	})
}

// --- the lands -----------------------------------------------------

func TestGap3IrrigatedFarmlandEntersTappedCyclesAndTapsForEither(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := playLandFromHand(t, g, "Irrigated Farmland", g3IrrigatedFarmlandOrcl)
	if card, ok := battlefieldCard(g, id); !ok || !card.Tapped {
		t.Fatal("Irrigated Farmland must enter tapped")
	}
	for _, color := range []string{"W", "U"} {
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(id) })
		me.ManaPool = nil
		b28TapForMana(t, g, me.ID, id, color)
		if got := poolColors(me); !reflect.DeepEqual(got, []string{color}) {
			t.Errorf("pool = %v, want {%s}", got, color)
		}
	}

	g2 := newCatalogGame(t)
	handBefore := g2.Seats[g2.Turn.ActiveSeat].Hand.Size()
	_, me2 := cycleFromHand(t, g2, "Irrigated Farmland", "Land — Plains Island", g3IrrigatedFarmlandOrcl, "{C}{C}")
	if got := me2.Hand.Size(); got != handBefore {
		t.Errorf("hand = %d after cycling, want %d", got, handBefore)
	}
	if me2.Graveyard.Size() == 0 {
		t.Error("the cycled land did not reach the graveyard")
	}
}

func TestGap3SunscorchedDivideFiltersOneManaIntoRedWhite(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := seedPermanentWithOracle(g, me.ID, "Sunscorched Divide", "Land", g3SunscorchedOracle)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("Sunscorched Divide made mana with nothing in the pool to pay {1}")
	}
	if c, _ := battlefieldCard(g, land); c.Tapped {
		t.Error("a refused activation tapped the land")
	}
	gl1bGivePool(me, "G")
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := gl1bSorted(poolColors(me)); !reflect.DeepEqual(got, []string{"R", "W"}) {
		t.Errorf("pool = %v, want {R}{W}", got)
	}
}

func TestGap3EscapeTunnelFetchesOrMakesASmallCreatureUnblockable(t *testing.T) {
	t.Run("fetch", func(t *testing.T) {
		g := newCatalogGame(t)
		advanceToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		tunnel := pushCatalogPermanent(g, me.ID, "Escape Tunnel", "Land", g3EscapeTunnelOracle, false)
		me.Library.Cards = nil
		pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
		b16Activate(t, g, me.ID, tunnel, 0, game.ActivateAbilityParams{})
		if g.Battlefield.Contains(tunnel) {
			t.Error("the Tunnel was not sacrificed")
		}
		var forest *game.Card
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].Name == "Forest" {
				forest = &g.Battlefield.Cards[i]
			}
		}
		if forest == nil || !forest.Tapped {
			t.Errorf("the fetched Forest is missing or untapped: %+v", forest)
		}
	})
	t.Run("unblockable", func(t *testing.T) {
		g := newCatalogGame(t)
		advanceToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		tunnel := pushCatalogPermanent(g, me.ID, "Escape Tunnel", "Land", g3EscapeTunnelOracle, false)
		small := pushRestrictionBear(g, me, "Small Beast") // 2/2
		b16Activate(t, g, me.ID, tunnel, 1, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: small}}})
		assertRestrictions(t, g, small, game.CantBeBlocked)
		if g.Battlefield.Contains(tunnel) {
			t.Error("the Tunnel was not sacrificed")
		}
	})
	t.Run("power three is not a legal target", func(t *testing.T) {
		g := newCatalogGame(t)
		advanceToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		tunnel := pushCatalogPermanent(g, me.ID, "Escape Tunnel", "Land", g3EscapeTunnelOracle, false)
		big := pushVanillaCreature(g, me.ID, "Big Beast", 3, 3)
		err := g.ActivateCatalogAbility(me.ID, tunnel, 1, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big}}})
		if !errors.Is(err, game.ErrIllegalTarget) {
			t.Fatalf("err = %v, want ErrIllegalTarget", err)
		}
		if !g.Battlefield.Contains(tunnel) {
			t.Error("a refused activation sacrificed the land")
		}
	})
}

// --- Tocasia's Welcome ---------------------------------------------

func TestGap3TocasiasWelcomeDrawsOncePerTurnForSmallCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Tocasia's Welcome", "Enchantment", g3TocasiasOracle, false)
	advanceTo(t, g, game.StepPrecombatMain)
	cast := func(name, cost string) {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Bear", ManaCost: cost,
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast %s: %v", name, err)
		}
		passPriorityAroundTable(t, g)
	}
	before := me.Hand.Size() // each cast puts its card in hand first, so only draws move this
	cast("Four Drop", "{3}{G}")
	if got := me.Hand.Size() - before; got != 0 {
		t.Fatalf("a mana value 4 creature drew %d cards", got)
	}
	cast("Three Drop", "{2}{G}")
	if got := me.Hand.Size() - before; got != 1 {
		t.Fatalf("a mana value 3 creature drew %d cards, want 1", got)
	}
	cast("Two Drop", "{1}{G}")
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("the second small creature drew again: total %d, want 1", got)
	}
}
