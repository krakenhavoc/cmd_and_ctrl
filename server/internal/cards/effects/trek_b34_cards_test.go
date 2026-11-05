package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b34_cards_test.go — Star Trek Commander, batch 34 (#4974). Each
// test plays the real catalog entry through the engine and asserts what its
// text says.

const (
	trekB34ReactorOracle  = "fea853c7-7dc8-45c7-829b-d857c65b6ee8"
	trekB34SphereOracle   = "4cfb6f06-137c-4b41-97ef-6710e899473b"
	trekB34OBrienOracle   = "c98d476a-2646-4171-bab7-4a4624fce917"
	trekB34CloneOracle    = "94fee2cf-7604-48fa-9d3c-03ca4af0edfd"
	trekB34AdjutantOracle = "caafecc0-adfa-478c-8a9d-234d4069300b"
	trekB34DataOracle     = "93bfb762-e69e-4d8e-a5b7-e13dd12cd42a"
	trekB34PikeOracle     = "e0a9a7a2-c996-4e3e-817d-9283be690c39"
	trekB34GeordiOracle   = "adbef171-3640-4f77-8c92-b50c109ca303"
	trekB34AmoebaOracle   = "ff9f3d23-a12f-428b-8001-90dafac0c6d7"
	trekB34GiftOracle     = "b44026ae-eaf1-4daf-99c6-2663db674ff7"
)

func TestTrekB34EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB34ReactorOracle:  "Ambulatory Reactor",
		trekB34SphereOracle:   "Borg Scout Sphere",
		trekB34OBrienOracle:   "Chief Miles O'Brien",
		trekB34CloneOracle:    "Colossal Clone Master",
		trekB34AdjutantOracle: "Co-Option Adjutant",
		trekB34DataOracle:     "Data, Exploring Humanity",
		trekB34PikeOracle:     "Fleet Captain Christopher Pike",
		trekB34GeordiOracle:   "Geordi La Forge, Innovator",
		trekB34AmoebaOracle:   "Giant Space Amoeba",
		trekB34GiftOracle:     "Gift of Perfection",
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

// trekB34Draw has p draw one card and settles what it triggered, answering
// a target prompt with pick.
func trekB34Draw(t *testing.T, g *game.Game, p *game.Player, pick *game.TargetRef) {
	t.Helper()
	g.WithWriteLock(func() { _ = g.DrawNForEffect(p.ID, 1) })
	assistCardSettle(t, g, p.ID, pick, true)
}

// trekB34Library replaces p's library with `cards`, the last one on top.
func trekB34Library(p *game.Player, cards ...game.Card) []uuid.UUID {
	ids := make([]uuid.UUID, len(cards))
	p.Library.Cards = nil
	for i, c := range cards {
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = p.ID, p.ID
		p.Library.Cards = append(p.Library.Cards, c)
		ids[i] = c.InstanceID
	}
	return ids
}

// --- Ambulatory Reactor ----------------------------------------------------

func TestTrekB34ReactorGivesArtifactCreaturesRestrictedColorless(t *testing.T) {
	g, me, opp := spendTable(t)
	reactor := b27Push(g, me.ID, "Ambulatory Reactor", "Artifact Creature — Borg", trekB34ReactorOracle, "{3}{U}", 2, 5, "U")
	drone := b12Creature(g, me.ID, "Drone", "Artifact Creature — Borg", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	theirs := b12Creature(g, opp.ID, "Their Drone", "Artifact Creature — Borg", 1, 1)

	for id, want := range map[uuid.UUID]bool{reactor: true, drone: true, bear: false, rock: false, theirs: false} {
		if _, _, got := grantedManaAbilityFor(g, id); got != want {
			t.Errorf("%s has the granted mana ability = %v, want %v", layeredCard(t, g, id).Name, got, want)
		}
	}

	idx, ref, _ := grantedManaAbilityFor(g, drone)
	if err := g.ActivateManaAbility(me.ID, drone, idx, game.ManaAbilityParams{Ref: ref}); err != nil {
		t.Fatalf("tap the Drone for mana: %v", err)
	}
	if !layeredCard(t, g, drone).Tapped {
		t.Error("the {T} cost did not tap the Drone")
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "C" {
		t.Fatalf("pool = %v, want {C}", got)
	}
	cost, _ := game.ParseCost("{1}")
	artifactSpell := game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{1}"}
	creatureSpell := game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}"}
	instant := game.Card{Name: "Opt", TypeLine: "Instant", ManaCost: "{U}"}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(artifactSpell)) {
		t.Error("refused an artifact spell")
	}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(artifactSpell)) {
		t.Error("refused an artifact source's ability")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(creatureSpell)) {
		t.Error("paid for a nonartifact creature spell")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(creatureSpell)) {
		t.Error("paid for a nonartifact source's ability")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(instant)) {
		t.Error("paid for an instant")
	}
}

// --- Borg Scout Sphere -----------------------------------------------------

func TestTrekB34SphereDrawsOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	hand := handSize(me)
	castCatalogSpell(t, g, "Borg Scout Sphere", "Artifact — Spacecraft", trekB34SphereOracle, nil)
	assistCardSettle(t, g, me.ID, nil, true)
	// castCatalogSpell puts the Sphere in the hand and casts it from there,
	// so the hand grows by the one card the enters trigger drew.
	if got := handSize(me); got != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, got)
	}
}

func TestTrekB34SphereGivesArtifactCreaturesFlyingAtSix(t *testing.T) {
	g, me, ship, crew := stationBoard(t, "Borg Scout Sphere", "Artifact — Spacecraft", trekB34SphereOracle, 5)
	drone := b12Creature(g, me.ID, "Drone", "Artifact Creature — Borg", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, "Their Drone", "Artifact Creature — Borg", 1, 1)

	station(t, g, me, ship, crew)
	passPriorityAroundTable(t, g)
	if layeredCard(t, g, ship).IsCreature() || hasEffectiveKeyword(t, g, drone, "flying") {
		t.Fatal("five charge counters already switched the 6+ line on")
	}
	second := b12Creature(g, me.ID, "Second Hand", "Creature — Human", 1, 1)
	station(t, g, me, ship, second)
	passPriorityAroundTable(t, g)
	sphere := layeredCard(t, g, ship)
	if !sphere.IsCreature() || sphere.CurrentPower() != 5 || sphere.CurrentToughness() != 5 {
		t.Fatalf("at 6+ the Sphere is a 5/5 artifact creature; creature=%v %d/%d",
			sphere.IsCreature(), sphere.CurrentPower(), sphere.CurrentToughness())
	}
	for id, want := range map[uuid.UUID]bool{ship: true, drone: true, bear: false, theirs: false} {
		if got := hasEffectiveKeyword(t, g, id, "flying"); got != want {
			t.Errorf("%s flying = %v, want %v", layeredCard(t, g, id).Name, got, want)
		}
	}
}

// --- Chief Miles O'Brien ---------------------------------------------------

func TestTrekB34OBrienDoublesEachCounterKindOnTheSecondDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	pushCatalogPermanent(g, me.ID, "Chief Miles O'Brien", "Legendary Creature — Human Artificer", trekB34OBrienOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if c := &g.Battlefield.Cards[i]; c.InstanceID == bear || c.InstanceID == theirs {
				c.Counters = map[string]int{game.CounterPlusOne: 2, "oil": 1}
			}
		}
	})

	trekB34Draw(t, g, opp, nil)
	trekB34Draw(t, g, opp, nil)
	pick := cardRef(bear)
	trekB34Draw(t, g, me, &pick)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 2 {
		t.Fatalf("the first draw changed the counters: %d +1/+1", got)
	}

	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt on the second draw")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, cardRef(theirs)); err == nil {
		t.Error("targeted a creature an opponent controls")
	}
	assistCardSettle(t, g, me.ID, &pick, true)
	if a, b := countersOn(g, bear, game.CounterPlusOne), countersOn(g, bear, "oil"); a != 4 || b != 2 {
		t.Errorf("after the second draw: %d +1/+1 and %d oil, want 4 and 2", a, b)
	}
	if got := countersOn(g, theirs, game.CounterPlusOne); got != 2 {
		t.Errorf("the opponent's creature changed: %d +1/+1", got)
	}

	trekB34Draw(t, g, me, &pick)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 4 {
		t.Errorf("the third draw doubled again: %d +1/+1", got)
	}
}

// --- Colossal Clone Master -------------------------------------------------

func TestTrekB34CloneMasterCopiesWithoutLegendAndCountsNamedLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	legend := seedCopyableCreature(g, me.ID, "Big Legend", "Legendary Creature — Human Soldier", 7, 7)
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	b12Permanent(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, "Swamp", "Basic Land — Swamp")

	id := castCatalogSpell(t, g, "Colossal Clone Master", "Creature — Giant", trekB34CloneOracle, nil)
	resolveWithCopyChoice(t, g, legend)
	g.RunStateChecksForTest()

	if !g.Battlefield.Contains(id) || !g.Battlefield.Contains(legend) {
		t.Fatal("the legend rule took one of them: the copy is not legendary")
	}
	c := layeredCard(t, g, id)
	if c.Name != "Big Legend" || c.HasSupertype("Legendary") || !c.HasSubtype("Soldier") {
		t.Errorf("copy = %q legendary=%v soldier=%v, want a nonlegendary Big Legend", c.Name, c.HasSupertype("Legendary"), c.HasSubtype("Soldier"))
	}
	if c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("P/T %d/%d, want 2/2 (Forest and Island; the opponent's Swamp does not count)", c.CurrentPower(), c.CurrentToughness())
	}
	b12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	if c := layeredCard(t, g, id); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("P/T %d/%d after a third name, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	if c := layeredCard(t, g, legend); c.CurrentPower() != 7 {
		t.Errorf("the original's power changed to %d", c.CurrentPower())
	}
}

func TestTrekB34CloneMasterThatDeclinesTheCopyDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	// A deck-imported card (it has a printing), so its printed 0/0 is real
	// and CR 704.5f applies, as for Phantasmal Image.
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Colossal Clone Master", TypeLine: "Creature — Giant", OracleID: trekB34CloneOracle,
		ScryfallID: "00000000-0000-0000-0000-0000000000b4", Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	resolveWithCopyChoice(t, g, uuid.Nil)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(id) {
		t.Error("a declined Clone Master is a 0/0 and dies (the land count is the copy's ability only)")
	}
}

// --- Co-Option Adjutant ----------------------------------------------------

func TestTrekB34AdjutantArtifactCreaturesDrawOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	adjutant := b27Push(g, me.ID, "Co-Option Adjutant", "Artifact Creature — Borg", trekB34AdjutantOracle, "{3}{U}", 3, 3, "U")
	drone := b12Creature(g, me.ID, "Drone", "Artifact Creature — Construct", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := handSize(me)
	trekB12HiveMindSwing(t, g, adjutant, drone, bear)
	assistCardSettle(t, g, me.ID, nil, true)
	if got := handSize(me); got != hand+2 {
		t.Errorf("hand %d -> %d, want two draws (the Adjutant and the Drone, not the Bear)", hand, got)
	}
}

// --- Data, Exploring Humanity ----------------------------------------------

func TestTrekB34DataDiscoversFourOnTheSecondDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Data, Exploring Humanity", "Legendary Artifact Creature — Android Scientist", trekB34DataOracle, false)
	hit := discoverLibrary(me, 4)
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})

	trekB34Draw(t, g, g.Seats[2], nil)
	trekB34Draw(t, g, me, nil)
	if c := discoverPromptFor(g, me.ID); c != nil {
		t.Fatal("the first draw discovered")
	}
	trekB34Draw(t, g, me, nil)
	expectDiscovered(t, g, me.ID, hit)
}

// --- Fleet Captain Christopher Pike ----------------------------------------

func TestTrekB34PikeUntapsAndTapsTwoOtherCreatures(t *testing.T) {
	g, me, opp := spendTable(t)
	pike := pushCatalogPermanent(g, me.ID, "Fleet Captain Christopher Pike", "Legendary Creature — Human Officer", trekB34PikeOracle, false)
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	c := b12Creature(g, opp.ID, "C", "Creature — Bear", 2, 2)
	setTapped := func(id uuid.UUID, tapped bool) {
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == id {
					g.Battlefield.Cards[i].Tapped = tapped
				}
			}
		})
	}
	setTapped(a, true)
	setTapped(b, true)

	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, pike, 0, game.ActivateAbilityParams{Targets: cardRefs(pike, a), Strict: true}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("Yes targeting Pike himself: %v, want an illegal target", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, pike, 0, game.ActivateAbilityParams{Targets: cardRefs(a), Strict: true}); err == nil {
		t.Fatal("Yes activated with one target")
	}
	b16Activate(t, g, me.ID, pike, 0, game.ActivateAbilityParams{Targets: cardRefs(a, b), Strict: true})
	if layeredCard(t, g, a).Tapped || layeredCard(t, g, b).Tapped {
		t.Error("Yes did not untap both targets")
	}
	if !layeredCard(t, g, pike).Tapped {
		t.Error("Pike did not tap for the cost")
	}

	setTapped(pike, false)
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, pike, 1, game.ActivateAbilityParams{Targets: cardRefs(b, c), Strict: true}); err == nil {
		t.Fatal("No activated with one mana")
	}
	floatMana(t, g, me, "{C}")
	b16Activate(t, g, me.ID, pike, 1, game.ActivateAbilityParams{Targets: cardRefs(b, c), Strict: true})
	if !layeredCard(t, g, b).Tapped || !layeredCard(t, g, c).Tapped || layeredCard(t, g, a).Tapped {
		t.Error("No should tap exactly its two targets")
	}
}

// --- Geordi La Forge, Innovator --------------------------------------------

func TestTrekB34GeordiIsUnblockableWithASpacecraft(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	geordi := pushCatalogPermanent(g, me.ID, "Geordi La Forge, Innovator", "Legendary Creature — Human Artificer", trekB34GeordiOracle, false)
	blocker := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 8)
	b12Permanent(g, opp.ID, "Their Ship", "Artifact — Spacecraft")
	b12Permanent(g, me.ID, "Shuttle", "Artifact — Spacecraft")
	brAttack(t, g, geordi)
	brRefusal(t, g.DeclareBlocker(blocker, geordi), game.BlockReasonCantBeBlocked)
}

func TestTrekB34GeordiIsBlockableWithoutASpacecraftAndDrawsOnDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	geordi := pushCatalogPermanent(g, me.ID, "Geordi La Forge, Innovator", "Legendary Creature — Human Artificer", trekB34GeordiOracle, false)
	blocker := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 8)
	b12Permanent(g, opp.ID, "Their Ship", "Artifact — Spacecraft")
	brAttack(t, g, geordi)
	if err := g.DeclareBlocker(blocker, geordi); err != nil {
		t.Fatalf("only an opponent controls a Spacecraft, so Geordi is blockable: %v", err)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	advanceToMain(t, g2)
	geordi2 := pushCatalogPermanent(g2, me2.ID, "Geordi La Forge, Innovator", "Legendary Creature — Human Artificer", trekB34GeordiOracle, false)
	hand := handSize(me2)
	trekB12HiveMindSwing(t, g2, geordi2)
	assistCardSettle(t, g2, me2.ID, nil, true)
	if got := handSize(me2); got != hand+1 {
		t.Errorf("hand %d -> %d, want one draw for the combat damage", hand, got)
	}
}

// --- Giant Space Amoeba ----------------------------------------------------

func TestTrekB34AmoebaShrinksADefenderAndCopiesItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	amoeba := b27Push(g, me.ID, "Giant Space Amoeba", "Creature — Germ", trekB34AmoebaOracle, "{4}{U}", 3, 3, "U")
	theirs := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	elsewhere := b12Creature(g, other.ID, "Other Ogre", "Creature — Ogre", 4, 4)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(amoeba, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the attack trigger")
	}
	for _, bad := range []uuid.UUID{elsewhere, mine} {
		if err := g.ResolvePickTarget(p.ID, me.ID, cardRef(bad)); err == nil {
			t.Errorf("targeted %s, which the defending player does not control", layeredCard(t, g, bad).Name)
		}
	}
	pick := cardRef(theirs)
	assistCardSettle(t, g, me.ID, &pick, true)

	if c := layeredCard(t, g, theirs); c.CurrentPower() != 1 || c.CurrentToughness() != 4 {
		t.Errorf("the target is %d/%d, want 1/4", c.CurrentPower(), c.CurrentToughness())
	}
	var copies []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.InstanceID != amoeba && c.Name == "Giant Space Amoeba" {
			copies = append(copies, layeredCard(t, g, c.InstanceID))
		}
	}
	if len(copies) != 1 {
		t.Fatalf("%d copies, want one token copy", len(copies))
	}
	if cp := copies[0]; !cp.IsToken() || !game.HasKeyword(&cp, "flying") || cp.CurrentPower() != 3 {
		t.Errorf("the copy: token=%v flying=%v power=%d, want a 3/3 flying token", cp.IsToken(), game.HasKeyword(&cp, "flying"), cp.CurrentPower())
	}
}

// --- Gift of Perfection ----------------------------------------------------

func trekB34LibraryOf(g *game.Game, p *game.Player) []uuid.UUID {
	var out []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range p.Library.Cards {
			out = append(out, c.InstanceID)
		}
	})
	return out
}

func TestTrekB34GiftAssimilatesTheChosenCreatureAndBottomsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	n := len(g.Seats)
	o1, o2, o3 := g.Seats[(g.Turn.ActiveSeat+1)%n], g.Seats[(g.Turn.ActiveSeat+2)%n], g.Seats[(g.Turn.ActiveSeat+3)%n]
	land := game.Card{Name: "Island", TypeLine: "Basic Land — Island"}
	elf := game.Card{Name: "Elf", TypeLine: "Creature — Elf Warrior", ManaCost: "{G}", Power: 1, Toughness: 1}
	ogre := game.Card{Name: "Ogre", TypeLine: "Creature — Ogre", ManaCost: "{2}{R}", Power: 3, Toughness: 3}
	l1 := trekB34Library(o1, land, elf, land, land) // top: land, land, then the Elf
	l2 := trekB34Library(o2, land, ogre)            // top: the Ogre
	l3 := trekB34Library(o3, land, land)            // no creature

	castCatalogSpell(t, g, "Gift of Perfection", "Sorcery", trekB34GiftOracle, nil)
	passPriorityAroundTable(t, g)
	rp := revealPickFor(g, me.ID)
	if rp == nil {
		t.Fatal("no pick among the revealed creature cards")
	}
	if len(rp.ChooseCards) != 2 || rp.ChooseMin != 1 || rp.ChooseMax != 1 {
		t.Fatalf("pick offers %d cards, %d..%d; want exactly one of the two creature cards", len(rp.ChooseCards), rp.ChooseMin, rp.ChooseMax)
	}
	if err := g.ResolveRevealPick(rp.ID, me.ID, []uuid.UUID{l1[2]}); err == nil {
		t.Error("picked a revealed land")
	}
	if err := g.ResolveRevealPick(rp.ID, me.ID, []uuid.UUID{l1[1]}); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	assistCardSettle(t, g, me.ID, nil, true)

	got := layeredCard(t, g, l1[1])
	if got.Controller != me.ID || !got.IsArtifact() || !got.IsCreature() || !got.HasSubtype("Borg") || got.HasSubtype("Elf") || got.HasSubtype("Warrior") {
		t.Errorf("the Elf: controller me=%v artifact=%v borg=%v elf=%v, want a Borg artifact creature under the caster",
			got.Controller == me.ID, got.IsArtifact(), got.HasSubtype("Borg"), got.HasSubtype("Elf"))
	}
	if got.Counters[game.CounterPlusOne] != 1 || got.CurrentPower() != 2 {
		t.Errorf("the Elf has %d +1/+1 counters and power %d, want 1 and 2", got.Counters[game.CounterPlusOne], got.CurrentPower())
	}
	// Opponent one: the two lands above the Elf went to the bottom, under
	// the land that was never revealed.
	if lib := trekB34LibraryOf(g, o1); len(lib) != 3 || lib[2] != l1[0] {
		t.Errorf("opponent one's library = %v, want the unrevealed land on top of the two revealed ones", lib)
	}
	// Opponent two: the Ogre that was not picked is on the bottom.
	if lib := trekB34LibraryOf(g, o2); len(lib) != 2 || lib[0] != l2[1] || lib[1] != l2[0] {
		t.Errorf("opponent two's library = %v, want the Ogre under the land", lib)
	}
	if g.Battlefield.Contains(l2[1]) {
		t.Error("the Ogre was not picked and must not enter")
	}
	if lib := trekB34LibraryOf(g, o3); len(lib) != 2 || len(l3) != 2 {
		t.Errorf("opponent three's library has %d cards, want both revealed lands back", len(lib))
	}
}

func TestTrekB34GiftWithOneCreatureNeedsNoChoice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	n := len(g.Seats)
	land := game.Card{Name: "Island", TypeLine: "Basic Land — Island"}
	ogre := game.Card{Name: "Ogre", TypeLine: "Creature — Ogre", ManaCost: "{2}{R}", Power: 3, Toughness: 3}
	var ogreID uuid.UUID
	for i := 1; i < n; i++ {
		p := g.Seats[(g.Turn.ActiveSeat+i)%n]
		if i == 2 {
			ogreID = trekB34Library(p, land, ogre)[1]
		} else {
			trekB34Library(p, land)
		}
	}
	castCatalogSpell(t, g, "Gift of Perfection", "Sorcery", trekB34GiftOracle, nil)
	assistCardSettle(t, g, me.ID, nil, true)
	if revealPickFor(g, me.ID) != nil {
		t.Fatal("asked to choose among one card")
	}
	if c := layeredCard(t, g, ogreID); c.Controller != me.ID || !c.HasSubtype("Borg") || c.HasSubtype("Ogre") {
		t.Errorf("the Ogre: under me=%v borg=%v ogre=%v", c.Controller == me.ID, c.HasSubtype("Borg"), c.HasSubtype("Ogre"))
	}
}
