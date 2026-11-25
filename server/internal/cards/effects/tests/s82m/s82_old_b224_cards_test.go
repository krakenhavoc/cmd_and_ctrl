package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b224_cards_test.go — Ixalan card draw and library, batch 224 (#5532).
// Each test plays the real catalog entry and asserts what its text says.

const (
	oldB224Sanctuary = "15bd8fa1-0858-4029-9973-e24c86e45af3"
	oldB224Rebirth   = "c35540ad-9ab1-4be9-b97d-603813394a26"
	oldB224Plunder   = "8cdfd9df-4783-4138-a286-828a97f516aa"
	oldB224Quarter   = "40abba38-533b-4f51-99e9-a2b337d65171"
	oldB224Mount     = "139b0fd5-8aa7-4b36-ab5d-cbf398381c6d"
	oldB224Ruse      = "58a55d14-b03b-4648-b168-ad35ff088b19"
	oldB224Current   = "a10a4bb2-c6fa-482b-9c8b-e64cbb8424ba"
	oldB224Priest    = "cc11efa2-27ca-4745-8178-31576539512a"
)

func TestOldB224CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB224Sanctuary: "Shapers' Sanctuary", oldB224Rebirth: "Verdant Rebirth", oldB224Plunder: "Costly Plunder",
		oldB224Quarter: "Deadeye Quartermaster", oldB224Mount: "Sun-Blessed Mount", oldB224Ruse: "Siren's Ruse",
		oldB224Current: "Grasping Current", oldB224Priest: "Priest of the Wakening Sun",
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

func oldB224Target(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func TestOldB224CostlyPlunderSacrificesAnArtifactOrCreatureAndDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	relic := effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")
	effectstest.SeedLibrary(me, "a", "b", "c")
	id := effectstest.B13HandCard(me, "Costly Plunder", "Instant", oldB224Plunder)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast with no sacrifice")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Fatal("sacrificed a land to pay for it")
	}
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, SacrificeIDs: []uuid.UUID{relic}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if g.Battlefield.Contains(relic) {
		t.Error("the artifact was not sacrificed as it was cast")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand-1+2 {
		t.Errorf("hand %d, want %d (cast one, drew two)", me.Hand.Size(), hand+1)
	}
}

func TestOldB224ShapersSanctuaryDrawsWhenAnOpponentTargetsYourCreature(t *testing.T) {
	cases := []struct {
		name     string
		actorOpp bool
		mine     bool
		typeLine string
		want     int
	}{
		{"an opponent's spell on my creature", true, true, "Creature — Bear", 1},
		{"my own spell on my creature", false, true, "Creature — Bear", 0},
		{"an opponent's spell on my artifact", true, true, "Artifact", 0},
		{"an opponent's spell on their own creature", true, false, "Creature — Bear", 0},
	}
	for _, tc := range cases {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.SeatsForCombat(g)
		effectstest.PushCatalogPermanent(g, me.ID, "Shapers' Sanctuary", "Enchantment", oldB224Sanctuary, false)
		effectstest.SeedLibrary(me, "a", "b", "c")
		owner := opp.ID
		if tc.mine {
			owner = me.ID
		}
		victim := effectstest.B12Permanent(g, owner, "Victim", tc.typeLine)
		if tc.typeLine != "Artifact" {
			victim = effectstest.B12Creature(g, owner, "Victim", tc.typeLine, 2, 2)
		}
		actor := me.ID
		if tc.actorOpp {
			actor = opp.ID
		}
		hand := me.Hand.Size()
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventBecomesTarget, StackItemID: uuid.New(), Actor: actor, CardID: victim})
		})
		effectstest.PassPriorityAroundTable(t, g)
		if tc.want > 0 {
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
			effectstest.PassPriorityAroundTable(t, g)
		}
		if got := me.Hand.Size() - hand; got != tc.want {
			t.Errorf("%s: drew %d, want %d", tc.name, got, tc.want)
		}
	}
}

func TestOldB224DeadeyeQuartermasterFetchesAnEquipmentOrVehicle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Sword", TypeLine: "Artifact — Equipment"},
		game.Card{Name: "Cart", TypeLine: "Artifact — Vehicle"})
	effectstest.CastCatalogSpell(t, g, "Deadeye Quartermaster", "Creature — Human Pirate", oldB224Quarter, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("a creature card was offered")
	}
	if effectstest.SearchOptionNamed(g, c, "Cart") == uuid.Nil || effectstest.SearchOptionNamed(g, c, "Sword") == uuid.Nil {
		t.Fatal("the Equipment and the Vehicle were not both offered")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ids[2])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[2]) {
		t.Error("the Vehicle did not reach the hand")
	}
}

func TestOldB224SunBlessedMountFetchesHuatliFromTheLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Huatli, Dinosaur Knight", TypeLine: "Legendary Planeswalker — Huatli"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.CastCatalogSpell(t, g, "Sun-Blessed Mount", "Creature — Dinosaur", oldB224Mount, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	if c := effectstest.SearchChoiceFor(g, me.ID); c != nil && c.SearchZonesPending() {
		effectstest.AnswerSearchZones(t, g, me.ID, "library")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ids[0])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[0]) {
		t.Error("Huatli did not reach the hand")
	}
}

func TestOldB224GraspingCurrentBouncesTwoAndFetchesJace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	a := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	keep := effectstest.B12Creature(g, opp.ID, "Stays", "Creature — Bear", 2, 2)
	jace := effectstest.PushGraveyardCardTyped(me, "Jace, Ingenious Mind-Mage", "Legendary Planeswalker — Jace")
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.CastCatalogSpell(t, g, "Grasping Current", "Sorcery", oldB224Current,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("a target stayed on the battlefield")
	}
	if !g.Battlefield.Contains(keep) {
		t.Error("an untargeted creature left")
	}
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	if c := effectstest.SearchChoiceFor(g, me.ID); c != nil && c.SearchZonesPending() {
		effectstest.AnswerSearchZones(t, g, me.ID, "graveyard")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, jace)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(jace) {
		t.Error("Jace did not reach the hand")
	}
}

func TestOldB224SirensRuseDrawsOnlyForAPirate(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		want     int
	}{{"Creature — Human Pirate", 1}, {"Creature — Bear", 0}} {
		g := effectstest.NewCatalogGame(t)
		me, _ := effectstest.SeatsForCombat(g)
		effectstest.SeedLibrary(me, "a", "b", "c")
		target := effectstest.B12Creature(g, me.ID, "Target", tc.typeLine, 2, 2)
		hand := me.Hand.Size()
		effectstest.CastCatalogSpell(t, g, "Siren's Ruse", "Instant", oldB224Ruse, oldB224Target(target))
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(target) || effectstest.NamedOnBattlefield(g, "Target", me.ID) != 1 {
			t.Errorf("%s: the creature did not leave and come back as a new object", tc.typeLine)
		}
		if got := me.Hand.Size() - hand; got != tc.want {
			t.Errorf("%s: hand changed by %d, want %d", tc.typeLine, got, tc.want)
		}
	}
}

func TestOldB224SirensRuseRefusesAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.SeatsForCombat(g)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Pirate", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Siren's Ruse", "Instant", oldB224Ruse, oldB224Target(theirs)); err == nil {
		t.Error("targeted a creature the caster doesn't control")
	}
}

func TestOldB224VerdantRebirthReturnsTheCreatureToHandIfItDiesThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	effectstest.SeedLibrary(me, "a", "b", "c")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Verdant Rebirth", "Instant", oldB224Rebirth, oldB224Target(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (the draw)", me.Hand.Size(), hand+1)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the creature did not return to its owner's hand")
	}
}

func TestOldB224PriestOfTheWakeningSunGainsLifeForARevealedDinosaur(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Priest of the Wakening Sun", "Creature — Human Cleric", oldB224Priest, false)
	dino := effectstest.B13HandCard(me, "Dino", "Creature — Dinosaur", "")
	life := me.Life
	effectstest.AdvanceToUpkeepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, dino)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
}

func TestOldB224PriestOfTheWakeningSunSacrificesToFetchADinosaur(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	priest := effectstest.B12Push(g, me.ID, "Priest of the Wakening Sun", "Creature — Human Cleric", oldB224Priest, 1, 1)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Raptor", TypeLine: "Creature — Dinosaur"})
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, priest, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana")
	}
	effectstest.B10AddMana(me, "W", "W", "W", "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, priest, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(priest) {
		t.Error("the Priest was not sacrificed")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) {
		t.Error("the Dinosaur did not reach the hand")
	}
}
