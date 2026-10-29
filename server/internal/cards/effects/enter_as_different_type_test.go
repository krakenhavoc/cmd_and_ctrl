package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// enter_as_different_type_test.go — ADR 0210 (#4506): a card put onto
// the battlefield as a different type. Supper for Spiders' Food
// artifacts and Tom, Bert, and William's artifact, each driven through
// the real catalog entry.

const (
	e4506SoulWardenOracle = "f3fad295-1af2-4ecc-8546-b121ad6be27b"
	e4506AuthorityOracle  = "55f3c721-e13a-406e-bc8e-d6cdc91ac477"
	e4506WrathOracle      = "34515b16-c9a4-4f98-8c77-416a7a523407"
	e4506CloneOracle      = "42226b87-0746-4ebf-9fd0-108d508462af"
)

func TestE4506EnterAsADifferentTypeCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		supperForSpidersOracle:  "Supper for Spiders",
		tomBertAndWilliamOracle: "Tom, Bert, and William",
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

// e4506View is a battlefield card with its layers settled.
func e4506View(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	return s58p6View(t, g, id)
}

// e4506AssertFood fails unless `id` is a noncreature Food artifact with
// no creature subtypes and Supper's granted ability.
func e4506AssertFood(t *testing.T, g *game.Game, id uuid.UUID, controller uuid.UUID) {
	t.Helper()
	c := e4506View(t, g, id)
	ch := c.Effective()
	if c.IsCreature() {
		t.Errorf("%s is a creature on the battlefield: types %v", c.Name, ch.Types)
	}
	if !slices.Equal(ch.Types, []string{"Artifact"}) {
		t.Errorf("%s: types %v, want [Artifact]", c.Name, ch.Types)
	}
	if !slices.Equal(ch.Subtypes, []string{"Food"}) {
		t.Errorf("%s: subtypes %v, want [Food] (they lose all other subtypes)", c.Name, ch.Subtypes)
	}
	if c.Controller != controller {
		t.Errorf("%s: controller %s, want %s", c.Name, c.Controller, controller)
	}
	if len(ch.GrantedAbilities) == 0 {
		t.Errorf("%s has no granted Food ability", c.Name)
	}
}

// Supper takes exactly the creature cards put into an OPPONENT's
// graveyard from the battlefield THIS turn: not a milled one, not last
// turn's, not yours, not a noncreature card.
func TestE4506SupperForSpidersTakesOnlyOpponentsCreaturesThatDiedThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]

	old := b12Creature(g, opp.ID, "Old Wolf", "Creature — Wolf", 2, 2)
	destroyNow(t, g, old)
	for range 4 {
		endTurn(t, g)
	}
	if g.Turn.ActiveSeat != me.Seat {
		t.Fatalf("not back to my turn: seat %d", g.Turn.ActiveSeat)
	}
	advanceToMain(t, g)

	wolf := b12Creature(g, opp.ID, "Warg", "Creature — Wolf", 3, 3)
	spider := b12Creature(g, opp2.ID, "Spider", "Creature — Spider", 2, 4)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	for _, id := range []uuid.UUID{wolf, spider, mine, rock} {
		destroyNow(t, g, id)
	}
	milled := seedLibraryTop(opp, "Milled Troll", "Creature — Troll", "{2}")
	g.WithWriteLock(func() {
		if err := g.MillNForEffect(opp.ID, 1); err != nil {
			t.Fatalf("mill: %v", err)
		}
	})
	if !opp.Graveyard.Contains(milled) {
		t.Fatal("setup: the Troll was not milled")
	}

	castCatalogSpell(t, g, "Supper for Spiders", "Instant", supperForSpidersOracle, nil)
	passPriorityAroundTable(t, g)

	for name, id := range map[string]uuid.UUID{"opponent's creature": wolf, "second opponent's creature": spider} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("the %s did not come back", name)
			continue
		}
		e4506AssertFood(t, g, id, me.ID)
	}
	if !opp.Graveyard.Contains(old) {
		t.Error("last turn's creature left the graveyard")
	}
	if !opp.Graveyard.Contains(milled) {
		t.Error("a milled creature card left the graveyard")
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("my own creature left my graveyard")
	}
	if !opp.Graveyard.Contains(rock) {
		t.Error("a noncreature card left the graveyard")
	}
}

// The returned cards are Food artifacts from the moment they land: a
// "whenever another creature enters" trigger does not fire, a
// legendary creature stays legendary (CR 205.4b), a 0/0 is not killed by
// CR 704.5f, and the granted Food ability gains 3 life.
func TestE4506SupperForSpidersFoodsAreNeverCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Soul Warden", "Creature — Human Cleric", e4506SoulWardenOracle, false)

	legend := b12Creature(g, opp.ID, "Smaug", "Legendary Creature — Dragon", 6, 6)
	hydra := b12Creature(g, opp.ID, "Hydra", "Creature — Hydra", 0, 0)
	for _, id := range []uuid.UUID{legend, hydra} {
		destroyNow(t, g, id)
	}
	life := me.Life

	castCatalogSpell(t, g, "Supper for Spiders", "Instant", supperForSpidersOracle, nil)
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{legend, hydra} {
		if !g.Battlefield.Contains(id) {
			t.Fatalf("a Food is not on the battlefield")
		}
		e4506AssertFood(t, g, id, me.ID)
	}
	if me.Life != life {
		t.Errorf("Soul Warden saw a creature enter: life %d, want %d", me.Life, life)
	}
	if ch := e4506View(t, g, legend).Effective(); !slices.Contains(ch.Supertypes, "Legendary") {
		t.Errorf("supertypes %v: a Food made from a legend is still legendary", ch.Supertypes)
	}

	b06AddMana(me, "C", "C")
	b16Activate(t, g, me.ID, legend, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d after the Food ability, want %d", me.Life, life+3)
	}
	if g.Battlefield.Contains(legend) || !opp.Graveyard.Contains(legend) {
		t.Error("the Food was not sacrificed to its owner's graveyard")
	}
}

// CR 614.12: the entry's replacements judge the card as the Food it
// enters as. An opponent's Authority of the Consuls taps "creatures your
// opponents control" as they enter, and gains life when one does: a Food
// is neither.
func TestE4506SupperForSpidersFoodsAreNotCreaturesToEntryReplacements(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, opp.ID, "Authority of the Consuls", "Enchantment", e4506AuthorityOracle, false)
	wolf := b12Creature(g, opp.ID, "Warg", "Creature — Wolf", 3, 3)
	destroyNow(t, g, wolf)
	oppLife := opp.Life

	castCatalogSpell(t, g, "Supper for Spiders", "Instant", supperForSpidersOracle, nil)
	passPriorityAroundTable(t, g)

	e4506AssertFood(t, g, wolf, me.ID)
	if c := e4506View(t, g, wolf); c.Tapped {
		t.Error("the Food entered tapped: Authority of the Consuls judged it a creature")
	}
	if opp.Life != oppLife {
		t.Errorf("Authority of the Consuls gained life for a Food: %d, want %d", opp.Life, oppLife)
	}
}

// The overwrite belongs to the object: a flickered Food returns as the
// creature card it is (CR 400.7), and Soul Warden sees that one. A token
// copy of a Food copies the printed creature (CR 707.2) with no Food
// ability.
func TestE4506SupperForSpidersOverwriteEndsWithTheObject(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Soul Warden", "Creature — Human Cleric", e4506SoulWardenOracle, false)
	wolf := b12Creature(g, opp.ID, "Warg", "Creature — Wolf", 3, 3)
	spider := b12Creature(g, opp.ID, "Spider", "Creature — Spider", 2, 4)
	destroyNow(t, g, wolf)
	destroyNow(t, g, spider)
	castCatalogSpell(t, g, "Supper for Spiders", "Instant", supperForSpidersOracle, nil)
	passPriorityAroundTable(t, g)
	e4506AssertFood(t, g, wolf, me.ID)
	e4506AssertFood(t, g, spider, me.ID)

	// The copy.
	g.WithWriteLock(func() {
		if err := (CreateTokenCopy{Controller: me.ID, Copy: spider, N: 1}).Apply(NewContext(g, nil)); err != nil {
			t.Fatalf("token copy: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	var copyID uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spider" && c.InstanceID != spider {
			copyID = c.InstanceID
		}
	}
	if copyID == uuid.Nil {
		t.Fatal("no token copy of the Food")
	}
	cp := e4506View(t, g, copyID)
	if !cp.IsCreature() || slices.Contains(cp.Effective().Subtypes, "Food") || len(cp.Effective().GrantedAbilities) != 0 {
		t.Errorf("the copy is %v %v with %d grants, want the printed creature",
			cp.Effective().Types, cp.Effective().Subtypes, len(cp.Effective().GrantedAbilities))
	}

	// The flicker.
	life := me.Life
	g.WithWriteLock(func() {
		if err := (Flicker{Target: wolf}).Apply(ctxFor(g, &game.StackItem{Controller: me.ID, SourceCardID: uuid.New()})); err != nil {
			t.Fatalf("flicker: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	var back uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Warg" {
			back = c.InstanceID
		}
	}
	if back == uuid.Nil {
		t.Fatal("the flickered Warg did not come back")
	}
	c := e4506View(t, g, back)
	if !c.IsCreature() || slices.Contains(c.Effective().Subtypes, "Food") {
		t.Errorf("the flickered card is %v %v, want the creature it is", c.Effective().Types, c.Effective().Subtypes)
	}
	if me.Life <= life {
		t.Errorf("Soul Warden did not see the flickered creature enter: life %d", me.Life)
	}
	if n := e4506OverwriteRecords(g); n != 1 {
		t.Errorf("%d overwrite records after the flicker, want 1 (the Spider's): the old object's record was not swept", n)
	}
}

// e4506OverwriteRecords counts the live ADR 0210 records.
func e4506OverwriteRecords(g *game.Game) int {
	n := 0
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		for _, e := range g.ScopedEffects {
			if e.Label == "Supper for Spiders — a Food artifact" ||
				e.Label == "Tom, Bert, and William — an artifact (no longer a creature)" {
				n++
			}
		}
	})
	return n
}

// A table with a Food on it is a restore point, and the Food is still a
// Food with its ability after the round trip.
func TestE4506SupperForSpidersFoodSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	wolf := b12Creature(g, opp.ID, "Warg", "Creature — Wolf", 3, 3)
	destroyNow(t, g, wolf)
	castCatalogSpell(t, g, "Supper for Spiders", "Instant", supperForSpidersOracle, nil)
	passPriorityAroundTable(t, g)

	r := restoreRoundTrip(t, g, true)
	rme := r.Seats[0]
	e4506AssertFood(t, r, wolf, rme.ID)
	life := rme.Life
	b06AddMana(rme, "C", "C")
	b16Activate(t, r, rme.ID, wolf, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, r)
	if rme.Life != life+3 {
		t.Errorf("restored Food gained %d life, want 3", rme.Life-life)
	}
	_ = me
}

// --- Tom, Bert, and William -----------------------------------------------

func e4506PushTrolls(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushCatalogPermanent(g, owner, "Tom, Bert, and William", "Legendary Creature — Troll", tomBertAndWilliamOracle, false)
}

// They die once and return as a legendary noncreature artifact with
// their abilities; a Wrath of God does not touch them; destroyed as an
// artifact they stay in the graveyard (the intervening if).
func TestE4506TomBertAndWilliamReturnAsAnArtifactOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Soul Warden", "Creature — Human Cleric", e4506SoulWardenOracle, false)
	trolls := e4506PushTrolls(g, me.ID)

	destroyNow(t, g, trolls)
	life := me.Life
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(trolls) {
		t.Fatal("Tom, Bert, and William did not return")
	}
	c := e4506View(t, g, trolls)
	ch := c.Effective()
	if c.IsCreature() || !c.IsArtifact() {
		t.Errorf("types %v, want a noncreature artifact", ch.Types)
	}
	if slices.Contains(ch.Subtypes, "Troll") {
		t.Errorf("subtypes %v: no creature type on an artifact", ch.Subtypes)
	}
	if !slices.Contains(ch.Supertypes, "Legendary") {
		t.Errorf("supertypes %v, want Legendary kept", ch.Supertypes)
	}
	if c.Controller != me.ID {
		t.Errorf("controller %s, want the owner", c.Controller)
	}
	if me.Life != life {
		t.Errorf("Soul Warden saw a creature enter: life %d, want %d", me.Life, life)
	}

	castCatalogSpell(t, g, "Wrath of God", "Sorcery", e4506WrathOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(trolls) {
		t.Fatal("Wrath of God destroyed the artifact as if it were a creature")
	}

	destroyNow(t, g, trolls)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(trolls) || !me.Graveyard.Contains(trolls) {
		t.Error("dying as an artifact returned them again")
	}
	if n := e4506OverwriteRecords(g); n != 0 {
		t.Errorf("%d overwrite records left after the artifact died, want 0", n)
	}

	// Reanimated by something else, the card is a new object (CR 400.7)
	// with the same instance: a creature again, with no overwrite.
	g.WithWriteLock(func() {
		if err := g.ReturnFromGraveyardForEffect(trolls, game.ZoneBattlefield); err != nil {
			t.Fatalf("reanimate: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(trolls) {
		t.Fatal("the reanimation did not land")
	}
	if c := e4506View(t, g, trolls); !c.IsCreature() || !slices.Contains(c.Effective().Subtypes, "Troll") {
		t.Errorf("reanimated as %v %v, want the Troll creature it is", c.Effective().Types, c.Effective().Subtypes)
	}
}

// The artifact keeps its abilities: {1}, sacrifice another creature
// draws cards equal to its power, then discards one.
func TestE4506TomBertAndWilliamSacrificeDrawsThePowerThenDiscards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	seedLibrary(me, "One", "Two", "Three", "Four")
	trolls := e4506PushTrolls(g, me.ID)
	destroyNow(t, g, trolls)
	passPriorityAroundTable(t, g)
	if c := e4506View(t, g, trolls); c.IsCreature() || !c.IsArtifact() {
		t.Fatalf("setup: Tom, Bert, and William did not return as an artifact")
	}
	fodder := b12Creature(g, me.ID, "Ogre", "Creature — Ogre", 3, 3)
	hand := me.Hand.Size()

	b06AddMana(me, "C")
	b16Activate(t, g, me.ID, trolls, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(fodder) {
		t.Fatal("the Ogre was not sacrificed")
	}
	if got := me.Hand.Size(); got != hand+3 {
		t.Fatalf("hand %d after the draw, want %d", got, hand+3)
	}
	answerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d after the discard, want %d", got, hand+2)
	}
}

// A Clone taken by Supper asks its "enter as a copy" question with the
// overwrite riding the paused entry. The table is a restore point, and
// after the answer the permanent is a copy (layer 1) that is a Food
// artifact (layer 4), not a creature.
func TestE4506SupperForSpidersPausedCloneEntryRestoresWithTheOverwrite(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	clone := pushCatalogPermanent(g, opp.ID, "Clone", "Creature — Shapeshifter", e4506CloneOracle, false)
	destroyNow(t, g, clone)
	if !opp.Graveyard.Contains(clone) {
		t.Fatal("setup: the Clone is not in the graveyard")
	}
	// A second card in the same entry, whose window opens after the
	// Clone's answer: the batch's own options carry the overwrite.
	wolf := b12Creature(g, opp.ID, "Warg", "Creature — Wolf", 3, 3)
	destroyNow(t, g, wolf)

	castCatalogSpell(t, g, "Supper for Spiders", "Instant", supperForSpidersOracle, nil)
	passPriorityAroundTable(t, g)
	if copyPrompt(g) == nil {
		t.Fatal("the Clone did not ask what to copy")
	}

	r := restoreRoundTrip(t, g, true)
	cp := copyPrompt(r)
	if cp == nil {
		t.Fatal("the copy question did not survive the restore")
	}
	if err := r.ResolveCopyTarget(cp.ID, cp.Chooser, bear); err != nil {
		t.Fatalf("ResolveCopyTarget: %v", err)
	}
	passPriorityAroundTable(t, r)
	if !r.Battlefield.Contains(clone) {
		t.Fatal("the Clone did not land")
	}
	if c := e4506View(t, r, clone); c.Name != "Bear" {
		t.Errorf("name %q, want the copied Bear", c.Name)
	}
	e4506AssertFood(t, r, clone, r.Seats[0].ID)
	e4506AssertFood(t, r, wolf, r.Seats[0].ID)
}
