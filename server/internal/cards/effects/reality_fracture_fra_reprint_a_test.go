package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_fra_reprint_a_test.go — slice fra-reprint-a (tracker
// #2795): older cards reprinted in Reality Fracture.

const (
	oracleRfrAkroma         = "2b80faaf-92fd-4fa0-a3f6-8bb263e7ff1d"
	oracleRfrBlazing        = "6344c96a-efa5-4125-8219-d333228391cf"
	oracleRfrBrainsurge     = "46c727cb-1f47-4775-8af4-0230ef53966b"
	oracleRfrContaminated   = "28196fd9-00c9-4cd0-b603-0eec8511ec79"
	oracleRfrGrandCrescendo = "dc600c06-8239-409e-b53d-20f813a3f5e7"
	oracleRfrLastGasp       = "a82c3860-4dd6-4ffd-aa8f-ab8df687db6c"
	oracleRfrMassPolymorph  = "b133129d-1ffb-4b78-9f5b-852d163cc9b7"
	oracleRfrOccult         = "6df1c314-b97b-4bbc-8b7e-a07785347a49"
	oracleRfrMistmoors      = "7e64b1dc-a238-4bff-98ff-2bea44340568"
	oracleRfrPerilous       = "e2b472dd-047d-47eb-9ebb-df6aa4b52dd4"
)

func rfrCreature(g *game.Game, owner uuid.UUID, name string, p, t int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear", Power: p, Toughness: t,
		Owner: owner, Controller: owner,
	})
}

func rfrTarget(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func TestLastGaspShrinksAndKills(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	big := rfrCreature(g, opp.ID, "Big", 4, 4)
	small := rfrCreature(g, opp.ID, "Small", 2, 3)

	castCatalogSpell(t, g, "Last Gasp", "Instant", oracleRfrLastGasp, rfrTarget(big))
	passPriorityAroundTable(t, g)
	if p, tt := effectivePower(t, g, big), effectiveToughness(t, g, big); p != 1 || tt != 1 {
		t.Errorf("target is %d/%d, want 1/1", p, tt)
	}

	castCatalogSpell(t, g, "Last Gasp", "Instant", oracleRfrLastGasp, rfrTarget(small))
	passPriorityAroundTable(t, g)
	if _, still := battlefieldCard(g, small); still {
		t.Error("a 2/3 hit by -3/-3 should have died")
	}
}

func TestBlazingCrescendoBoostsAndExilesTop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := rfrCreature(g, me.ID, "Bear", 2, 2)
	top := pushLibraryCardForTest(me, game.Card{Name: "Top Land", TypeLine: "Basic Land — Forest"})
	// pushLibraryCardForTest pushes to the bottom; move it to the top.
	g.WithWriteLock(func() {
		for i, c := range me.Library.Cards {
			if c.InstanceID == top {
				me.Library.Cards = append(append([]game.Card{}, me.Library.Cards[:i]...), me.Library.Cards[i+1:]...)
				me.Library.Cards = append(me.Library.Cards, c)
				break
			}
		}
	})

	castCatalogSpell(t, g, "Blazing Crescendo", "Instant", oracleRfrBlazing, rfrTarget(bear))
	passPriorityAroundTable(t, g)

	if p, tt := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 5 || tt != 3 {
		t.Errorf("target is %d/%d, want 5/3", p, tt)
	}
	if !g.Exile.Contains(top) {
		t.Error("the top card of the library should be exiled")
	}
}

func TestGrandCrescendoMakesCitizensThatAreIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castFromHand(t, g, game.Card{
		Name: "Grand Crescendo", TypeLine: "Instant", OracleID: oracleRfrGrandCrescendo, ManaCost: "{X}{W}{W}",
	}, game.CastSpellParams{XValue: 3})
	passPriorityAroundTable(t, g)

	var citizens []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Citizen" && c.Controller == me.ID {
			citizens = append(citizens, c.InstanceID)
		}
	}
	if len(citizens) != 3 {
		t.Fatalf("%d Citizens, want 3", len(citizens))
	}
	for _, id := range citizens {
		if !effectiveAbilitiesContain(t, g, id, "indestructible") {
			t.Error("a new Citizen should be indestructible")
		}
	}
}

func TestBrainsurgeDrawsFourAndPutsTwoBack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Brainsurge", "Instant", oracleRfrBrainsurge, nil)
	passPriorityAroundTable(t, g)

	// The spell left the hand and four cards came in.
	if n := me.Hand.Size(); n != before+4 {
		t.Fatalf("hand size %d, want %d before the put-back", n, before+4)
	}
	prompt := chooseCardsChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no put-back prompt for the caster")
	}
	if prompt.ChooseMax != 2 {
		t.Errorf("prompt asks for %d cards, want 2", prompt.ChooseMax)
	}
}

func TestOccultEpiphanyDrawsDiscardsAndMakesSpiritsPerType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	// Empty the hand so the discard is exactly the cards we choose among.
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	land := uuid.New()
	sorc := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: land, Name: "Forest A", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
		me.Hand.PushTop(game.Card{InstanceID: sorc, Name: "Spell A", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	})
	castFromHand(t, g, game.Card{
		Name: "Occult Epiphany", TypeLine: "Instant", OracleID: oracleRfrOccult, ManaCost: "{X}{U}",
	}, game.CastSpellParams{XValue: 2})
	passPriorityAroundTable(t, g)

	// Two cards were drawn; discard the land and the sorcery.
	answerDiscard(t, g, me.ID, land, sorc)
	passPriorityAroundTable(t, g)

	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spirit" && c.Controller == me.ID {
			n++
		}
	}
	if n != 2 {
		t.Errorf("%d Spirits, want 2 (a land and a sorcery)", n)
	}
}

func TestOccultEpiphanySharedTypeCountsOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	a, b := uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: a, Name: "Forest A", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
		me.Hand.PushTop(game.Card{InstanceID: b, Name: "Forest B", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	})
	castFromHand(t, g, game.Card{
		Name: "Occult Epiphany", TypeLine: "Instant", OracleID: oracleRfrOccult, ManaCost: "{X}{U}",
	}, game.CastSpellParams{XValue: 2})
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID, a, b)
	passPriorityAroundTable(t, g)

	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spirit" && c.Controller == me.ID {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d Spirits, want 1 (two lands share one type)", n)
	}
}

func TestMassPolymorphExilesYourCreaturesAndRevealsThatMany(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine1 := rfrCreature(g, me.ID, "Mine One", 1, 1)
	mine2 := rfrCreature(g, me.ID, "Mine Two", 1, 1)
	theirs := rfrCreature(g, opp.ID, "Theirs", 1, 1)
	// Library top-first: land, creature, land, creature, creature.
	g.WithWriteLock(func() {
		me.Library.Cards = nil
		for _, c := range []game.Card{
			{Name: "Deep Bear", TypeLine: "Creature — Bear"},
			{Name: "Bear C", TypeLine: "Creature — Bear"},
			{Name: "Land B", TypeLine: "Basic Land — Island"},
			{Name: "Bear A", TypeLine: "Creature — Bear"},
			{Name: "Land A", TypeLine: "Basic Land — Forest"},
		} {
			c.InstanceID = uuid.New()
			c.Owner, c.Controller = me.ID, me.ID
			me.Library.Cards = append(me.Library.Cards, c)
		}
	})

	castCatalogSpell(t, g, "Mass Polymorph", "Sorcery", oracleRfrMassPolymorph, nil)
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{mine1, mine2} {
		if _, still := battlefieldCard(g, id); still || !g.Exile.Contains(id) {
			t.Error("each of my creatures should be exiled")
		}
	}
	if _, ok := battlefieldCard(g, theirs); !ok {
		t.Error("an opponent's creature must stay")
	}
	names := map[string]bool{}
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID {
			names[c.Name] = true
		}
	}
	if !names["Bear A"] || !names["Bear C"] || names["Deep Bear"] {
		t.Errorf("battlefield after Mass Polymorph: %v, want Bear A and Bear C only", names)
	}
	if me.Library.Size() != 3 {
		t.Errorf("library has %d cards, want 3 (the lands and the unrevealed bear)", me.Library.Size())
	}
}

func TestMassPolymorphWithNoCreaturesRevealsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	libBefore := me.Library.Size()
	castCatalogSpell(t, g, "Mass Polymorph", "Sorcery", oracleRfrMassPolymorph, nil)
	passPriorityAroundTable(t, g)
	if me.Library.Size() != libBefore {
		t.Errorf("library %d -> %d, want unchanged", libBefore, me.Library.Size())
	}
}

func TestAkromaAngelOfFuryShape(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := castFromHand(t, g, game.Card{
		Name: "Akroma, Angel of Fury", TypeLine: "Legendary Creature — Angel", OracleID: oracleRfrAkroma,
		ManaCost: "{5}{R}{R}{R}", Power: 6, Toughness: 6,
	}, game.CastSpellParams{})
	var uncounterable bool
	g.ReadSnapshot(func() { uncounterable = g.SpellCantBeCounteredForEffect(id) })
	if !uncounterable {
		t.Error("Akroma should be uncounterable on the stack")
	}
	passPriorityAroundTable(t, g)
	for _, kw := range []string{"flying", "trample", game.ProtectionFromColor("W"), game.ProtectionFromColor("U")} {
		if !effectiveAbilitiesContain(t, g, id, kw) {
			t.Errorf("Akroma lacks %q", kw)
		}
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}{R}"); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("firebreathing: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	if p := effectivePower(t, g, id); p != 8 {
		t.Errorf("power %d after two pumps, want 8", p)
	}
}

func TestAkromaCanBeCastFaceDown(t *testing.T) {
	g := newCatalogGame(t)
	id := castFromHand(t, g, game.Card{
		Name: "Akroma, Angel of Fury", TypeLine: "Legendary Creature — Angel", OracleID: oracleRfrAkroma,
		ManaCost: "{5}{R}{R}{R}", Power: 6, Toughness: 6,
	}, game.CastSpellParams{AlternativeCost: "morph"})
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, id)
	if !ok || !c.FaceDown {
		t.Fatalf("Akroma should be on the battlefield face down: ok=%v", ok)
	}
	if p := effectivePower(t, g, id); p != 2 {
		t.Errorf("face-down power %d, want 2", p)
	}
}

func TestOverlordOfTheMistmoorsMakesTwoInsectsImpending(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	overlord := castFromHand(t, g, game.Card{
		Name: "Overlord of the Mistmoors", TypeLine: "Enchantment Creature — Avatar Horror",
		OracleID: oracleRfrMistmoors, ManaCost: "{5}{W}{W}", Power: 6, Toughness: 6,
	}, game.CastSpellParams{AlternativeCost: AltCostKeyImpending})
	passPriorityAroundTable(t, g)

	if n := countersOn(g, overlord, game.CounterTime); n != 4 {
		t.Errorf("%d time counters, want 4", n)
	}
	if containsString(effectiveTypes(t, g, overlord), "Creature") {
		t.Error("the impending Overlord is not a creature")
	}
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Insect" && c.Controller == me.ID {
			n++
			if c.Power != 2 || c.Toughness != 1 {
				t.Errorf("Insect is %d/%d, want 2/1", c.Power, c.Toughness)
			}
			if !effectiveAbilitiesContain(t, g, c.InstanceID, "flying") {
				t.Error("an Insect should fly")
			}
		}
	}
	if n != 2 {
		t.Errorf("%d Insects, want 2", n)
	}
}

func TestLandscapesFetchTheirThreeBasicsTapped(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		fetch        []string
		refuse       string
	}{
		{"Perilous Landscape", oracleRfrPerilous, []string{"Island", "Mountain", "Plains"}, "Swamp"},
		{"Contaminated Landscape", oracleRfrContaminated, []string{"Plains", "Island", "Swamp"}, "Mountain"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			land := pushCatalogPermanent(g, me.ID, tc.name, "Land", tc.oracle, false)
			for _, sub := range append(append([]string{}, tc.fetch...), tc.refuse) {
				stapleLibraryCard(me, "Basic "+sub, "Basic Land — "+sub)
			}
			if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{}); err != nil {
				t.Fatalf("fetch: %v", err)
			}
			if _, still := battlefieldCard(g, land); still {
				t.Error("the land is sacrificed as a cost")
			}
			passPriorityAroundTable(t, g)
			c := searchChoiceFor(g, me.ID)
			if c == nil {
				t.Fatal("no search prompt")
			}
			if searchOptionNamed(g, c, "Basic "+tc.refuse) != uuid.Nil {
				t.Errorf("the search offered a basic %s", tc.refuse)
			}
			for _, sub := range tc.fetch {
				if searchOptionNamed(g, c, "Basic "+sub) == uuid.Nil {
					t.Errorf("the search did not offer a basic %s", sub)
				}
			}
			answerSearchNamed(t, g, me.ID, "Basic "+tc.fetch[0])
			passPriorityAroundTable(t, g)
			found := false
			for _, bc := range g.Battlefield.Cards {
				if bc.Name == "Basic "+tc.fetch[0] {
					found = true
					if !bc.Tapped {
						t.Error("the fetched land should enter tapped")
					}
				}
			}
			if !found {
				t.Error("the fetched land did not reach the battlefield")
			}
		})
	}
}

func TestLandscapesCycleForTheirCost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	id := pushCatalogHandCard(me, "Perilous Landscape", "Land", oracleRfrPerilous)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{R}{W}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	handBefore := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id) {
		t.Error("the cycled land should be in the graveyard")
	}
	if me.Hand.Size() != handBefore {
		t.Errorf("hand %d -> %d, want the cycled card replaced by a draw", handBefore, me.Hand.Size())
	}
}
