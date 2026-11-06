package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// eoe_cd_cards_test.go — Edge of Eternities cards whose names start with C
// or D (#4461). Each test plays the real catalog entry and asserts what its
// text says, and what it refuses.

const (
	eoeCDChromeCompanionOracle  = "1ac6fabb-71db-442f-8c14-ba9e659f906c"
	eoeCDCloudsculptOracle      = "7c5183be-cddf-48fb-a4b5-dc5f6faf26de"
	eoeCDCosmograndZenithOracle = "e2cd65f6-7cbc-497c-bdc9-698ff17da31e"
	eoeCDDarkEnduranceOracle    = "ded4009d-be23-4e36-a7a3-9e07230712ee"
	eoeCDDawnstrikeOracle       = "69892dc7-4087-447c-a41c-99cf3d5c0c12"
	eoeCDDepressurizeOracle     = "abcb5321-c295-4abf-8622-7de79679d27a"
	eoeCDDiplomaticOracle       = "57b75be4-9229-4a48-a59e-a0f812f38637"
	eoeCDDrillTooDeepOracle     = "0413e0e4-c8f8-44c3-a749-c799766bc8cd"
	eoeCDDualSunAdeptsOracle    = "155e01a1-5809-4be9-ad38-c11553a4a723"
	eoeCDDualSunTechniqueOracle = "89ec0f03-47d0-4657-8987-95e680404f24"
	eoeCDDubiousDelicacyOracle  = "697b8045-0edb-4f7b-a933-4d41ec2b53d5"
)

func TestEoeCDEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		eoeCDChromeCompanionOracle:  "Chrome Companion",
		eoeCDCloudsculptOracle:      "Cloudsculpt Technician",
		eoeCDCosmograndZenithOracle: "Cosmogrand Zenith",
		eoeCDDarkEnduranceOracle:    "Dark Endurance",
		eoeCDDawnstrikeOracle:       "Dawnstrike Vanguard",
		eoeCDDepressurizeOracle:     "Depressurize",
		eoeCDDiplomaticOracle:       "Diplomatic Relations",
		eoeCDDrillTooDeepOracle:     "Drill Too Deep",
		eoeCDDualSunAdeptsOracle:    "Dual-Sun Adepts",
		eoeCDDualSunTechniqueOracle: "Dual-Sun Technique",
		eoeCDDubiousDelicacyOracle:  "Dubious Delicacy",
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

// --- helpers ---------------------------------------------------------------

// eoeCDCard is a target ref to a card.
func eoeCDCard(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

// eoeCDPT is a permanent's power and toughness after the layers, with its
// counters.
func eoeCDPT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.PowerForComparison(), c.CurrentToughness()
}

// eoeCDSet edits a battlefield permanent in place under the write lock.
func eoeCDSet(g *game.Game, id uuid.UUID, edit func(c *game.Card)) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				edit(&g.Battlefield.Cards[i])
			}
		}
	})
}

// eoeCDActivate activates a catalog row strictly, paying from the pool.
func eoeCDActivate(g *game.Game, who, source uuid.UUID, idx int, targets ...game.TargetRef) error {
	return g.ActivateCatalogAbility(who, source, idx, game.ActivateAbilityParams{Strict: true, Targets: targets})
}

// eoeCDCastModal casts a modal catalog spell from a main phase with its
// modes and targets announced together (CR 601.2b–c).
func eoeCDCastModal(t *testing.T, g *game.Game, name, typeLine, oracle string, modes []int, targets []game.TargetRef) error {
	t.Helper()
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: me.ID, Controller: me.ID})
	return g.CastSpell(me.ID, id, game.CastSpellParams{Modes: modes, Targets: targets})
}

// eoeCDCount counts the permanents a player controls with a name.
func eoeCDCount(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			n++
		}
	}
	return n
}

// --- Chrome Companion ------------------------------------------------------

func TestEoeCDChromeCompanionTucksAGraveyardCardAndGainsLifeWhenTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dog := b12Push(g, me.ID, "Chrome Companion", "Artifact Creature — Dog", eoeCDChromeCompanionOracle, 2, 1)
	dead := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	life := me.Life
	advanceToMain(t, g)
	fillPoolColored(me, "C", 2)
	if err := eoeCDActivate(g, me.ID, dog, 0, eoeCDCard(dead)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	if opp.Graveyard.Contains(dead) {
		t.Fatal("the card is still in the graveyard")
	}
	if cards := opp.Library.Cards; len(cards) == 0 || cards[0].InstanceID != dead {
		t.Error("the card is not on the bottom of its owner's library")
	}
	if me.Life != life+1 {
		t.Errorf("life %d, want %d: tapping for the ability taps the Companion", me.Life, life+1)
	}
}

func TestEoeCDChromeCompanionGainsLifeWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dog := b12Push(g, me.ID, "Chrome Companion", "Artifact Creature — Dog", eoeCDChromeCompanionOracle, 2, 1)
	life := me.Life
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(dog, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
}

// --- Cloudsculpt Technician ------------------------------------------------

func TestEoeCDCloudsculptTechnicianGetsPlusOneWhileYouControlAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tech := b12Push(g, me.ID, "Cloudsculpt Technician", "Creature — Jellyfish Artificer", eoeCDCloudsculptOracle, 1, 4)
	if !hasKeywordOnBattlefield(t, g, tech, "flying") {
		t.Error("no flying")
	}
	b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if p, tough := eoeCDPT(t, g, tech); p != 1 || tough != 4 {
		t.Errorf("with only an opponent's artifact: %d/%d, want 1/4", p, tough)
	}
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	if p, tough := eoeCDPT(t, g, tech); p != 2 || tough != 4 {
		t.Errorf("with my artifact: %d/%d, want 2/4", p, tough)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rock, game.DestroyOptions{}) })
	if p, _ := eoeCDPT(t, g, tech); p != 1 {
		t.Errorf("after my artifact left: power %d, want 1", p)
	}
}

// --- Cosmogrand Zenith -----------------------------------------------------

// eoeCDZenithSecondSpell casts the Zenith as the turn's first spell, then a
// second spell, and answers the trigger's mode prompt with mode.
func eoeCDZenithSecondSpell(t *testing.T, g *game.Game, mode int) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	zenith := castCatalogSpell(t, g, "Cosmogrand Zenith", "Creature — Human Soldier", eoeCDCosmograndZenithOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(zenith) {
		t.Fatal("the Zenith did not resolve")
	}
	if latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID) != nil {
		t.Fatal("the first spell of the turn triggered")
	}
	b24CastNoncreature(t, g, "Second", "Instant", "{U}")
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatal("no mode prompt for the second spell")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
	return zenith
}

func TestEoeCDCosmograndZenithMakesTwoSoldiersOnTheSecondSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	eoeCDZenithSecondSpell(t, g, 0)
	if n := eoeCDCount(g, me.ID, "Human Soldier"); n != 2 {
		t.Fatalf("%d Human Soldier tokens, want 2", n)
	}
	soldier := findBattlefieldByName(g, "Human Soldier")
	if p, tough := eoeCDPT(t, g, soldier); p != 1 || tough != 1 {
		t.Errorf("token %d/%d, want 1/1", p, tough)
	}
	b24CastNoncreature(t, g, "Third", "Instant", "{U}")
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID) != nil || eoeCDCount(g, me.ID, "Human Soldier") != 2 {
		t.Error("the third spell triggered")
	}
}

func TestEoeCDCosmograndZenithCountersEachCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	zenith := eoeCDZenithSecondSpell(t, g, 1)
	for _, id := range []uuid.UUID{zenith, bear} {
		if n := b12Counter(t, g, id, game.CounterPlusOne); n != 1 {
			t.Errorf("my creature %s has %d +1/+1 counters, want 1", id, n)
		}
	}
	if n := b12Counter(t, g, theirs, game.CounterPlusOne); n != 0 {
		t.Errorf("the opponent's creature has %d counters, want 0", n)
	}
	if eoeCDCount(g, me.ID, "Human Soldier") != 0 {
		t.Error("the counter mode also made tokens")
	}
}

// --- Dark Endurance --------------------------------------------------------

func TestEoeCDDarkEnduranceCostsOneLessOnlyForABlockingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	blocker := b12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	idle := b12Creature(g, opp.ID, "Idle", "Creature — Bear", 2, 2)
	pr7aBlockedCombat(t, g, opp.ID, attacker, blocker)
	finishBlockDeclarations(t, g)
	card := game.Card{Name: "Dark Endurance", TypeLine: "Instant", ManaCost: "{1}{B}", OracleID: eoeCDDarkEnduranceOracle}
	if err := socB19CastWithPool(g, me, card, "B", 1, game.CastSpellParams{Targets: cardRefs(idle)}); err == nil {
		t.Fatal("cast for {B} at a creature that is not blocking")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	card.InstanceID = uuid.New()
	if err := socB19CastWithPool(g, me, card, "B", 1, game.CastSpellParams{Targets: cardRefs(blocker)}); err != nil {
		t.Fatalf("cast for {B} at a blocking creature: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left over, want the one spent", me.ManaPool.Total())
	}
}

func TestEoeCDDarkEnduranceGivesPlusTwoAndIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Dark Endurance", "Instant", eoeCDDarkEnduranceOracle, []game.TargetRef{eoeCDCard(bear)})
	passPriorityAroundTable(t, g)
	if p, tough := eoeCDPT(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("%d/%d, want 4/2", p, tough)
	}
	if !hasKeywordOnBattlefield(t, g, bear, "indestructible") {
		t.Error("no indestructible")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear, game.DestroyOptions{}) })
	if !g.Battlefield.Contains(bear) {
		t.Error("an indestructible creature was destroyed")
	}
	endTurn(t, g)
	if hasKeywordOnBattlefield(t, g, bear, "indestructible") {
		t.Error("indestructible outlived the turn")
	}
}

// --- Dawnstrike Vanguard ---------------------------------------------------

func TestEoeCDDawnstrikeVanguardCountersTheOthersWithTwoTappedCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vanguard := b12Push(g, me.ID, "Dawnstrike Vanguard", "Creature — Human Knight", eoeCDDawnstrikeOracle, 4, 5)
	tapped := b12Creature(g, me.ID, "Tapped Bear", "Creature — Bear", 2, 2)
	untapped := b12Creature(g, me.ID, "Untapped Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if !hasKeywordOnBattlefield(t, g, vanguard, "lifelink") {
		t.Error("no lifelink")
	}
	for _, id := range []uuid.UUID{vanguard, tapped} {
		eoeCDSet(g, id, func(c *game.Card) { c.Tapped = true })
	}
	advanceToEndStep(t, g)
	if triggerOnStack(g, vanguard) == nil && len(g.PendingTriggers) == 0 {
		t.Fatal("no end step trigger with two tapped creatures (the Vanguard counts)")
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{tapped, untapped} {
		if n := b12Counter(t, g, id, game.CounterPlusOne); n != 1 {
			t.Errorf("my other creature %s has %d counters, want 1", id, n)
		}
	}
	if n := b12Counter(t, g, vanguard, game.CounterPlusOne); n != 0 {
		t.Errorf("the Vanguard has %d counters; it puts them on the others", n)
	}
	if n := b12Counter(t, g, theirs, game.CounterPlusOne); n != 0 {
		t.Errorf("the opponent's creature has %d counters", n)
	}
}

func TestEoeCDDawnstrikeVanguardNeedsTwoTappedCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vanguard := b12Push(g, me.ID, "Dawnstrike Vanguard", "Creature — Human Knight", eoeCDDawnstrikeOracle, 4, 5)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	eoeCDSet(g, vanguard, func(c *game.Card) { c.Tapped = true })
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if n := b12Counter(t, g, bear, game.CounterPlusOne); n != 0 {
		t.Errorf("one tapped creature put %d counters", n)
	}
}

func TestEoeCDDawnstrikeVanguardRechecksAsItResolves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vanguard := b12Push(g, me.ID, "Dawnstrike Vanguard", "Creature — Human Knight", eoeCDDawnstrikeOracle, 4, 5)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{vanguard, bear} {
		eoeCDSet(g, id, func(c *game.Card) { c.Tapped = true })
	}
	advanceToEndStep(t, g)
	eoeCDSet(g, bear, func(c *game.Card) { c.Tapped = false })
	passPriorityAroundTable(t, g)
	if n := b12Counter(t, g, bear, game.CounterPlusOne); n != 0 {
		t.Errorf("the condition failed before it resolved, yet it put %d counters", n)
	}
}

// --- Depressurize ----------------------------------------------------------

func TestEoeCDDepressurizeDestroysACreatureLeftAtZeroPower(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	three := b12Creature(g, opp.ID, "Three", "Creature — Bear", 3, 3)
	castCatalogSpell(t, g, "Depressurize", "Instant", eoeCDDepressurizeOracle, []game.TargetRef{eoeCDCard(three)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(three) {
		t.Error("a 3-power creature at 0 power survived")
	}
	if !opp.Graveyard.Contains(three) {
		t.Error("the creature is not in its owner's graveyard")
	}
}

func TestEoeCDDepressurizeOnlyShrinksABiggerCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	four := b12Creature(g, opp.ID, "Four", "Creature — Bear", 4, 4)
	castCatalogSpell(t, g, "Depressurize", "Instant", eoeCDDepressurizeOracle, []game.TargetRef{eoeCDCard(four)})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(four) {
		t.Fatal("a creature left at 1 power was destroyed")
	}
	if p, tough := eoeCDPT(t, g, four); p != 1 || tough != 4 {
		t.Errorf("%d/%d, want 1/4", p, tough)
	}
}

// --- Diplomatic Relations --------------------------------------------------

func TestEoeCDDiplomaticRelationsPumpsGrantsVigilanceAndBites(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Wolf", "Creature — Wolf", 2, 5)
	castCatalogSpell(t, g, "Diplomatic Relations", "Instant", eoeCDDiplomaticOracle, []game.TargetRef{eoeCDCard(mine), eoeCDCard(theirs)})
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, theirs); c.DamageMarked != 3 {
		t.Errorf("their creature has %d damage, want 3 (power after +1/+0)", c.DamageMarked)
	}
	if c, _ := battlefieldCard(g, mine); c.DamageMarked != 0 {
		t.Errorf("my creature took %d damage; the fight is one-sided", c.DamageMarked)
	}
	if !hasKeywordOnBattlefield(t, g, mine, "vigilance") {
		t.Error("my creature did not gain vigilance")
	}
	if p, _ := eoeCDPT(t, g, mine); p != 3 {
		t.Errorf("my creature's power %d, want 3", p)
	}
}

func TestEoeCDDiplomaticRelationsRefusesTheWrongSides(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	mine2 := b12Creature(g, me.ID, "My Other Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Wolf", "Creature — Wolf", 2, 5)
	if err := castCatalogSpellErr(t, g, "Diplomatic Relations", "Instant", eoeCDDiplomaticOracle, []game.TargetRef{eoeCDCard(theirs), eoeCDCard(mine)}); err == nil {
		t.Error("the slots were accepted reversed")
	}
	if err := castCatalogSpellErr(t, g, "Diplomatic Relations", "Instant", eoeCDDiplomaticOracle, []game.TargetRef{eoeCDCard(mine), eoeCDCard(mine2)}); err == nil {
		t.Error("a creature I control was accepted as the creature an opponent controls")
	}
}

func TestEoeCDDiplomaticRelationsKeepsThePumpWhenTheVictimLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Wolf", "Creature — Wolf", 2, 5)
	castCatalogSpell(t, g, "Diplomatic Relations", "Instant", eoeCDDiplomaticOracle, []game.TargetRef{eoeCDCard(mine), eoeCDCard(theirs)})
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(theirs) })
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, mine, "vigilance") {
		t.Error("my creature lost the vigilance when the other target left")
	}
	if p, _ := eoeCDPT(t, g, mine); p != 3 {
		t.Errorf("my creature's power %d, want 3", p)
	}
}

// --- Drill Too Deep --------------------------------------------------------

func TestEoeCDDrillTooDeepPutsFiveChargeCountersOnYourSpacecraft(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ship := b12Permanent(g, me.ID, "My Ship", "Artifact — Spacecraft")
	if err := eoeCDCastModal(t, g, "Drill Too Deep", "Instant", eoeCDDrillTooDeepOracle, []int{0}, []game.TargetRef{eoeCDCard(ship)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := b12Counter(t, g, ship, game.CounterCharge); n != 5 {
		t.Errorf("%d charge counters, want 5", n)
	}
}

func TestEoeCDDrillTooDeepCounterModeRefusesOtherTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirShip := b12Permanent(g, opp.ID, "Their Ship", "Artifact — Spacecraft")
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	if err := eoeCDCastModal(t, g, "Drill Too Deep", "Instant", eoeCDDrillTooDeepOracle, []int{0}, []game.TargetRef{eoeCDCard(theirShip)}); err == nil {
		t.Error("an opponent's Spacecraft was accepted")
	}
	if err := eoeCDCastModal(t, g, "Drill Too Deep", "Instant", eoeCDDrillTooDeepOracle, []int{0}, []game.TargetRef{eoeCDCard(rock)}); err == nil {
		t.Error("an artifact that is not a Spacecraft or Planet was accepted")
	}
}

func TestEoeCDDrillTooDeepDestroysTargetArtifact(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := eoeCDCastModal(t, g, "Drill Too Deep", "Instant", eoeCDDrillTooDeepOracle, []int{1}, []game.TargetRef{eoeCDCard(bear)}); err == nil {
		t.Error("a creature was accepted as the artifact")
	}
	if err := eoeCDCastModal(t, g, "Drill Too Deep", "Instant", eoeCDDrillTooDeepOracle, []int{1}, []game.TargetRef{eoeCDCard(rock)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
}

// --- Dual-Sun Adepts -------------------------------------------------------

func TestEoeCDDualSunAdeptsPumpsYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	adepts := b12Push(g, me.ID, "Dual-Sun Adepts", "Creature — Human Soldier", eoeCDDualSunAdeptsOracle, 2, 2)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if !hasKeywordOnBattlefield(t, g, adepts, "double strike") {
		t.Error("no double strike")
	}
	advanceToMain(t, g)
	fillPoolColored(me, "C", 4)
	if err := eoeCDActivate(g, me.ID, adepts, 0); err == nil {
		t.Fatal("activated for four mana")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	fillPoolColored(me, "C", 5)
	if err := eoeCDActivate(g, me.ID, adepts, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{adepts, bear} {
		if p, tough := eoeCDPT(t, g, id); p != 3 || tough != 3 {
			t.Errorf("my creature %s is %d/%d, want 3/3", id, p, tough)
		}
	}
	if p, tough := eoeCDPT(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("the opponent's creature is %d/%d, want 2/2", p, tough)
	}
}

// --- Dual-Sun Technique ----------------------------------------------------

func TestEoeCDDualSunTechniqueGrantsDoubleStrikeAndDrawsWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
	advanceToMain(t, g)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Dual-Sun Technique", "Instant", eoeCDDualSunTechniqueOracle, []game.TargetRef{eoeCDCard(bear)})
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, bear, "double strike") {
		t.Error("no double strike")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}
}

func TestEoeCDDualSunTechniqueDrawsNothingWithoutACounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Dual-Sun Technique", "Instant", eoeCDDualSunTechniqueOracle, []game.TargetRef{eoeCDCard(theirs)}); err == nil {
		t.Error("an opponent's creature was accepted")
	}
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Dual-Sun Technique", "Instant", eoeCDDualSunTechniqueOracle, []game.TargetRef{eoeCDCard(bear)})
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, bear, "double strike") {
		t.Error("no double strike")
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d, want no draw", hand, me.Hand.Size())
	}
}

// --- Dubious Delicacy ------------------------------------------------------

func TestEoeCDDubiousDelicacyShrinksUpToOneCreatureAsItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	three := b12Creature(g, opp.ID, "Three", "Creature — Bear", 3, 3)
	castCatalogSpell(t, g, "Dubious Delicacy", "Artifact — Food", eoeCDDubiousDelicacyOracle, nil)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, three)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(three) {
		t.Error("a 3/3 survived -3/-3")
	}
	if findBattlefieldByName(g, "Dubious Delicacy") == uuid.Nil {
		t.Error("the Delicacy is not on the battlefield")
	}
}

func TestEoeCDDubiousDelicacyHasFlash(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if g.Turn.Step == game.StepPrecombatMain || g.Turn.Step == game.StepPostcombatMain {
		t.Fatal("the harness is in a main phase; the test needs another step")
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Dubious Delicacy", TypeLine: "Artifact — Food",
		OracleID: eoeCDDubiousDelicacyOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast outside a main phase: %v", err)
	}
}

func TestEoeCDDubiousDelicacySacrificesForLifeOrForAnOpponentsLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life, theirLife := me.Life, opp.Life
	gain := b12Permanent(g, me.ID, "Dubious Delicacy", "Artifact — Food")
	eoeCDSet(g, gain, func(c *game.Card) { c.OracleID = eoeCDDubiousDelicacyOracle })
	drain := b12Permanent(g, me.ID, "Dubious Delicacy", "Artifact — Food")
	eoeCDSet(g, drain, func(c *game.Card) { c.OracleID = eoeCDDubiousDelicacyOracle })
	advanceToMain(t, g)

	fillPoolColored(me, "C", 2)
	if err := eoeCDActivate(g, me.ID, gain, 0); err != nil {
		t.Fatalf("gain: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
	if g.Battlefield.Contains(gain) {
		t.Error("the first Delicacy was not sacrificed")
	}

	fillPoolColored(me, "C", 2)
	if err := eoeCDActivate(g, me.ID, drain, 1, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err == nil {
		t.Fatal("I was accepted as the target opponent")
	}
	if err := eoeCDActivate(g, me.ID, drain, 1, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("drain: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != theirLife-3 {
		t.Errorf("opponent life %d, want %d", opp.Life, theirLife-3)
	}
	if me.Life != life+3 {
		t.Errorf("my life moved to %d on the drain", me.Life)
	}
	if g.Battlefield.Contains(drain) {
		t.Error("the second Delicacy was not sacrificed")
	}
}
