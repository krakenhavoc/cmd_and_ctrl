package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	neoB13Swift      = "5d47e820-913f-441a-a6cc-37ab3181d79a"
	neoB13Mech       = "f434e266-d9d6-4058-96ff-a1bd4029d8dc"
	neoB13Kotori     = "0659a5b0-9807-4e4c-8af3-b6e233051ad6"
	neoB13Katsumasa  = "80ba76a6-cc48-4820-bc0b-c9ec18088a3b"
	neoB13SuitUp     = "29bbbbfe-77cb-411f-9c1b-f7e4724685c6"
	neoB13Hangar     = "abc04775-171d-41f3-83ea-4b4eb72723d5"
	neoB13Core       = "f1eb489d-104c-4801-b6d2-3f1a7ee73a75"
	neoB13Greasefang = "0129ffe2-389f-4e05-8a04-96461bac39a5"
)

func TestNeoB13EveryRegisteredCardIsFull(t *testing.T) {
	for oracle, name := range map[string]string{
		neoB13Swift: "Swift Reconfiguration", neoB13Mech: "Imposter Mech", neoB13Kotori: "Kotori, Pilot Prodigy",
		neoB13Katsumasa: "Katsumasa, the Animator", neoB13SuitUp: "Suit Up", neoB13Hangar: "Mech Hangar",
		neoB13Core: "Mechtitan Core", neoB13Greasefang: "Greasefang, Okiba Boss",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func neoB13Vehicle(g *game.Game, owner uuid.UUID, name string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Artifact — Vehicle",
		Power: power, Toughness: toughness, Owner: owner, Controller: owner})
}

// --- Swift Reconfiguration ---

func neoB13CastSwift(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) error {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Swift Reconfiguration", TypeLine: "Enchantment — Aura", OracleID: neoB13Swift,
			ManaCost: "{W}", Colors: []string{"W"}, Keywords: []string{"flash"}, Owner: me.ID, Controller: me.ID})
	})
	floatForTest(g, me, "W")
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: b16TargetCard(target)})
}

func TestNeoB13SwiftReconfigurationMakesACreatureACrew5Vehicle(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	big := b12Creature(g, me.ID, "Big", "Creature — Giant", 5, 5)
	land := pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	advanceToMain(t, g)
	if err := neoB13CastSwift(t, g, me, land); err == nil {
		t.Fatal("enchanted a land")
	}
	if err := neoB13CastSwift(t, g, me, bear); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, bear)
	e := c.Effective()
	if len(e.Types) != 1 || e.Types[0] != "Artifact" || c.IsCreature() {
		t.Fatalf("types %v, want only Artifact", e.Types)
	}
	if !c.HasSubtype("Vehicle") || c.HasSubtype("Bear") {
		t.Errorf("subtypes %v, want only Vehicle", e.Subtypes)
	}
	idx, ref := beGrantedRow(t, g, bear, "Crew 5")
	if idx < 0 {
		t.Fatal("no crew 5")
	}
	// Crew 5 is paid with power 5: the 5/5 taps and the Vehicle is a creature again.
	if err := g.ActivateCatalogAbility(me.ID, bear, idx, game.ActivateAbilityParams{Strict: true, Ref: ref, CrewIDs: []uuid.UUID{big}}); err != nil {
		t.Fatalf("crew 5: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, bear); !c.IsCreature() || !c.IsArtifact() {
		t.Errorf("the crewed Vehicle is %v", c.Effective().Types)
	}
	if b, _ := battlefieldCard(g, big); !b.Tapped {
		t.Error("the crew was not tapped")
	}
}

func TestNeoB13SwiftReconfigurationCanEnchantAVehicleAndKeepsItsOtherAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cart := neoB13Vehicle(g, me.ID, "Cart", 3, 3)
	advanceToMain(t, g)
	if err := neoB13CastSwift(t, g, me, cart); err != nil {
		t.Fatalf("a Vehicle is a legal host: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, cart); !c.HasSubtype("Vehicle") || c.IsCreature() {
		t.Errorf("types %v", c.Effective().Types)
	}
	if idx, _ := beGrantedRow(t, g, cart, "Crew 5"); idx < 0 {
		t.Error("no crew 5 on the enchanted Vehicle")
	}
}

// --- Kotori, Pilot Prodigy ---

func TestNeoB13KotoriGivesYourVehiclesCrew2AndPumpsAnArtifactCreatureAtCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Kotori, Pilot Prodigy", "Legendary Creature — Moonfolk Pilot", neoB13Kotori, false)
	cart := neoB13Vehicle(g, me.ID, "Cart", 3, 3)
	theirs := neoB13Vehicle(g, opp.ID, "Their Cart", 3, 3)
	golem := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Golem", TypeLine: "Artifact Creature — Golem",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if idx, _ := beGrantedRow(t, g, cart, "Crew 2"); idx < 0 {
		t.Error("my Vehicle has no crew 2")
	}
	if idx, _ := beGrantedRow(t, g, theirs, "Crew 2"); idx >= 0 {
		t.Error("an opponent's Vehicle has crew 2")
	}
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		pickTriggerTarget(t, g, me.ID, golem)
		passPriorityAroundTable(t, g)
	}
	for _, kw := range []string{"lifelink", "vigilance"} {
		if !hasEffectiveKeyword(t, g, golem, kw) {
			t.Errorf("the artifact creature lacks %s", kw)
		}
	}
}

// --- Suit Up ---

func neoB13CastSuitUp(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Suit Up", TypeLine: "Instant", OracleID: neoB13SuitUp,
			ManaCost: "{2}{U}", Colors: []string{"U"}, Owner: me.ID, Controller: me.ID})
	})
	floatForTest(g, me, "UUU")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: b16TargetCard(target)}); err != nil {
		t.Fatalf("cast Suit Up: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestNeoB13SuitUpMakesACreatureAndAVehicleA4By5ArtifactCreatureAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	stockLibraryFor(g, me, 5)
	ench := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Spirit", TypeLine: "Enchantment Creature — Spirit",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	cart := neoB13Vehicle(g, me.ID, "Cart", 2, 2)
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	hand := me.Hand.Size()
	neoB13CastSuitUp(t, g, me, ench)
	c, _ := battlefieldCard(g, ench)
	if !c.IsArtifact() || !c.IsCreature() || c.IsEnchantment() {
		t.Errorf("the Spirit is %v, want artifact creature only", c.Effective().Types)
	}
	if p, tt := ptOf(t, g, ench); p != 4 || tt != 5 {
		t.Errorf("Spirit is %d/%d, want 4/5", p, tt)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	neoB13CastSuitUp(t, g, me, cart)
	if c, _ := battlefieldCard(g, cart); !c.IsCreature() || !c.HasSubtype("Vehicle") {
		t.Errorf("the Vehicle is %v", c.Effective().Types)
	}
	if p, tt := ptOf(t, g, cart); p != 4 || tt != 5 {
		t.Errorf("Vehicle is %d/%d, want 4/5", p, tt)
	}
	// A plain artifact is not a legal target, and the effect ends with the turn.
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Suit Up", TypeLine: "Instant", OracleID: neoB13SuitUp,
			ManaCost: "{2}{U}", Colors: []string{"U"}, Owner: me.ID, Controller: me.ID})
	})
	floatForTest(g, me, "UUU")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: b16TargetCard(rock)}); err == nil {
		t.Error("targeted a noncreature non-Vehicle artifact")
	}
	ecl6NextTurn(t, g)
	if c, _ := battlefieldCard(g, ench); c.IsArtifact() || !c.IsEnchantment() {
		t.Errorf("the Spirit did not revert: %v", c.Effective().Types)
	}
}

// --- Mech Hangar ---

func TestNeoB13MechHangarAnimatesAVehicleAndRestrictsItsColoredMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hangar := pushCatalogPermanent(g, me.ID, "Mech Hangar", "Land", neoB13Hangar, false)
	cart := neoB13Vehicle(g, me.ID, "Cart", 3, 3)
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	floatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, hangar, 0, game.ActivateAbilityParams{Strict: true, Targets: b16TargetCard(rock)}); err == nil {
		t.Fatal("targeted a non-Vehicle")
	}
	b16Activate(t, g, me.ID, hangar, 0, game.ActivateAbilityParams{Strict: true, Targets: b16TargetCard(cart)})
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, cart); !c.IsCreature() || !c.IsArtifact() {
		t.Errorf("the Vehicle is %v", c.Effective().Types)
	}
	if !b16Tapped(t, g, hangar) {
		t.Error("the Hangar did not tap")
	}
	ecl6NextTurn(t, g)
	if c, _ := battlefieldCard(g, cart); c.IsCreature() {
		t.Error("the Vehicle stayed a creature past the turn")
	}
}

func TestNeoB13MechHangarHasColorlessAndRestrictedAnyColorMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hangar := pushCatalogPermanent(g, me.ID, "Mech Hangar", "Land", neoB13Hangar, false)
	advanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, hangar, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("{T}: Add {C}: %v", err)
	}
	if n := me.ManaPool.Total(); n != 1 {
		t.Errorf("pool has %d, want 1", n)
	}
	untapForTest(g, hangar)
	spec, _ := Lookup(neoB13Hangar)
	if len(spec.ManaAbilities) != 2 || len(spec.ManaAbilities[1].Restrictions) != 2 {
		t.Fatalf("mana abilities %+v", spec.ManaAbilities)
	}
}

// --- Katsumasa, the Animator ---

func TestNeoB13KatsumasaAnimatesAnArtifactAsAOneByOneUnlessItIsAVehicle(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	kat := pushCatalogPermanent(g, me.ID, "Katsumasa, the Animator", "Legendary Creature — Moonfolk Artificer", neoB13Katsumasa, false)
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	cart := neoB13Vehicle(g, me.ID, "Cart", 3, 3)
	advanceToMain(t, g)
	floatForTest(g, me, "UUUUUU")
	if err := g.ActivateCatalogAbility(me.ID, kat, 0, game.ActivateAbilityParams{Strict: true, Targets: b16TargetCard(kat)}); err == nil {
		t.Fatal("targeted a creature")
	}
	b16Activate(t, g, me.ID, kat, 0, game.ActivateAbilityParams{Strict: true, Targets: b16TargetCard(rock)})
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, rock); !c.IsCreature() || !c.IsArtifact() {
		t.Errorf("the artifact is %v", c.Effective().Types)
	}
	if p, tt := ptOf(t, g, rock); p != 1 || tt != 1 {
		t.Errorf("artifact is %d/%d, want 1/1", p, tt)
	}
	if !hasEffectiveKeyword(t, g, rock, "flying") {
		t.Error("no flying")
	}
	b16Activate(t, g, me.ID, kat, 0, game.ActivateAbilityParams{Strict: true, Targets: b16TargetCard(cart)})
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, cart); p != 3 || tt != 3 {
		t.Errorf("Vehicle is %d/%d, want its printed 3/3", p, tt)
	}
	if !hasEffectiveKeyword(t, g, cart, "flying") {
		t.Error("the Vehicle has no flying")
	}
}

func TestNeoB13KatsumasaUpkeepPutsACounterOnEachChosenNoncreatureArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Katsumasa, the Animator", "Legendary Creature — Moonfolk Artificer", neoB13Katsumasa, false)
	mine := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID})
	golem := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Golem", TypeLine: "Artifact Creature — Golem",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	ecl6NextTurn(t, g)
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no pick_target prompt")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if countersOn(g, mine, game.CounterPlusOne) != 1 || countersOn(g, theirs, game.CounterPlusOne) != 1 {
		t.Errorf("counters: mine %d theirs %d, want 1 each", countersOn(g, mine, game.CounterPlusOne), countersOn(g, theirs, game.CounterPlusOne))
	}
	if countersOn(g, golem, game.CounterPlusOne) != 0 {
		t.Error("a creature got a counter")
	}
}

// --- Imposter Mech ---

func neoB13CastMech(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Imposter Mech", TypeLine: "Artifact — Vehicle", OracleID: neoB13Mech,
			ManaCost: "{1}{U}", Colors: []string{"U"}, Power: 3, Toughness: 1, Owner: me.ID, Controller: me.ID})
	})
	floatForTest(g, me, "UU")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast Imposter Mech: %v", err)
	}
	return id
}

func TestNeoB13ImposterMechCopiesAnOpponentsCreatureAsACrew3Vehicle(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant Warrior", 5, 4)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	id := neoB13CastMech(t, g, me)
	for i := 0; i < 8 && copyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	pc := copyPrompt(g)
	if pc == nil {
		t.Fatal("no copy prompt")
	}
	if len(pc.CopyOptions) != 1 || pc.CopyOptions[0] != theirs {
		t.Errorf("CopyOptions = %v, want only the opponent's creature (not mine %s)", pc.CopyOptions, mine)
	}
	resolveWithCopyChoice(t, g, theirs)
	c := copyBattlefieldCard(t, g, id)
	if c.Name != "Their Giant" {
		t.Fatalf("name %q, want the copied creature's", c.Name)
	}
	e := c.Effective()
	if len(e.Types) != 1 || e.Types[0] != "Artifact" || c.IsCreature() {
		t.Errorf("types %v, want Artifact alone", e.Types)
	}
	if !c.HasSubtype("Vehicle") || c.HasSubtype("Giant") {
		t.Errorf("subtypes %v, want Vehicle only", e.Subtypes)
	}
	if p, tt := ptOf(t, g, id); p != 5 || tt != 4 {
		t.Errorf("it is %d/%d, want the copied 5/4", p, tt)
	}
	crew := b12Creature(g, me.ID, "Crew", "Creature — Bear", 3, 3)
	idx := -1
	for i, r := range game.ActivatedAbilitiesForCard(gaLayered(t, g, id)) {
		if r.Label == "Crew 3" {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatal("the copy has no crew 3")
	}
	if err := g.ActivateCatalogAbility(me.ID, id, idx, game.ActivateAbilityParams{Strict: true, CrewIDs: []uuid.UUID{crew}}); err != nil {
		t.Fatalf("crew 3: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, id); !c.IsCreature() {
		t.Error("the crewed copy is not a creature")
	}
}

func TestNeoB13ImposterMechDeclinedIsAPlain3By1VehicleWithCrew3(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 4)
	advanceToMain(t, g)
	id := neoB13CastMech(t, g, me)
	resolveWithCopyChoice(t, g, uuid.Nil)
	c := copyBattlefieldCard(t, g, id)
	if c.Name != "Imposter Mech" || c.IsCreature() || !c.HasSubtype("Vehicle") {
		t.Fatalf("declined: %q %v", c.Name, c.Effective().Types)
	}
	if p, tt := ptOf(t, g, id); p != 3 || tt != 1 {
		t.Errorf("it is %d/%d, want 3/1", p, tt)
	}
	crew := b12Creature(g, me.ID, "Crew", "Creature — Bear", 3, 3)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, CrewIDs: []uuid.UUID{crew}}); err != nil {
		t.Fatalf("printed crew 3: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, id); !c.IsCreature() {
		t.Error("crewed Mech is not a creature")
	}
}

// --- Greasefang, Okiba Boss ---

func TestNeoB13GreasefangReturnsAVehicleWithHasteAndBouncesItAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Greasefang, Okiba Boss", "Legendary Creature — Rat Pilot", neoB13Greasefang, false)
	cart := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: cart, Name: "Cart", TypeLine: "Artifact — Vehicle", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	})
	advanceToMain(t, g)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		pickTriggerTarget(t, g, me.ID, cart)
		passPriorityAroundTable(t, g)
	}
	var back uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Cart" {
			back = c.InstanceID
		}
	}
	if back == uuid.Nil {
		t.Fatal("the Vehicle did not return to the battlefield")
	}
	if !hasEffectiveKeyword(t, g, back, "haste") {
		t.Error("the returned Vehicle lacks haste")
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, back); ok {
		t.Error("the Vehicle is still on the battlefield after your end step")
	}
	if !dmuB13InHand(me, back) {
		t.Error("the Vehicle was not returned to its owner's hand")
	}
}

func TestNeoB13GreasefangOffersOnlyVehicleCardsFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Greasefang, Okiba Boss", "Legendary Creature — Rat Pilot", neoB13Greasefang, false)
	var bear, theirs uuid.UUID
	bear, theirs = uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
		opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Their Cart", TypeLine: "Artifact — Vehicle", Owner: opp.ID, Controller: opp.ID})
	})
	advanceToMain(t, g)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if pick := latestPickTarget(g, me.ID); pick != nil {
		for _, bad := range []uuid.UUID{bear, theirs} {
			if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
				t.Errorf("picked an illegal graveyard card %s", bad)
			}
		}
	}
}

// --- Mechtitan Core ---

func neoB13CoreBoard(g *game.Game, me *game.Player) (core uuid.UUID, mats []uuid.UUID) {
	core = pushCatalogPermanent(g, me.ID, "Mechtitan Core", "Artifact — Vehicle", neoB13Core, false)
	for i, line := range []string{"Artifact Creature — Golem", "Artifact — Vehicle", "Artifact Creature — Construct", "Artifact — Vehicle"} {
		mats = append(mats, pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Mat " + string(rune('A'+i)),
			TypeLine: line, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID}))
	}
	return
}

func TestNeoB13MechtitanCoreExilesFiveAndMakesAnAllColorsTenTenThatReturnsThemTappedWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	core, mats := neoB13CoreBoard(g, me)
	plain := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	floatForTest(g, me, "CCCCC")
	// Refusals: a non-artifact creature, a plain artifact, and too few materials.
	for name, bad := range map[string][]uuid.UUID{
		"a plain creature":  {mats[0], mats[1], mats[2], plain},
		"a plain artifact":  {mats[0], mats[1], mats[2], rock},
		"only three":        mats[:3],
		"the source itself": {mats[0], mats[1], mats[2], core},
	} {
		if err := g.ActivateCatalogAbility(me.ID, core, 1, game.ActivateAbilityParams{Strict: true, ExilePermanentIDs: bad}); err == nil {
			t.Errorf("activated with %s", name)
		}
	}
	b16Activate(t, g, me.ID, core, 1, game.ActivateAbilityParams{Strict: true, ExilePermanentIDs: mats})
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, core); ok || !g.Exile.Contains(core) {
		t.Fatal("the Core was not exiled")
	}
	var token game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Mechtitan" {
			token = c
		}
	}
	if token.InstanceID == uuid.Nil {
		t.Fatal("no Mechtitan token")
	}
	if p, tt := ptOf(t, g, token.InstanceID); p != 10 || tt != 10 {
		t.Errorf("Mechtitan is %d/%d", p, tt)
	}
	for _, kw := range []string{"flying", "vigilance", "trample", "lifelink", "haste"} {
		if !hasEffectiveKeyword(t, g, token.InstanceID, kw) {
			t.Errorf("Mechtitan lacks %s", kw)
		}
	}
	if len(token.Effective().Colors) != 5 || !token.IsLegendary() || !token.IsArtifact() || !token.IsCreature() {
		t.Errorf("Mechtitan colors %v types %v", token.Effective().Colors, token.Effective().Types)
	}
	for _, m := range mats {
		if !g.Exile.Contains(m) {
			t.Errorf("material %s not exiled", m)
		}
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(token.InstanceID) })
	passPriorityAroundTable(t, g)
	back := 0
	for _, c := range g.Battlefield.Cards {
		if len(c.Name) == 5 && c.Name[:4] == "Mat " {
			back++
			if !c.Tapped {
				t.Errorf("%s returned untapped", c.Name)
			}
		}
	}
	if back != 4 {
		t.Errorf("%d materials returned, want 4", back)
	}
	if g.Exile.Contains(core) == false {
		t.Error("the Core itself must stay exiled")
	}
}
