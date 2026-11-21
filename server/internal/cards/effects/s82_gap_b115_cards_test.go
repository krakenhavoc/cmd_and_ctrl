package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b115_cards_test.go — S82 GAP batch 115 (Tales of Middle-earth
// Commander, #5948). Every test plays the real catalog card through the engine.

const (
	gapB115Crown     = "f95021cf-14c4-4792-8746-432bf9e7c1a9"
	gapB115Rope      = "76dfa5c2-1ea6-4f89-9903-26998636845e"
	gapB115Blade     = "a2a39123-8b70-4678-8675-ba79f37a5db8"
	gapB115Pan       = "0c9e3e6a-0bab-4619-b135-d2dbe824b1f0"
	gapB115Rammas    = "53d58401-c318-47e0-aab0-60001c071dbb"
	gapB115FellBeast = "3b64e0a8-7d11-41c5-b439-edec19928b7d"
	gapB115Gollum    = "95d59a06-5144-4da3-8a9c-846b087975b9"
	gapB115Pippin    = "421ec8a8-2b20-4296-b8f7-010acc3f8097"
	gapB115Bilbo     = "5789520c-e4b2-44b2-b756-3f6b468dc55b"
)

func TestGapB115AllAreFull(t *testing.T) {
	for o, name := range map[string]string{
		gapB115Crown: "Crown of Gondor", gapB115Rope: "Hithlain Rope", gapB115Blade: "Lothlórien Blade",
		gapB115Pan: "Field-Tested Frying Pan", gapB115Rammas: "Rammas Echor, Ancient Shield",
		gapB115FellBeast: "Fell Beast of Mordor", gapB115Gollum: "Gollum, Obsessed Stalker",
		gapB115Pippin: "Pippin, Warden of Isengard", gapB115Bilbo: "Bilbo, Birthday Celebrant",
	} {
		spec, ok := Lookup(o)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v, name %q, completeness %v, caveats %v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Crown of Gondor ---------------------------------------------------------

func TestGapB115CrownOfGondorPumpsByCreaturesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	crown := b12Push(g, me.ID, "Crown of Gondor", "Legendary Artifact — Equipment", gapB115Crown, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Other", "Creature — Bear", 1, 1)
	b12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 1, 1)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}{C}{C}")
	b16Activate(t, g, me.ID, crown, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(bear)})
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("equipped Bear is %d/%d, want 4/4 (+1/+1 for each of my two creatures)", p, tough)
	}
}

func TestGapB115CrownOfGondorEquipCostsThreeLessForTheMonarch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	crown := b12Push(g, me.ID, "Crown of Gondor", "Legendary Artifact — Equipment", gapB115Crown, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, crown, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(bear)}); err == nil {
		t.Fatal("equipped for {1} while not the monarch")
	}
	g.WithWriteLock(func() { _ = g.SetMonarchForEffect(me.ID) })
	b16Activate(t, g, me.ID, crown, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(bear)})
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, crown); c.AttachedTo.ID != bear {
		t.Error("the Crown did not equip for {1} as the monarch")
	}
}

func TestGapB115CrownOfGondorCrownsYouOnlyWhenThereIsNoMonarch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	b12Push(g, me.ID, "Crown of Gondor", "Legendary Artifact — Equipment", gapB115Crown, 0, 0)
	advanceToMain(t, g)
	b12PlayFromHand(t, g, "Plain Bear", "Creature — Bear", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if g.Monarch != uuid.Nil {
		t.Fatal("a nonlegendary creature made me the monarch")
	}
	b12PlayFromHand(t, g, "Legend", "Legendary Creature — Bear", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if g.Monarch != me.ID {
		t.Fatalf("monarch is %v, want me", g.Monarch)
	}
	g.WithWriteLock(func() { _ = g.SetMonarchForEffect(opp.ID) })
	b12PlayFromHand(t, g, "Legend Two", "Legendary Creature — Bear", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if g.Monarch != opp.ID {
		t.Error("took the crown from another monarch")
	}
}

// --- Hithlain Rope -----------------------------------------------------------

func TestGapB115HithlainRopeDrawsThenHandsItToTheRight(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rope := pushCatalogPermanent(g, me.ID, "Hithlain Rope", "Artifact", gapB115Rope, false)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}")
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, rope, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
	var right uuid.UUID
	g.ReadSnapshot(func() { right = g.NearestOpponentForEffect(me.ID, game.AttackDirectionRight) })
	c, ok := battlefieldCard(g, rope)
	if !ok || right == uuid.Nil || c.Controller != right {
		t.Errorf("the Rope is controlled by %v, want the player to my right %v", c.Controller, right)
	}
}

func TestGapB115HithlainRopeFetchesABasicTappedThenHandsItToTheRight(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rope := pushCatalogPermanent(g, me.ID, "Hithlain Rope", "Artifact", gapB115Rope, false)
	forest := pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	pushLibraryCardForTest(me, game.Card{Name: "Nonbasic", TypeLine: "Land"})
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}")
	b16Activate(t, g, me.ID, rope, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, forest)
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, forest); !ok || !c.Tapped {
		t.Errorf("the basic land: on battlefield %v, tapped %v, want tapped", ok, c.Tapped)
	}
	if c, _ := battlefieldCard(g, rope); c.Controller == me.ID {
		t.Error("I still control the Rope after the search")
	}
}

func TestGapB115HithlainRopeCantBeSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rope := pushCatalogPermanent(g, me.ID, "Hithlain Rope", "Artifact", gapB115Rope, false)
	var cant bool
	g.ReadSnapshot(func() { cant = g.CantBeSacrificedForEffect(rope) })
	if !cant {
		t.Error("the Rope can be sacrificed")
	}
}

// --- Lothlórien Blade --------------------------------------------------------

func TestGapB115LothlorienBladeDealsItsBearersPowerToACreatureOfTheDefender(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	third := g.Seats[2]
	blade := b12Push(g, me.ID, "Lothlórien Blade", "Artifact — Equipment", gapB115Blade, 0, 0)
	elf := b12Creature(g, me.ID, "Elf", "Creature — Elf Warrior", 3, 3)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 3)
	bystander := b12Creature(g, third.ID, "Bystander", "Creature — Bear", 2, 3)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}")
	b16Activate(t, g, me.ID, blade, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(elf)})
	passPriorityAroundTable(t, g)
	declareAttack(t, g, opp.ID, elf)
	pickTriggerTarget(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the defending player's 3-toughness creature survived 3 damage")
	}
	if !g.Battlefield.Contains(bystander) {
		t.Error("a creature the defending player does not control was hit")
	}
}

func TestGapB115LothlorienBladeEquipsAnElfForTwoAndAnyCreatureForFive(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	blade := b12Push(g, me.ID, "Lothlórien Blade", "Artifact — Equipment", gapB115Blade, 0, 0)
	elf := b12Creature(g, me.ID, "Elf", "Creature — Elf Warrior", 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, blade, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(bear)}); err == nil {
		t.Error("Equip Elf {2} equipped a Bear")
	}
	b16Activate(t, g, me.ID, blade, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(elf)})
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, blade); c.AttachedTo.ID != elf {
		t.Error("the Blade did not equip the Elf for {2}")
	}
	floatMana(t, g, me, "{C}{C}{C}{C}{C}")
	b16Activate(t, g, me.ID, blade, 1, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(bear)})
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, blade); c.AttachedTo.ID != bear {
		t.Error("the Blade did not equip the Bear for {5}")
	}
}

// --- Field-Tested Frying Pan -------------------------------------------------

func TestGapB115FryingPanMakesAFoodAndAnArmedHalfling(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pan := b12PlayFromHand(t, g, "Field-Tested Frying Pan", "Artifact — Equipment", gapB115Pan, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	food := findBattlefieldByName(g, "Food")
	halfling := findBattlefieldByName(g, "Halfling")
	if food == uuid.Nil || halfling == uuid.Nil {
		t.Fatalf("Food %v, Halfling %v: want both tokens", food, halfling)
	}
	if c, _ := battlefieldCard(g, pan); c.AttachedTo.ID != halfling {
		t.Error("the Frying Pan is not attached to the Halfling")
	}
	if c, _ := battlefieldCard(g, halfling); c.Controller != me.ID || !c.IsCreature() {
		t.Errorf("the Halfling is %+v", c)
	}
}

func TestGapB115FryingPanEquippedCreatureGrowsByTheLifeGained(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pan := b12Push(g, me.ID, "Field-Tested Frying Pan", "Artifact — Equipment", gapB115Pan, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}")
	b16Activate(t, g, me.ID, pan, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(bear)})
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("equipped Bear is %d/%d, want 5/5", p, tough)
	}
	if p, tough := ptOf(t, g, other); p != 2 || tough != 2 {
		t.Errorf("the other Bear is %d/%d, want unchanged 2/2", p, tough)
	}
}

// --- Rammas Echor, Ancient Shield --------------------------------------------

func TestGapB115RammasEchorDrawsAndMakesAWallOnTheSecondSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Rammas Echor, Ancient Shield", "Legendary Artifact", gapB115Rammas, 0, 0)
	castCatalogSpell(t, g, "First", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Wall") != uuid.Nil {
		t.Fatal("the first spell made a Wall")
	}
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Second", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	wall := findBattlefieldByName(g, "Wall")
	if wall == uuid.Nil {
		t.Fatal("the second spell made no Wall")
	}
	if c, _ := battlefieldCard(g, wall); c.Power != 0 || c.Toughness != 3 || !game.HasKeyword(&c, "defender") {
		t.Errorf("Wall is %d/%d defender=%v, want 0/3 with defender", c.Power, c.Toughness, game.HasKeyword(&c, "defender"))
	}
	// castCatalogSpell puts the spell in hand and casts it (net 0); the trigger drew one card.
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand changed by %d, want 1 (the drawn card)", got)
	}
	castCatalogSpell(t, g, "Third", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if n := countNamed(g, "Wall"); n != 1 {
		t.Errorf("%d Walls after a third spell, want 1", n)
	}
}

func TestGapB115RammasEchorGivesDefendersExaltedAtBeginningOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	b12Push(g, me.ID, "Rammas Echor, Ancient Shield", "Legendary Artifact", gapB115Rammas, 0, 0)
	wall := b12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 3)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, wall).Keywords = []string{"defender"} })
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c, _ := battlefieldCard(g, wall)
	if !game.HasKeyword(&c, "exalted") {
		t.Fatal("the defender did not gain exalted")
	}
	if a, _ := battlefieldCard(g, attacker); game.HasKeyword(&a, "exalted") {
		t.Error("a creature without defender gained exalted")
	}
	life := opp.Life
	attackWith(t, g, opp.ID, attacker)
	if got := life - opp.Life; got != 3 {
		t.Errorf("the lone attacker dealt %d, want 3 (2 and +1/+1 from the Wall's exalted)", got)
	}
}

// --- Fell Beast of Mordor ----------------------------------------------------

func TestGapB115FellBeastDevoursAndDrainsByItsCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	a := pushDevourFodder(g, me.ID, "Goblin A")
	b := pushDevourFodder(g, me.ID, "Goblin B")
	id := castDevourer(t, g, "Fell Beast of Mordor", gapB115FellBeast)
	answerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, a, b)
	pickTriggerTargetPlayer(t, g, me.ID, opp.ID)
	life, mine := opp.Life, me.Life
	passPriorityAroundTable(t, g)
	if got := devourPlusOnes(g, id); got != 2 {
		t.Fatalf("%d +1/+1 counters, want 2", got)
	}
	if got := life - opp.Life; got != 2 {
		t.Errorf("the opponent lost %d, want 2", got)
	}
	if got := me.Life - mine; got != 2 {
		t.Errorf("I gained %d, want 2", got)
	}
}

func TestGapB115FellBeastWithNoCountersDrainsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	pushDevourFodder(g, me.ID, "Goblin A")
	castDevourer(t, g, "Fell Beast of Mordor", gapB115FellBeast)
	answerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID)
	life := opp.Life
	passPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("the opponent lost %d with no counters on the Beast", life-opp.Life)
	}
}

// --- Gollum, Obsessed Stalker ------------------------------------------------

func TestGapB115GollumDrainsOnlyThoseHeHasHitByTheLifeYouGained(t *testing.T) {
	g := newCatalogGame(t)
	me, hit, safe := g.Seats[g.Turn.ActiveSeat], g.Seats[1], g.Seats[2]
	gollum := pushCatalogPermanent(g, me.ID, "Gollum, Obsessed Stalker", "Legendary Creature — Halfling Horror", gapB115Gollum, false)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, gollum).Keywords = []string{"skulk"} })
	advanceToMain(t, g)
	attackWith(t, g, hit.ID, gollum)
	if hit.Life != 39 {
		t.Fatalf("Gollum dealt %d combat damage, want 1", 40-hit.Life)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 5) })
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if hit.Life != 34 {
		t.Errorf("the hit opponent is at %d, want 34 (39 - 5 gained)", hit.Life)
	}
	if safe.Life != 40 {
		t.Errorf("an opponent Gollum never hit lost %d", 40-safe.Life)
	}
}

func TestGapB115GollumWithNoLifeGainedDrainsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, hit := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	gollum := pushCatalogPermanent(g, me.ID, "Gollum, Obsessed Stalker", "Legendary Creature — Halfling Horror", gapB115Gollum, false)
	advanceToMain(t, g)
	attackWith(t, g, hit.ID, gollum)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if hit.Life != 39 {
		t.Errorf("life %d, want 39", hit.Life)
	}
}

// --- Pippin, Warden of Isengard ----------------------------------------------

func TestGapB115PippinMakesAFoodForOneAndTap(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pippin := pushCatalogPermanent(g, me.ID, "Pippin, Warden of Isengard", "Legendary Creature — Halfling Advisor", gapB115Pippin, false)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}")
	b16Activate(t, g, me.ID, pippin, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Food") == uuid.Nil {
		t.Error("no Food token")
	}
}

func TestGapB115PippinSacrificesFourFoodsToPumpTheOthers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pippin := pushCatalogPermanent(g, me.ID, "Pippin, Warden of Isengard", "Legendary Creature — Halfling Advisor", gapB115Pippin, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	foods := make([]uuid.UUID, 4)
	for i := range foods {
		foods[i] = b12Permanent(g, me.ID, "Food", "Artifact — Food")
	}
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, pippin, 1, game.ActivateAbilityParams{Strict: true, SacrificeIDs: foods[:3]}); err == nil {
		t.Fatal("activated with three Foods")
	}
	b16Activate(t, g, me.ID, pippin, 1, game.ActivateAbilityParams{Strict: true, SacrificeIDs: foods})
	passPriorityAroundTable(t, g)
	for _, f := range foods {
		if g.Battlefield.Contains(f) {
			t.Error("a Food was not sacrificed")
		}
	}
	if p, tough := ptOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("my Bear is %d/%d, want 5/5", p, tough)
	}
	if p, tough := ptOf(t, g, pippin); p != 1 || tough != 1 {
		t.Errorf("Pippin is %d/%d, want unchanged: only other creatures", p, tough)
	}
	if p, _ := ptOf(t, g, theirs); p != 2 {
		t.Errorf("their Bear got the pump: power %d", p)
	}
	if c, _ := battlefieldCard(g, bear); !game.HasKeyword(&c, "haste") {
		t.Error("my Bear did not gain haste")
	}
}

// --- Bilbo, Birthday Celebrant -----------------------------------------------

func TestGapB115BilboAddsOneToEveryLifeGain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Bilbo, Birthday Celebrant", "Legendary Creature — Halfling Rogue", gapB115Bilbo, false)
	life := me.Life
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	if got := me.Life - life; got != 4 {
		t.Errorf("gained %d, want 3 plus 1", got)
	}
}

func TestGapB115BilboNeedsOneHundredElevenLifeToSearch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bilbo := pushCatalogPermanent(g, me.ID, "Bilbo, Birthday Celebrant", "Legendary Creature — Halfling Rogue", gapB115Bilbo, false)
	bear := pushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	wurm := pushLibraryCardForTest(me, game.Card{Name: "Wurm", TypeLine: "Creature — Wurm", Power: 5, Toughness: 5})
	pushLibraryCardForTest(me, game.Card{Name: "Spell", TypeLine: "Sorcery"})
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}{W}{B}{G}")
	me.Life = 110
	if err := g.ActivateCatalogAbility(me.ID, bilbo, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with 110 life")
	}
	me.Life = 111
	b16Activate(t, g, me.ID, bilbo, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, bear, wurm)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bilbo) {
		t.Error("Bilbo was not exiled")
	}
	if !g.Exile.Contains(bilbo) {
		t.Error("Bilbo is not in exile")
	}
	if !g.Battlefield.Contains(bear) || !g.Battlefield.Contains(wurm) {
		t.Error("the searched creatures are not on the battlefield")
	}
}
