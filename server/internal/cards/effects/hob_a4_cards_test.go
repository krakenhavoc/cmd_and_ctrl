package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hob_a4_cards_test.go — The Hobbit / The Hobbit Commander, half A, fourth
// batch (#4454): phasing, flashback and amass. Each test casts or plays the
// real catalog entry and asserts what the card's text says.

const (
	hobA4GaladrielsDismissalOracle = "e46d0b57-5eec-4ab6-b9a7-a5c5cc01cc34"
	hobA4StoneBySunlightOracle     = "7907bbf0-f6d9-4e5b-b5de-73cd94979f98"
	hobA4MomentOfGloryOracle       = "3d39db09-67b7-4a4e-9667-4e1b806f51a5"
	hobA4PlunderOracle             = "526f520e-b4bb-4215-9c6b-afc298744f06"
	hobA4TidingsOfWarOracle        = "5753f4d6-08d0-4e12-85c5-875ee6441626"
	hobA4GatheringOracle           = "85e1556d-c4da-4979-a5db-4ce3842c7797"
	hobA4RageOracle                = "8c7641b6-3a6c-4441-a98c-3b5d61d91f76"
	hobA4PlateMailOracle           = "eda99a16-6a7c-4f39-8a6b-a284e6afd3fc"
	hobA4GreatGoblinOracle         = "080f504f-20c8-4975-837c-d704e879d7f5"
	hobA4RaiderOracle              = "c394b767-620c-4310-b78f-ece312259e46"
)

func TestHobA4EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		hobA4GaladrielsDismissalOracle: "Galadriel's Dismissal",
		hobA4StoneBySunlightOracle:     "Stone by Sunlight",
		hobA4MomentOfGloryOracle:       "Moment of Glory",
		hobA4PlunderOracle:             "Plunder the Trollshaws",
		hobA4TidingsOfWarOracle:        "Tidings of War",
		hobA4GatheringOracle:           "Gathering of Darkness",
		hobA4RageOracle:                "Rage into the Valley",
		hobA4PlateMailOracle:           "Goblin Plate Mail",
		hobA4GreatGoblinOracle:         "Great Goblin, Foul-Hearted",
		hobA4RaiderOracle:              "Misty Mountains Raider",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull {
			t.Errorf("%s: name %q, completeness %s", name, spec.Name, spec.Completeness)
		}
	}
}

// hobA4Armies lists the Army creatures `owner` controls, with the layers
// freshly computed so a counter's size and an amassed subtype are visible.
func hobA4Armies(t *testing.T, g *game.Game, owner uuid.UUID) []game.Card {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.HasSubtype(game.ArmySubtype) {
			out = append(out, c)
		}
	}
	return out
}

// hobA4OnlyArmy fails the test unless `owner` controls exactly one Army.
func hobA4OnlyArmy(t *testing.T, g *game.Game, owner uuid.UUID) game.Card {
	t.Helper()
	armies := hobA4Armies(t, g, owner)
	if len(armies) != 1 {
		t.Fatalf("Armies = %d, want exactly 1", len(armies))
	}
	return armies[0]
}

// hobA4Power reads a battlefield creature's power after the layers settle.
func hobA4Power(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := findBattlefieldCardByID(g, id)
	if c == nil {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.CurrentPower()
}

// hobA4Cast puts a card in the active player's hand, walks to a main phase
// and casts it with the given parameters.
func hobA4Cast(t *testing.T, g *game.Game, name, typeLine, oracle string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: active.ID, Controller: active.ID,
	})
	advanceToMain(t, g)
	return id, g.CastSpell(active.ID, id, params)
}

// hobA4Flashback casts a card out of the active player's graveyard for its
// flashback cost.
func hobA4Flashback(t *testing.T, g *game.Game, name, typeLine, oracle string, targets []game.TargetRef) {
	t.Helper()
	id := seedGraveyardCard(t, g, name, typeLine, oracle)
	if err := g.CastSpell(g.Seats[g.Turn.ActiveSeat].ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback", Targets: targets,
	}); err != nil {
		t.Fatalf("flashback %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
}

func hobA4PlayerRef(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: id}}
}

// Galadriel's Dismissal unkicked: the one target creature phases out.
func TestHobA4GaladrielsDismissalPhasesOutOneCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	target := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, opp.ID, "Their Wolf", "Creature — Wolf", 2, 2)

	castCatalogSpell(t, g, "Galadriel's Dismissal", "Instant", hobA4GaladrielsDismissalOracle, s58Refs(target))
	passPriorityAroundTable(t, g)

	if !g.IsPhasedOutForEffect(target) {
		t.Error("the target creature should be phased out")
	}
	if g.IsPhasedOutForEffect(other) {
		t.Error("an untargeted creature must stay in play")
	}
}

// Unkicked, the spell names a creature and cannot name a player.
func TestHobA4GaladrielsDismissalUnkickedRefusesAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	if err := castCatalogSpellErr(t, g, "Galadriel's Dismissal", "Instant", hobA4GaladrielsDismissalOracle, hobA4PlayerRef(g.Seats[1].ID)); err == nil {
		t.Error("an unkicked Dismissal has no player target")
	}
}

// Kicked: each creature the target player controls phases out instead, and
// the caster's own creatures stay.
func TestHobA4GaladrielsDismissalKickedPhasesOutEachCreatureOfAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Beast", "Creature — Beast", 3, 3)
	a := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Their Wolf", "Creature — Wolf", 2, 2)
	land := b12Permanent(g, opp.ID, "Their Relic", "Artifact")

	if _, err := hobA4Cast(t, g, "Galadriel's Dismissal", "Instant", hobA4GaladrielsDismissalOracle, game.CastSpellParams{
		Targets: hobA4PlayerRef(opp.ID), OptionalCosts: []int{0},
	}); err != nil {
		t.Fatalf("kicked cast: %v", err)
	}
	passPriorityAroundTable(t, g)

	if !g.IsPhasedOutForEffect(a) || !g.IsPhasedOutForEffect(b) {
		t.Error("every creature the target player controls should phase out")
	}
	if g.IsPhasedOutForEffect(mine) {
		t.Error("the caster's creature must stay in play")
	}
	if g.IsPhasedOutForEffect(land) {
		t.Error("only creatures phase out, not the player's artifact")
	}
}

// Stone by Sunlight, first bullet: destroys a creature with power 4 or more
// and refuses a smaller one.
func TestHobA4StoneBySunlightDestroysABigCreatureOnly(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	big := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)
	small := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)

	if _, err := hobA4Cast(t, g, "Stone by Sunlight", "Instant", hobA4StoneBySunlightOracle, game.CastSpellParams{
		Modes: []int{0}, Targets: s58Refs(small),
	}); err == nil {
		t.Fatal("a power-3 creature is not a legal target of the destroy bullet")
	}
	if _, err := hobA4Cast(t, g, "Stone by Sunlight", "Instant", hobA4StoneBySunlightOracle, game.CastSpellParams{
		Modes: []int{0}, Targets: s58Refs(big),
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(big) {
		t.Error("the power-4 creature should be destroyed")
	}
	if !g.Battlefield.Contains(small) {
		t.Error("the power-3 creature must be untouched")
	}
}

// Stone by Sunlight, second bullet: the creature is an artifact and cannot
// be destroyed this turn, and the grant ends with the turn.
func TestHobA4StoneBySunlightMakesAnIndestructibleArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b12Creature(g, me.ID, "My Beast", "Creature — Beast", 2, 2)

	if _, err := hobA4Cast(t, g, "Stone by Sunlight", "Instant", hobA4StoneBySunlightOracle, game.CastSpellParams{
		Modes: []int{1}, Targets: s58Refs(mine),
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)

	if !hasKeywordOnBattlefield(t, g, mine, "indestructible") {
		t.Error("the creature should be indestructible")
	}
	c := findBattlefieldCardByID(g, mine)
	if c == nil || !c.IsArtifact() || !c.IsCreature() {
		t.Fatalf("the creature should be an artifact creature, got %+v", c)
	}
	if err := g.DestroyPermanentForEffect(mine); err != nil {
		t.Fatalf("destroy: %v", err)
	}
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(mine) {
		t.Error("an indestructible creature must survive a destroy effect")
	}

	advanceToNextSeatsTurn(t, g)
	if hasKeywordOnBattlefield(t, g, mine, "indestructible") {
		t.Error("the indestructible grant should end at end of turn")
	}
	if c := findBattlefieldCardByID(g, mine); c == nil || c.IsArtifact() {
		t.Error("the artifact type should end at end of turn")
	}
}

// Moment of Glory from the hand: one counter on the target and nothing else.
func TestHobA4MomentOfGloryFromHandPutsOneCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	target := b12Creature(g, me.ID, "My Beast", "Creature — Beast", 2, 2)
	other := b12Creature(g, me.ID, "My Wolf", "Creature — Wolf", 2, 2)

	castCatalogSpell(t, g, "Moment of Glory", "Sorcery", hobA4MomentOfGloryOracle, s58Refs(target))
	passPriorityAroundTable(t, g)

	if got := hobA4Power(t, g, target); got != 3 {
		t.Errorf("target power = %d, want 3", got)
	}
	if got := hobA4Power(t, g, other); got != 2 {
		t.Errorf("other creature power = %d, want 2 (no counter from the hand)", got)
	}
}

// Moment of Glory from a graveyard: a counter on the target and on each
// other creature you control, but not on an opponent's.
func TestHobA4MomentOfGloryFlashbackPutsACounterOnEachOtherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := b12Creature(g, me.ID, "My Beast", "Creature — Beast", 2, 2)
	other := b12Creature(g, me.ID, "My Wolf", "Creature — Wolf", 2, 2)
	third := b12Creature(g, me.ID, "My Cat", "Creature — Cat", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	hobA4Flashback(t, g, "Moment of Glory", "Sorcery", hobA4MomentOfGloryOracle, s58Refs(target))

	if got := hobA4Power(t, g, target); got != 3 {
		t.Errorf("target power = %d, want 3 (one counter, not two)", got)
	}
	if got := hobA4Power(t, g, other); got != 3 {
		t.Errorf("other creature power = %d, want 3", got)
	}
	if got := hobA4Power(t, g, third); got != 2 {
		t.Errorf("third creature power = %d, want 2", got)
	}
	if got := hobA4Power(t, g, theirs); got != 2 {
		t.Errorf("opponent's creature power = %d, want 2", got)
	}
	if hobA4GraveyardHas(g, "Moment of Glory") {
		t.Error("the flashed-back card should be exiled, not in the graveyard")
	}
}

func hobA4GraveyardHas(g *game.Game, name string) bool {
	for _, c := range g.Seats[0].Graveyard.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}

// Plunder the Trollshaws draws one card from the hand and two from a graveyard.
func TestHobA4PlunderDrawsOneThenTwoOnFlashback(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Plunder the Trollshaws", "Instant", hobA4PlunderOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand after the normal cast = %d, want %d (one card drawn)", got, before+1)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	before2 := me2.Hand.Size()
	hobA4Flashback(t, g2, "Plunder the Trollshaws", "Instant", hobA4PlunderOracle, nil)
	if got := me2.Hand.Size(); got != before2+2 {
		t.Errorf("hand after the flashback = %d, want %d (two cards drawn)", got, before2+2)
	}
	if hobA4GraveyardHas(g2, "Plunder the Trollshaws") {
		t.Error("the flashed-back card should be exiled")
	}
}

// Tidings of War amasses 1 from the hand and 3 from a graveyard, onto the
// same Army, which becomes a Goblin.
func TestHobA4TidingsOfWarAmassesOneThenThreeOnFlashback(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Tidings of War", "Sorcery", hobA4TidingsOfWarOracle, nil)
	passPriorityAroundTable(t, g)
	army := hobA4OnlyArmy(t, g, me.ID)
	if army.CurrentPower() != 1 || !army.HasSubtype("Goblin") {
		t.Fatalf("after the normal cast: power %d, Goblin %v; want a 1/1 Goblin Army", army.CurrentPower(), army.HasSubtype("Goblin"))
	}

	hobA4Flashback(t, g, "Tidings of War", "Sorcery", hobA4TidingsOfWarOracle, nil)
	army = hobA4OnlyArmy(t, g, me.ID)
	if army.CurrentPower() != 4 {
		t.Errorf("Army power after the flashback = %d, want 4 (1 + 3)", army.CurrentPower())
	}
}

// Gathering of Darkness returns the chosen creature card and amasses 3.
func TestHobA4GatheringOfDarknessReturnsACreatureAndAmasses(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dead := pushGraveyardPermanent(me, "Fallen Bear", "Creature — Bear", "{1}{G}")

	castCatalogSpell(t, g, "Gathering of Darkness", "Sorcery", hobA4GatheringOracle, rfCardTarget(dead))
	passPriorityAroundTable(t, g)

	if !me.Hand.Contains(dead) {
		t.Error("the creature card should return to hand")
	}
	army := hobA4OnlyArmy(t, g, me.ID)
	if army.CurrentPower() != 3 || !army.HasSubtype("Goblin") {
		t.Errorf("Army = %d power, Goblin %v; want a 3/3 Goblin Army", army.CurrentPower(), army.HasSubtype("Goblin"))
	}
}

// With no target the spell still amasses; a noncreature card is no target.
func TestHobA4GatheringOfDarknessWithoutATargetStillAmasses(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spell := pushGraveyardCardTyped(me, "Old Bolt", "Instant")
	if err := castCatalogSpellErr(t, g, "Gathering of Darkness", "Sorcery", hobA4GatheringOracle, rfCardTarget(spell)); err == nil {
		t.Error("an instant card is not a creature card")
	}

	g2 := newCatalogGame(t)
	castCatalogSpell(t, g2, "Gathering of Darkness", "Sorcery", hobA4GatheringOracle, nil)
	passPriorityAroundTable(t, g2)
	if army := hobA4OnlyArmy(t, g2, g2.Seats[0].ID); army.CurrentPower() != 3 {
		t.Errorf("Army power = %d, want 3", army.CurrentPower())
	}
}

// Rage into the Valley draws a card, loses 1 life and amasses 2.
func TestHobA4RageIntoTheValleyDrawsLosesLifeAndAmasses(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand, life := me.Hand.Size(), me.Life

	castCatalogSpell(t, g, "Rage into the Valley", "Sorcery", hobA4RageOracle, nil)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand = %d, want %d", got, hand+1)
	}
	if me.Life != life-1 {
		t.Errorf("life = %d, want %d", me.Life, life-1)
	}
	if army := hobA4OnlyArmy(t, g, me.ID); army.CurrentPower() != 2 || !army.HasSubtype("Goblin") {
		t.Errorf("Army = %d power, Goblin %v; want a 2/2 Goblin Army", army.CurrentPower(), army.HasSubtype("Goblin"))
	}
}

// Goblin Plate Mail: entering amasses 1 and attaches itself to the Army, which
// then has +1/+0 and menace.
func TestHobA4GoblinPlateMailAmassesAndAttachesToTheArmy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]

	mail, err := hobA4Cast(t, g, "Goblin Plate Mail", "Artifact — Equipment", hobA4PlateMailOracle, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)

	army := hobA4OnlyArmy(t, g, me.ID)
	if got := hobA4Power(t, g, army.InstanceID); got != 2 {
		t.Errorf("Army power = %d, want 2 (one counter plus the +1/+0)", got)
	}
	if !hasKeywordOnBattlefield(t, g, army.InstanceID, "menace") {
		t.Error("the equipped Army should have menace")
	}
	eq := findBattlefieldCardByID(g, mail)
	if eq == nil || !eq.IsAttachedTo(army.InstanceID) {
		t.Fatal("the Equipment should be attached to the amassed Army")
	}
}

// With an Army already out, the Plate Mail amasses onto it and attaches there.
func TestHobA4GoblinPlateMailAttachesToAnExistingArmy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	old := b12Creature(g, me.ID, "Zombie Army", "Creature — Zombie Army", 2, 2)
	bystander := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	mail, err := hobA4Cast(t, g, "Goblin Plate Mail", "Artifact — Equipment", hobA4PlateMailOracle, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := hobA4Power(t, g, old); got != 4 {
		t.Errorf("existing Army power = %d, want 4 (2 + counter + 1)", got)
	}
	if eq := findBattlefieldCardByID(g, mail); eq == nil || !eq.IsAttachedTo(old) {
		t.Error("the Equipment should be attached to the existing Army")
	}
	if got := hobA4Power(t, g, bystander); got != 2 {
		t.Errorf("bystander power = %d, want 2", got)
	}
	if len(hobA4Armies(t, g, me.ID)) != 1 {
		t.Error("no new Army should be made when one is already out")
	}
}

// The equip ability ({4}) moves the Plate Mail's bonus to another creature.
func TestHobA4GoblinPlateMailEquipMovesTheBonus(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mail, err := hobA4Cast(t, g, "Goblin Plate Mail", "Artifact — Equipment", hobA4PlateMailOracle, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	equipTo(t, g, me.ID, mail, bear)

	if got := hobA4Power(t, g, bear); got != 3 {
		t.Errorf("equipped Bear power = %d, want 3", got)
	}
	if !hasKeywordOnBattlefield(t, g, bear, "menace") {
		t.Error("the equipped Bear should have menace")
	}
	army := hobA4OnlyArmy(t, g, me.ID)
	if got := hobA4Power(t, g, army.InstanceID); got != 1 {
		t.Errorf("Army power = %d, want 1 once the Equipment has moved", got)
	}
}

// Great Goblin: amass 3 on entering, amass 3 more on attacking, and every Army
// its controller has tramples while it is out.
func TestHobA4GreatGoblinAmassesOnEnteringAndAttackingAndGivesTheArmyTrample(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirArmy := b12Creature(g, opp.ID, "Zombie Army", "Creature — Zombie Army", 2, 2)

	goblin := castHoldCreature(t, g, "Great Goblin, Foul-Hearted", "Legendary Creature — Goblin Noble", hobA4GreatGoblinOracle, 3, 3)

	army := hobA4OnlyArmy(t, g, me.ID)
	if army.CurrentPower() != 3 {
		t.Fatalf("Army power after entering = %d, want 3", army.CurrentPower())
	}
	if !hasKeywordOnBattlefield(t, g, army.InstanceID, "trample") {
		t.Error("the Army should have trample")
	}
	if hasKeywordOnBattlefield(t, g, theirArmy, "trample") {
		t.Error("an opponent's Army must not gain trample")
	}

	// Summoning sickness is a rule about the turn it entered; clear it so the
	// test can attack in the same turn.
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == goblin {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	}
	declareAttack(t, g, opp.ID, goblin)
	passPriorityAroundTable(t, g)

	army = hobA4OnlyArmy(t, g, me.ID)
	if army.CurrentPower() != 6 {
		t.Errorf("Army power after attacking = %d, want 6", army.CurrentPower())
	}
}

func TestHobA4GreatGoblinArmyLosesTrampleWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	goblin := castHoldCreature(t, g, "Great Goblin, Foul-Hearted", "Legendary Creature — Goblin Noble", hobA4GreatGoblinOracle, 3, 3)
	army := hobA4OnlyArmy(t, g, me.ID)

	if err := g.DestroyPermanentForEffect(goblin); err != nil {
		t.Fatalf("destroy: %v", err)
	}
	g.RunStateChecksForTest()
	if hasKeywordOnBattlefield(t, g, army.InstanceID, "trample") {
		t.Error("the Army should lose trample with Great Goblin gone")
	}
}

// Misty Mountains Raider: one amass 2 for the whole attack, however many
// creatures attack.
func TestHobA4MistyMountainsRaiderAmassesOncePerAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raider := pushCatalogPermanent(g, me.ID, "Misty Mountains Raider", "Creature — Goblin Soldier", hobA4RaiderOracle, false)
	friend := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	declareAttack(t, g, opp.ID, raider, friend)
	passPriorityAroundTable(t, g)

	army := hobA4OnlyArmy(t, g, me.ID)
	if army.CurrentPower() != 2 || !army.HasSubtype("Goblin") {
		t.Errorf("Army = %d power, Goblin %v; want a 2/2 Goblin Army from one trigger", army.CurrentPower(), army.HasSubtype("Goblin"))
	}
}

// An opponent's attack does not trigger it.
func TestHobA4MistyMountainsRaiderIgnoresAnOpponentsAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Misty Mountains Raider", "Creature — Goblin Soldier", hobA4RaiderOracle, false)
	attacker := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	advanceToNextSeatsTurn(t, g)
	declareAttack(t, g, me.ID, attacker)
	passPriorityAroundTable(t, g)

	if n := len(hobA4Armies(t, g, me.ID)); n != 0 {
		t.Errorf("Armies = %d, want none from an opponent's attack", n)
	}
}
