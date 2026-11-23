package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Monument to Perfection -----------------------------------------------------

func TestS82OneB18MonumentTutorsABasicSphereOrLocusLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mon := effectstest.PushCatalogPermanent(g, me.ID, "Monument to Perfection", "Artifact", s82OneB18MonumentOracle, false)
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Mirrex Sphere", "Land — Sphere"),
		effectstest.SearchTestLand("Locus Place", "Land — Locus"),
		effectstest.SearchTestLand("Mountain Gate", "Land — Gate"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, mon, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	for _, n := range []string{"Forest", "Mirrex Sphere", "Locus Place"} {
		if effectstest.SearchOptionNamed(g, c, n) == uuid.Nil {
			t.Errorf("the search did not offer %q", n)
		}
	}
	for _, n := range []string{"Mountain Gate", "Bear"} {
		if effectstest.SearchOptionNamed(g, c, n) != uuid.Nil {
			t.Errorf("the search offered %q", n)
		}
	}
	before := me.Hand.Size()
	effectstest.AnswerSearchNamed(t, g, me.ID, "Mirrex Sphere")
	if me.Hand.Size() != before+1 {
		t.Errorf("hand %d -> %d, want +1", before, me.Hand.Size())
	}
	if !effectstest.LayeredCard(t, g, mon).Tapped {
		t.Error("the Monument did not tap")
	}
}

func TestS82OneB18MonumentBecomesAnIndestructibleToxicNineNineOnlyWithNineLandNames(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mon := effectstest.PushCatalogPermanent(g, me.ID, "Monument to Perfection", "Artifact", s82OneB18MonumentOracle, false)
	// Eight names (and a repeated one) is not enough.
	for _, l := range []struct{ name, tl string }{
		{"Plains", "Basic Land — Plains"}, {"Plains", "Basic Land — Plains"}, {"Island", "Basic Land — Island"},
		{"Swamp", "Basic Land — Swamp"}, {"Mountain", "Basic Land — Mountain"},
		{"Forest", "Basic Land — Forest"}, {"Wastes", "Basic Land"},
		{"Snow-Covered Plains", "Basic Snow Land — Plains"}, {"Snow-Covered Island", "Basic Snow Land — Island"},
	} {
		effectstest.B12Permanent(g, me.ID, l.name, l.tl)
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, mon, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with only eight differently named lands")
	}
	// A ninth name that is not basic, Sphere or Locus doesn't count.
	effectstest.B12Permanent(g, me.ID, "Mountain Gate", "Land — Gate")
	if err := g.ActivateCatalogAbility(me.ID, mon, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("a Gate counted toward the nine")
	}
	// A Sphere does.
	effectstest.B12Permanent(g, me.ID, "Mirrex Sphere", "Land — Sphere")
	if err := g.ActivateCatalogAbility(me.ID, mon, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate with nine names: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, mon)
	if !c.IsCreature() || !c.IsArtifact() || !c.HasSubtype("Construct") || !c.HasSubtype("Phyrexian") {
		t.Errorf("types after animating: %q", c.Effective().Types)
	}
	if c.CurrentPower() != 9 || c.CurrentToughness() != 9 {
		t.Errorf("animated %d/%d, want 9/9", c.CurrentPower(), c.CurrentToughness())
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, mon, "indestructible") {
		t.Error("lacks indestructible")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, mon, "toxic 9") {
		t.Errorf("lacks toxic 9: %v", c.Effective().Abilities)
	}
	// It lost its abilities (the tutor is gone) and the animation is permanent.
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, mon, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the Monument still has its tutor ability")
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	if c := effectstest.LayeredCard(t, g, mon); !c.IsCreature() || c.CurrentPower() != 9 {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Kethek, Crucible Goliath", "Legendary Creature — Phyrexian Beast", s82OneB18KethekOracle, false)
	fodder := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Fodder", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	me.Library.Cards = nil
	// Top first: a spell, a legendary of lesser value, an equal-value
	// creature, a greater one, the hit (MV 2), then a card below it.
	below := s82OneB18LibraryCard(me, "Below", "Creature — Elf", "{1}")
	hit := s82OneB18LibraryCard(me, "The Hit", "Creature — Elf", "{1}{G}")
	big := s82OneB18LibraryCard(me, "Too Big", "Creature — Elf", "{4}")
	equal := s82OneB18LibraryCard(me, "Equal Value", "Creature — Elf", "{2}{G}")
	legend := s82OneB18LibraryCard(me, "Legendary Elf", "Legendary Creature — Elf", "{1}")
	spell := s82OneB18LibraryCard(me, "A Spell", "Instant", "{G}")

	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)

	if effects.OnBattlefield(g, fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if !effects.OnBattlefield(g, hit) {
		t.Fatal("the creature with lesser mana value did not enter")
	}
	for name, id := range map[string]uuid.UUID{"spell": spell, "legendary": legend, "equal": equal, "too big": big} {
		if effects.OnBattlefield(g, id) {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Kethek, Crucible Goliath", "Legendary Creature — Phyrexian Beast", s82OneB18KethekOracle, false)
	fodder := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Fodder", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	hit := s82OneB18LibraryCard(me, "The Hit", "Creature — Elf", "{1}")
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, fodder) || effects.OnBattlefield(g, hit) {
		t.Error("declining the sacrifice still revealed and put a creature onto the battlefield")
	}
}

func TestS82OneB18KethekCantSacrificeItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Kethek, Crucible Goliath", "Legendary Creature — Phyrexian Beast", s82OneB18KethekOracle, false)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 {
		t.Error("asked for a sacrifice with no other creature")
	}
}

// --- Unctus, Grand Metatect -----------------------------------------------------

func TestS82OneB18UnctusPumpsOtherArtifactCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	unctus := effectstest.PushCatalogPermanent(g, me.ID, "Unctus, Grand Metatect", "Legendary Artifact Creature — Phyrexian Vedalken", s82OneB18UnctusOracle, false)
	golem := effectstest.B12Creature(g, me.ID, "My Golem", "Artifact Creature — Golem", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if got := effectstest.LayeredCard(t, g, golem).CurrentPower(); got != 3 {
		t.Errorf("my artifact creature has power %d, want 3", got)
	}
	if got := effectstest.LayeredCard(t, g, bear).CurrentPower(); got != 2 {
		t.Errorf("my nonartifact creature has power %d, want 2", got)
	}
	if got := effectstest.LayeredCard(t, g, theirs).CurrentPower(); got != 2 {
		t.Errorf("an opponent's artifact creature has power %d, want 2", got)
	}
	if got := effectstest.LayeredCard(t, g, unctus).CurrentPower(); got != 1 {
		t.Errorf("Unctus pumped itself: power %d, want its own 1", got)
	}
}

func TestS82OneB18UnctusGivesOtherBlueCreaturesALootOnTap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Unctus, Grand Metatect", "Legendary Artifact Creature — Phyrexian Vedalken", s82OneB18UnctusOracle, false)
	blue := effectstest.B27Push(g, me.ID, "Blue Bear", "Creature — Bear", "", "{1}{U}", 2, 2, "U")
	effectstest.SeedLibraryTop(me, "Top Card", "Sorcery", "")
	hand := effectstest.HandCardForTest(me, "Held Card", "Sorcery", "")
	libBefore := me.Library.Size()
	handBefore := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, blue)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) != nil {
		effectstest.DiscardFromHand(t, g, me.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != libBefore-1 {
		t.Errorf("library %d -> %d, want one card drawn", libBefore, got)
	}
	if got := me.Hand.Size(); got != handBefore {
		t.Errorf("hand size %d, want %d (drew one, discarded one)", got, handBefore)
	}
	_ = hand
}

func TestS82OneB18UnctusLootsNothingForOpponentsOrNonblueCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Unctus, Grand Metatect", "Legendary Artifact Creature — Phyrexian Vedalken", s82OneB18UnctusOracle, false)
	red := effectstest.B27Push(g, me.ID, "Red Bear", "Creature — Bear", "", "{1}{R}", 2, 2, "R")
	effectstest.SeedLibraryTop(me, "Top Card", "Sorcery", "")
	libBefore := me.Library.Size()
	effectstest.DeclareAttack(t, g, opp.ID, red)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != libBefore {
		t.Errorf("a nonblue creature looted: library %d -> %d", libBefore, got)
	}
}

func TestS82OneB18UnctusMakesATargetCreatureABlueArtifactAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	unctus := effectstest.PushCatalogPermanent(g, me.ID, "Unctus, Grand Metatect", "Legendary Artifact Creature — Phyrexian Vedalken", s82OneB18UnctusOracle, false)
	bear := effectstest.B27Push(g, me.ID, "Red Bear", "Creature — Bear", "", "{1}{R}", 2, 2, "R")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{U}")
	if err := g.ActivateCatalogAbility(me.ID, unctus, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, bear)
	if !c.IsArtifact() || !c.IsCreature() || !c.HasColor("U") || !c.HasColor("R") {
		t.Errorf("after the ability: artifact %v creature %v blue %v red %v", c.IsArtifact(), c.IsCreature(), c.HasColor("U"), c.HasColor("R"))
	}
	if got := c.CurrentPower(); got != 3 {
		t.Errorf("the new artifact creature has power %d, want 3 from Unctus's own anthem", got)
	}
	// It ends at the turn's end.
	effectstest.AdvanceToMainOf(t, g, 1)
	if c := effectstest.LayeredCard(t, g, bear); c.IsArtifact() || c.HasColor("U") {
		t.Error("the change outlasted the turn")
	}
}

func TestS82OneB18UnctusAbilityIsRefusedAtInstantSpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	unctus := effectstest.PushCatalogPermanent(g, me.ID, "Unctus, Grand Metatect", "Legendary Artifact Creature — Phyrexian Vedalken", s82OneB18UnctusOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FloatMana(t, g, me, "{U}")
	if err := g.ActivateCatalogAbility(me.ID, unctus, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err == nil {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Ria Ivor, Bane of Bladehold", "Legendary Creature — Phyrexian Knight", s82OneB18RiaOracle, false)
	ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 4, 4)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PickTriggerTarget(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	life := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, mine)
	_ = ogre
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if opp.Life != life {
		t.Errorf("the opponent lost %d life, want the damage prevented", life-opp.Life)
	}
	if got := s82OneB18Mites(g, me.ID); got != 4 {
		t.Errorf("%d Phyrexian Mites, want 4 (one per point prevented)", got)
	}
}
