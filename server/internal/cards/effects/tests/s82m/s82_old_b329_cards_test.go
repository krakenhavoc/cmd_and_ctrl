package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 Dragons of Tarkir batch 329 (#5551): card draw and library.

const (
	dtk329Scholar   = "070bd1e8-6c61-4b0e-b179-e375b843410f"
	dtk329Prerog    = "46bc0aaf-987f-4c8d-b98b-50b90f97c9a3"
	dtk329Tapestry  = "9248f9b9-ddaa-459c-91e6-9045a65b0f07"
	dtk329Scribe    = "a1841bcc-738e-4f4b-a88f-571a93c59ad7"
	dtk329Contradic = "a48f259a-99a0-49c7-a11b-6e796f8a12f4"
	dtk329Meditate  = "d7c18aa7-4184-4658-9a5a-0b64992348ee"
)

func TestDtk329CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		dtk329Scholar: "Youthful Scholar", dtk329Prerog: "Dragonlord's Prerogative", dtk329Tapestry: "Tapestry of the Ages",
		dtk329Scribe: "Zephyr Scribe", dtk329Contradic: "Contradict", dtk329Meditate: "Mystic Meditation",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestDtk329YouthfulScholarDrawsTwoWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogPermanent(g, me.ID, "Youthful Scholar", "Creature — Human Wizard", dtk329Scholar, false)
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "A", TypeLine: "Land"})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "B", TypeLine: "Land"})
	hand := me.Hand.Size()
	effectstest.KillCreature(t, g, me.ID, id)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+2)
	}
}

func TestDtk329ContradictCountersAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B13OpponentCasts(t, g, opp, "Their Spell", "Instant", "", "{R}", nil)
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Drawn", TypeLine: "Land"})
	c := effectstest.B13HandCard(me, "Contradict", "Instant", dtk329Contradic)
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, c, game.CastSpellParams{Targets: effectstest.B16TargetCard(theirs)}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(theirs) {
		t.Error("the spell was not countered")
	}
	if got := me.Hand.Size() - hand; got != 0 {
		// Contradict left the hand (-1) and a card was drawn (+1).
		t.Errorf("net hand change %d, want 0", got)
	}
}

func TestDtk329ContradictNeedsASpellTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Contradict", "Instant", dtk329Contradic, effectstest.B16TargetCard(bear)); err == nil {
		t.Error("Contradict targeted a creature")
	}
}

func TestDtk329MysticMeditationDiscardsOneCreatureOrTwoOthers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 3; i++ {
		effectstest.PushLibraryCard(me, game.Card{Name: "Filler", TypeLine: "Land"})
	}
	var creature uuid.UUID
	effectstest.PushLibraryCard(me, game.Card{Name: "Beast", TypeLine: "Creature — Beast"})
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Mystic Meditation", "Sorcery", dtk329Meditate, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range me.Hand.Cards {
		if c.Name == "Beast" {
			creature = c.InstanceID
		}
	}
	if creature == uuid.Nil {
		t.Fatal("the creature card was not drawn")
	}
	effectstest.AnswerDiscard(t, g, me.ID, creature)
	// The spell arrived (+1) and was cast (-1), three were drawn (+3) and one creature card discarded (-1).
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+2)
	}
}

func TestDtk329DragonlordsPrerogativeDrawsFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 5; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Land"})
	}
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Dragonlord's Prerogative", "Instant", dtk329Prerog, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+4 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+4)
	}
}

func TestDtk329DragonlordsPrerogativeCantBeCounteredOnlyWithADragon(t *testing.T) {
	for _, tc := range []string{"nothing", "controlled", "revealed", "controlled but it left in response"} {
		t.Run(tc, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			params := game.CastSpellParams{}
			var dragonOnBf uuid.UUID
			switch tc {
			case "controlled", "controlled but it left in response":
				dragonOnBf = effectstest.B12Creature(g, me.ID, "Dragon", "Creature — Dragon", 4, 4)
			case "revealed":
				dragon := effectstest.PutInHand(me, game.Card{Name: "Hand Dragon", TypeLine: "Creature — Dragon", Power: 4, Toughness: 4})
				params = game.CastSpellParams{OptionalCosts: []int{0}, RevealIDs: []uuid.UUID{dragon}}
			}
			id := effectstest.B13HandCard(me, "Dragonlord's Prerogative", "Instant", dtk329Prerog)
			if err := g.CastSpell(me.ID, id, params); err != nil {
				t.Fatalf("cast: %v", err)
			}
			if tc == "controlled but it left in response" {
				effectstest.KillCreature(t, g, me.ID, dragonOnBf)
			}
			var uncounterable bool
			g.ReadSnapshot(func() { uncounterable = g.SpellCantBeCounteredForEffect(id) })
			if want := tc != "nothing"; uncounterable != want {
				t.Errorf("%s: can't be countered = %v, want %v", tc, uncounterable, want)
			}
		})
	}
}

func TestDtk329TapestryDrawsOnlyAfterANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	tap := effectstest.PushCatalogPermanent(g, me.ID, "Tapestry of the Ages", "Artifact", dtk329Tapestry, false)
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Drawn", TypeLine: "Land"})
	effectstest.B10AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, tap, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("Tapestry activated before a noncreature spell was cast")
	}
	effectstest.CastCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, tap, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("Tapestry after a noncreature spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestDtk329TapestryNeedsItsManaEvenAfterANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	tap := effectstest.PushCatalogPermanent(g, me.ID, "Tapestry of the Ages", "Artifact", dtk329Tapestry, false)
	effectstest.CastCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, tap, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("Tapestry activated with no mana to pay {2}")
	}
}

func TestDtk329ZephyrScribeLootsAndUntapsOnANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	scribe := effectstest.PushCatalogPermanent(g, me.ID, "Zephyr Scribe", "Creature — Human Monk", dtk329Scribe, false)
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Drawn", TypeLine: "Land"})
	effectstest.B10AddMana(me, "U")
	if err := g.ActivateCatalogAbility(me.ID, scribe, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("Scribe activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, scribe); !c.Tapped {
		t.Fatal("the Scribe should be tapped")
	}
	if me.Graveyard.Size() != 0 {
		t.Fatal("the loot discarded before the player chose")
	}
	effectstest.AnswerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	if me.Graveyard.Size() != 1 {
		t.Errorf("graveyard %d after the loot, want 1", me.Graveyard.Size())
	}
	effectstest.CastCatalogSpell(t, g, "Test Creature", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, scribe); !c.Tapped {
		t.Error("a creature spell untapped the Scribe")
	}
	effectstest.CastCatalogSpell(t, g, "Test Instant", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, scribe); c.Tapped {
		t.Error("a noncreature spell did not untap the Scribe")
	}
}
