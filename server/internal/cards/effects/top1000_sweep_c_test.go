package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// top1000_sweep_c_test.go — the third slice of the missing top-1,000
// cards (#3654): the three MDFC fronts (Bala Ged Recovery, Revitalizing
// Repast, Sejiri Shelter), Dusk // Dawn, Xorn, Welcoming Vampire and
// Green Sun's Zenith.

const (
	sweepCBalaGedRecoveryOracle    = "d2075f58-b0e9-4e85-b7e6-0523a27a1d5b"
	sweepCRevitalizingRepastOracle = "8dd6d060-d023-48a6-85cb-7a5521b6257b"
	sweepCSejiriShelterOracle      = "d54e4e37-042b-44a5-918d-757308545d4d"
	sweepCDuskDawnOracle           = "7bc3f92f-68a2-4934-afc4-89f6d0e8cf98"
	sweepCXornOracle               = "11d8fab8-93af-4291-80bd-cf9436e99f4b"
	sweepCWelcomingVampireOracle   = "605c1ee0-5e8a-4e0a-a99b-42a38873f822"
	sweepCGreenSunsZenithOracle    = "0d96b60b-a060-48ee-bb83-93f1c4a10669"
)

func TestSweepCBalaGedRecoveryReturnsAnyCardFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	card := pushGraveyardCardForTest(me, "Old Friend")

	castCatalogSpell(t, g, "Bala Ged Recovery", "Sorcery", sweepCBalaGedRecoveryOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: card}})
	passPriorityAroundTable(t, g)
	if !inHandOf(me, card) {
		t.Error("the targeted card is not back in your hand")
	}
	if me.Graveyard.Contains(card) {
		t.Error("the card is still in the graveyard")
	}
}

func TestSweepCRevitalizingRepastCountersAndProtectsOneCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "Revitalizing Repast", "Instant", sweepCRevitalizingRepastOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	c := s58p6View(t, g, bear)
	if c.Counters["+1/+1"] != 1 || c.CurrentPower() != 3 {
		t.Errorf("target: counters %v power %d, want one +1/+1 counter and power 3", c.Counters, c.CurrentPower())
	}
	if !game.HasKeyword(&c, "indestructible") {
		t.Error("the target should be indestructible this turn")
	}
	o := s58p6View(t, g, other)
	if o.Counters["+1/+1"] != 0 || game.HasKeyword(&o, "indestructible") {
		t.Error("an untargeted creature was affected")
	}
}

func TestSweepCSejiriShelterGrantsProtectionFromTheChosenColour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "Sejiri Shelter", "Instant", sweepCSejiriShelterOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	answerColor(t, g, me.ID, "R")

	c := s58p6View(t, g, bear)
	if !game.HasKeyword(&c, game.ProtectionFromColor("R")) {
		t.Error("want protection from red")
	}
	if game.HasKeyword(&c, game.ProtectionFromColor("B")) {
		t.Error("protection from an unchosen colour")
	}
}

func TestSweepCDuskDestroysCreaturesWithPowerThreeOrMore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := b12Creature(g, me.ID, "Big", "Creature — Giant", 3, 3)
	small := b12Creature(g, me.ID, "Small", "Creature — Bear", 2, 4)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Giant", 5, 5)

	castCatalogSpell(t, g, "Dusk", "Sorcery", sweepCDuskDawnOracle, nil)
	passPriorityAroundTable(t, g)
	if onBattlefield(g, big) || onBattlefield(g, theirs) {
		t.Error("a creature with power 3 or greater survived Dusk")
	}
	if !onBattlefield(g, small) {
		t.Error("a power 2 creature was destroyed")
	}
}

func TestSweepCDawnReturnsSmallCreatureCardsFromYourGraveyardOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mkCard := func(owner *game.Player, name, typeLine string, power int) uuid.UUID {
		c := game.Card{
			InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Power: power, Toughness: 2,
			Owner: owner.ID, Controller: owner.ID,
		}
		owner.Graveyard.PushTop(c)
		return c.InstanceID
	}
	small := mkCard(me, "Small", "Creature — Bear", 2)
	zero := mkCard(me, "Zero", "Creature — Wall", 0)
	big := mkCard(me, "Big", "Creature — Giant", 3)
	spell := mkCard(me, "Not A Creature", "Instant", 1)
	theirs := mkCard(opp, "Theirs", "Creature — Bear", 1)

	card := game.Card{
		InstanceID: uuid.New(), OracleID: sweepCDuskDawnOracle, Layout: game.LayoutSplit,
		Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: "Dusk", TypeLine: "Sorcery", ManaCost: "{2}{W}{W}",
				OracleText: "Destroy all creatures with power 3 or greater."},
			{Name: "Dawn", TypeLine: "Sorcery", ManaCost: "{3}{W}{W}",
				OracleText: "Aftermath (Cast this spell only from your graveyard. Then exile it.)\nReturn all creature cards with power 2 or less from your graveyard to your hand."},
		},
	}
	card.SettleImported()
	me.Graveyard.PushTop(card)
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast Dawn from the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{small, zero} {
		if !inHandOf(me, id) {
			t.Errorf("a creature card with power 2 or less stayed in the graveyard")
		}
	}
	if inHandOf(me, big) || inHandOf(me, spell) {
		t.Error("Dawn returned a card it should not have")
	}
	if !opp.Graveyard.Contains(theirs) {
		t.Error("Dawn touched an opponent's graveyard")
	}
	if me.Graveyard.Contains(card.InstanceID) {
		t.Error("Dawn is still in the graveyard; aftermath exiles it")
	}
}

func TestSweepCXornAddsOneTreasureToATreasureCreation(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Xorn", "Creature — Elemental", sweepCXornOracle, false)

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TreasureToken(), 2) })
	if got := countTreasures(g, me.ID); got != 3 {
		t.Errorf("creating 2 Treasures gave %d, want 3", got)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1) })
	if got := countTreasures(g, me.ID); got != 3 {
		t.Errorf("a Goblin-only creation changed the Treasures to %d", got)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, TreasureToken(), 1) })
	if got := countTreasures(g, opp.ID); got != 1 {
		t.Errorf("an opponent's Treasure creation gave %d, want 1 (Xorn is yours only)", got)
	}
}

// castSweepCCreature casts a creature of the given power from the
// active seat's hand in a main phase and lets it resolve, so it enters
// through the real path and its enters triggers are harvested.
func castSweepCCreature(t *testing.T, g *game.Game, name string, power int) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Bear", Power: power, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	// Pass until the spell has resolved and entered, stopping with its
	// enters triggers still on the stack.
	for i := 0; i < 8 && !onBattlefield(g, id); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !onBattlefield(g, id) {
		t.Fatalf("%s never entered the battlefield", name)
	}
}

func TestSweepCWelcomingVampireDrawsOncePerTurnForSmallCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vamp := pushCatalogPermanent(g, me.ID, "Welcoming Vampire", "Creature — Vampire", sweepCWelcomingVampireOracle, false)
	advanceTo(t, g, game.StepPrecombatMain)

	// A power 3 creature does not trigger it.
	castSweepCCreature(t, g, "Big Bear", 3)
	if triggerOnStack(g, vamp) != nil {
		t.Fatal("a power 3 creature triggered Welcoming Vampire")
	}
	passPriorityAroundTable(t, g)

	// A power 2 creature does.
	before := me.Hand.Size()
	castSweepCCreature(t, g, "Small One", 2)
	if triggerOnStack(g, vamp) == nil {
		t.Fatal("a power 2 creature did not trigger Welcoming Vampire")
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("hand delta %d after the trigger, want 1", got)
	}

	// A second small creature in the same turn does not.
	castSweepCCreature(t, g, "Small Two", 1)
	if triggerOnStack(g, vamp) != nil {
		t.Error("the ability triggered a second time in one turn")
	}
}

func TestSweepCGreenSunsZenithFindsAGreenCreatureAndShufflesItselfIn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elf := pushLibraryCardForTest(me, game.Card{Name: "Elf", TypeLine: "Creature — Elf", ManaCost: "{G}", Colors: []string{"G"}})
	red := pushLibraryCardForTest(me, game.Card{Name: "Goblin", TypeLine: "Creature — Goblin", ManaCost: "{R}", Colors: []string{"R"}})
	titan := pushLibraryCardForTest(me, game.Card{Name: "Titan", TypeLine: "Creature — Giant", ManaCost: "{4}{G}{G}", Colors: []string{"G"}})

	zenith := castXSpell(t, g, "Green Sun's Zenith", "Sorcery", sweepCGreenSunsZenithOracle, "{X}{G}", 1, nil)
	passPriorityAroundTable(t, g)
	takeSearchPrompts(t, g)
	passPriorityAroundTable(t, g)

	if _, ok := battlefieldCard(g, elf); !ok {
		t.Error("the green creature with mana value 1 was not put onto the battlefield")
	}
	for _, id := range []uuid.UUID{red, titan} {
		if _, wrong := battlefieldCard(g, id); wrong {
			t.Error("a nongreen creature or one over the mana value came into play")
		}
	}
	if me.Graveyard.Contains(zenith) {
		t.Error("Green Sun's Zenith went to the graveyard instead of into the library")
	}
	found := false
	for _, c := range me.Library.Cards {
		if c.InstanceID == zenith || c.Name == "Green Sun's Zenith" {
			found = true
		}
	}
	if !found {
		t.Error("Green Sun's Zenith is not in its owner's library")
	}
}

func TestSweepCGreenSunsZenithGoesBackEvenWhenNothingIsFound(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castXSpell(t, g, "Green Sun's Zenith", "Sorcery", sweepCGreenSunsZenithOracle, "{X}{G}", 0, nil)
	passPriorityAroundTable(t, g)
	takeSearchPrompts(t, g)
	passPriorityAroundTable(t, g)
	for _, c := range me.Graveyard.Cards {
		if c.Name == "Green Sun's Zenith" {
			t.Error("the spell stayed in the graveyard after a fruitless search")
		}
	}
	found := false
	for _, c := range me.Library.Cards {
		if c.Name == "Green Sun's Zenith" {
			found = true
		}
	}
	if !found {
		t.Error("Green Sun's Zenith is not in the library")
	}
}
