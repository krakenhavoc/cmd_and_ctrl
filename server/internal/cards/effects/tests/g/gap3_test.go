package testsg

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestGap3CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.G3MangaraOracle:         "Mangara, the Diplomat",
		effectstest.G3IcetillOracle:         "Icetill Explorer",
		effectstest.G3PuresteelOracle:       "Puresteel Paladin",
		effectstest.G3MirarisWakeOracle:     "Mirari's Wake",
		effectstest.G3LifeFromLoamOracle:    "Life from the Loam",
		effectstest.G3FlareOracle:           "Flare of Fortitude",
		effectstest.G3FloodcallerOracle:     "Valley Floodcaller",
		effectstest.G3NivParunOracle:        "Niv-Mizzet, Parun",
		effectstest.G3RangerCaptainOracle:   "Ranger-Captain of Eos",
		effectstest.G3CoruscationOracle:     "Coruscation Mage",
		effectstest.G3CankerbloomOracle:     "Cankerbloom",
		effectstest.G3IrrigatedFarmlandOrcl: "Irrigated Farmland",
		effectstest.G3EscapeTunnelOracle:    "Escape Tunnel",
		effectstest.G3SunscorchedOracle:     "Sunscorched Divide",
		effectstest.G3TocasiasOracle:        "Tocasia's Welcome",
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
		} else if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
}

func g3Spec(t *testing.T, oracle string) effects.Spec {
	t.Helper()
	spec, ok := effects.Lookup(oracle)
	if !ok {
		t.Fatalf("%s is not registered", oracle)
	}
	return spec
}

// --- Mangara, the Diplomat -----------------------------------------

func TestGap3MangaraLifelinkAndSecondSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[0]
	mangara := effectstest.SeedEnchantment(g, effectstest.G3MangaraOracle, "Mangara, the Diplomat", "Legendary Creature — Human Cleric", me.ID)
	if !effectstest.HasEffectiveKeyword(t, g, mangara, "lifelink") {
		t.Error("Mangara has lifelink")
	}
	effectstest.E7Main(t, g)
	before := me.Hand.Size()

	if err := effectstest.E7CastInstant(g, opp); err != nil {
		t.Fatalf("first spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 0 {
		t.Fatalf("the first spell drew %d cards, want 0", got)
	}
	if err := effectstest.E7CastInstant(g, opp); err != nil {
		t.Fatalf("second spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Fatalf("the second spell drew %d cards, want 1", got)
	}
	if err := effectstest.E7CastInstant(g, opp); err != nil {
		t.Fatalf("third spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
			g := effectstest.NewCatalogGame(t)
			me, opp, other := g.Seats[1], g.Seats[0], g.Seats[2]
			effectstest.SeedEnchantment(g, effectstest.G3MangaraOracle, "Mangara, the Diplomat", "Legendary Creature — Human Cleric", me.ID)
			var ids []uuid.UUID
			for i := 0; i < tc.attackers; i++ {
				ids = append(ids, effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2))
			}
			defender := other.ID
			if tc.atMe {
				defender = me.ID
			}
			before := me.Hand.Size()
			effectstest.DeclareAttack(t, g, defender, ids...)
			effectstest.PassPriorityAroundTable(t, g)
			if got := me.Hand.Size() - before; got != tc.want {
				t.Errorf("drew %d cards, want %d", got, tc.want)
			}
		})
	}
}

// #4864: Mangara prints "attacking you and/or planeswalkers you
// control", so two creatures at your planeswalker draw, and so do one
// at you and one at it. Two at a battle you protect do not.
func TestGap3MangaraCountsYourPlaneswalkersButNotYourBattles(t *testing.T) {
	for _, tc := range []struct {
		name string
		at   func(me, walker, battle uuid.UUID) [2]uuid.UUID
		want int
	}{
		{"two at your planeswalker draw one", func(_, w, _ uuid.UUID) [2]uuid.UUID { return [2]uuid.UUID{w, w} }, 1},
		{"one at you and one at your planeswalker draw one", func(m, w, _ uuid.UUID) [2]uuid.UUID { return [2]uuid.UUID{m, w} }, 1},
		{"two at a battle you protect draw nothing", func(_, _, b uuid.UUID) [2]uuid.UUID { return [2]uuid.UUID{b, b} }, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp, other := g.Seats[1], g.Seats[0], g.Seats[2]
			effectstest.SeedEnchantment(g, effectstest.G3MangaraOracle, "Mangara, the Diplomat", "Legendary Creature — Human Cleric", me.ID)
			walker := effectstest.P1g2Planeswalker(g, me.ID, 5)
			battle := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "Test Siege", TypeLine: "Battle — Siege",
				Owner: other.ID, Controller: other.ID, ProtectorPlayerID: me.ID,
				Counters: map[string]int{game.CounterDefense: 5},
			})
			ids := []uuid.UUID{effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2), effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)}
			at := tc.at(me.ID, walker, battle)
			before := me.Hand.Size()
			effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
			for i, id := range ids {
				if err := g.DeclareAttacker(id, at[i]); err != nil {
					t.Fatalf("DeclareAttacker: %v", err)
				}
			}
			effectstest.LockInAttacks(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			if got := me.Hand.Size() - before; got != tc.want {
				t.Errorf("drew %d cards, want %d", got, tc.want)
			}
		})
	}
}

// --- Icetill Explorer ----------------------------------------------

func TestGap3IcetillExplorerExtraLandGraveyardPlayAndMill(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Icetill Explorer", "Creature — Insect Scout", effectstest.G3IcetillOracle, false)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)

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
	effectstest.PassPriorityAroundTable(t, g)
	if got := libBefore - me.Library.Size(); got != 1 {
		t.Errorf("landfall milled %d, want 1", got)
	}
	if _, err := land(); err != nil {
		t.Fatalf("the extra land drop was refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, err := land(); err == nil {
		t.Error("a third land drop was allowed")
	}
	// A graveyard land is playable, but a graveyard spell is not.
	if !effectstest.GrantedPermissionOn(g, me.ID, effectstest.PushGraveyardCardTyped(me, "Dead Forest", "Basic Land — Forest"), game.ZoneGraveyard).Granted() {
		t.Error("lands in the graveyard are playable")
	}
	if effectstest.GrantedPermissionOn(g, me.ID, effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear"), game.ZoneGraveyard).Granted() {
		t.Error("creature cards in the graveyard are not castable")
	}
}

// --- Puresteel Paladin ---------------------------------------------

func TestGap3PuresteelPaladinDrawsOnEquipmentOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Puresteel Paladin", "Creature — Human Knight", effectstest.G3PuresteelOracle, false)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)

	before := me.Hand.Size()
	effectstest.CastAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 0 {
		t.Fatalf("a creature drew %d cards", got)
	}

	effectstest.CastAndResolveCreature(t, g, "Sword", "Artifact — Equipment", "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("the Equipment drew %d cards, want 1", got)
	}

	effectstest.CastAndResolveCreature(t, g, "Sword Two", "Artifact — Equipment", "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("declining still drew: total %d, want 1", got)
	}
}

// --- Mirari's Wake -------------------------------------------------

func TestGap3MirarisWakeAnthemAndDoubledMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Mirari's Wake", "Enchantment", effectstest.G3MirarisWakeOracle, false)
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	if got := effectstest.EffectivePower(t, g, mine); got != 3 {
		t.Errorf("my creature power = %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != 2 {
		t.Errorf("their creature power = %d, want 2", got)
	}
	forest := effectstest.SeedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	effectstest.TapForMana(t, g, me.ID, forest)
	if got := effectstest.SortedPool(me); !reflect.DeepEqual(got, []string{"G", "G"}) {
		t.Errorf("pool = %v, want {G}{G}", got)
	}
}

// --- Life from the Loam --------------------------------------------

func TestGap3LifeFromTheLoamReturnsUpToThreeLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.PushGraveyardCardTyped(me, "Forest", "Basic Land — Forest")
	b := effectstest.PushGraveyardCardTyped(me, "Island", "Basic Land — Island")
	c := effectstest.PushGraveyardCardTyped(me, "Swamp", "Basic Land — Swamp")
	d := effectstest.PushGraveyardCardTyped(me, "Mountain", "Basic Land — Mountain")
	bear := effectstest.PushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	refs := func(ids ...uuid.UUID) []game.TargetRef {
		var out []game.TargetRef
		for _, id := range ids {
			out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
		}
		return out
	}

	if err := effectstest.CastCatalogSpellErr(t, g, "Life from the Loam", "Sorcery", effectstest.G3LifeFromLoamOracle, refs(bear)); err == nil {
		t.Fatal("a creature card is not a land card")
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Life from the Loam", "Sorcery", effectstest.G3LifeFromLoamOracle, refs(a, b, c, d)); err == nil {
		t.Fatal("four targets are over the limit")
	}
	effectstest.CastCatalogSpell(t, g, "Life from the Loam", "Sorcery", effectstest.G3LifeFromLoamOracle, refs(a, b, c))
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"Forest": a, "Island": b, "Swamp": c} {
		if !me.Hand.Contains(id) {
			t.Errorf("%s did not return to hand", name)
		}
	}
	if !me.Graveyard.Contains(d) || !me.Graveyard.Contains(bear) {
		t.Error("an untargeted card left the graveyard")
	}
	if spec := g3Spec(t, effectstest.G3LifeFromLoamOracle); spec.Completeness != effects.CompletenessFull {
		t.Errorf("dredge landed in #2127, so the card is %v, want full", spec.Completeness)
	}
}

// --- Flare of Fortitude --------------------------------------------

func TestGap3FlareOfFortitudeSacrificeCostAndEffects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	flare := effectstest.HandCardAtMain(t, g, "Flare of Fortitude", "Instant", effectstest.G3FlareOracle)
	green := effectstest.SeedColoredCreature(g, me.ID, "Elf", []string{"G"}, 0)
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Soldier",
		TypeLine: "Token Creature — Soldier", Colors: []string{"W"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	white := effectstest.SeedColoredCreature(g, me.ID, "Cleric", []string{"W"}, 0)
	theirs := effectstest.SeedColoredCreature(g, opp.ID, "Their Cleric", []string{"W"}, 0)

	for name, id := range map[string]uuid.UUID{"a green creature": green, "a white token": token, "an opposing creature": theirs} {
		if err := effectstest.CastForAltCost(g, flare, "", "sacrifice", []uuid.UUID{id}, nil); err == nil {
			t.Fatalf("%s paid \"sacrifice a nontoken white creature\"", name)
		}
	}
	if err := effectstest.CastForAltCost(g, flare, "", "sacrifice", []uuid.UUID{white}, nil); err != nil {
		t.Fatalf("Flare for a nontoken white creature: %v", err)
	}
	if g.Battlefield.Contains(white) {
		t.Error("the white creature was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)

	for _, kw := range []string{"hexproof", "indestructible"} {
		if !effectstest.HasEffectiveKeyword(t, g, green, kw) {
			t.Errorf("my creature lacks %s", kw)
		}
		if effectstest.HasEffectiveKeyword(t, g, theirs, kw) {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Valley Floodcaller", "Creature — Otter Wizard", effectstest.G3FloodcallerOracle, false)
	otter := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Otter",
		TypeLine: "Creature — Otter", Power: 1, Toughness: 1, Tapped: true, Owner: me.ID, Controller: me.ID})
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bear",
		TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Tapped: true, Owner: me.ID, Controller: me.ID})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Rat",
		TypeLine: "Creature — Rat", Power: 1, Toughness: 1, Tapped: true, Owner: opp.ID, Controller: opp.ID})

	effectstest.CastAndResolveCreature(t, g, "Bear Cub", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, otter); got != 1 {
		t.Fatalf("a creature spell pumped the Otter: power %d", got)
	}

	if err := effectstest.E7CastInstant(g, me); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, otter); got != 2 {
		t.Errorf("Otter power = %d, want 2", got)
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, otter); tapped {
		t.Error("the Otter stayed tapped")
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("a Bear was pumped: power %d", got)
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, bear); !tapped {
		t.Error("a Bear was untapped")
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != 1 {
		t.Errorf("an opposing Rat was pumped: power %d", got)
	}
}

func TestGap3ValleyFloodcallerLetsYouCastANoncreatureSpellAtInstantSpeed(t *testing.T) {
	cast := func(floodcaller bool) error {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		if floodcaller {
			effectstest.PushCatalogPermanent(g, me.ID, "Valley Floodcaller", "Creature — Otter Wizard", effectstest.G3FloodcallerOracle, false)
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
	if !g3Spec(t, effectstest.G3NivParunOracle).CantBeCountered {
		t.Error("Niv-Mizzet can't be countered")
	}
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[0]
	niv := effectstest.B21Push(g, me.ID, "Niv-Mizzet, Parun", "Legendary Creature — Dragon Wizard", effectstest.G3NivParunOracle, 5, 5, "U", "R")
	if !effectstest.HasEffectiveKeyword(t, g, niv, "flying") {
		t.Error("Niv-Mizzet flies")
	}
	effectstest.E7Main(t, g)

	// A creature spell by anyone draws nothing.
	hand := me.Hand.Size()
	effectstest.CastAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatalf("a creature spell drew me a card")
	}

	// An opponent's instant draws me a card, and that draw pings.
	if err := effectstest.E7CastInstant(g, opp); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards off the instant, want 1", got)
	}
	if opp.Life != 39 {
		t.Errorf("the drawn card pinged for %d, want 1", 40-opp.Life)
	}
}

// --- Ranger-Captain of Eos -----------------------------------------

func TestGap3RangerCaptainTutorsAOneDropAndBansNoncreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Big Guy", TypeLine: "Creature — Giant", ManaCost: "{3}{G}", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID},
		game.Card{Name: "Cheap Elf", TypeLine: "Creature — Elf", ManaCost: "{G}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID},
		game.Card{Name: "Cheap Spell", TypeLine: "Instant", ManaCost: "{G}", Owner: me.ID, Controller: me.ID},
	)
	captain := effectstest.CastAndResolveCreature(t, g, "Ranger-Captain of Eos", "Creature — Human Soldier Ranger", effectstest.G3RangerCaptainOracle)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Big Guy") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Cheap Spell") != uuid.Nil {
		t.Error("only a creature card with mana value 1 or less is a legal find")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Cheap Elf")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasNamed(me, "Cheap Elf") {
		t.Error("the Elf did not reach the hand")
	}

	spell := effectstest.InstantCardFor(t, g)
	creature := game.NewCard("Test Creature", g.Seats[0].ID)
	creature.TypeLine = "Creature — Bear"
	if err := effectstest.CastGateFor(g, opp.ID, spell, game.ZoneHand); err != nil {
		t.Fatalf("an opponent's instant was refused before the sacrifice: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, captain, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("sacrifice ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(captain) {
		t.Error("the Captain was not sacrificed")
	}
	var banErr *game.CantCastError
	if err := effectstest.CastGateFor(g, opp.ID, spell, game.ZoneHand); !errors.As(err, &banErr) {
		t.Errorf("an opponent's noncreature spell: got %v, want a cast refusal", err)
	}
	if err := effectstest.CastGateFor(g, opp.ID, creature, game.ZoneHand); err != nil {
		t.Errorf("an opponent's creature spell was refused: %v", err)
	}
	if err := effectstest.CastGateFor(g, me.ID, spell, game.ZoneHand); err != nil {
		t.Errorf("the controller's own spell was refused: %v", err)
	}
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
			g := effectstest.NewCatalogGame(t)
			if _, err := effectstest.CastWithOptionalCosts(t, g, "Coruscation Mage", "Creature — Otter Wizard", effectstest.G3CoruscationOracle, nil, tc.optional, nil); err != nil {
				t.Fatalf("cast: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
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

	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Coruscation Mage", "Creature — Otter Wizard", effectstest.G3CoruscationOracle, false)
	before := effectstest.B29Lives(g)
	effectstest.CastAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Life != before[i] {
			t.Errorf("a creature spell changed seat %d's life", i)
		}
	}
	if err := effectstest.E7CastInstant(g, me); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		bloom := effectstest.PushCatalogPermanent(g, me.ID, "Cankerbloom", "Creature — Phyrexian Fungus", effectstest.G3CankerbloomOracle, false)
		rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
		aura := effectstest.B12Permanent(g, opp.ID, "Their Anthem", "Enchantment")
		if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
			t.Fatalf("AddManaForEffect: %v", err)
		}
		return g, me, bloom, rock, aura
	}
	t.Run("destroy an artifact", func(t *testing.T) {
		g, me, bloom, rock, aura := setup(t)
		effectstest.B16Activate(t, g, me.ID, bloom, 0, game.ActivateAbilityParams{Modes: []int{0},
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
		effectstest.B16Activate(t, g, me.ID, bloom, 0, game.ActivateAbilityParams{Modes: []int{1},
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
		mine := effectstest.SeedCreatureWithCounters(g, me.ID, map[string]int{"+1/+1": 1})
		effectstest.B16Activate(t, g, me.ID, bloom, 0, game.ActivateAbilityParams{Modes: []int{2}})
		if got := effectstest.CountersOn(g, mine, "+1/+1"); got != 2 {
			t.Errorf("+1/+1 counters = %d, want 2", got)
		}
		if g.Battlefield.Contains(bloom) {
			t.Error("the Cankerbloom was not sacrificed")
		}
	})
}

// --- the lands -----------------------------------------------------

func TestGap3IrrigatedFarmlandEntersTappedCyclesAndTapsForEither(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PlayLandFromHand(t, g, "Irrigated Farmland", effectstest.G3IrrigatedFarmlandOrcl)
	if card, ok := effectstest.BattlefieldCard(g, id); !ok || !card.Tapped {
		t.Fatal("Irrigated Farmland must enter tapped")
	}
	for _, color := range []string{"W", "U"} {
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(id) })
		me.ManaPool = nil
		effectstest.B28TapForMana(t, g, me.ID, id, color)
		if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{color}) {
			t.Errorf("pool = %v, want {%s}", got, color)
		}
	}

	g2 := effectstest.NewCatalogGame(t)
	handBefore := g2.Seats[g2.Turn.ActiveSeat].Hand.Size()
	_, me2 := effectstest.CycleFromHand(t, g2, "Irrigated Farmland", "Land — Plains Island", effectstest.G3IrrigatedFarmlandOrcl, "{C}{C}")
	if got := me2.Hand.Size(); got != handBefore {
		t.Errorf("hand = %d after cycling, want %d", got, handBefore)
	}
	if me2.Graveyard.Size() == 0 {
		t.Error("the cycled land did not reach the graveyard")
	}
}

func TestGap3SunscorchedDivideFiltersOneManaIntoRedWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.SeedPermanentWithOracle(g, me.ID, "Sunscorched Divide", "Land", effectstest.G3SunscorchedOracle)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("Sunscorched Divide made mana with nothing in the pool to pay {1}")
	}
	if c, _ := effectstest.BattlefieldCard(g, land); c.Tapped {
		t.Error("a refused activation tapped the land")
	}
	gl1bGivePool(me, "G")
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := gl1bSorted(effectstest.PoolColors(me)); !reflect.DeepEqual(got, []string{"R", "W"}) {
		t.Errorf("pool = %v, want {R}{W}", got)
	}
}

func TestGap3EscapeTunnelFetchesOrMakesASmallCreatureUnblockable(t *testing.T) {
	t.Run("fetch", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		tunnel := effectstest.PushCatalogPermanent(g, me.ID, "Escape Tunnel", "Land", effectstest.G3EscapeTunnelOracle, false)
		me.Library.Cards = nil
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
		effectstest.B16Activate(t, g, me.ID, tunnel, 0, game.ActivateAbilityParams{})
		effectstest.TakeSearchPrompts(t, g)
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
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		tunnel := effectstest.PushCatalogPermanent(g, me.ID, "Escape Tunnel", "Land", effectstest.G3EscapeTunnelOracle, false)
		small := effectstest.PushRestrictionBear(g, me, "Small Beast") // 2/2
		effectstest.B16Activate(t, g, me.ID, tunnel, 1, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: small}}})
		effectstest.AssertRestrictions(t, g, small, game.CantBeBlocked)
		if g.Battlefield.Contains(tunnel) {
			t.Error("the Tunnel was not sacrificed")
		}
	})
	t.Run("power three is not a legal target", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		tunnel := effectstest.PushCatalogPermanent(g, me.ID, "Escape Tunnel", "Land", effectstest.G3EscapeTunnelOracle, false)
		big := effectstest.PushVanillaCreature(g, me.ID, "Big Beast", 3, 3)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Tocasia's Welcome", "Enchantment", effectstest.G3TocasiasOracle, false)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	cast := func(name, cost string) {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Bear", ManaCost: cost,
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast %s: %v", name, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
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
