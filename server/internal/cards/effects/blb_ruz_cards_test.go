package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blb_ruz_cards_test.go — Bloomburrow cards starting with R or U–Z
// (#4467). Each test plays the real catalog entry and asserts what its
// text says.

const (
	blbRUZResponseOracle    = "9f7abd7d-1087-495c-8a52-2a2368cc1032"
	blbRUZBiteOracle        = "6d5dc34b-3eea-4b77-8db7-94bc30b14c4c"
	blbRUZGnawOracle        = "077ae62d-ed94-4365-a8b7-97fd99cff028"
	blbRUZRallierOracle     = "50c038a4-6546-4a15-a117-176849edaa3d"
	blbRUZRaiderOracle      = "4c2e84dc-b4f5-4eca-af07-1c53a1caae7c"
	blbRUZVillageOracle     = "7e103748-3f76-42ce-a063-d0256b2dce2b"
	blbRUZDuoOracle         = "0cc89c15-5a48-4998-b6df-21791b6ed9a3"
	blbRUZRampagerOracle    = "23878990-dcbd-4a2a-b4f7-ee7edb396c71"
	blbRUZFlamecallerOracle = "60654004-7d9a-47ac-a70d-b246dd1728d3"
	blbRUZRotcallerOracle   = "8da0599a-8f63-4dba-857f-b4e10b7191b6"
	blbRUZMentorOracle      = "9d14c7bd-7802-401c-b02b-c7f3c51b007f"
	blbRUZSqueakOracle      = "54702a84-3125-41a7-a875-fe2ec870836d"
	blbRUZWardenOracle      = "5e11386f-4667-47f7-b280-eabd927e4a46"
	blbRUZWitnessOracle     = "426dbe89-0cbb-4a3b-8812-d9761d063373"
	blbRUZHowlOracle        = "04fd8c1f-80e5-4b3d-b63e-7af6ccafbbaa"
)

func TestBlbRUZEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		blbRUZResponseOracle:    "Rabbit Response",
		blbRUZBiteOracle:        "Rabid Bite",
		blbRUZGnawOracle:        "Rabid Gnaw",
		blbRUZRallierOracle:     "Raccoon Rallier",
		blbRUZRaiderOracle:      "Ravine Raider",
		blbRUZVillageOracle:     "Rockface Village",
		blbRUZDuoOracle:         "Roughshod Duo",
		blbRUZRampagerOracle:    "Rust-Shield Rampager",
		blbRUZFlamecallerOracle: "Valley Flamecaller",
		blbRUZRotcallerOracle:   "Valley Rotcaller",
		blbRUZMentorOracle:      "Wandertale Mentor",
		blbRUZSqueakOracle:      "War Squeak",
		blbRUZWardenOracle:      "Waterspout Warden",
		blbRUZWitnessOracle:     "Wax-Wane Witness",
		blbRUZHowlOracle:        "Wildfire Howl",
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

// blbRUZGone reports whether a permanent has left the battlefield.
func blbRUZGone(g *game.Game, id uuid.UUID) bool {
	_, ok := battlefieldCard(g, id)
	return !ok
}

// --- Rabbit Response ----------------------------------------------------------

func TestBlbRUZRabbitResponsePumpsAndScriesOnlyWithARabbit(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	seedLibrary(me, "A", "B", "C")
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	castCatalogSpell(t, g, "Rabbit Response", "Instant", blbRUZResponseOracle, nil)
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) != nil {
		t.Fatal("no Rabbit: the spell should not scry")
	}
	if p, tough := ptOf(t, g, bear); p != 4 || tough != 3 {
		t.Errorf("my bear is %d/%d, want 4/3", p, tough)
	}
	if p, tough := ptOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("the opponent's bear is %d/%d, want 2/2", p, tough)
	}

	g = newCatalogGame(t)
	me, _, _ = blbB1Seats(g)
	seedLibrary(me, "A", "B", "C")
	rabbit := blbB1Push(g, me.ID, "Rabbit", "Creature — Rabbit", "", 1, 1)
	castCatalogSpell(t, g, "Rabbit Response", "Instant", blbRUZResponseOracle, nil)
	passPriorityAroundTable(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("with a Rabbit the spell should scry 2, got %+v", c)
	}
	answerScryKeepAll(t, g, me.ID)
	if p, tough := ptOf(t, g, rabbit); p != 3 || tough != 2 {
		t.Errorf("the Rabbit is %d/%d, want 3/2", p, tough)
	}
}

// --- Rabid Bite ---------------------------------------------------------------

func TestBlbRUZRabidBiteDealsYourCreaturesPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	mine := blbB1Push(g, me.ID, "Big Bear", "Creature — Bear", "", 3, 3)
	other := blbB1Push(g, me.ID, "Other Bear", "Creature — Bear", "", 2, 2)
	theirs := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 3, 3)
	advanceToMain(t, g)
	if err := castCatalogSpellErr(t, g, "Rabid Bite", "Sorcery", blbRUZBiteOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: other, Slot: 1},
	}); err == nil {
		t.Fatal("a creature you control is not a legal second target")
	}
	castCatalogSpell(t, g, "Rabid Bite", "Sorcery", blbRUZBiteOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: theirs, Slot: 1},
	})
	passPriorityAroundTable(t, g)
	if !blbRUZGone(g, theirs) {
		t.Error("3 damage should destroy the opponent's 3/3")
	}
	if blbRUZGone(g, mine) {
		t.Error("a one-sided fight deals no damage back")
	}
}

// --- Rabid Gnaw ---------------------------------------------------------------

func TestBlbRUZRabidGnawBoostsThenBites(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	mine := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := blbB1Push(g, opp.ID, "Their Ox", "Creature — Ox", "", 3, 3)
	castCatalogSpell(t, g, "Rabid Gnaw", "Instant", blbRUZGnawOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: theirs, Slot: 1},
	})
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, mine); p != 3 || tough != 2 {
		t.Errorf("my bear is %d/%d, want 3/2", p, tough)
	}
	if !blbRUZGone(g, theirs) {
		t.Error("the boosted bear deals 3, which should destroy the 3/3")
	}
}

// --- Raccoon Rallier ----------------------------------------------------------

func TestBlbRUZRaccoonRallierGivesHasteAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	rallier := blbB1Push(g, me.ID, "Raccoon Rallier", "Creature — Raccoon Bard", blbRUZRallierOracle, 2, 2)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, rallier, 0, game.ActivateAbilityParams{Targets: cardRefs(theirs)}); err == nil {
		t.Fatal("an opponent's creature is not a legal target")
	}
	b16Activate(t, g, me.ID, rallier, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, bear, "haste") {
		t.Error("the bear should have haste")
	}
	if !b16Tapped(t, g, rallier) {
		t.Error("the Rallier should be tapped")
	}
	spec, _ := Lookup(blbRUZRallierOracle)
	if !spec.Activated[0].SorcerySpeed {
		t.Error("the ability is activated only as a sorcery")
	}
}

// --- Ravine Raider ------------------------------------------------------------

func TestBlbRUZRavineRaiderHasMenaceAndPumps(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	raider := blbB1Push(g, me.ID, "Ravine Raider", "Creature — Lizard Rogue", blbRUZRaiderOracle, 1, 1)
	if !hasKeywordOnBattlefield(t, g, raider, "menace") {
		t.Error("the Raider lacks menace")
	}
	advanceToMain(t, g)
	b06AddMana(me, "C", "B", "C", "B")
	b16Activate(t, g, me.ID, raider, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	b16Activate(t, g, me.ID, raider, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, raider); p != 3 || tough != 3 {
		t.Errorf("after two activations the Raider is %d/%d, want 3/3", p, tough)
	}
}

// --- Rockface Village ---------------------------------------------------------

func TestBlbRUZRockfaceVillagePumpsOnlyItsFourTypes(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	village := pushCatalogPermanent(g, me.ID, "Rockface Village", "Land", blbRUZVillageOracle, false)
	mouse := blbB1Push(g, me.ID, "Mouse", "Creature — Mouse Soldier", "", 1, 1)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	advanceToMain(t, g)
	b06AddMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, village, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err == nil {
		t.Fatal("a Bear is not a Lizard, Mouse, Otter, or Raccoon")
	}
	b16Activate(t, g, me.ID, village, 0, game.ActivateAbilityParams{Targets: cardRefs(mouse)})
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, mouse); p != 2 || tough != 1 {
		t.Errorf("the Mouse is %d/%d, want 2/1", p, tough)
	}
	if !hasKeywordOnBattlefield(t, g, mouse, "haste") {
		t.Error("the Mouse should have haste")
	}
	spec, _ := Lookup(blbRUZVillageOracle)
	if len(spec.ManaAbilities) != 2 || len(spec.ManaAbilities[1].Restrictions) != 2 {
		t.Errorf("mana abilities %+v, want {C} and a creature-only {R}", spec.ManaAbilities)
	}
}

// --- Roughshod Duo ------------------------------------------------------------

func TestBlbRUZRoughshodDuoPumpsATargetOnExpendFour(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	duo := blbB1Push(g, me.ID, "Roughshod Duo", "Creature — Mouse Raccoon", blbRUZDuoOracle, 3, 2)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	if !hasKeywordOnBattlefield(t, g, duo, "trample") {
		t.Error("the Duo lacks trample")
	}
	toMainForCost(t, g)
	castPaid(t, g, me, "Three", "{2}{R}", "CCR")
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("three mana spent: no expend 4 yet")
	}
	passPriorityAroundTable(t, g)
	castPaid(t, g, me, "One", "{R}", "R")
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the bear is %d/%d, want 3/3", p, tough)
	}
	if !hasKeywordOnBattlefield(t, g, bear, "trample") {
		t.Error("the bear should have trample")
	}
}

// --- Rust-Shield Rampager -----------------------------------------------------

func TestBlbRUZRustShieldRampagerEvadesSmallBlockers(t *testing.T) {
	spec, _ := Lookup(blbRUZRampagerOracle)
	if len(spec.OptionalCosts) != 1 || spec.OptionalCosts[0].Key != OffspringKey || spec.OptionalCosts[0].ManaCost != "{2}" {
		t.Fatalf("optional costs %+v, want Offspring {2}", spec.OptionalCosts)
	}

	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rampager := b12Push(g, me.ID, "Rust-Shield Rampager", "Creature — Raccoon Warrior", blbRUZRampagerOracle, 4, 4)
	b12Creature(g, opp.ID, "Wall", "Creature — Wall", 2, 4)
	brAttack(t, g, rampager)
	if n := len(brOffered(t, g, opp.ID)); n != 0 {
		t.Errorf("a power-2 creature is offered %d blocks", n)
	}

	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	rampager = b12Push(g, me.ID, "Rust-Shield Rampager", "Creature — Raccoon Warrior", blbRUZRampagerOracle, 4, 4)
	b12Creature(g, opp.ID, "Ox", "Creature — Ox", 3, 3)
	brAttack(t, g, rampager)
	if n := len(brOffered(t, g, opp.ID)); n == 0 {
		t.Error("a power-3 creature should be able to block")
	}
}

func TestBlbRUZRustShieldRampagerOffspringMakesAOneOneCopy(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	blbGLCastCreature(t, g, "Rust-Shield Rampager", "Creature — Raccoon Warrior", blbRUZRampagerOracle, 4, 4, []int{0})
	for i := 0; i < 4; i++ {
		blbGLSettleTriggerOrder(g, me.ID)
		passPriorityAroundTable(t, g)
	}
	got := blbMP2Named(g, "Rust-Shield Rampager")
	if len(got) != 2 {
		t.Fatalf("%d Rampagers, want 2 (offspring paid)", len(got))
	}
	for _, c := range got {
		if c.IsToken() && (c.Power != 1 || c.Toughness != 1) {
			t.Errorf("the offspring token is %d/%d, want 1/1", c.Power, c.Toughness)
		}
	}
}

// --- Valley Flamecaller -------------------------------------------------------

func TestBlbRUZValleyFlamecallerAddsOneForItsFourTypes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flamecaller := pushCatalogPermanent(g, me.ID, "Valley Flamecaller", "Creature — Lizard Warlock", blbRUZFlamecallerOracle, false)
	otter := b16Creature(g, me.ID, "Otter", "Creature — Otter Wizard", 1, 1, "U")
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := b16Creature(g, opp.ID, "Their Mouse", "Creature — Mouse", 1, 1, "W")

	for _, c := range []struct {
		name     string
		src, hit uuid.UUID
		victim   *game.Player
		want     int
	}{
		{"your Otter", otter, opp.ID, opp, 3},
		{"the Flamecaller, a Lizard", flamecaller, opp.ID, opp, 3},
		{"your Bear", bear, opp.ID, opp, 2},
		{"an opponent's Mouse", theirs, me.ID, me, 2},
		{"your Otter, at you", otter, me.ID, me, 3},
	} {
		before := c.victim.Life
		g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(c.src, c.hit, 2) })
		if got := before - c.victim.Life; got != c.want {
			t.Errorf("%s dealt %d, want %d", c.name, got, c.want)
		}
	}
}

// --- Valley Rotcaller ---------------------------------------------------------

func TestBlbRUZValleyRotcallerDrainsForItsOtherTypes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := blbB1Seats(g)
	rotcaller := blbB1Push(g, me.ID, "Valley Rotcaller", "Creature — Squirrel Warlock", blbRUZRotcallerOracle, 1, 3)
	blbB1Push(g, me.ID, "Bat", "Creature — Bat", "", 1, 1)
	blbB1Push(g, me.ID, "Rat", "Creature — Rat Rogue", "", 1, 1)
	blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	blbB1Push(g, opp.ID, "Their Lizard", "Creature — Lizard", "", 1, 1)
	if !hasKeywordOnBattlefield(t, g, rotcaller, "menace") {
		t.Error("the Rotcaller lacks menace")
	}
	myLife, oppLife, otherLife := lifeOf(g, me.ID), lifeOf(g, opp.ID), lifeOf(g, other.ID)
	declareAttack(t, g, opp.ID, rotcaller)
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != oppLife-2 {
		t.Errorf("the defending opponent is at %d, want %d", got, oppLife-2)
	}
	if got := lifeOf(g, other.ID); got != otherLife-2 {
		t.Errorf("each opponent loses: the other opponent is at %d, want %d", got, otherLife-2)
	}
	if got := lifeOf(g, me.ID); got != myLife+2 {
		t.Errorf("I am at %d, want %d", got, myLife+2)
	}
}

// --- Wandertale Mentor --------------------------------------------------------

func TestBlbRUZWandertaleMentorGrowsOnExpendFourAndTapsForMana(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	mentor := blbB1Push(g, me.ID, "Wandertale Mentor", "Creature — Raccoon Bard", blbRUZMentorOracle, 2, 2)
	toMainForCost(t, g)
	castPaid(t, g, me, "Three", "{2}{G}", "CCG")
	passPriorityAroundTable(t, g)
	if n := plusOnes(g, mentor); n != 0 {
		t.Fatalf("three mana spent: %d counters, want 0", n)
	}
	castPaid(t, g, me, "One", "{G}", "G")
	passPriorityAroundTable(t, g)
	if n := plusOnes(g, mentor); n != 1 {
		t.Errorf("the fourth mana was spent: %d counters, want 1", n)
	}
	castPaid(t, g, me, "Five", "{G}", "G")
	passPriorityAroundTable(t, g)
	if n := plusOnes(g, mentor); n != 1 {
		t.Errorf("expend 4 happens once a turn: %d counters, want 1", n)
	}
	if err := g.ActivateManaAbility(me.ID, mentor, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if !b16Tapped(t, g, mentor) {
		t.Error("the Mentor should be tapped")
	}
}

// --- War Squeak ---------------------------------------------------------------

func TestBlbRUZWarSqueakPumpsHastesAndStopsABlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	castCatalogSpell(t, g, "War Squeak", "Enchantment — Aura", blbRUZSqueakOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the entry trigger should ask for its target")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
		t.Error("a creature you control is not a legal target")
	}
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the enchanted bear is %d/%d, want 3/3", p, tough)
	}
	if !hasKeywordOnBattlefield(t, g, bear, "haste") {
		t.Error("the enchanted bear should have haste")
	}
	assertRestrictions(t, g, theirs, game.CantBlock)
	assertRestrictions(t, g, bear, 0)
}

// --- Waterspout Warden --------------------------------------------------------

func TestBlbRUZWaterspoutWardenFliesOnlyAfterAnotherCreatureEntered(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	warden := blbB1Push(g, me.ID, "Waterspout Warden", "Creature — Frog Soldier", blbRUZWardenOracle, 3, 2)
	declareAttack(t, g, opp.ID, warden)
	passPriorityAroundTable(t, g)
	if hasKeywordOnBattlefield(t, g, warden, "flying") {
		t.Error("no other creature entered: the Warden should not fly")
	}

	g = newCatalogGame(t)
	me, opp, _ = blbB1Seats(g)
	warden = blbB1Push(g, me.ID, "Waterspout Warden", "Creature — Frog Soldier", blbRUZWardenOracle, 3, 2)
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	declareAttack(t, g, opp.ID, warden)
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, warden, "flying") {
		t.Error("another creature entered this turn: the Warden should fly")
	}
}

// --- Wax-Wane Witness ---------------------------------------------------------

func TestBlbRUZWaxWaneWitnessGrowsOnEachLifeChangeDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	witness := blbB1Push(g, me.ID, "Wax-Wane Witness", "Creature — Bat Cleric", blbRUZWitnessOracle, 2, 4)
	theirs := blbB1Push(g, opp.ID, "Their Witness", "Creature — Bat Cleric", blbRUZWitnessOracle, 2, 4)
	for _, kw := range []string{"flying", "vigilance"} {
		if !hasKeywordOnBattlefield(t, g, witness, kw) {
			t.Errorf("the Witness lacks %s", kw)
		}
	}
	blbST2ChangeLife(t, g, me.ID, 2)
	passPriorityAroundTable(t, g)
	blbST2ChangeLife(t, g, me.ID, -1)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, witness); p != 4 || tough != 4 {
		t.Errorf("after a gain and a loss on my turn the Witness is %d/%d, want 4/4", p, tough)
	}
	blbST2ChangeLife(t, g, opp.ID, 3)
	passPriorityAroundTable(t, g)
	if p, _ := ptOf(t, g, theirs); p != 2 {
		t.Errorf("a gain outside its controller's turn pumped it to power %d, want 2", p)
	}
}

// --- Wildfire Howl ------------------------------------------------------------

func blbRUZCastHowl(t *testing.T, g *game.Game, me *game.Player, gift uuid.UUID, targets []game.TargetRef) {
	t.Helper()
	advanceToMain(t, g)
	id := pushCatalogHandCard(me, "Wildfire Howl", "Sorcery", blbRUZHowlOracle)
	params := game.CastSpellParams{Targets: targets}
	if gift != uuid.Nil {
		params.OptionalCosts = []int{giftIndex(t, blbRUZHowlOracle)}
		params.GiftOpponent = gift
	}
	if err := g.CastSpell(me.ID, id, params); err != nil {
		t.Fatalf("cast Wildfire Howl: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestBlbRUZWildfireHowlDealsTwoToEachCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	ox := blbB1Push(g, opp.ID, "Their Ox", "Creature — Ox", "", 3, 3)
	oppLife := lifeOf(g, opp.ID)
	blbRUZCastHowl(t, g, me, uuid.Nil, nil)
	if !blbRUZGone(g, bear) {
		t.Error("2 damage should destroy the 2/2")
	}
	if blbRUZGone(g, ox) {
		t.Error("2 damage should not destroy the 3/3")
	}
	if got := lifeOf(g, opp.ID); got != oppLife {
		t.Errorf("no gift: the opponent is at %d, want %d", got, oppLife)
	}
}

func TestBlbRUZWildfireHowlPromisedAddsOneToAnyTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	seedLibrary(opp, "Gift")
	ox := blbB1Push(g, opp.ID, "Their Ox", "Creature — Ox", "", 3, 3)
	big := blbB1Push(g, opp.ID, "Their Giant", "Creature — Giant", "", 4, 4)
	hand := opp.Hand.Size()
	blbRUZCastHowl(t, g, me, opp.ID, []game.TargetRef{{Kind: game.TargetCard, ID: ox, Slot: 0}})
	if got := opp.Hand.Size(); got != hand+1 {
		t.Errorf("the gift: the opponent's hand is %d, want %d", got, hand+1)
	}
	if !blbRUZGone(g, ox) {
		t.Error("the target takes 1 and 2: the 3/3 should be destroyed")
	}
	if blbRUZGone(g, big) {
		t.Error("the 4/4 takes only 2")
	}

	g = newCatalogGame(t)
	me, opp, _ = blbB1Seats(g)
	seedLibrary(opp, "Gift")
	oppLife := lifeOf(g, opp.ID)
	blbRUZCastHowl(t, g, me, opp.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0}})
	if got := lifeOf(g, opp.ID); got != oppLife-1 {
		t.Errorf("the target player is at %d, want %d", got, oppLife-1)
	}
}
