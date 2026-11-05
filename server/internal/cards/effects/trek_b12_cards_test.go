package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b12_cards_test.go — Star Trek, batch 12 (#4973). Each test plays
// the real catalog entry through the engine and asserts what its text
// says.

const (
	trekB12GoodDayOracle   = "4c4559a5-a7a4-4a40-8a8f-e7375c741748"
	trekB12ZekOracle       = "654ce8c5-b329-4046-9728-7fa276bc8695"
	trekB12DukatOracle     = "f41f0290-0f9d-4c51-935e-f69d2120e4a7"
	trekB12DeadJimOracle   = "ee7150d1-8b38-4387-9b9c-a40180430170"
	trekB12HiveMindOracle  = "8e6ae6ed-065c-47bd-a3c6-69b2608e7ddc"
	trekB12ImpossibleOracl = "e277ed07-2113-4465-a45c-a64490814c1d"
	trekB12MoonlightOracle = "90c78f79-982f-4acd-a8ba-0c6ce115a248"
	trekB12KaiWinnOracle   = "16a5ed42-ee86-4dcb-b8ea-239ac2b13d67"
)

func TestTrekB12EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB12GoodDayOracle:   "A Good Day to Die",
		trekB12ZekOracle:       "Grand Nagus Zek",
		trekB12DukatOracle:     "Gul Dukat, Colonial Prefect",
		trekB12DeadJimOracle:   "He's Dead, Jim",
		trekB12HiveMindOracle:  "Hive Mind Coprocessor",
		trekB12ImpossibleOracl: "Impossible Choice",
		trekB12MoonlightOracle: "In the Pale Moonlight",
		trekB12KaiWinnOracle:   "Kai Winn, Faithless Leader",
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

// --- A Good Day to Die -------------------------------------------------------

func TestTrekB12GoodDayPumpsAndDrawsWhenThatCreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Other", "Creature — Ox", 1, 1)
	castCatalogSpell(t, g, "A Good Day to Die", "Instant", trekB12GoodDayOracle, []game.TargetRef{cardRef(bear)})
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, bear)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 2 {
		t.Fatalf("the target is %d/%d, want 4/2", c.CurrentPower(), c.CurrentToughness())
	}

	hand := handSize(me)
	destroyAndSettle(t, g, other)
	if got := handSize(me); got != hand {
		t.Fatalf("another creature died and the hand went %d -> %d", hand, got)
	}
	destroyAndSettle(t, g, bear)
	if got := handSize(me); got != hand+1 {
		t.Errorf("the pumped creature died: hand %d, want %d", got, hand+1)
	}
}

func TestTrekB12GoodDayDrawsForTheCasterWhenAnOpponentsCreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "A Good Day to Die", "Instant", trekB12GoodDayOracle, []game.TargetRef{cardRef(theirs)})
	passPriorityAroundTable(t, g)
	mine, theirHand := handSize(me), handSize(opp)
	destroyAndSettle(t, g, theirs)
	if handSize(me) != mine+1 || handSize(opp) != theirHand {
		t.Errorf("hands me %d->%d, opponent %d->%d; want the caster to draw", mine, handSize(me), theirHand, handSize(opp))
	}
}

func TestTrekB12GoodDayRefusesANoncreatureTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	fillPoolColored(me, "B", 1)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "A Good Day to Die", TypeLine: "Instant", ManaCost: "{B}",
			OracleID: trekB12GoodDayOracle, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{cardRef(rock)}}); err == nil {
		t.Fatal("A Good Day to Die was cast at an artifact")
	}
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{cardRef(bear)}}); err != nil {
		t.Fatalf("a strict cast at a creature for {B}: %v", err)
	}
}

// --- Grand Nagus Zek ---------------------------------------------------------

func TestTrekB12ZekMakesOneTreasureATurnAndGrowsWithThem(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zek := b27Push(g, me.ID, "Grand Nagus Zek", "Legendary Creature — Ferengi Noble", trekB12ZekOracle, "{2}{B}", 1, 1, "B")
	first := b12Creature(g, opp.ID, "First", "Creature — Bear", 2, 2)
	second := b12Creature(g, me.ID, "Second", "Creature — Bear", 2, 2)
	advanceToMain(t, g)

	destroyAndSettle(t, g, first)
	if n, _ := countTokensNamed(g, me.ID, "Treasure"); n != 1 {
		t.Fatalf("%d Treasures after an opponent's creature died, want 1", n)
	}
	c := layeredCard(t, g, zek)
	if c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("Zek with one Treasure is %d/%d, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}

	destroyAndSettle(t, g, second)
	if n, _ := countTokensNamed(g, me.ID, "Treasure"); n != 1 {
		t.Fatalf("%d Treasures after a second death this turn, want still 1", n)
	}

	// Next turn the ability is fresh.
	third := b12Creature(g, opp.ID, "Third", "Creature — Bear", 2, 2)
	advanceToUpkeepOf(t, g, 1)
	destroyAndSettle(t, g, third)
	if n, _ := countTokensNamed(g, me.ID, "Treasure"); n != 2 {
		t.Fatalf("%d Treasures after a death on the next turn, want 2", n)
	}
	c = layeredCard(t, g, zek)
	if c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("Zek with two Treasures is %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestTrekB12ZekOnlyOneTreasureWhenTwoDieAtOnce(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Grand Nagus Zek", "Legendary Creature — Ferengi Noble", trekB12ZekOracle, "{2}{B}", 1, 1, "B")
	a := b12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{a, b} {
			if err := g.DestroyPermanentForEffect(id); err != nil {
				t.Fatalf("destroy: %v", err)
			}
		}
	})
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("%d Treasures after two deaths at once, want 1", n)
	}
}

func TestTrekB12ZekIgnoresHisOwnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	zek := b27Push(g, me.ID, "Grand Nagus Zek", "Legendary Creature — Ferengi Noble", trekB12ZekOracle, "{2}{B}", 1, 1, "B")
	advanceToMain(t, g)
	destroyAndSettle(t, g, zek)
	if n, _ := countTokensNamed(g, me.ID, "Treasure"); n != 0 {
		t.Errorf("%d Treasures after Zek himself died, want 0", n)
	}
}

// --- Gul Dukat, Colonial Prefect ---------------------------------------------

func TestTrekB12DukatDrainsTheLifeEachOpponentLost(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b, c := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	dukat := b27Push(g, me.ID, "Gul Dukat, Colonial Prefect", "Legendary Creature — Cardassian Rogue", trekB12DukatOracle, "{4}{B}{B}{B}", 6, 6, "B")
	if v := layeredCard(t, g, dukat); !game.HasKeyword(&v, "menace") {
		t.Error("Gul Dukat has no menace")
	}
	advanceToMain(t, g)
	for _, step := range []struct {
		p *game.Player
		n int
	}{{a, 5}, {b, 2}} {
		if _, err := g.ChangePlayerLife(step.p.ID, -step.n); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := g.ChangePlayerLife(me.ID, -3); err != nil {
		t.Fatal(err)
	}
	aLife, bLife, cLife, myLife := a.Life, b.Life, c.Life, me.Life
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if a.Life != aLife-5 || b.Life != bLife-2 || c.Life != cLife {
		t.Errorf("opponents at %d/%d/%d, want %d/%d/%d", a.Life, b.Life, c.Life, aLife-5, bLife-2, cLife)
	}
	if me.Life != myLife+7 {
		t.Errorf("you are at %d, want %d (the 7 life lost this way; your own loss is not counted)", me.Life, myLife+7)
	}
}

func TestTrekB12DukatTriggersOnAnOpponentsEndStepToo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Gul Dukat, Colonial Prefect", "Legendary Creature — Cardassian Rogue", trekB12DukatOracle, "{4}{B}{B}{B}", 6, 6, "B")
	advanceToUpkeepOf(t, g, 1)
	if _, err := g.ChangePlayerLife(opp.ID, -4); err != nil {
		t.Fatal(err)
	}
	oppLife, myLife := opp.Life, me.Life
	advanceToEndStepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-4 || me.Life != myLife+4 {
		t.Errorf("opponent %d (want %d), you %d (want %d)", opp.Life, oppLife-4, me.Life, myLife+4)
	}
}

// --- He's Dead, Jim ----------------------------------------------------------

func TestTrekB12HesDeadJimDestroysThenSurveils(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "Bin Me", "Keep Me")
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "He's Dead, Jim", "Instant", trekB12DeadJimOracle, []game.TargetRef{cardRef(victim)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the target creature survived")
	}
	c := surveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("want a surveil 1 prompt, got %+v", c)
	}
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if !me.Graveyard.Contains(c.ScryCards[0]) {
		t.Error("the surveilled card did not go to the graveyard")
	}
}

// --- Hive Mind Coprocessor ---------------------------------------------------

// trekB12HiveMindSwing attacks seat 1 with `attackers` and walks into the
// combat damage step.
func trekB12HiveMindSwing(t *testing.T, g *game.Game, attackers ...uuid.UUID) {
	t.Helper()
	opp := g.Seats[1]
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, a := range attackers {
		if err := g.DeclareAttacker(a, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
}

func TestTrekB12HiveMindArtifactCreaturesPayLifeToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hive := b27Push(g, me.ID, "Hive Mind Coprocessor", "Artifact Creature — Borg", trekB12HiveMindOracle, "{4}{B}", 4, 4, "B")
	drone := b12Creature(g, me.ID, "Drone", "Artifact Creature — Construct", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life, hand := me.Life, handSize(me)
	trekB12HiveMindSwing(t, g, hive, drone, bear)

	// Two artifact creatures connected: two prompts, the bear none.
	answerAnyTriggerOrderPrompt(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if latestConfirmFor(g, me.ID) != nil {
		t.Fatal("a third pay-1-life prompt opened; the non-artifact Bear has no such ability")
	}
	if me.Life != life-1 || handSize(me) != hand+1 {
		t.Errorf("life %d (want %d), hand %d (want %d): one paid, one declined", me.Life, life-1, handSize(me), hand+1)
	}
}

func TestTrekB12HiveMindGivesNothingToAnOpponentsArtifactCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, opp.ID, "Hive Mind Coprocessor", "Artifact Creature — Borg", trekB12HiveMindOracle, "{4}{B}", 4, 4, "B")
	drone := b12Creature(g, me.ID, "Drone", "Artifact Creature — Construct", 1, 1)
	trekB12HiveMindSwing(t, g, drone)
	if latestConfirmFor(g, me.ID) != nil {
		t.Error("an opponent's Coprocessor granted my artifact creature the draw")
	}
}

// --- Impossible Choice -------------------------------------------------------

func TestTrekB12ImpossibleChoiceSeparationDestroysAndGivesOfficers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 3, 3)
	castModal(t, g, "Impossible Choice", "Instant", trekB12ImpossibleOracl, []int{0}, []game.TargetRef{cardRef(victim)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the target survived")
	}
	officers := trekB12Tokens(t, g, opp.ID, "Officer")
	if len(officers) != 2 {
		t.Fatalf("the victim's controller has %d Officers, want 2", len(officers))
	}
	for _, o := range officers {
		if o.CurrentPower() != 1 || o.CurrentToughness() != 1 || !o.IsCreature() || len(o.Colors) != 1 || o.Colors[0] != "R" {
			t.Errorf("Officer %+v, want a 1/1 red creature", o)
		}
	}
	if n := len(trekB12Tokens(t, g, me.ID, "Officer")); n != 0 {
		t.Errorf("the caster got %d Officers", n)
	}
}

func TestTrekB12ImpossibleChoiceRemainFusedDrawsTwoAndEveryoneLosesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lives := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		lives[i] = p.Life
	}
	castModal(t, g, "Impossible Choice", "Instant", trekB12ImpossibleOracl, []int{1}, nil)
	hand := handSize(me)
	passPriorityAroundTable(t, g)
	if handSize(me) != hand+2 {
		t.Errorf("hand %d, want %d", handSize(me), hand+2)
	}
	for i, p := range g.Seats {
		if p.Life != lives[i]-2 {
			t.Errorf("seat %d at %d, want %d", i, p.Life, lives[i]-2)
		}
	}
}

func TestTrekB12ImpossibleChoiceSeparationRefusesALand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	land := b12Permanent(g, opp.ID, "Swamp", "Basic Land — Swamp")
	fillPoolColored(me, "B", 3)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Impossible Choice", TypeLine: "Instant", ManaCost: "{2}{B}",
			OracleID: trekB12ImpossibleOracl, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{0}, Targets: []game.TargetRef{cardRef(land)}}); err == nil {
		t.Fatal("Surgical Separation was cast at a land")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{1}}); err != nil {
		t.Fatalf("a strict Remain Fused cast for {2}{B}: %v", err)
	}
}

// trekB12Tokens is the tokens named `name` that `controller` controls, as
// the layers see them.
func trekB12Tokens(t *testing.T, g *game.Game, controller uuid.UUID, name string) []game.Card {
	t.Helper()
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name && IsToken(c) {
			out = append(out, layeredCard(t, g, c.InstanceID))
		}
	}
	return out
}

// --- In the Pale Moonlight ---------------------------------------------------

func TestTrekB12PaleMoonlightChapters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(other.ID, RedGoblinToken(), 1) })
	goblin := findBattlefieldByName(g, "Goblin")

	saga := castCatalogSpell(t, g, "In the Pale Moonlight", "Enchantment — Saga", trekB12MoonlightOracle, nil)
	passPriorityAroundTable(t, g)
	// I: each opponent sacrifices a nontoken creature.
	if g.Battlefield.Contains(theirs) {
		t.Error("chapter I: the opponent's nontoken creature survived")
	}
	if !g.Battlefield.Contains(goblin) {
		t.Error("chapter I took a creature token")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("chapter I took the controller's creature")
	}

	// II: investigate.
	advanceToPrecombatMainOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Clue"); n != 1 {
		t.Fatalf("chapter II: %d Clues, want 1", n)
	}
	clue := findBattlefieldByName(g, "Clue")

	// III: sacrifice the Clue for two Romulans.
	advanceToPrecombatMainOf(t, g, 0)
	passPriorityAroundTable(t, g)
	answerOwnPermanents(t, g, me.ID, clue)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(clue) {
		t.Error("chapter III: the Clue was not sacrificed")
	}
	romulans := trekB12Tokens(t, g, me.ID, "Romulan")
	if len(romulans) != 2 {
		t.Fatalf("chapter III: %d Romulans, want 2", len(romulans))
	}
	for _, r := range romulans {
		if r.CurrentPower() != 2 || r.CurrentToughness() != 2 || len(r.Colors) != 1 || r.Colors[0] != "B" {
			t.Errorf("Romulan %+v, want a 2/2 black creature", r)
		}
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}

func TestTrekB12PaleMoonlightChapterThreeMayBeDeclined(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "In the Pale Moonlight", "Enchantment — Saga", trekB12MoonlightOracle, nil)
	passPriorityAroundTable(t, g)
	advanceToPrecombatMainOf(t, g, 0)
	passPriorityAroundTable(t, g)
	advanceToPrecombatMainOf(t, g, 0)
	passPriorityAroundTable(t, g)
	answerOwnPermanents(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) {
		t.Error("declining still sacrificed the creature")
	}
	if n := len(trekB12Tokens(t, g, me.ID, "Romulan")); n != 0 {
		t.Errorf("%d Romulans after declining, want 0", n)
	}
}

// --- Kai Winn, Faithless Leader ----------------------------------------------

// trekB12CastKaiWinn casts Kai Winn and names `name`.
func trekB12CastKaiWinn(t *testing.T, g *game.Game, name string) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	kai := castCatalogSpell(t, g, "Kai Winn, Faithless Leader", "Legendary Creature — Bajoran Cleric", trekB12KaiWinnOracle, nil)
	passPriorityAroundTable(t, g)
	actNameFor(t, g, me.ID, name)
	return kai
}

func TestTrekB12KaiWinnPunishesAnOpponentsNamedActivation(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	trekB12CastKaiWinn(t, g, "Krenko, Mob Boss")
	theirKrenko := actPushKrenko(g, them.ID)
	myKrenko := actPushKrenko(g, me.ID)
	theirSolRing := b21Push(g, them.ID, "Sol Ring", "Artifact", actSolRingOracle, 0, 0)

	life := them.Life
	if err := g.ActivateCatalogAbility(them.ID, theirKrenko, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the opponent's Krenko: %v", err)
	}
	passPriorityAroundTable(t, g)
	if them.Life != life-3 {
		t.Fatalf("the opponent activated the named source and is at %d, want %d", them.Life, life-3)
	}

	life = them.Life
	if err := g.ActivateManaAbility(them.ID, theirSolRing, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the opponent's Sol Ring: %v", err)
	}
	passPriorityAroundTable(t, g)
	if them.Life != life {
		t.Errorf("an unnamed source cost the opponent %d life", life-them.Life)
	}

	myLife := me.Life
	if err := g.ActivateCatalogAbility(me.ID, myKrenko, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("my Krenko: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != myLife {
		t.Errorf("Kai Winn's controller lost %d life for their own activation", myLife-me.Life)
	}
}

func TestTrekB12KaiWinnCountsAManaAbility(t *testing.T) {
	g := newCatalogGame(t)
	them := g.Seats[1]
	trekB12CastKaiWinn(t, g, "Sol Ring")
	theirSolRing := b21Push(g, them.ID, "Sol Ring", "Artifact", actSolRingOracle, 0, 0)
	life := them.Life
	if err := g.ActivateManaAbility(them.ID, theirSolRing, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the opponent's Sol Ring: %v", err)
	}
	passPriorityAroundTable(t, g)
	if them.Life != life-3 {
		t.Errorf("the opponent tapped the named Sol Ring and is at %d, want %d", them.Life, life-3)
	}
}

func TestTrekB12KaiWinnPunishesAnOpponentsNamedSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	trekB12CastKaiWinn(t, g, "Shock")
	life := them.Life
	castFromHandForTest(t, g, them, "Shock", "Instant", "", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if them.Life != life-3 {
		t.Fatalf("the opponent cast the named spell and is at %d, want %d", them.Life, life-3)
	}
	life = them.Life
	castFromHandForTest(t, g, them, "Opt", "Instant", "", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if them.Life != life {
		t.Errorf("an unnamed spell cost the opponent %d life", life-them.Life)
	}
	myLife := me.Life
	castFromHandForTest(t, g, me, "Shock", "Instant", "", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if me.Life != myLife {
		t.Errorf("Kai Winn's controller lost %d life for their own spell", myLife-me.Life)
	}
}
