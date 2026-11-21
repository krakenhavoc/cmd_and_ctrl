package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b77_cards_test.go — S82 GAP batch 77 (Ravnica Remastered
// creatures and enchantments, #5948). Every test plays the real catalog entry
// through the engine.

var gapB77Cards = map[string]string{
	"Mistral Charger":        "a084b37a-bfb2-4dde-9faa-4c55ad894300",
	"Rising Populace":        "340de754-9d64-4824-936f-643be4a939db",
	"Ill-Gotten Inheritance": "84e8ca5e-730e-468b-943c-7c6358ade95c",
	"Fists of Ironwood":      "2f6a5278-23bf-47b2-a2b9-dfff7d96f3dd",
	"Forced Adaptation":      "1c859c97-74df-4360-baf8-0f79eae97c45",
	"Perilous Forays":        "d552f5cb-19b6-4def-90d5-6d2d9570f93d",
	"Wurmweaver Coil":        "83419229-c266-4022-8297-d1b05ec5ee20",
	"Cindervines":            "92d95e12-fa51-4360-88e4-9078ccc5d8d3",
	"Taste for Mayhem":       "60e7517f-7cc5-4073-9f2c-3e6acb891128",
}

func TestGapB77EveryCardIsRegisteredFull(t *testing.T) {
	for name, oracle := range gapB77Cards {
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

func TestGapB77MistralChargerFlies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b12Push(g, me.ID, "Mistral Charger", "Creature — Pegasus", gapB77Cards["Mistral Charger"], 2, 1)
	if !hasEffectiveKeyword(t, g, id, "flying") {
		t.Error("Mistral Charger lacks flying")
	}
}

func TestGapB77RisingPopulaceGrowsOnlyForOthersThatDie(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pop := b12Push(g, me.ID, "Rising Populace", "Creature — Human", gapB77Cards["Rising Populace"], 2, 2)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	killCreature(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, pop, game.CounterPlusOne); n != 0 {
		t.Fatalf("%d counters after an opponent's creature died, want 0", n)
	}
	killCreature(t, g, me.ID, mine)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, pop, game.CounterPlusOne); n != 1 {
		t.Fatalf("%d counters after my creature died, want 1", n)
	}
	var walker uuid.UUID
	g.WithWriteLock(func() {
		walker = pushPermanent(g, me.ID, game.Card{Name: "My Walker", TypeLine: "Legendary Planeswalker — Test",
			Counters: map[string]int{game.CounterLoyalty: 3}})
	})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield},
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, walker); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if n := counterCount(g, pop, game.CounterPlusOne); n != 2 {
		t.Errorf("%d counters after my planeswalker died, want 2", n)
	}
	killCreature(t, g, me.ID, pop)
	passPriorityAroundTable(t, g)
}

func TestGapB77FistsOfIronwoodMakesTwoSaprolingsAndGrantsTrample(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	host := b12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Fists of Ironwood", "Enchantment — Aura", gapB77Cards["Fists of Ironwood"], cardTarget(host))
	passPriorityAroundTable(t, g)
	if n := namedOnBattlefield(g, "Saproling", me.ID); n != 2 {
		t.Errorf("%d Saprolings, want 2", n)
	}
	if !hasEffectiveKeyword(t, g, host, "trample") {
		t.Error("the enchanted creature lacks trample")
	}
	if p, to := effectivePTOf(t, g, host); p != 2 || to != 2 {
		t.Errorf("host is %d/%d, want 2/2", p, to)
	}
}

func TestGapB77FistsOfIronwoodNeedsACreature(t *testing.T) {
	g := newCatalogGame(t)
	rock := b12Permanent(g, g.Seats[0].ID, "Rock", "Artifact")
	if err := castCatalogSpellErrStrict(t, g, "Fists of Ironwood", "Enchantment — Aura", gapB77Cards["Fists of Ironwood"], cardTarget(rock)); err == nil {
		t.Error("the Aura was cast on an artifact")
	}
}

func TestGapB77ForcedAdaptationGrowsTheHostEachOfMyUpkeeps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := b12Creature(g, opp.ID, "Host", "Creature — Bear", 2, 2)
	cmdr2B19Attach(g, me.ID, host, "Forced Adaptation", gapB77Cards["Forced Adaptation"])
	advanceToUpkeepOfSeat(t, g, 0)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, host, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters on the host, want 1", n)
	}
}

func TestGapB77TasteForMayhemIsBiggerWhenHellbent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	host := b12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	cmdr2B19Attach(g, me.ID, host, "Taste for Mayhem", gapB77Cards["Taste for Mayhem"])
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	if p, to := effectivePTOf(t, g, host); p != 4 || to != 2 {
		t.Errorf("with a card in hand the host is %d/%d, want 4/2", p, to)
	}
	for me.Hand.Size() > 0 {
		id := me.Hand.Cards[0].InstanceID
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID},
			game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, id); err != nil {
			t.Fatal(err)
		}
	}
	if p, to := effectivePTOf(t, g, host); p != 6 || to != 2 {
		t.Errorf("hellbent the host is %d/%d, want 6/2", p, to)
	}
}

func TestGapB77WurmweaverCoilPumpsAGreenCreatureAndMakesAWurm(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	green := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear",
		Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	red := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Red Bear", TypeLine: "Creature — Bear",
		Colors: []string{"R"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if err := castCatalogSpellErrStrict(t, g, "Wurmweaver Coil", "Enchantment — Aura", gapB77Cards["Wurmweaver Coil"], cardTarget(red)); err == nil {
		t.Fatal("the Aura was cast on a red creature")
	}
	coil := cmdr2B19Attach(g, me.ID, green, "Wurmweaver Coil", gapB77Cards["Wurmweaver Coil"])
	if p, to := effectivePTOf(t, g, green); p != 8 || to != 8 {
		t.Fatalf("host is %d/%d, want 8/8", p, to)
	}
	apaMana(me, "G", "G", "G")
	apaActivate(t, g, me, coil, 0, game.ActivateAbilityParams{})
	if battlefieldHas(g, coil) {
		t.Error("the Aura was not sacrificed")
	}
	if n := namedOnBattlefield(g, "Wurm", me.ID); n != 1 {
		t.Errorf("%d Wurm tokens, want 1", n)
	}
	if p, to := effectivePTOf(t, g, green); p != 2 || to != 2 {
		t.Errorf("host is %d/%d after the Aura left, want 2/2", p, to)
	}
	if err := g.ActivateCatalogAbility(me.ID, coil, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated a sacrificed Aura")
	}
}

func TestGapB77IllGottenInheritance(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	inh := b12Push(g, me.ID, "Ill-Gotten Inheritance", "Enchantment", gapB77Cards["Ill-Gotten Inheritance"], 0, 0)
	myLife, oppLife := me.Life, opp.Life
	if err := g.ActivateCatalogAbility(me.ID, inh, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{playerTarget(opp.ID)}}); err == nil {
		t.Fatal("activated with no mana")
	}
	apaMana(me, "B", "B", "B", "B", "B", "B")
	apaActivate(t, g, me, inh, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{playerTarget(opp.ID)}})
	if opp.Life != oppLife-4 || me.Life != myLife+4 {
		t.Errorf("life %d/%d, want %d/%d", me.Life, opp.Life, myLife+4, oppLife-4)
	}
	if battlefieldHas(g, inh) {
		t.Error("the enchantment was not sacrificed")
	}
}

func TestGapB77IllGottenInheritanceDrainsEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Ill-Gotten Inheritance", "Enchantment", gapB77Cards["Ill-Gotten Inheritance"], 0, 0)
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	advanceToUpkeepOfSeat(t, g, 0)
	passPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		want := lives[s.ID] - 1
		if s.ID == me.ID {
			want = lives[s.ID] + 1
		}
		if s.Life != want {
			t.Errorf("%s is at %d, want %d", s.Name, s.Life, want)
		}
	}
}

func TestGapB77PerilousForaysFetchesATappedLandWithABasicType(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	forays := b12Push(g, me.ID, "Perilous Forays", "Enchantment", gapB77Cards["Perilous Forays"], 0, 0)
	victim := b12Creature(g, me.ID, "Victim", "Creature — Bear", 1, 1)
	wastes := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: wastes, Name: "Wastes", TypeLine: "Basic Land", Owner: me.ID, Controller: me.ID})
	shock := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: shock, Name: "Breeding Pool", TypeLine: "Land — Forest Island", Owner: me.ID, Controller: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, forays, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{victim}}); err == nil {
		t.Fatal("activated with no mana")
	}
	apaMana(me, "G")
	apaActivate(t, g, me, forays, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{victim}})
	if battlefieldHas(g, victim) {
		t.Error("the creature was not sacrificed")
	}
	search := latestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, me.ID)
	if search == nil {
		t.Fatalf("no search prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, []uuid.UUID{wastes}); err == nil {
		t.Fatal("a Wastes (no basic land type) was accepted")
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, []uuid.UUID{shock}); err != nil {
		t.Fatalf("search: %v", err)
	}
	g.SettleResolution()
	passPriorityAroundTable(t, g)
	if !battlefieldHas(g, shock) {
		t.Error("the Forest Island did not enter the battlefield")
	} else if c := findBattlefieldCardForTest(g, shock); !c.Tapped {
		t.Error("the fetched land entered untapped")
	}
}

func TestGapB77CindervinesPingsAnOpponentsNoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	var owner *game.Player
	for _, s := range g.Seats {
		if s.ID != active.ID {
			owner = s
			break
		}
	}
	b12Push(g, owner.ID, "Cindervines", "Enchantment", gapB77Cards["Cindervines"], 0, 0)
	life := active.Life
	castCatalogSpell(t, g, "Plain Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if active.Life != life-1 {
		t.Errorf("the caster is at %d, want %d", active.Life, life-1)
	}
	castCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if active.Life != life-1 {
		t.Errorf("a creature spell cost %d life", life-1-active.Life)
	}
}

func TestGapB77CindervinesDoesNotPingItsOwnersSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Cindervines", "Enchantment", gapB77Cards["Cindervines"], 0, 0)
	life := me.Life
	castCatalogSpell(t, g, "Plain Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("my own spell cost me %d life", life-me.Life)
	}
}

func TestGapB77CindervinesDestroysAndHurtsThatPermanentsController(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	vines := b12Push(g, me.ID, "Cindervines", "Enchantment", gapB77Cards["Cindervines"], 0, 0)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	apaMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, vines, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(bear)}); err == nil {
		t.Fatal("targeted a creature")
	}
	life := opp.Life
	apaActivate(t, g, me, vines, 0, game.ActivateAbilityParams{Targets: cardTarget(rock)})
	if battlefieldHas(g, rock) {
		t.Error("the artifact survived")
	}
	if battlefieldHas(g, vines) {
		t.Error("Cindervines was not sacrificed")
	}
	if opp.Life != life-2 {
		t.Errorf("the controller is at %d, want %d", opp.Life, life-2)
	}
}
