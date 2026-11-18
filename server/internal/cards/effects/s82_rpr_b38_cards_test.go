package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_rpr_b38_cards_test.go — reprint pass batch 38 (#5908): Cytoplast
// Manipulator, Plague of Vermin, Silverclad Ferocidons, Sword of Sinew and
// Steel, Thada Adel, Ashcoat of the Shadow Swarm and Sisay, each driven
// through a real cast, activation or trigger.

const (
	cytoplastManipulatorOracle = "026ef98e-4f0a-4a6d-b74f-d17ac8b10e37"
	plagueOfVerminOracle       = "efc12acf-ffef-40a6-b2ce-fe93e89ad840"
	silvercladFerocidonsOracle = "b3aa5889-0d98-4a8c-a8b1-ada1e9b169e1"
	swordSinewSteelOracle      = "ccab4509-f189-4610-9597-547e7f6b0775"
	thadaAdelOracle            = "a9a84fbb-0972-4a3d-bdae-a3b01219b35d"
	ashcoatOracle              = "9cc69ea5-42a2-4306-ac6d-cff5adb20bcb"
	sisayOracle                = "fb777610-7562-4d9d-8497-62c562e6d7cb"
)

func TestS82RprB38AllSevenAreFullWithNoCaveats(t *testing.T) {
	for _, id := range []string{cytoplastManipulatorOracle, plagueOfVerminOracle, silvercladFerocidonsOracle,
		swordSinewSteelOracle, thadaAdelOracle, ashcoatOracle, sisayOracle} {
		spec, ok := Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Cytoplast Manipulator ---

func TestS82RprB38CytoplastManipulatorEntersWithTwoCountersAndGraftsOntoANewCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	addPool(t, g, me.ID, "{U}{U}{C}{C}")
	id := castCatalogSpell(t, g, "Cytoplast Manipulator", "Creature — Human Wizard Mutant", cytoplastManipulatorOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Manipulator did not resolve")
	}
	if n := countersOn(g, id, game.CounterPlusOne); n != 2 {
		t.Fatalf("%d +1/+1 counters on entry, want 2", n)
	}
	if p := effectivePowerWithCounters(t, g, id); p != 2 {
		t.Errorf("power %d, want 2 (0/0 with two counters)", p)
	}
	bear := castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if countersOn(g, id, game.CounterPlusOne) != 1 || countersOn(g, bear, game.CounterPlusOne) != 1 {
		t.Errorf("graft moved a counter: manipulator %d, bear %d; want 1 and 1",
			countersOn(g, id, game.CounterPlusOne), countersOn(g, bear, game.CounterPlusOne))
	}
}

func TestS82RprB38CytoplastManipulatorStealsACreatureWithACounterWhileItRemains(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	manip := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Cytoplast Manipulator", TypeLine: "Creature — Human Wizard Mutant",
		OracleID: cytoplastManipulatorOracle, Counters: map[string]int{game.CounterPlusOne: 2},
		Owner: me.ID, Controller: me.ID,
	})
	withCounter := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Counter Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Counters: map[string]int{game.CounterPlusOne: 1}, Owner: opp.ID, Controller: opp.ID,
	})
	plain := b12Creature(g, opp.ID, "Plain Bear", "Creature — Bear", 2, 2)
	b06AddMana(me, "U")
	if err := g.ActivateCatalogAbility(me.ID, manip, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(plain)}); err == nil {
		t.Fatal("a creature without a +1/+1 counter is not a legal target")
	}
	b16Activate(t, g, me.ID, manip, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(withCounter)})
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, withCounter); got != me.ID {
		t.Fatalf("controller %v, want the Manipulator's controller", got)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(manip); err != nil {
			t.Fatal(err)
		}
	})
	if got := controllerOf(t, g, withCounter); got != opp.ID {
		t.Errorf("controller %v after the Manipulator left, want the owner back", got)
	}
}

// --- Plague of Vermin ---

func TestS82RprB38PlagueOfVerminRepeatsUntilNoonePaysThenEachPlayerMakesTheirRats(t *testing.T) {
	g := newCatalogGame(t)
	me, o1 := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	addPool(t, g, me.ID, "{B}{C}{C}{C}{C}{C}{C}")
	castCatalogSpell(t, g, "Plague of Vermin", "Sorcery", plagueOfVerminOracle, nil)
	passPriorityAroundTable(t, g)
	answers := map[uuid.UUID][]int{me.ID: {3, 0}, o1.ID: {2, 0}}
	for i := 0; i < 20; i++ {
		var c *game.PendingChoice
		for _, pc := range g.PendingChoices {
			if pc != nil && pc.PayAmount != nil {
				c = pc
				break
			}
		}
		if c == nil {
			break
		}
		ans := 0
		if q := answers[c.Chooser]; len(q) > 0 {
			ans, answers[c.Chooser] = q[0], q[1:]
		}
		if err := g.ResolvePayAmount(c.ID, c.Chooser, ans); err != nil {
			t.Fatalf("pay %d: %v", ans, err)
		}
	}
	passPriorityAroundTable(t, g)
	if me.Life != 37 || o1.Life != 38 {
		t.Errorf("life %d and %d, want 37 and 38", me.Life, o1.Life)
	}
	if n := b43TokensNamed(g, me.ID, "Rat"); n != 3 {
		t.Errorf("your Rats = %d, want 3", n)
	}
	if n := b43TokensNamed(g, o1.ID, "Rat"); n != 2 {
		t.Errorf("their Rats = %d, want 2", n)
	}
	if n := b43TokensNamed(g, g.Seats[2].ID, "Rat"); n != 0 {
		t.Errorf("a player who paid nothing made %d Rats", n)
	}
}

func TestS82RprB38PlagueOfVerminWhenNobodyPaysMakesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	addPool(t, g, me.ID, "{B}{C}{C}{C}{C}{C}{C}")
	castCatalogSpell(t, g, "Plague of Vermin", "Sorcery", plagueOfVerminOracle, nil)
	passPriorityAroundTable(t, g)
	for i := 0; i < 20; i++ {
		var c *game.PendingChoice
		for _, pc := range g.PendingChoices {
			if pc != nil && pc.PayAmount != nil {
				c = pc
				break
			}
		}
		if c == nil {
			break
		}
		if err := g.ResolvePayAmount(c.ID, c.Chooser, 0); err != nil {
			t.Fatal(err)
		}
	}
	for _, s := range g.Seats {
		if n := b43TokensNamed(g, s.ID, "Rat"); n != 0 || s.Life != 40 {
			t.Errorf("seat %v: %d Rats, life %d; want none and 40", s.ID, n, s.Life)
		}
	}
}

// --- Silverclad Ferocidons ---

func TestS82RprB38SilvercladFerocidonsEnrageMakesEachOpponentSacrificeAPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, o1 := g.Seats[0], g.Seats[1]
	dino := b41CatalogCreature(g, me.ID, "Silverclad Ferocidons", "Creature — Dinosaur", silvercladFerocidonsOracle, 8, 5)
	mine := b12Permanent(g, me.ID, "My Rock", "Artifact")
	theirs := b12Permanent(g, o1.ID, "Their Rock", "Artifact")
	g.WithWriteLock(func() {
		if err := (DealDamage{Source: theirs, Target: dino, Amount: 1}).Apply(NewContext(g, &game.StackItem{Controller: o1.ID, SourceCardID: theirs})); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	for i := 0; i < 6; i++ {
		var c *game.PendingChoice
		for _, pc := range g.PendingChoices {
			if pc != nil && pc.Kind == game.PendingChoiceChooseCards {
				c = pc
				break
			}
		}
		if c == nil {
			break
		}
		if err := g.ResolveChooseCards(c.ID, c.Chooser, c.ChooseCards[:1]); err != nil {
			t.Fatal(err)
		}
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(dino) {
		t.Error("the controller sacrifices nothing")
	}
	if got := len(g.Seats[1].Graveyard.Cards); got != 1 {
		t.Errorf("opponent graveyard holds %d cards, want the one sacrificed permanent", got)
	}
}

// --- Sword of Sinew and Steel ---

func TestS82RprB38SwordOfSinewAndSteelGivesProtectionAndDestroysAPlaneswalkerAndAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	sword := b41Catalog(g, me.ID, "Sword of Sinew and Steel", "Artifact — Equipment", swordSinewSteelOracle)
	pw := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: 3}, Owner: opp.ID, Controller: opp.ID,
	})
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	other := b12Permanent(g, opp.ID, "Their Land", "Land")
	advanceToMain(t, g)
	b06AddMana(me, "C", "C")
	b16Activate(t, g, me.ID, sword, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("+2/+2: power %d, want 4", got)
	}
	if got := protectionsOn(t, g, bear); !hasString(got, "black") || !hasString(got, "red") {
		t.Errorf("protections %q, want black and red", got)
	}
	dealCombatDamageToPlayer(g, bear, opp.ID, 4)
	passPriorityAroundTable(t, g)
	// Each "up to one" clause is its own pick.
	answerTriggerTargets(t, g, me.ID, pw)
	answerTriggerTargets(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(pw) || g.Battlefield.Contains(rock) {
		t.Error("both the planeswalker and the artifact should be destroyed")
	}
	if !g.Battlefield.Contains(other) {
		t.Error("only a planeswalker and an artifact are destroyed")
	}
}

// --- Thada Adel, Acquisitor ---

func TestS82RprB38ThadaAdelExilesAnArtifactFromTheLibraryAndLetsYouPlayItThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thada := b41CatalogCreature(g, me.ID, "Thada Adel, Acquisitor", "Legendary Creature — Merfolk Rogue", thadaAdelOracle, 2, 2)
	art := seedLibraryTop(opp, "Free Rock", "Artifact", "")
	land := seedLibraryTop(opp, "Plain Forest", "Basic Land — Forest", "")
	advanceToMain(t, g)
	dealCombatDamageToPlayer(g, thada, opp.ID, 2)
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, art)
	passPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(art); z == nil || z.Kind != game.ZoneExile {
		t.Fatalf("the artifact is in %v, want exile", z)
	}
	if opp.Library.Contains(land) == false {
		t.Error("the land stays in their library")
	}
	if g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		advanceTo(t, g, game.StepPostcombatMain)
	}
	if err := g.CastSpell(me.ID, art, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Errorf("casting the exiled artifact this turn: %v", err)
	}
}

// --- Ashcoat of the Shadow Swarm ---

func TestS82RprB38AshcoatPumpsOtherRatsByTheNumberOfRatsWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ash := b41CatalogCreature(g, me.ID, "Ashcoat of the Shadow Swarm", "Legendary Creature — Rat Warlock", ashcoatOracle, 3, 4)
	rat := b12Creature(g, me.ID, "Rat", "Creature — Rat", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirRat := b12Creature(g, opp.ID, "Their Rat", "Creature — Rat", 1, 1)
	declareAttack(t, g, opp.ID, ash)
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, rat); p != 3 {
		t.Errorf("another Rat has power %d, want 3 (+2/+2: two Rats)", p)
	}
	if p := effectivePower(t, g, ash); p != 3 {
		t.Errorf("Ashcoat has power %d, want 3 (only other Rats)", p)
	}
	if effectivePower(t, g, bear) != 2 || effectivePower(t, g, theirRat) != 1 {
		t.Error("a non-Rat and an opponent's Rat are not pumped")
	}
}

func TestS82RprB38AshcoatMillsFourAtEndStepThenReturnsUpToTwoRats(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b41CatalogCreature(g, me.ID, "Ashcoat of the Shadow Swarm", "Legendary Creature — Rat Warlock", ashcoatOracle, 3, 4)
	var rats []uuid.UUID
	for i := 0; i < 3; i++ {
		id := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Dead Rat", TypeLine: "Creature — Rat", Owner: me.ID, Controller: me.ID})
		rats = append(rats, id)
	}
	bear := pushGraveCreature(me, "Dead Bear")
	before := me.Library.Size()
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	if got := before - me.Library.Size(); got != 4 {
		t.Fatalf("milled %d, want 4", got)
	}
	answerChooseCards(t, g, me.ID, rats[0], rats[1])
	if !me.Hand.Contains(rats[0]) || !me.Hand.Contains(rats[1]) {
		t.Error("the two chosen Rats did not reach the hand")
	}
	if me.Hand.Contains(rats[2]) || me.Hand.Contains(bear) {
		t.Error("only up to two Rat creature cards return")
	}
}

func TestS82RprB38AshcoatDecliningTheMillReturnsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b41CatalogCreature(g, me.ID, "Ashcoat of the Shadow Swarm", "Legendary Creature — Rat Warlock", ashcoatOracle, 3, 4)
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Dead Rat", TypeLine: "Creature — Rat", Owner: me.ID, Controller: me.ID})
	before := me.Library.Size()
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	if me.Library.Size() != before || me.Hand.Contains(id) {
		t.Error("declining the mill must neither mill nor return")
	}
}

// --- Sisay, Weatherlight Captain ---

func TestS82RprB38SisayGrowsByTheColorsAmongOtherLegendaryPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sisay := b41CatalogCreature(g, me.ID, "Sisay, Weatherlight Captain", "Legendary Creature — Human Soldier", sisayOracle, 2, 2)
	if p := effectivePower(t, g, sisay); p != 2 {
		t.Fatalf("alone: power %d, want 2 (her own colour doesn't count)", p)
	}
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Gold Legend", TypeLine: "Legendary Creature — Test", Colors: []string{"U", "B"},
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Blue Legend", TypeLine: "Legendary Artifact", Colors: []string{"U"},
		Owner: me.ID, Controller: me.ID,
	})
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Legend", TypeLine: "Legendary Creature — Test", Colors: []string{"R"},
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Plain Green", TypeLine: "Creature — Test", Colors: []string{"G"},
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	if p := effectivePower(t, g, sisay); p != 4 {
		t.Errorf("power %d, want 4 (blue and black among your other legends)", p)
	}
}

func TestS82RprB38SisayFetchesALegendWithManaValueLessThanHerPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	sisay := b41CatalogCreature(g, me.ID, "Sisay, Weatherlight Captain", "Legendary Creature — Human Soldier", sisayOracle, 2, 2)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Gold Legend", TypeLine: "Legendary Creature — Test", Colors: []string{"U", "B"},
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	}) // Sisay is now 4/4
	ok := game.Card{InstanceID: uuid.New(), Name: "Three Drop", TypeLine: "Legendary Creature — Test", ManaCost: "{2}{U}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID}
	big := game.Card{InstanceID: uuid.New(), Name: "Four Drop", TypeLine: "Legendary Creature — Test", ManaCost: "{3}{U}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID}
	nonlegend := game.Card{InstanceID: uuid.New(), Name: "One Drop", TypeLine: "Creature — Test", ManaCost: "{U}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID}
	me.Library.PushTop(ok)
	me.Library.PushTop(big)
	me.Library.PushTop(nonlegend)
	addPool(t, g, me.ID, "{W}{U}{B}{R}{G}")
	b16Activate(t, g, me.ID, sisay, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	pick := searchChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no search prompt")
	}
	for _, id := range pick.ChooseCards {
		if id == big.InstanceID || id == nonlegend.InstanceID {
			t.Errorf("offered %v, which is not a legendary permanent with mana value less than 4", id)
		}
	}
	answerSearchByID(t, g, me.ID, ok.InstanceID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ok.InstanceID) {
		t.Error("the chosen legend should be on the battlefield")
	}
}
