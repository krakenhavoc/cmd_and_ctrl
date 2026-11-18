package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	rprVivisection   = "46f78c42-85d1-4cd4-9fc3-d5fdec04a4c8"
	rprAinokBondKin  = "8a4028cd-6004-4fb0-a044-0bee662a78e1"
	rprAngelGlory    = "c6a606af-ea08-4cd9-bfd9-60d76e232e85"
	rprBestialMenace = "ef08d371-9d6c-4a2d-a444-d475fbf1b633"
	rprDarkSalvation = "f5c2b700-da52-4436-b234-e32ba7759e17"
	rprDearlyDeparts = "539be0f3-0863-4d44-87ee-5912dd003475"
	rprDeathsPres    = "d0db5996-f749-48d2-8fb3-a69104113828"
	rprGrowthSpasm   = "6c022656-2b97-46c5-8f7b-72a71996457a"
	rprRunebinder    = "e91e7f62-0788-445e-ba93-2c9d233ec3ab"
	rprHeronChampion = "86b5a853-9035-4490-9d0f-7c719bedfd58"
	rprHourEternity  = "6fbaac3b-37ec-4815-a4ea-544c547e90b1"
)

func TestRprB31CardsAreFull(t *testing.T) {
	for _, id := range []string{rprVivisection, rprAinokBondKin, rprAngelGlory, rprBestialMenace, rprDarkSalvation,
		rprDearlyDeparts, rprDeathsPres, rprGrowthSpasm, rprRunebinder, rprHeronChampion, rprHourEternity} {
		spec, ok := Lookup(id)
		if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v, completeness %v, caveats %v", id, ok, spec.Completeness, spec.Caveats)
		}
	}
}

// rprCastCreature casts a creature card from the active seat's hand.
func rprCastCreature(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	return id
}

func rprBattlefieldCount(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			n++
		}
	}
	return n
}

// --- Vivisection ----------------------------------------------------------

func TestRprVivisectionSacrificesACreatureAndDrawsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	seedTypedLibrary(me, "Land", "Island", "Island", "Island", "Island")
	hand := me.Hand.Size()
	paCast(t, g, "Vivisection", "Sorcery", "{3}{U}", rprVivisection, game.CastSpellParams{SacrificeIDs: []uuid.UUID{bear}})
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Bear") != uuid.Nil {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("hand grew by %d, want 3 (the spell left the hand, three were drawn)", got+1)
	}
}

func TestRprVivisectionRefusedWithoutASacrifice(t *testing.T) {
	g := newCatalogGame(t)
	_, err := castWithTapParams(t, g, "Vivisection", "Sorcery", "{3}{U}", rprVivisection, game.CastSpellParams{Strict: true})
	if err == nil {
		t.Fatal("cast without sacrificing a creature")
	}
}

// --- Bestial Menace -------------------------------------------------------

func TestRprBestialMenaceMakesThreeTokens(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Bestial Menace", "Sorcery", rprBestialMenace, nil)
	passPriorityAroundTable(t, g)
	for _, w := range []struct {
		name string
		p    int
	}{{"Snake", 1}, {"Wolf", 2}, {"Elephant", 3}} {
		id := findBattlefieldByName(g, w.name)
		if id == uuid.Nil {
			t.Fatalf("no %s token", w.name)
		}
		if got := effectivePower(t, g, id); got != w.p {
			t.Errorf("%s power %d, want %d", w.name, got, w.p)
		}
	}
	if rprBattlefieldCount(g, me.ID, "Snake")+rprBattlefieldCount(g, me.ID, "Wolf")+rprBattlefieldCount(g, me.ID, "Elephant") != 3 {
		t.Error("want exactly one of each token under the caster's control")
	}
}

// --- Growth Spasm ---------------------------------------------------------

func TestRprGrowthSpasmFetchesATappedBasicAndMakesASpawn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Bolt", "Instant"),
	)
	castCatalogSpell(t, g, "Growth Spasm", "Sorcery", rprGrowthSpasm, nil)
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, ids[0])
	passPriorityAroundTable(t, g)
	card, ok := battlefieldCard(g, ids[0])
	if !ok {
		t.Fatal("no basic land fetched")
	}
	if !card.Tapped {
		t.Error("the land entered untapped")
	}
	if findBattlefieldByName(g, "Eldrazi Spawn") == uuid.Nil {
		t.Error("no Eldrazi Spawn token")
	}
}

func TestRprGrowthSpasmStillMakesASpawnWhenTheSearchFindsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me, searchTestLand("Forest", "Basic Land — Forest"))
	castCatalogSpell(t, g, "Growth Spasm", "Sorcery", rprGrowthSpasm, nil)
	passPriorityAroundTable(t, g)
	answerSearchFailToFind(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Forest") != uuid.Nil {
		t.Error("a land was fetched after failing to find")
	}
	if findBattlefieldByName(g, "Eldrazi Spawn") == uuid.Nil {
		t.Error("no Eldrazi Spawn token")
	}
}

// --- Ainok Bond-Kin -------------------------------------------------------

func TestRprAinokBondKinGivesFirstStrikeToCounteredCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kin := b21Push(g, me.ID, "Ainok Bond-Kin", "Creature — Dog Soldier", rprAinokBondKin, 2, 1)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	if hasKeywordOnBattlefield(t, g, bear, "first strike") || hasKeywordOnBattlefield(t, g, kin, "first strike") {
		t.Fatal("first strike before any counter")
	}
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1)
		_ = g.AddCounterForEffect(theirs, game.CounterPlusOne, 1)
	})
	if !hasKeywordOnBattlefield(t, g, bear, "first strike") {
		t.Error("a countered creature you control lacks first strike")
	}
	if hasKeywordOnBattlefield(t, g, theirs, "first strike") {
		t.Error("an opponent's countered creature has first strike")
	}
}

func TestRprAinokBondKinOutlastAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	kin := b21Push(g, me.ID, "Ainok Bond-Kin", "Creature — Dog Soldier", rprAinokBondKin, 2, 1)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, kin, 0, game.ActivateAbilityParams{})
	if got := countersOn(g, kin, game.CounterPlusOne); got != 1 {
		t.Fatalf("%d counters after outlast, want 1", got)
	}
	if !hasKeywordOnBattlefield(t, g, kin, "first strike") {
		t.Error("the Bond-Kin did not give itself first strike")
	}
	// Tapped now: a second outlast is refused.
	if err := g.ActivateCatalogAbility(me.ID, kin, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("outlast activated while tapped")
	}
}

// --- Death's Presence -----------------------------------------------------

func TestRprDeathsPresencePutsLastKnownPowerCountersOnATarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b21Push(g, me.ID, "Death's Presence", "Enchantment", rprDeathsPres, 0, 0)
	fodder := b16Creature(g, me.ID, "Fodder", "Creature — Bear", 3, 3)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(fodder, game.CounterPlusOne, 1) })
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(fodder) })
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 4 {
		t.Errorf("%d counters, want 4 (the dead creature was a 4/4)", got)
	}
}

func TestRprDeathsPresenceIgnoresAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b21Push(g, me.ID, "Death's Presence", "Enchantment", rprDeathsPres, 0, 0)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("%d counters, want 0", got)
	}
}

// --- Heron's Grace Champion -----------------------------------------------

func TestRprHeronsGraceChampionPumpsOtherHumansOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	human := b16Creature(g, me.ID, "Villager", "Creature — Human Peasant", 1, 1)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b16Creature(g, opp.ID, "Their Human", "Creature — Human", 1, 1)
	champ := rprCastCreature(t, g, "Heron's Grace Champion", "Creature — Human Knight", rprHeronChampion, 3, 3)
	if got := effectivePower(t, g, human); got != 2 {
		t.Errorf("your other Human has power %d, want 2", got)
	}
	if !hasKeywordOnBattlefield(t, g, human, "lifelink") {
		t.Error("your other Human lacks lifelink")
	}
	if got := effectivePower(t, g, champ); got != 3 {
		t.Errorf("the Champion's power %d, want 3 (\"other\")", got)
	}
	if got := effectivePower(t, g, bear); got != 2 || hasKeywordOnBattlefield(t, g, bear, "lifelink") {
		t.Errorf("a non-Human was affected (power %d)", got)
	}
	if got := effectivePower(t, g, theirs); got != 1 || hasKeywordOnBattlefield(t, g, theirs, "lifelink") {
		t.Errorf("an opponent's Human was affected (power %d)", got)
	}
}

// --- Dearly Departed ------------------------------------------------------

func TestRprDearlyDepartedBuffsYourHumansFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedGraveyardCard(t, g, "Dearly Departed", "Creature — Spirit", rprDearlyDeparts)
	human := rprCastCreature(t, g, "Villager", "Creature — Human Peasant", "", 1, 1)
	bear := rprCastCreature(t, g, "Bear", "Creature — Bear", "", 2, 2)
	if got := countersOn(g, human, game.CounterPlusOne); got != 1 {
		t.Errorf("a Human entered with %d counters, want 1", got)
	}
	if got := countersOn(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("a non-Human entered with %d counters, want 0", got)
	}
	_ = me
}

func TestRprDearlyDepartedDoesNothingFromTheBattlefieldOrAnOpponentsGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dearly Departed", TypeLine: "Creature — Spirit",
		OracleID: rprDearlyDeparts, Owner: opp.ID, Controller: opp.ID})
	human := rprCastCreature(t, g, "Villager", "Creature — Human Peasant", "", 1, 1)
	if got := countersOn(g, human, game.CounterPlusOne); got != 0 {
		t.Errorf("an opponent's Dearly Departed gave %d counters", got)
	}
}

func TestRprDearlyDepartedTwoCopiesAddTwoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 2; i++ {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dearly Departed", TypeLine: "Creature — Spirit",
			OracleID: rprDearlyDeparts, Owner: me.ID, Controller: me.ID})
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Villager", TypeLine: "Creature — Human Peasant",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	// Two replacements of one entry: the controller orders them (CR 616.1).
	for len(g.PendingChoices) > 0 {
		p := g.PendingChoices[0]
		if p.Kind != game.PendingChoiceReplacementOrder {
			t.Fatalf("unexpected prompt %q", p.Kind)
		}
		if err := g.ResolveReplacementOrder(p.ID, p.Chooser, append([]game.ReplacementEffectID(nil), p.ReplacementEffectIDs...)); err != nil {
			t.Fatalf("ResolveReplacementOrder: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	if got := countersOn(g, id, game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters, want 2", got)
	}
}

// --- Havengul Runebinder --------------------------------------------------

func TestRprHavengulRunebinderExilesACreatureCardForAZombieAndCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	binder := b21Push(g, me.ID, "Havengul Runebinder", "Creature — Human Wizard", rprRunebinder, 2, 2)
	old := b16Creature(g, me.ID, "Old Zombie", "Creature — Zombie", 2, 2, "B")
	opp := g.Seats[1]
	theirs := b16Creature(g, opp.ID, "Their Zombie", "Creature — Zombie", 2, 2, "B")
	card := seedGraveyardCard(t, g, "Dead Bear", "Creature — Bear", "")
	b16Activate(t, g, me.ID, binder, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{card}})
	if !g.Exile.Contains(card) {
		t.Error("the creature card was not exiled")
	}
	token := findBattlefieldByName(g, "Zombie")
	if token == uuid.Nil {
		t.Fatal("no Zombie token")
	}
	if got := countersOn(g, token, game.CounterPlusOne); got != 1 {
		t.Errorf("the new token has %d counters, want 1", got)
	}
	if got := countersOn(g, old, game.CounterPlusOne); got != 1 {
		t.Errorf("your other Zombie has %d counters, want 1", got)
	}
	if got := countersOn(g, theirs, game.CounterPlusOne); got != 0 {
		t.Errorf("an opponent's Zombie got %d counters", got)
	}
	if got := countersOn(g, binder, game.CounterPlusOne); got != 0 {
		t.Errorf("the Wizard is not a Zombie but got %d counters", got)
	}
}

func TestRprHavengulRunebinderNeedsACreatureCardInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	binder := b21Push(g, me.ID, "Havengul Runebinder", "Creature — Human Wizard", rprRunebinder, 2, 2)
	seedGraveyardCard(t, g, "Dead Spell", "Sorcery", "")
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, binder, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with no creature card to exile")
	}
}

// --- Dark Salvation -------------------------------------------------------

func TestRprDarkSalvationMakesXZombiesThenShrinksByTheirCount(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b16Creature(g, opp.ID, "Existing Zombie", "Creature — Zombie", 2, 2, "B")
	victim := b16Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4, "R")
	paCast(t, g, "Dark Salvation", "Sorcery", "{X}{X}{B}", rprDarkSalvation, game.CastSpellParams{
		XValue: 2,
		Targets: []game.TargetRef{
			{Kind: game.TargetPlayer, ID: opp.ID},
			{Kind: game.TargetCard, ID: victim},
		},
	})
	passPriorityAroundTable(t, g)
	zombies := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == opp.ID && c.Name == "Zombie" {
			zombies++
		}
	}
	if zombies != 2 {
		t.Errorf("the targeted player controls %d Zombie tokens, want 2", zombies)
	}
	if got := effectivePower(t, g, victim); got != 1 {
		t.Errorf("the Ogre is %d power, want 4-3 = 1 (three Zombies)", got)
	}
	_ = me
}

func TestRprDarkSalvationWorksOnYourselfWithNoCreatureTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	paCast(t, g, "Dark Salvation", "Sorcery", "{X}{X}{B}", rprDarkSalvation, game.CastSpellParams{
		XValue:  1,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	})
	passPriorityAroundTable(t, g)
	if rprBattlefieldCount(g, me.ID, "Zombie") != 1 {
		t.Error("no Zombie token for the targeted player")
	}
}

// --- Angel of Glory's Rise ------------------------------------------------

func TestRprAngelOfGlorysRiseExilesZombiesThenReturnsHumans(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b16Creature(g, me.ID, "My Zombie", "Creature — Zombie", 2, 2)
	theirs := b16Creature(g, opp.ID, "Their Zombie", "Creature — Zombie", 2, 2)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	human := seedGraveyardCard(t, g, "Dead Human", "Creature — Human Soldier", "")
	zombieHuman := uuid.New() // a Human Zombie card in the graveyard comes back, THEN is not exiled
	me.Graveyard.PushTop(game.Card{InstanceID: zombieHuman, Name: "Dead Zombie Human", TypeLine: "Creature — Zombie Human",
		Owner: me.ID, Controller: me.ID, Power: 1, Toughness: 1})
	bearCard := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bearCard, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	oppHuman := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: oppHuman, Name: "Their Dead Human", TypeLine: "Creature — Human", Owner: opp.ID, Controller: opp.ID})
	rprCastCreature(t, g, "Angel of Glory's Rise", "Creature — Angel", rprAngelGlory, 4, 6)
	if !g.Exile.Contains(mine) || !g.Exile.Contains(theirs) {
		t.Error("not every Zombie was exiled")
	}
	if findBattlefieldByName(g, "Bear") != bear {
		t.Error("a non-Zombie was exiled")
	}
	if findBattlefieldByName(g, "Dead Human") == uuid.Nil {
		t.Error("the Human creature card did not return")
	}
	if findBattlefieldByName(g, "Dead Zombie Human") == uuid.Nil {
		t.Error("a Human that is also a Zombie did not return (the exile came first)")
	}
	if findBattlefieldByName(g, "Dead Bear") != uuid.Nil || findBattlefieldByName(g, "Their Dead Human") != uuid.Nil {
		t.Error("a non-Human or an opponent's card returned")
	}
	_ = human
}

// --- Hour of Eternity -----------------------------------------------------

func TestRprHourOfEternityExilesXCardsAndMakesFourFourZombieCopies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := uuid.New()
	b := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: a, Name: "Elf Scout", TypeLine: "Creature — Elf Scout", Power: 1, Toughness: 1,
		Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: b, Name: "Goblin Raider", TypeLine: "Creature — Goblin", Power: 2, Toughness: 1,
		Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	paCast(t, g, "Hour of Eternity", "Sorcery", "{X}{X}{U}{U}{U}", rprHourEternity, game.CastSpellParams{
		XValue: 2,
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b},
		},
	})
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(a) || !g.Exile.Contains(b) {
		t.Error("the targeted cards were not exiled")
	}
	for _, name := range []string{"Elf Scout", "Goblin Raider"} {
		id := findBattlefieldByName(g, name)
		if id == uuid.Nil {
			t.Fatalf("no token copy of %s", name)
		}
		c := findBattlefieldCardForTest(g, id)
		eff := c.Effective()
		if eff.Power != 4 || eff.Toughness != 4 || len(eff.Colors) != 1 || eff.Colors[0] != "B" || !c.HasSubtype("Zombie") || c.HasSubtype("Elf") || c.HasSubtype("Goblin") {
			t.Errorf("%s copy: %d/%d colors %v type %q, want a 4/4 black Zombie", name, eff.Power, eff.Toughness, eff.Colors, c.TypeLine)
		}
	}
}

func TestRprHourOfEternityRefusesANonCreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spell := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Old Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	_, err := castWithTapParams(t, g, "Hour of Eternity", "Sorcery", "{X}{X}{U}{U}{U}", rprHourEternity, game.CastSpellParams{
		Strict: true, XValue: 1, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}},
	})
	if err == nil {
		t.Fatal("a sorcery card was a legal target")
	}
}
