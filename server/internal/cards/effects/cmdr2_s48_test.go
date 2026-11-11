package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	courtOfEmberethOracle   = "87a82441-9dfb-49bd-8b9e-eaab2c552843"
	courtOfArdenvaleOracle  = "be14d3be-295e-431c-b891-590164b730dd"
	courtOfGarenbrigOracle  = "1f3f6336-6f99-4181-be3c-1814f7b8ef36"
	graveVenerationsOracle  = "1cb71772-592d-4dd1-91af-816933af760c"
	azureFleetAdmiralOracle = "b39f228f-d146-49c6-bab4-5304e0f742d8"
)

func TestCmdr2S48MonarchCardsAreFull(t *testing.T) {
	for _, id := range []string{courtOfEmberethOracle, courtOfArdenvaleOracle, courtOfGarenbrigOracle, graveVenerationsOracle, azureFleetAdmiralOracle} {
		spec, ok := Lookup(id)
		if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v %v", id, spec, ok)
		}
	}
}

// Court of Embereth: the Knight counts toward X, and X is dealt only to
// opponents, only while the controller is the monarch.
func TestCourtOfEmberethKnightThenDamageForTheMonarch(t *testing.T) {
	for _, tc := range []struct {
		name    string
		monarch int
		want    int
	}{{"monarch", 1, 3}, {"not the monarch", 0, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			owner := g.Seats[1]
			pushPermanentForTest(g, owner.ID, "Court of Embereth", courtOfEmberethOracle, "Enchantment")
			pushVanillaCreature(g, owner.ID, "Bear", 2, 2)
			pushVanillaCreature(g, owner.ID, "Cub", 1, 1)
			monCrown(t, g, g.Seats[tc.monarch].ID)
			lives := map[uuid.UUID]int{}
			for _, p := range g.Seats {
				lives[p.ID] = p.Life
			}
			advanceToUpkeepOf(t, g, 1)
			monSettle(t, g)
			if n := monTokens(g, owner.ID, "Knight"); n != 1 {
				t.Fatalf("%d Knight tokens, want 1", n)
			}
			for _, c := range g.Battlefield.Cards {
				if c.Name == "Knight" && (c.Power != 3 || c.Toughness != 1) {
					t.Errorf("Knight is %d/%d, want 3/1", c.Power, c.Toughness)
				}
			}
			for i, p := range g.Seats {
				want := 0
				if p.ID != owner.ID {
					want = tc.want
				}
				if got := lives[p.ID] - p.Life; got != want {
					t.Errorf("seat %d lost %d, want %d", i, got, want)
				}
			}
		})
	}
}

// Court of Ardenvale: hand normally, battlefield for the monarch, and
// only a permanent card with mana value 3 or less.
func TestCourtOfArdenvaleReturnsToHandOrBattlefield(t *testing.T) {
	for _, tc := range []struct {
		name    string
		monarch int
		zone    game.ZoneKind
	}{{"monarch", 1, game.ZoneBattlefield}, {"not the monarch", 2, game.ZoneHand}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			owner := g.Seats[1]
			pushPermanentForTest(g, owner.ID, "Court of Ardenvale", courtOfArdenvaleOracle, "Enchantment")
			cheap := pushGraveyardPermanent(owner, "Cheap Bear", "Creature — Bear", "{1}{G}")
			dear := pushGraveyardPermanent(owner, "Dear Dragon", "Creature — Dragon", "{4}{R}{R}")
			spell := pushGraveyardPermanent(owner, "A Bolt", "Instant", "{R}")
			monCrown(t, g, g.Seats[tc.monarch].ID)
			advanceToUpkeepOf(t, g, 1)
			monSettle(t, g)
			p := latestPickTarget(g, owner.ID)
			if p == nil {
				t.Fatal("no target prompt")
			}
			if !hasID(p.PickTargetCards, cheap) || hasID(p.PickTargetCards, dear) || hasID(p.PickTargetCards, spell) {
				t.Fatalf("offered %v; want only the cheap permanent card", p.PickTargetCards)
			}
			pickCard(t, g, owner.ID, cheap)
			monSettle(t, g)
			if z := e2Zone(g, cheap); z != tc.zone {
				t.Errorf("Cheap Bear is in %q, want %q", z, tc.zone)
			}
		})
	}
}

// Court of Garenbrig: two counters split among up to two creatures, then the
// monarch doubles every +1/+1 counter on their creatures.
func TestCourtOfGarenbrigDistributesThenDoublesForTheMonarch(t *testing.T) {
	for _, tc := range []struct {
		name         string
		monarch      int
		wantA, wantB int
	}{{"monarch", 1, 4, 2}, {"not the monarch", 0, 2, 1}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			owner := g.Seats[1]
			pushPermanentForTest(g, owner.ID, "Court of Garenbrig", courtOfGarenbrigOracle, "Enchantment")
			a := pushVanillaCreature(g, owner.ID, "Alpha", 2, 2)
			b := pushVanillaCreature(g, owner.ID, "Beta", 2, 2)
			other := pushVanillaCreature(g, g.Seats[2].ID, "Theirs", 2, 2)
			g.WithWriteLock(func() { _ = g.AddCounterForEffect(a, game.CounterPlusOne, 1) })
			monCrown(t, g, g.Seats[tc.monarch].ID)
			advanceToUpkeepOf(t, g, 1)
			monSettle(t, g)
			pickTargetRefs(t, g, owner.ID, map[uuid.UUID]int{a: 1, b: 1},
				game.TargetRef{Kind: game.TargetCard, ID: a}, game.TargetRef{Kind: game.TargetCard, ID: b})
			monSettle(t, g)
			if got := counterOn(g, a, game.CounterPlusOne); got != tc.wantA {
				t.Errorf("Alpha has %d counters, want %d", got, tc.wantA)
			}
			if got := counterOn(g, b, game.CounterPlusOne); got != tc.wantB {
				t.Errorf("Beta has %d counters, want %d", got, tc.wantB)
			}
			if got := counterOn(g, other, game.CounterPlusOne); got != 0 {
				t.Errorf("their creature has %d counters, want 0", got)
			}
		})
	}
}

// Grave Venerations: end-step return only for the monarch; the drain on a
// death is unconditional.
func TestGraveVenerationsReturnsACreatureOnlyAsMonarch(t *testing.T) {
	for _, tc := range []struct {
		name    string
		monarch int
		want    bool
	}{{"monarch", 1, true}, {"not the monarch", 0, false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			owner := g.Seats[1]
			pushPermanentForTest(g, owner.ID, "Grave Venerations", graveVenerationsOracle, "Enchantment")
			dead := pushGraveyardCreature(g, owner.ID, "Fallen Bear", "{1}{G}")
			monCrown(t, g, g.Seats[tc.monarch].ID)
			advanceToEndStepOf(t, g, 1)
			if tc.want {
				monSettle(t, g)
				pickCard(t, g, owner.ID, dead)
			}
			monSettle(t, g)
			inHand := e2Zone(g, dead) == game.ZoneHand
			if inHand != tc.want {
				t.Errorf("Fallen Bear in hand = %v, want %v", inHand, tc.want)
			}
		})
	}
}

func TestGraveVenerationsDrainsWhenYourCreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, me.ID, "Grave Venerations", graveVenerationsOracle, "Enchantment")
	mine := pushVanillaCreature(g, me.ID, "Mine", 1, 1)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 1, 1)
	myLife, theirLife := me.Life, opp.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	monSettle(t, g)
	if me.Life != myLife || opp.Life != theirLife {
		t.Fatalf("an opponent's creature dying drained: %d/%d", me.Life, opp.Life)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })
	monSettle(t, g)
	if me.Life != myLife+1 || opp.Life != theirLife-1 {
		t.Errorf("after my creature died: me %d (was %d), opp %d (was %d)", me.Life, myLife, opp.Life, theirLife)
	}
}

// Azure Fleet Admiral can't be blocked by creatures the monarch controls.
func TestAzureFleetAdmiralCantBeBlockedByTheMonarchsCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	admiral := dskB1Creature(g, me.ID, "Azure Fleet Admiral", "Creature — Human Pirate", azureFleetAdmiralOracle, 3, 3)
	kings := b12Creature(g, opp.ID, "King's Guard", "Creature — Bear", 2, 2)
	commoner := b12Creature(g, third.ID, "Commoner", "Creature — Bear", 2, 2)
	monCrown(t, g, opp.ID)
	declareAttack(t, g, opp.ID, admiral)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(kings, admiral); err == nil {
		t.Error("the monarch's creature blocked the Admiral")
	}
	_ = commoner
}

func TestAzureFleetAdmiralIsBlockableWhenNobodyIsTheMonarch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	admiral := dskB1Creature(g, me.ID, "Azure Fleet Admiral", "Creature — Human Pirate", azureFleetAdmiralOracle, 3, 3)
	guard := b12Creature(g, opp.ID, "Guard", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, admiral)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(guard, admiral); err != nil {
		t.Errorf("a creature could not block with no monarch: %v", err)
	}
}

func TestAzureFleetAdmiralEntersAndCrownsItsController(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castAndResolveCreature(t, g, "Azure Fleet Admiral", "Creature — Human Pirate", azureFleetAdmiralOracle)
	monSettle(t, g)
	if g.Monarch != me.ID {
		t.Fatalf("monarch = %v, want %v", g.Monarch, me.ID)
	}
}

const (
	courtOfVantressOracle   = "ad82b43b-b15e-48f4-86ab-5ee2c448cf14"
	mirrormadeOracle        = "79a71fe5-38ce-4bb5-aea0-c9b9a856d397"
	malleableImpostorOracle = "acd209d1-37e9-4713-b320-ee52854790bb"
	sunfrillImitatorOracle  = "2f618a4a-8b76-4c87-baa7-fb6a17f6f51c"
)

func TestCmdr2S48ClonesAreFull(t *testing.T) {
	for _, id := range []string{courtOfVantressOracle, mirrormadeOracle, malleableImpostorOracle, sunfrillImitatorOracle} {
		spec, ok := Lookup(id)
		if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v %v", id, spec, ok)
		}
	}
}

func TestMirrormadeCopiesAnArtifactOrAnEnchantmentButNotACreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	ring := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact", OracleID: "oracle-Sol Ring", Owner: opp.ID, Controller: opp.ID})
	ench := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Pacifism Field", TypeLine: "Enchantment", OracleID: "oracle-Pacifism Field", Owner: opp.ID, Controller: opp.ID})
	bear := pushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	id := castCatalogSpell(t, g, "Mirrormade", "Enchantment", mirrormadeOracle, nil)
	for i := 0; i < 32 && copyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	p := copyPrompt(g)
	if p == nil {
		t.Fatal("no copy prompt")
	}
	if !hasID(p.CopyOptions, ring) || !hasID(p.CopyOptions, ench) || hasID(p.CopyOptions, bear) {
		t.Fatalf("copy options %v: want the artifact and the enchantment, not the creature", p.CopyOptions)
	}
	resolveWithCopyChoice(t, g, ring)
	got := copyBattlefieldCard(t, g, id)
	if got.Name != "Sol Ring" || !got.IsArtifact() || !got.IsCopy() {
		t.Errorf("Mirrormade is %q artifact=%v copy=%v, want a copy of Sol Ring", got.Name, got.IsArtifact(), got.IsCopy())
	}
}

func TestMalleableImpostorCopiesOnlyAnOpponentsCreatureAndGainsFlyingAndTypes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Ogre", TypeLine: "Creature — Ogre", Power: 4, Toughness: 4, OracleID: "oracle-Their Ogre", Owner: opp.ID, Controller: opp.ID})
	id := castCatalogSpell(t, g, "Malleable Impostor", "Creature — Faerie Shapeshifter", malleableImpostorOracle, nil)
	for i := 0; i < 32 && copyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	p := copyPrompt(g)
	if p == nil {
		t.Fatal("no copy prompt")
	}
	if !hasID(p.CopyOptions, theirs) || hasID(p.CopyOptions, mine) {
		t.Fatalf("copy options %v: want only the opponent's creature", p.CopyOptions)
	}
	resolveWithCopyChoice(t, g, theirs)
	got := copyBattlefieldCard(t, g, id)
	if got.Name != "Their Ogre" {
		t.Errorf("name = %q, want Their Ogre", got.Name)
	}
	if e := got.Effective(); e.Power != 4 || e.Toughness != 4 {
		t.Errorf("copy is %d/%d, want 4/4", e.Power, e.Toughness)
	}
	if !game.HasKeyword(&got, "flying") {
		t.Error("the copy has no flying")
	}
	if !got.HasSubtype("Faerie") || !got.HasSubtype("Shapeshifter") || !got.HasSubtype("Ogre") {
		t.Errorf("types = %q, want Ogre plus Faerie Shapeshifter", got.TypeLine)
	}
}

func TestMalleableImpostorWithNoOpposingCreatureEntersAsItselfWithoutACopyPrompt(t *testing.T) {
	g := newCatalogGame(t)
	pushVanillaCreature(g, g.Seats[0].ID, "My Bear", 2, 2)
	id := castCatalogSpell(t, g, "Malleable Impostor", "Creature — Faerie Shapeshifter", malleableImpostorOracle, nil)
	passPriorityAroundTable(t, g)
	if copyPrompt(g) != nil {
		t.Fatal("offered a copy with no opposing creature")
	}
	// (The 0/0 dies to CR 704.5f only for a card with a real printing behind
	// it, which a seeded test card has not.)
	if c, ok := battlefieldCard(g, id); ok && c.IsCopy() {
		t.Error("it copied something with no opposing creature on the battlefield")
	}
}

func TestSunfrillImitatorBecomesACopyOfAnotherDinosaurAndKeepsItsName(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	imitator := dskB1Creature(g, me.ID, "Sunfrill Imitator", "Creature — Dinosaur", sunfrillImitatorOracle, 3, 3)
	rex := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Big Rex", TypeLine: "Creature — Dinosaur", Power: 7, Toughness: 7, Keywords: []string{"trample"}, OracleID: "oracle-Big Rex", Owner: me.ID, Controller: me.ID})
	bear := b12Creature(g, me.ID, "Not A Dino", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, imitator)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if !hasID(p.PickTargetCards, rex) || hasID(p.PickTargetCards, bear) || hasID(p.PickTargetCards, imitator) {
		t.Fatalf("targets %v: want only the other Dinosaur", p.PickTargetCards)
	}
	answerPickTarget(t, g, rex)
	ask := awaitTriggerPrompt(t, g, me.ID)
	if err := g.ResolveTriggerPrompt(ask.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, imitator)
	if c.Name != "Sunfrill Imitator" {
		t.Errorf("name = %q, want Sunfrill Imitator", c.Name)
	}
	if e := c.Effective(); e.Power != 7 || e.Toughness != 7 {
		t.Errorf("copy is %d/%d, want 7/7", e.Power, e.Toughness)
	}
	if !game.HasKeyword(&c, "trample") {
		t.Error("the copy lacks the Dinosaur's trample")
	}
}

func TestSunfrillImitatorDeclinedStaysAThreeThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	imitator := dskB1Creature(g, me.ID, "Sunfrill Imitator", "Creature — Dinosaur", sunfrillImitatorOracle, 3, 3)
	rex := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Big Rex", TypeLine: "Creature — Dinosaur", Power: 7, Toughness: 7, OracleID: "oracle-Big Rex", Owner: me.ID, Controller: me.ID})
	declareAttack(t, g, opp.ID, imitator)
	answerPickTarget(t, g, rex)
	ask := awaitTriggerPrompt(t, g, me.ID)
	if err := g.ResolveTriggerPrompt(ask.ID, me.ID, false); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, imitator)
	if e := c.Effective(); c.Name != "Sunfrill Imitator" || e.Power != 3 {
		t.Errorf("declined, yet it is %s %d/%d", c.Name, e.Power, e.Toughness)
	}
}

func TestSunfrillImitatorKeepsTheAbilityAfterCopying(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	imitator := dskB1Creature(g, me.ID, "Sunfrill Imitator", "Creature — Dinosaur", sunfrillImitatorOracle, 3, 3)
	rex := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Big Rex", TypeLine: "Creature — Dinosaur", Power: 7, Toughness: 7, OracleID: "oracle-Big Rex", Owner: me.ID, Controller: me.ID})
	declareAttack(t, g, opp.ID, imitator)
	answerPickTarget(t, g, rex)
	ask := awaitTriggerPrompt(t, g, me.ID)
	if err := g.ResolveTriggerPrompt(ask.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	// Next turn of mine: it attacks again and the granted trigger asks for a target.
	advanceToStepOf(t, g, 1, game.StepUpkeep)
	advanceToStepOf(t, g, 0, game.StepUpkeep)
	declareAttack(t, g, opp.ID, imitator)
	if latestPickTarget(g, me.ID) == nil && triggerOnStack(g, imitator) == nil && len(g.PendingTriggers) == 0 {
		t.Error("the copy lost the attack trigger")
	}
}

func TestCourtOfVantressMonarchMakesATokenCopyAndOtherwiseBecomesTheCopy(t *testing.T) {
	for _, tc := range []struct {
		name    string
		monarch int
	}{{"monarch", 1}, {"not the monarch", 2}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			owner := g.Seats[1]
			court := pushPermanentForTest(g, owner.ID, "Court of Vantress", courtOfVantressOracle, "Enchantment")
			ring := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact", OracleID: "oracle-Sol Ring", Owner: g.Seats[2].ID, Controller: g.Seats[2].ID})
			bear := pushVanillaCreature(g, owner.ID, "Bear", 2, 2)
			monCrown(t, g, g.Seats[tc.monarch].ID)
			advanceToUpkeepOf(t, g, 1)
			monSettle(t, g)
			p := latestPickTarget(g, owner.ID)
			if p == nil {
				t.Fatal("no target prompt")
			}
			if !hasID(p.PickTargetCards, ring) || hasID(p.PickTargetCards, court) || hasID(p.PickTargetCards, bear) {
				t.Fatalf("targets %v: want the artifact only (not the Court, not a creature)", p.PickTargetCards)
			}
			pickCard(t, g, owner.ID, ring)
			ask := awaitTriggerPrompt(t, g, owner.ID)
			if err := g.ResolveTriggerPrompt(ask.ID, owner.ID, true); err != nil {
				t.Fatal(err)
			}
			passPriorityAroundTable(t, g)
			if tc.monarch == 1 {
				if n := monTokens(g, owner.ID, "Sol Ring"); n != 1 {
					t.Errorf("%d Sol Ring token copies, want 1", n)
				}
				c, _ := battlefieldCard(g, court)
				if c.Name != "Court of Vantress" {
					t.Errorf("the monarch's Court became %q", c.Name)
				}
				return
			}
			if n := monTokens(g, owner.ID, "Sol Ring"); n != 0 {
				t.Errorf("a token was made while not the monarch")
			}
			c, _ := battlefieldCard(g, court)
			if c.Name != "Sol Ring" {
				t.Errorf("the Court is %q, want a copy of Sol Ring", c.Name)
			}
			// It keeps the ability: the next upkeep asks for a target again.
			monSettle(t, g)
			advanceToUpkeepOf(t, g, 0)
			advanceToUpkeepOf(t, g, 1)
			monSettle(t, g)
			if latestPickTarget(g, owner.ID) == nil {
				t.Error("the copy lost the Court's upkeep ability")
			}
		})
	}
}

func TestCourtOfVantressDeclinedDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	pushPermanentForTest(g, owner.ID, "Court of Vantress", courtOfVantressOracle, "Enchantment")
	ring := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact", OracleID: "oracle-Sol Ring", Owner: owner.ID, Controller: owner.ID})
	monCrown(t, g, owner.ID)
	advanceToUpkeepOf(t, g, 1)
	monSettle(t, g)
	pickCard(t, g, owner.ID, ring)
	ask := awaitTriggerPrompt(t, g, owner.ID)
	if err := g.ResolveTriggerPrompt(ask.ID, owner.ID, false); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if n := monTokens(g, owner.ID, "Sol Ring"); n != 0 {
		t.Errorf("%d tokens after declining", n)
	}
}
