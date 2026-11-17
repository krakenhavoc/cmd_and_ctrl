package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Monument to Perfection -----------------------------------------------------

func s82OneB18NineBasics(g *game.Game, owner uuid.UUID) {
	for _, l := range []struct{ name, tl string }{
		{"Plains", "Basic Land — Plains"}, {"Island", "Basic Land — Island"},
		{"Swamp", "Basic Land — Swamp"}, {"Mountain", "Basic Land — Mountain"},
		{"Forest", "Basic Land — Forest"}, {"Wastes", "Basic Land"},
		{"Snow-Covered Plains", "Basic Snow Land — Plains"},
		{"Snow-Covered Island", "Basic Snow Land — Island"},
		{"Snow-Covered Swamp", "Basic Snow Land — Swamp"},
	} {
		b12Permanent(g, owner, l.name, l.tl)
	}
}

func TestS82OneB18MonumentTutorsABasicSphereOrLocusLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mon := pushCatalogPermanent(g, me.ID, "Monument to Perfection", "Artifact", s82OneB18MonumentOracle, false)
	seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Mirrex Sphere", "Land — Sphere"),
		searchTestLand("Locus Place", "Land — Locus"),
		searchTestLand("Mountain Gate", "Land — Gate"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	advanceToMain(t, g)
	floatMana(t, g, me, "{3}")
	if err := g.ActivateCatalogAbility(me.ID, mon, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	for _, n := range []string{"Forest", "Mirrex Sphere", "Locus Place"} {
		if searchOptionNamed(g, c, n) == uuid.Nil {
			t.Errorf("the search did not offer %q", n)
		}
	}
	for _, n := range []string{"Mountain Gate", "Bear"} {
		if searchOptionNamed(g, c, n) != uuid.Nil {
			t.Errorf("the search offered %q", n)
		}
	}
	before := me.Hand.Size()
	answerSearchNamed(t, g, me.ID, "Mirrex Sphere")
	if me.Hand.Size() != before+1 {
		t.Errorf("hand %d -> %d, want +1", before, me.Hand.Size())
	}
	if !layeredCard(t, g, mon).Tapped {
		t.Error("the Monument did not tap")
	}
}

func TestS82OneB18MonumentBecomesAnIndestructibleToxicNineNineOnlyWithNineLandNames(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mon := pushCatalogPermanent(g, me.ID, "Monument to Perfection", "Artifact", s82OneB18MonumentOracle, false)
	// Eight names (and a repeated one) is not enough.
	for _, l := range []struct{ name, tl string }{
		{"Plains", "Basic Land — Plains"}, {"Plains", "Basic Land — Plains"}, {"Island", "Basic Land — Island"},
		{"Swamp", "Basic Land — Swamp"}, {"Mountain", "Basic Land — Mountain"},
		{"Forest", "Basic Land — Forest"}, {"Wastes", "Basic Land"},
		{"Snow-Covered Plains", "Basic Snow Land — Plains"}, {"Snow-Covered Island", "Basic Snow Land — Island"},
	} {
		b12Permanent(g, me.ID, l.name, l.tl)
	}
	advanceToMain(t, g)
	floatMana(t, g, me, "{3}")
	if err := g.ActivateCatalogAbility(me.ID, mon, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with only eight differently named lands")
	}
	// A ninth name that is not basic, Sphere or Locus doesn't count.
	b12Permanent(g, me.ID, "Mountain Gate", "Land — Gate")
	if err := g.ActivateCatalogAbility(me.ID, mon, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("a Gate counted toward the nine")
	}
	// A Sphere does.
	b12Permanent(g, me.ID, "Mirrex Sphere", "Land — Sphere")
	if err := g.ActivateCatalogAbility(me.ID, mon, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate with nine names: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, mon)
	if !c.IsCreature() || !c.IsArtifact() || !c.HasSubtype("Construct") || !c.HasSubtype("Phyrexian") {
		t.Errorf("types after animating: %q", c.Effective().Types)
	}
	if c.CurrentPower() != 9 || c.CurrentToughness() != 9 {
		t.Errorf("animated %d/%d, want 9/9", c.CurrentPower(), c.CurrentToughness())
	}
	if !hasKeywordOnBattlefield(t, g, mon, "indestructible") {
		t.Error("lacks indestructible")
	}
	if !hasKeywordOnBattlefield(t, g, mon, "toxic 9") {
		t.Errorf("lacks toxic 9: %v", c.Effective().Abilities)
	}
	// It lost its abilities (the tutor is gone) and the animation is permanent.
	floatMana(t, g, me, "{3}")
	if err := g.ActivateCatalogAbility(me.ID, mon, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the Monument still has its tutor ability")
	}
	advanceToMainOf(t, g, 1)
	if c := layeredCard(t, g, mon); !c.IsCreature() || c.CurrentPower() != 9 {
		t.Error("the animation ended with the turn")
	}
}

// --- Kethek, Crucible Goliath ---------------------------------------------------

func s82OneB18LibraryCard(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Power: 1, Toughness: 1, Owner: p.ID, Controller: p.ID})
	return id
}

func TestS82OneB18KethekSacrificesForANonlegendaryCreatureWithLesserManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Kethek, Crucible Goliath", "Legendary Creature — Phyrexian Beast", s82OneB18KethekOracle, false)
	fodder := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Fodder", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	me.Library.Cards = nil
	// Top first: a spell, a legendary of lesser value, an equal-value
	// creature, a greater one, the hit (MV 2), then a card below it.
	below := s82OneB18LibraryCard(me, "Below", "Creature — Elf", "{1}")
	hit := s82OneB18LibraryCard(me, "The Hit", "Creature — Elf", "{1}{G}")
	big := s82OneB18LibraryCard(me, "Too Big", "Creature — Elf", "{4}")
	equal := s82OneB18LibraryCard(me, "Equal Value", "Creature — Elf", "{2}{G}")
	legend := s82OneB18LibraryCard(me, "Legendary Elf", "Legendary Creature — Elf", "{1}")
	spell := s82OneB18LibraryCard(me, "A Spell", "Instant", "{G}")

	advanceTo(t, g, game.StepEnd)
	answerChooseCards(t, g, me.ID, fodder)
	passPriorityAroundTable(t, g)

	if onBattlefield(g, fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if !onBattlefield(g, hit) {
		t.Fatal("the creature with lesser mana value did not enter")
	}
	for name, id := range map[string]uuid.UUID{"spell": spell, "legendary": legend, "equal": equal, "too big": big} {
		if onBattlefield(g, id) {
			t.Errorf("the %s card entered", name)
		}
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneLibrary {
			t.Errorf("the revealed %s card is in %v, want the library bottom", name, z)
		}
	}
	// The rest went to the bottom: the card that was never revealed is on top.
	if top, err := me.Library.Top(); err != nil || top.InstanceID != below {
		t.Errorf("the unrevealed card is not on top of the library (top %q)", top.Name)
	}
}

func TestS82OneB18KethekDoesNothingIfYouDeclineTheSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Kethek, Crucible Goliath", "Legendary Creature — Phyrexian Beast", s82OneB18KethekOracle, false)
	fodder := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Fodder", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	hit := s82OneB18LibraryCard(me, "The Hit", "Creature — Elf", "{1}")
	advanceTo(t, g, game.StepEnd)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, fodder) || onBattlefield(g, hit) {
		t.Error("declining the sacrifice still revealed and put a creature onto the battlefield")
	}
}

func TestS82OneB18KethekCantSacrificeItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Kethek, Crucible Goliath", "Legendary Creature — Phyrexian Beast", s82OneB18KethekOracle, false)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if c := discardChoiceFor(g, me.ID); c != nil {
		t.Error("asked for a sacrifice with no other creature")
	}
}

// --- Unctus, Grand Metatect -----------------------------------------------------

func TestS82OneB18UnctusPumpsOtherArtifactCreaturesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	unctus := pushCatalogPermanent(g, me.ID, "Unctus, Grand Metatect", "Legendary Artifact Creature — Phyrexian Vedalken", s82OneB18UnctusOracle, false)
	golem := b12Creature(g, me.ID, "My Golem", "Artifact Creature — Golem", 2, 2)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 2, 2)
	advanceToMain(t, g)
	if got := layeredCard(t, g, golem).CurrentPower(); got != 3 {
		t.Errorf("my artifact creature has power %d, want 3", got)
	}
	if got := layeredCard(t, g, bear).CurrentPower(); got != 2 {
		t.Errorf("my nonartifact creature has power %d, want 2", got)
	}
	if got := layeredCard(t, g, theirs).CurrentPower(); got != 2 {
		t.Errorf("an opponent's artifact creature has power %d, want 2", got)
	}
	if got := layeredCard(t, g, unctus).CurrentPower(); got != 1 {
		t.Errorf("Unctus pumped itself: power %d, want its own 1", got)
	}
}

func TestS82OneB18UnctusGivesOtherBlueCreaturesALootOnTap(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Unctus, Grand Metatect", "Legendary Artifact Creature — Phyrexian Vedalken", s82OneB18UnctusOracle, false)
	blue := b27Push(g, me.ID, "Blue Bear", "Creature — Bear", "", "{1}{U}", 2, 2, "U")
	seedLibraryTop(me, "Top Card", "Sorcery", "")
	hand := handCardForTest(me, "Held Card", "Sorcery", "")
	libBefore := me.Library.Size()
	declareAttack(t, g, opp.ID, blue)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) != nil {
		discardFromHand(t, g, me.ID)
	}
	passPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != libBefore-1 {
		t.Errorf("library %d -> %d, want one card drawn", libBefore, got)
	}
	if got := me.Hand.Size(); got != 1 {
		t.Errorf("hand size %d, want 1 (drew one, discarded one)", got)
	}
	_ = hand
}

func TestS82OneB18UnctusLootsNothingForOpponentsOrNonblueCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Unctus, Grand Metatect", "Legendary Artifact Creature — Phyrexian Vedalken", s82OneB18UnctusOracle, false)
	red := b27Push(g, me.ID, "Red Bear", "Creature — Bear", "", "{1}{R}", 2, 2, "R")
	seedLibraryTop(me, "Top Card", "Sorcery", "")
	libBefore := me.Library.Size()
	declareAttack(t, g, opp.ID, red)
	passPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != libBefore {
		t.Errorf("a nonblue creature looted: library %d -> %d", libBefore, got)
	}
}

func TestS82OneB18UnctusMakesATargetCreatureABlueArtifactAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	unctus := pushCatalogPermanent(g, me.ID, "Unctus, Grand Metatect", "Legendary Artifact Creature — Phyrexian Vedalken", s82OneB18UnctusOracle, false)
	bear := b27Push(g, me.ID, "Red Bear", "Creature — Bear", "", "{1}{R}", 2, 2, "R")
	advanceToMain(t, g)
	floatMana(t, g, me, "{U}")
	if err := g.ActivateCatalogAbility(me.ID, unctus, 0, game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, bear)
	if !c.IsArtifact() || !c.IsCreature() || !c.HasColor("U") || !c.HasColor("R") {
		t.Errorf("after the ability: artifact %v creature %v blue %v red %v", c.IsArtifact(), c.IsCreature(), c.HasColor("U"), c.HasColor("R"))
	}
	if got := c.CurrentPower(); got != 3 {
		t.Errorf("the new artifact creature has power %d, want 3 from Unctus's own anthem", got)
	}
	// It ends at the turn's end.
	advanceToMainOf(t, g, 1)
	if c := layeredCard(t, g, bear); c.IsArtifact() || c.HasColor("U") {
		t.Error("the change outlasted the turn")
	}
}

func TestS82OneB18UnctusAbilityIsRefusedAtInstantSpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	unctus := pushCatalogPermanent(g, me.ID, "Unctus, Grand Metatect", "Legendary Artifact Creature — Phyrexian Vedalken", s82OneB18UnctusOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	floatMana(t, g, me, "{U}")
	if err := g.ActivateCatalogAbility(me.ID, unctus, 0, game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true}); err == nil {
		t.Error("activated outside a main phase")
	}
}

// --- Ria Ivor, Bane of Bladehold ------------------------------------------------

func s82OneB18Mites(g *game.Game, owner uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Phyrexian Mite" && c.Controller == owner {
			n++
		}
	}
	return n
}

func TestS82OneB18RiaPreventsTheTargetsNextCombatDamageToAPlayerAndMakesMites(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Ria Ivor, Bane of Bladehold", "Legendary Creature — Phyrexian Knight", s82OneB18RiaOracle, false)
	ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	mine := b12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 4, 4)
	advanceTo(t, g, game.StepBeginCombat)
	// The trigger's target: an opposing creature, so its damage to me is prevented.
	if len(g.StackMeta) == 0 {
		t.Fatal("Ria's beginning-of-combat trigger is not on the stack")
	}
	g.StackMeta[len(g.StackMeta)-1].Targets = cardRefs(mine)
	passPriorityAroundTable(t, g)
	life := opp.Life
	declareAttack(t, g, opp.ID, mine)
	_ = ogre
	advanceTo(t, g, game.StepEndCombat)
	if opp.Life != life {
		t.Errorf("the opponent lost %d life, want the damage prevented", life-opp.Life)
	}
	if got := s82OneB18Mites(g, me.ID); got != 4 {
		t.Errorf("%d Phyrexian Mites, want 4 (one per point prevented)", got)
	}
}
