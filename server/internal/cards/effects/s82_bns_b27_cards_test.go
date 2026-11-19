package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 BNS batch 27 — Foundations Jumpstart (#5929).

const (
	bnsB27Brushstrider = "6dfcc4bb-7cd5-4b3e-ad79-3c246a3af2a8"
	bnsB27Courier      = "e186cc6f-8c11-4475-a704-0b4c3a9eade9"
	bnsB27Baloth       = "4493828f-144f-49cf-9fd2-61de86e632ec"
	bnsB27Unicorn      = "22005b01-84b1-4d4b-9193-3c561bd422ce"
	bnsB27Cavalier     = "08e2fead-53e9-445b-8fd1-c108bb09d81c"
	bnsB27Enforcer     = "81f5c70f-fbb1-4b3a-a048-1f4602adfa0e"
	bnsB27Spitfire     = "5a0eb270-b142-45da-87a2-2f1c4e25db17"
	bnsB27Hatchling    = "bd7111bc-9b8f-4c4b-b61c-7841a857ce6b"
	bnsB27Crystacean   = "baddf7ee-f6b9-41e3-b930-57129952cc1d"
	bnsB27Cyclops      = "638ecc20-a275-4810-9669-89a0730c5362"
)

func TestBnsB27CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		bnsB27Brushstrider: "Brushstrider", bnsB27Courier: "Byway Courier",
		bnsB27Baloth: "Canopy Baloth", bnsB27Unicorn: "Captivating Unicorn",
		bnsB27Cavalier: "Cavalier of Gales", bnsB27Enforcer: "Celestial Enforcer",
		bnsB27Spitfire: "Chandra's Spitfire", bnsB27Hatchling: "Cherished Hatchling",
		bnsB27Crystacean: "Crystacean", bnsB27Cyclops: "Cyclops Electromancer",
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

func TestBnsB27KeywordBodies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	assertKeywords(t, g, b12Push(g, me.ID, "Brushstrider", "Creature — Beast", bnsB27Brushstrider, 3, 1), "vigilance")
	assertKeywords(t, g, b12Push(g, me.ID, "Crystacean", "Creature — Crab", bnsB27Crystacean, 1, 6), "flash")
	assertKeywords(t, g, b12Push(g, me.ID, "Chandra's Spitfire", "Creature — Elemental", bnsB27Spitfire, 1, 3), "flying")
	assertKeywords(t, g, b12Push(g, me.ID, "Cavalier of Gales", "Creature — Elemental Knight", bnsB27Cavalier, 5, 5), "flying")
}

func TestBnsB27BywayCourierInvestigatesWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	c := b12Push(g, me.ID, "Byway Courier", "Creature — Human Scout", bnsB27Courier, 3, 2)
	killCreature(t, g, me.ID, c)
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Clue") == uuid.Nil {
		t.Error("no Clue after the Courier died")
	}
}

func TestBnsB27CanopyBalothGrowsOnYourLandsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	baloth := b12Push(g, me.ID, "Canopy Baloth", "Creature — Beast", bnsB27Baloth, 4, 3)
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if p, tough := rfPT(t, g, baloth); p != 6 || tough != 5 {
		t.Errorf("after a land: %d/%d, want 6/5", p, tough)
	}
	// A land entering under an opponent's control does nothing.
	opp := g.Seats[1]
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID})
	})
	if p, _ := rfPT(t, g, baloth); p != 6 {
		t.Errorf("power %d after nothing happened", p)
	}
}

func TestBnsB27ChandrasSpitfireGrowsOnNoncombatDamageToAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spit := b12Push(g, me.ID, "Chandra's Spitfire", "Creature — Elemental", bnsB27Spitfire, 1, 3)
	src := b16Creature(g, me.ID, "Pinger", "Creature — Wizard", 1, 1)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, opp.ID, 2) })
	passPriorityAroundTable(t, g)
	if p, _ := rfPT(t, g, spit); p != 4 {
		t.Errorf("power %d, want 4 after noncombat damage to an opponent", p)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 2) })
	passPriorityAroundTable(t, g)
	if p, _ := rfPT(t, g, spit); p != 4 {
		t.Errorf("power %d: damage to its controller must not grow it", p)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, opp.ID, 1) })
	passPriorityAroundTable(t, g)
	if p, _ := rfPT(t, g, spit); p != 7 {
		t.Errorf("power %d, want 7 after a second hit", p)
	}
}

func TestBnsB27CaptivatingUnicornTapsOnYourEnchantmentsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Captivating Unicorn", "Creature — Unicorn", bnsB27Unicorn, 4, 4)
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	// An enchantment under the opponent's control does nothing.
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Their Ench", TypeLine: "Enchantment", Owner: opp.ID, Controller: opp.ID})
	})
	passPriorityAroundTable(t, g)
	if tappedNow(g, victim) {
		t.Fatal("an opponent's enchantment tapped their own creature")
	}
	castCatalogSpell(t, g, "My Ench", "Enchantment", "", nil)
	passPriorityAroundTable(t, g)
	if pick := latestPickTarget(g, me.ID); pick != nil {
		pickTriggerTarget(t, g, me.ID, victim)
		passPriorityAroundTable(t, g)
	}
	if !tappedNow(g, victim) {
		t.Error("the target creature was not tapped")
	}
}

func TestBnsB27CelestialEnforcerNeedsAFlyer(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	enf := pushCatalogPermanent(g, me.ID, "Celestial Enforcer", "Creature — Human Cleric", bnsB27Enforcer, false)
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	floatForTest(g, me, "1W")
	if err := g.ActivateCatalogAbility(me.ID, enf, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(victim)}); err == nil {
		t.Fatal("activated without controlling a creature with flying")
	}
	floatForTest(g, me, "1W")
	flyer := b12Creature(g, me.ID, "Flyer", "Creature — Bird", 1, 1)
	g.WithWriteLock(func() {
		if c := findBattlefieldCardByID(g, flyer); c != nil {
			c.Keywords = append(c.Keywords, "flying")
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, enf, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(victim)}); err != nil {
		t.Fatalf("activate with a flyer: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !tappedNow(g, victim) {
		t.Error("target not tapped")
	}
	if !tappedNow(g, enf) {
		t.Error("the Enforcer did not tap for its cost")
	}
}

func TestBnsB27CyclopsElectromancerDealsXFromInstantsAndSorceries(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	big := b12Creature(g, opp.ID, "Behemoth", "Creature — Beast", 6, 6)
	for _, tl := range []string{"Instant", "Sorcery", "Creature — Bear", "Land"} {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "G " + tl, TypeLine: tl, Owner: me.ID, Controller: me.ID})
	}
	castCatalogSpell(t, g, "Cyclops Electromancer", "Creature — Cyclops Wizard", bnsB27Cyclops, nil)
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, big)
	passPriorityAroundTable(t, g)
	if got := pr6Marked(g, big); got != 2 {
		t.Errorf("target has %d damage, want 2 (one instant, one sorcery)", got)
	}
}

func TestBnsB27CavalierOfGalesDrawsThreeAndPutsTwoBack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Cavalier of Gales", "Creature — Elemental Knight", bnsB27Cavalier, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+3 {
		t.Fatalf("hand after the draw = %d, want %d", got, before+3)
	}
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil || pick.ChooseMin != 2 || pick.ChooseMax != 2 {
		t.Fatalf("put-back prompt = %+v, want exactly two", pick)
	}
	first, second := me.Hand.Cards[0].InstanceID, me.Hand.Cards[1].InstanceID
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{first, second}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	order := putInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatal("no order prompt")
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, []uuid.UUID{first, second}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand after putting two back = %d, want %d", got, before+1)
	}
}

func TestBnsB27CavalierOfGalesShufflesItselfAwayAndScries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cav := b12Push(g, me.ID, "Cavalier of Gales", "Creature — Elemental Knight", bnsB27Cavalier, 5, 5)
	seedLibrary(me, "A", "B", "C")
	lib := me.Library.Size()
	killCreature(t, g, me.ID, cav)
	passPriorityAroundTable(t, g)
	answerScryKeepAll(t, g, me.ID)
	passPriorityAroundTable(t, g)
	z := g.FindCardZoneForEffect(cav)
	if z == nil || z.Kind != game.ZoneLibrary {
		t.Fatalf("the Cavalier is in %+v, want the library", z)
	}
	if got := me.Library.Size(); got != lib+1 {
		t.Errorf("library %d, want %d", got, lib+1)
	}
}

func TestBnsB27CherishedHatchlingGivesDinosaurSpellsTheFight(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	hatch := b12Push(g, me.ID, "Cherished Hatchling", "Creature — Dinosaur", bnsB27Hatchling, 2, 1)
	victim := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 2, 6)
	killCreature(t, g, me.ID, hatch)
	passPriorityAroundTable(t, g)
	// A non-Dinosaur spell gains nothing.
	bear := castCatalogSpell(t, g, "My Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("a non-Dinosaur spell triggered the fight")
	}
	_ = bear
	// A Dinosaur spell does, and it may fight another target creature.
	dino := castCreatureCard(t, g, "My Dino", "Creature — Dinosaur", "", 3, 5)
	if !g.Battlefield.Contains(dino) {
		t.Fatal("the Dinosaur did not resolve")
	}
	pickTriggerTarget(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	prompt := assistCardTriggerPrompt(g, me.ID)
	if prompt == nil {
		t.Fatal("the granted enters trigger did not ask whether to fight")
	}
	if err := g.ResolveTriggerPrompt(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := pr6Marked(g, victim); got != 3 {
		t.Errorf("the fought creature has %d damage, want 3", got)
	}
	if got := pr6Marked(g, dino); got != 2 {
		t.Errorf("the Dinosaur has %d damage, want 2", got)
	}
}

func TestBnsB27CherishedHatchlingLetsYouFlashInDinosaursOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	hatch := b12Push(g, me.ID, "Cherished Hatchling", "Creature — Dinosaur", bnsB27Hatchling, 2, 1)
	dino, bear := uuid.New(), uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: dino, Name: "Hand Dino", TypeLine: "Creature — Dinosaur", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Hand Bear", TypeLine: "Creature — Bear", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.CastSpell(me.ID, dino, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a Dinosaur was cast at instant speed before the Hatchling died")
	}
	killCreature(t, g, me.ID, hatch)
	passPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a non-Dinosaur creature was cast at instant speed")
	}
	if err := g.CastSpell(me.ID, dino, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("a Dinosaur could not be cast at instant speed: %v", err)
	}
}
