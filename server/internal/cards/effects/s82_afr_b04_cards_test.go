package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_afr_b04_cards_test.go — Adventures in the Forgotten Realms, S82 batch 4
// (#5499): the twelve modal "You ..." spells and their kin. Each test casts
// the real catalog entry through the engine.

const (
	afrB04GnollCampOracle = "e0945337-1833-4545-a83f-fca673e7f0d0"
	afrB04PrisonersOracle = "467c526a-02a7-405d-98de-c632ea83519b"
	afrB04VillainsOracle  = "41cdeb52-e90f-44f8-94d2-23aa032c4870"
	afrB04WeaponOracle    = "a3acb58d-3fb1-4602-8575-d78b716b0ea5"
	afrB04GuardOracle     = "49aa87c7-034b-409d-b37f-50db6a42e533"
	afrB04GoblinsOracle   = "77d016e3-f652-4827-b49a-8044f6d2d45b"
	afrB04WildShapeOracle = "99f4a22a-ba33-43f0-bc58-9df8388706cd"
	afrB04TavernOracle    = "c735077a-fae8-4c5f-a5bf-35969696460c"
	afrB04AmbushedOracle  = "934ef52e-3e97-418a-bbcf-114badbb2c8a"
	afrB04WatchOracle     = "55d94c09-b6aa-4af5-923a-b4efc2a245df"
	afrB04RiverOracle     = "9913da11-07f0-494a-8c47-11629dcb5be2"
	afrB04GladeOracle     = "88e5300d-8339-41cb-85bf-325909116b61"
)

func TestAfrB04CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		afrB04GnollCampOracle: "You Come to the Gnoll Camp",
		afrB04PrisonersOracle: "You Find Some Prisoners",
		afrB04VillainsOracle:  "You Find the Villains' Lair",
		afrB04WeaponOracle:    "Choose Your Weapon",
		afrB04GuardOracle:     "You See a Guard Approach",
		afrB04GoblinsOracle:   "You See a Pair of Goblins",
		afrB04WildShapeOracle: "Wild Shape",
		afrB04TavernOracle:    "You Meet in a Tavern",
		afrB04AmbushedOracle:  "You're Ambushed on the Road",
		afrB04WatchOracle:     "You Hear Something on Watch",
		afrB04RiverOracle:     "You Come to a River",
		afrB04GladeOracle:     "You Happen On a Glade",
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

func afrB04Target(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func TestAfrB04GnollCampKeepsUpToTwoCreaturesFromBlocking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	w1 := b12Creature(g, opp.ID, "Wall One", "Creature — Wall", 0, 4)
	w2 := b12Creature(g, opp.ID, "Wall Two", "Creature — Wall", 0, 4)
	w3 := b12Creature(g, opp.ID, "Wall Three", "Creature — Wall", 0, 4)
	castModal(t, g, "You Come to the Gnoll Camp", "Instant", afrB04GnollCampOracle, []int{0}, cardRefs(w1, w2))
	passPriorityAroundTable(t, g)
	declareAttack(t, g, opp.ID, bear)
	if blockRefusal(t, g, w1, bear) == "" || blockRefusal(t, g, w2, bear) == "" {
		t.Error("a chosen creature can block")
	}
	if blockRefusal(t, g, w3, bear) != "" {
		t.Error("the unchosen creature cannot block")
	}
}

func TestAfrB04GnollCampFendThemOffPumps(t *testing.T) {
	g := newCatalogGame(t)
	bear := b12Creature(g, g.Seats[0].ID, "Bear", "Creature — Bear", 2, 2)
	castModal(t, g, "You Come to the Gnoll Camp", "Instant", afrB04GnollCampOracle, []int{1}, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 5 || tough != 3 {
		t.Errorf("the Bear is %d/%d, want 5/3", p, tough)
	}
}

func TestAfrB04PrisonersDestroysAnArtifactAndRefusesAnythingElse(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rock := pushArtifactForTest(g, opp.ID, "Rock")
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	toMain(t, g)
	hand := g.Seats[g.Turn.ActiveSeat].Hand
	id := uuid.New()
	hand.PushTop(game.Card{InstanceID: id, Name: "You Find Some Prisoners", TypeLine: "Instant", OracleID: afrB04PrisonersOracle,
		Owner: g.Seats[g.Turn.ActiveSeat].ID, Controller: g.Seats[g.Turn.ActiveSeat].ID})
	if err := g.CastSpell(g.Seats[g.Turn.ActiveSeat].ID, id, game.CastSpellParams{Modes: []int{0}, Targets: cardRefs(bear), Strict: true}); err == nil {
		t.Fatal("Break Their Chains targeted a nonartifact creature")
	}
	if err := g.CastSpell(g.Seats[g.Turn.ActiveSeat].ID, id, game.CastSpellParams{Modes: []int{0}, Targets: cardRefs(rock), Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
}

func TestAfrB04PrisonersInterrogateExilesThreeAndOnePlayableWithAnyColor(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ids := seedSearchLibrary(opp,
		game.Card{Name: "Red Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{R}", Colors: []string{"R"}, Power: 2, Toughness: 2},
		game.Card{Name: "Second", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Third", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Deep", TypeLine: "Creature — Bear", Power: 1, Toughness: 1},
	)
	castModal(t, g, "You Find Some Prisoners", "Instant", afrB04PrisonersOracle, []int{1},
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	for _, id := range ids[:3] {
		if !g.Exile.Contains(id) {
			t.Fatalf("one of the top three cards is not exiled")
		}
	}
	if g.Exile.Contains(ids[3]) || !opp.Library.Contains(ids[3]) {
		t.Error("more than three cards were exiled")
	}
	answerChooseCards(t, g, me.ID, ids[0])
	passPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, ids[1], game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("an unchosen card can be played")
	}
	floatMana(t, g, me, "{G}{G}")
	if err := g.CastSpell(me.ID, ids[0], game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast the chosen red card with green mana: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ids[0]) {
		t.Error("the chosen creature did not enter")
	}
}

func TestAfrB04VillainsLairCountersASpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	victim := castCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	castModal(t, g, "You Find the Villains' Lair", "Instant", afrB04VillainsOracle, []int{0}, afrB04Target(victim))
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(victim) {
		t.Error("the targeted spell was not countered")
	}
}

func TestAfrB04VillainsLairLoots(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B", "C")
	before := me.Hand.Size()
	castModal(t, g, "You Find the Villains' Lair", "Instant", afrB04VillainsOracle, []int{1}, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != before+2 {
		t.Errorf("drew two, then a discard prompt: hand %d → %d", before, me.Hand.Size())
	}
}

func TestAfrB04ChooseYourWeaponDoublesAndShoots(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, g.Seats[0].ID, "Bear", "Creature — Bear", 2, 3)
	bird := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird", Keywords: []string{"flying"},
		Power: 2, Toughness: 5, Owner: opp.ID, Controller: opp.ID,
	})
	castModal(t, g, "Choose Your Weapon", "Instant", afrB04WeaponOracle, []int{0}, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tough != 6 {
		t.Errorf("the Bear is %d/%d, want 4/6", p, tough)
	}
	castModal(t, g, "Choose Your Weapon", "Instant", afrB04WeaponOracle, []int{1}, cardRefs(bird))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bird) {
		t.Error("the flier survived 5 damage")
	}
}

func TestAfrB04GuardTapsAndHides(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	castModal(t, g, "You See a Guard Approach", "Instant", afrB04GuardOracle, []int{0}, cardRefs(theirs))
	passPriorityAroundTable(t, g)
	if !isTapped(g, theirs) {
		t.Error("the creature was not tapped")
	}
	castModal(t, g, "You See a Guard Approach", "Instant", afrB04GuardOracle, []int{1}, cardRefs(mine))
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, mine, "hexproof") {
		t.Error("the creature did not gain hexproof")
	}
}

func TestAfrB04GuardHideRefusesAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	theirs := b12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "You See a Guard Approach", TypeLine: "Instant", OracleID: afrB04GuardOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{1}, Targets: cardRefs(theirs), Strict: true}); err == nil {
		t.Error("Hide targeted a creature the caster does not control")
	}
}

func TestAfrB04GoblinsPumpOrMakeTokens(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	castModal(t, g, "You See a Pair of Goblins", "Instant", afrB04GoblinsOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)
	if effectivePower(t, g, a) != 4 || effectivePower(t, g, theirs) != 2 {
		t.Error("Charge Them should pump only the caster's creatures")
	}
	castModal(t, g, "You See a Pair of Goblins", "Instant", afrB04GoblinsOracle, []int{1}, nil)
	passPriorityAroundTable(t, g)
	gobs := socB8Named(g, me.ID, "Goblin")
	if len(gobs) != 2 {
		t.Fatalf("%d Goblins, want 2", len(gobs))
	}
}

func TestAfrB04WildShapeTurnsACreatureIntoEachForm(t *testing.T) {
	for _, tc := range []struct {
		mode     int
		p, tough int
		kind, kw string
	}{
		{0, 1, 3, "Turtle", "hexproof"},
		{1, 1, 5, "Spider", "reach"},
		{2, 3, 3, "Elephant", "trample"},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 4, 4)
		castModal(t, g, "Wild Shape", "Instant", afrB04WildShapeOracle, []int{tc.mode}, cardRefs(bear))
		passPriorityAroundTable(t, g)
		if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != tc.p || tough != tc.tough {
			t.Errorf("%s: %d/%d, want %d/%d", tc.kind, p, tough, tc.p, tc.tough)
		}
		if !hasKeywordOnBattlefield(t, g, bear, tc.kw) {
			t.Errorf("%s: no %s", tc.kind, tc.kw)
		}
		c := findBattlefieldCardByID(g, bear)
		if c == nil || !c.HasSubtype(tc.kind) || c.HasSubtype("Bear") {
			t.Errorf("%s: should be that type and no longer a Bear", tc.kind)
		}
	}
}

func TestAfrB04TavernFormAPartyTakesAnyNumberOfCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Bear One", TypeLine: "Creature — Bear", Power: 1, Toughness: 1},
		game.Card{Name: "Land", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Bear Two", TypeLine: "Creature — Bear", Power: 1, Toughness: 1},
		game.Card{Name: "Spell", TypeLine: "Sorcery"},
		game.Card{Name: "Bear Three", TypeLine: "Creature — Bear", Power: 1, Toughness: 1},
		game.Card{Name: "Deep Bear", TypeLine: "Creature — Bear", Power: 1, Toughness: 1},
	)
	castModal(t, g, "You Meet in a Tavern", "Sorcery", afrB04TavernOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, ids[0], ids[2], ids[4])
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{ids[0], ids[2], ids[4]} {
		if !me.Hand.Contains(id) {
			t.Error("a chosen creature card is not in hand")
		}
	}
	if !me.Library.Contains(ids[1]) || !me.Library.Contains(ids[3]) || !me.Library.Contains(ids[5]) {
		t.Error("the rest should be on the library (bottom) and the sixth card untouched")
	}
	if me.Library.Size() != 3 {
		t.Errorf("library %d, want 3", me.Library.Size())
	}
}

func TestAfrB04TavernBrawlPumps(t *testing.T) {
	g := newCatalogGame(t)
	a := b12Creature(g, g.Seats[0].ID, "A", "Creature — Bear", 2, 2)
	castModal(t, g, "You Meet in a Tavern", "Sorcery", afrB04TavernOracle, []int{1}, nil)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, a), effectiveToughness(t, g, a); p != 4 || tough != 4 {
		t.Errorf("%d/%d, want 4/4", p, tough)
	}
}

func TestAfrB04AmbushedRetreatsOrPumps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	castModal(t, g, "You're Ambushed on the Road", "Instant", afrB04AmbushedOracle, []int{1}, cardRefs(a))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, a), effectiveToughness(t, g, a); p != 3 || tough != 5 {
		t.Errorf("%d/%d, want 3/5", p, tough)
	}
	castModal(t, g, "You're Ambushed on the Road", "Instant", afrB04AmbushedOracle, []int{0}, cardRefs(a))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || !me.Hand.Contains(a) {
		t.Error("the creature did not return to its owner's hand")
	}
}

func TestAfrB04WatchRousesAndSetsOffTraps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	other := b12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 2)
	castModal(t, g, "You Hear Something on Watch", "Instant", afrB04WatchOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)
	if effectivePower(t, g, a) != 3 || effectivePower(t, g, other) != 2 {
		t.Error("Rouse the Party should pump only the caster's creatures")
	}
	attacker := a
	declareAttack(t, g, opp.ID, attacker)
	lockInAttacks(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "You Hear Something on Watch", TypeLine: "Instant", OracleID: afrB04WatchOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{1}, Targets: cardRefs(other), Strict: true}); err == nil {
		t.Error("Set Off Traps targeted a creature that is not attacking")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{1}, Targets: cardRefs(attacker), Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("the attacking creature survived 5 damage")
	}
}

func TestAfrB04RiverBouncesNonlandsOnly(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rock := pushArtifactForTest(g, opp.ID, "Rock")
	land := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID})
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "You Come to a River", TypeLine: "Instant", OracleID: afrB04RiverOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0}, Targets: cardRefs(land), Strict: true}); err == nil {
		t.Fatal("Fight the Current targeted a land")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0}, Targets: cardRefs(rock), Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(rock) {
		t.Error("the artifact did not return to its owner's hand")
	}
}

func TestAfrB04RiverFindACrossing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	castModal(t, g, "You Come to a River", "Instant", afrB04RiverOracle, []int{1}, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, bear); p != 3 {
		t.Errorf("power %d, want 3", p)
	}
	declareAttack(t, g, opp.ID, bear)
	if blockRefusal(t, g, wall, bear) == "" {
		t.Error("the creature can be blocked")
	}
}

func TestAfrB04GladeSearchesTwoBasicsOrReturnsAPermanentCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 1, Toughness: 1},
	)
	castModal(t, g, "You Happen On a Glade", "Instant", afrB04GladeOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, ids[0], ids[1])
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[0]) || !me.Hand.Contains(ids[1]) {
		t.Error("the two basic lands are not in hand")
	}
}

func TestAfrB04GladeMakeCampReturnsAPermanentCardOnly(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := seedGraveyardCard(t, g, "Bear", "Creature — Bear", "")
	spell := seedGraveyardCard(t, g, "Shock", "Instant", "")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "You Happen On a Glade", TypeLine: "Instant", OracleID: afrB04GladeOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{1}, Targets: cardRefs(spell), Strict: true}); err == nil {
		t.Fatal("Make Camp targeted an instant card")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{1}, Targets: cardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the permanent card did not return to hand")
	}
}
