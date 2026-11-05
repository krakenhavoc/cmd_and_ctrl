package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Star Trek Commander batch 36 (#4974). Each test plays the real catalog
// entry through the engine and asserts what its text says.

const (
	trekB36WoundedOracle     = "2147101f-d24a-45bb-a141-b69ef7841f53"
	trekB36AssessorOracle    = "48848570-00c4-4a89-a84c-e3df47a8b0d2"
	trekB36WolfOracle        = "bb203cf5-1545-43c7-8278-3720616c22f4"
	trekB36ChampionOracle    = "9d08cc9b-2b26-49de-80aa-962a358bfc1a"
	trekB36IntegrantOracle   = "24964ce8-1eec-49b3-9db1-b895bc686ec5"
	trekB36KurnOracle        = "2a078705-4afe-4887-94cc-f12f4690fb01"
	trekB36PromiseOracle     = "a5866d5e-ee03-4bca-9964-0d391eae9011"
	trekB36ReclamationOracle = "7c348d24-af26-4171-bfa7-44ce9bd36f06"
	trekB36BeaconOracle      = "66fc3c1e-acea-46eb-8af8-b0740f006add"
	trekB36ColonistOracle    = "fd549938-8919-45ac-bf52-92492d9d1d49"
)

func TestTrekB36EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB36WoundedOracle:     "Assimilate the Wounded",
		trekB36AssessorOracle:    "Assimilation Assessor",
		trekB36WolfOracle:        "Battle of Wolf 359",
		trekB36ChampionOracle:    "Blood-Sport Champion",
		trekB36IntegrantOracle:   "Extraction Integrant",
		trekB36KurnOracle:        "Kurn, Son of Mogh",
		trekB36PromiseOracle:     "The Promise of Kahless",
		trekB36ReclamationOracle: "Reclamation Specialist",
		trekB36BeaconOracle:      "The Beacon of Kahless",
		trekB36ColonistOracle:    "Ferocious Colonist",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func trekB36Me(g *game.Game) *game.Player { return g.Seats[g.Turn.ActiveSeat] }

func trekB36Opp(g *game.Game, n int) *game.Player {
	return g.Seats[(g.Turn.ActiveSeat+n)%len(g.Seats)]
}

// trekB36Cast puts `c` in the active seat's hand in a main phase and casts
// it under `params`, without settling the stack.
func trekB36Cast(t *testing.T, g *game.Game, c game.Card, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	me := trekB36Me(g)
	advanceToMain(t, g)
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = me.ID, me.ID
	g.WithWriteLock(func() { me.Hand.PushTop(c) })
	if err := g.CastSpell(me.ID, c.InstanceID, params); err != nil {
		t.Fatalf("cast %s: %v", c.Name, err)
	}
	return c.InstanceID
}

// trekB36Ready lets a creature that entered this turn attack.
func trekB36Ready(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		if c := findBattlefieldCardForTest(g, id); c != nil {
			c.SummonedThisTurn = false
		}
	})
}

// trekB36GraveyardCreature puts a creature card into `p`'s graveyard
// directly (not from the battlefield).
func trekB36GraveyardCreature(g *game.Game, p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	})
	return id
}

// trekB36Assimilated asserts `id` is on the battlefield under `me` as an
// assimilated card: one +1/+1 counter, an artifact creature whose only
// creature type is Borg.
func trekB36Assimilated(t *testing.T, g *game.Game, id, me uuid.UUID, lost string) {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	c = tmtB1Card(t, g, id)
	if c.Controller != me {
		t.Errorf("%s is controlled by %s, want the caster", c.Name, c.Controller)
	}
	if !c.IsArtifact() || !c.IsCreature() || !c.HasSubtypeWord("Borg") || c.HasSubtypeWord(lost) {
		t.Errorf("%s has type line %q, want a Borg artifact creature without %s", c.Name, c.TypeLine, lost)
	}
	if got := counterCount(g, id, game.CounterPlusOne); got != 1 {
		t.Errorf("%s has %d +1/+1 counters, want 1", c.Name, got)
	}
}

// trekB36OnBattlefieldNamed is the one permanent named `name`.
func trekB36OnBattlefieldNamed(t *testing.T, g *game.Game, name string) uuid.UUID {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			return c.InstanceID
		}
	}
	t.Fatalf("%s is not on the battlefield", name)
	return uuid.Nil
}

// --- Assimilate the Wounded ---------------------------------------------------

func TestTrekB36AssimilateTheWoundedTakesOneDeadCreaturePerOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1, opp2 := trekB36Me(g), trekB36Opp(g, 1), trekB36Opp(g, 2)
	bear := b12Creature(g, opp1.ID, "Their Bear", "Creature — Bear", 2, 2)
	elf := b12Creature(g, opp2.ID, "Their Elf", "Creature — Elf", 1, 1)
	advanceToMain(t, g)
	b27Kill(g, bear)
	b27Kill(g, elf)
	passPriorityAroundTable(t, g)
	if !inZone(opp1.Graveyard, bear) || !inZone(opp2.Graveyard, elf) {
		t.Fatal("the creatures did not die")
	}
	castCatalogSpell(t, g, "Assimilate the Wounded", "Instant", trekB36WoundedOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetCard, ID: elf},
	})
	passPriorityAroundTable(t, g)
	trekB36Assimilated(t, g, bear, me.ID, "Bear")
	trekB36Assimilated(t, g, elf, me.ID, "Elf")
}

func TestTrekB36AssimilateTheWoundedRefusesCardsThatDidNotDieThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	opp1 := trekB36Opp(g, 1)
	advanceToMain(t, g)
	milled := trekB36GraveyardCreature(g, opp1, "Milled Bear", "Creature — Bear")
	if err := castCatalogSpellErr(t, g, "Assimilate the Wounded", "Instant", trekB36WoundedOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: milled}}); err == nil {
		t.Error("a creature card that never left the battlefield this turn must not be a legal target")
	}
}

func TestTrekB36AssimilateTheWoundedTakesAtMostOnePerOpponent(t *testing.T) {
	g := newCatalogGame(t)
	opp1 := trekB36Opp(g, 1)
	a := b12Creature(g, opp1.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp1.ID, "Bear B", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	b27Kill(g, a)
	b27Kill(g, b)
	passPriorityAroundTable(t, g)
	if err := castCatalogSpellErr(t, g, "Assimilate the Wounded", "Instant", trekB36WoundedOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}); err == nil {
		t.Error("two cards from one opponent's graveyard must be refused")
	}
	// The caster's own dead creature is not an opponent's.
	mine := b12Creature(g, trekB36Me(g).ID, "My Bear", "Creature — Bear", 2, 2)
	b27Kill(g, mine)
	passPriorityAroundTable(t, g)
	if err := castCatalogSpellErr(t, g, "Assimilate the Wounded", "Instant", trekB36WoundedOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}}); err == nil {
		t.Error("a card in the caster's own graveyard must be refused")
	}
}

// --- Assimilation Assessor ----------------------------------------------------

func TestTrekB36AssessorExilesAGraveyardAndAssimilatesFromIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1, opp2 := trekB36Me(g), trekB36Opp(g, 1), trekB36Opp(g, 2)
	bear := trekB36GraveyardCreature(g, opp1, "Their Bear", "Creature — Bear")
	elf := trekB36GraveyardCreature(g, opp1, "Their Elf", "Creature — Elf")
	bolt := uuid.New()
	g.WithWriteLock(func() {
		opp1.Graveyard.PushTop(game.Card{InstanceID: bolt, Name: "Bolt", TypeLine: "Instant", Owner: opp1.ID, Controller: opp1.ID})
	})
	other := trekB36GraveyardCreature(g, opp2, "Other Graveyard Ox", "Creature — Ox")

	assessor := trekB36Cast(t, g, game.Card{Name: "Assimilation Assessor", TypeLine: "Artifact Creature — Borg",
		OracleID: trekB36AssessorOracle, Power: 5, Toughness: 5}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	answerPickTargetPlayer(t, g, opp1.ID)
	passPriorityAroundTable(t, g)
	if opp1.Graveyard.Size() != 0 {
		t.Fatalf("the target opponent's graveyard has %d cards, want it exiled", opp1.Graveyard.Size())
	}
	if !inZone(opp2.Graveyard, other) {
		t.Fatal("another opponent's graveyard was touched")
	}
	for _, id := range []uuid.UUID{bear, elf, bolt} {
		if !inZone(g.Exile, id) {
			t.Fatalf("%s is not in exile", id)
		}
	}
	// An unrelated creature card in exile is not "exiled with it".
	stranger := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.PushTop(game.Card{InstanceID: stranger, Name: "Stranger", TypeLine: "Creature — Human", Power: 1, Toughness: 1, Owner: opp2.ID, Controller: opp2.ID})
	})

	trekB36Ready(g, assessor)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(assessor, opp1.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	pick := pendingChooseCards(t, g, me.ID)
	if pick == nil {
		t.Fatal("two linked creature cards should ask which to assimilate")
	}
	if !slices.Contains(pick.ChooseCards, bear) || !slices.Contains(pick.ChooseCards, elf) ||
		slices.Contains(pick.ChooseCards, bolt) || slices.Contains(pick.ChooseCards, stranger) {
		t.Fatalf("offered %v, want exactly the two creature cards exiled with it", pick.ChooseCards)
	}
	if pick.ChooseMin != 1 {
		t.Errorf("ChooseMin = %d, want 1: assimilate is not optional", pick.ChooseMin)
	}
	answerChooseCards(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	// A card leaving exile is a new object with a new ID.
	trekB36Assimilated(t, g, trekB36OnBattlefieldNamed(t, g, "Their Bear"), me.ID, "Bear")
	if !inZone(g.Exile, elf) {
		t.Error("the card not chosen stays in exile")
	}
}

func TestTrekB36AssessorAttackWithNothingExiledDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1 := trekB36Me(g), trekB36Opp(g, 1)
	assessor := b27Push(g, me.ID, "Assimilation Assessor", "Artifact Creature — Borg", trekB36AssessorOracle, "{4}{B}", 5, 5, "B")
	stranger := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.PushTop(game.Card{InstanceID: stranger, Name: "Stranger", TypeLine: "Creature — Human", Power: 1, Toughness: 1, Owner: opp1.ID, Controller: opp1.ID})
	})
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(assessor, opp1.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if !inZone(g.Exile, stranger) {
		t.Error("a card not exiled with the Assessor must not be assimilated")
	}
}

// --- Battle of Wolf 359 -----------------------------------------------------------

func TestTrekB36BattleOfWolf359ShrinksNonartifactCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1 := trekB36Me(g), trekB36Opp(g, 1)
	cube1 := b12Creature(g, me.ID, "Drone A", "Artifact Creature — Borg", 1, 1)
	cube2 := b12Creature(g, me.ID, "Drone B", "Artifact Creature — Borg", 1, 1)
	mine := b12Creature(g, me.ID, "My Ox", "Creature — Ox", 3, 3)
	theirs := b12Creature(g, opp1.ID, "Their Bear", "Creature — Bear", 2, 2)
	theirArtifact := b12Creature(g, opp1.ID, "Their Golem", "Artifact Creature — Golem", 1, 1)

	castCatalogSpell(t, g, "Battle of Wolf 359", "Sorcery", trekB36WolfOracle, nil)
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()

	if c := tmtB1Card(t, g, mine); c.CurrentPower() != 1 || c.CurrentToughness() != 1 {
		t.Errorf("my nonartifact creature is %d/%d, want 1/1 (X = 2, my artifact creatures only)", c.CurrentPower(), c.CurrentToughness())
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the opponent's 2/2 should have died to -2/-2")
	}
	for _, id := range []uuid.UUID{cube1, cube2, theirArtifact} {
		if c := tmtB1Card(t, g, id); c.CurrentPower() != 1 || c.CurrentToughness() != 1 {
			t.Errorf("%s is %d/%d, artifact creatures are spared", c.Name, c.CurrentPower(), c.CurrentToughness())
		}
	}
	// The set was locked at resolution.
	late := b12Creature(g, opp1.ID, "Late Bear", "Creature — Bear", 2, 2)
	if c := tmtB1Card(t, g, late); c.CurrentPower() != 2 {
		t.Errorf("a creature that entered afterwards has power %d, want 2", c.CurrentPower())
	}
}

// --- Blood-Sport Champion -----------------------------------------------------

func trekB36Champion() game.Card {
	return game.Card{Name: "Blood-Sport Champion", TypeLine: "Creature — Klingon Warrior",
		OracleID: trekB36ChampionOracle, Power: 2, Toughness: 2}
}

func TestTrekB36ChampionKickedGivesOthersDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1 := trekB36Me(g), trekB36Opp(g, 1)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp1.ID, "Their Bear", "Creature — Bear", 2, 2)
	champ := trekB36Cast(t, g, trekB36Champion(), game.CastSpellParams{OptionalCosts: []int{0}})
	passPriorityAroundTable(t, g)
	c := tmtB1Card(t, g, bear)
	if !game.HasKeyword(&c, "deathtouch") {
		t.Error("another creature you control should gain deathtouch")
	}
	o := tmtB1Card(t, g, theirs)
	if game.HasKeyword(&o, "deathtouch") {
		t.Error("an opponent's creature must not gain deathtouch")
	}
	ch := tmtB1Card(t, g, champ)
	if !game.HasKeyword(&ch, "deathtouch") {
		t.Error("the Champion has deathtouch")
	}
}

func TestTrekB36ChampionUnkickedTriggersNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := trekB36Me(g)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	champ := trekB36Cast(t, g, trekB36Champion(), game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if triggerOnStack(g, champ) != nil || len(g.PendingTriggers) != 0 {
		t.Fatal("an unkicked Champion must not trigger")
	}
	c := tmtB1Card(t, g, bear)
	if game.HasKeyword(&c, "deathtouch") {
		t.Error("unkicked, other creatures gain nothing")
	}
}

func TestTrekB36ChampionKickerCostsTwoMore(t *testing.T) {
	g := newCatalogGame(t)
	me := trekB36Me(g)
	advanceToMain(t, g)
	floatForTest(g, me, "BC")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Blood-Sport Champion", TypeLine: "Creature — Klingon Warrior",
			ManaCost: "{1}{B}", OracleID: trekB36ChampionOracle, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{OptionalCosts: []int{0}, Strict: true}); err == nil {
		t.Error("two mana cannot pay {1}{B} plus kicker {2}")
	}
}

// --- Extraction Integrant -----------------------------------------------------

func TestTrekB36IntegrantGrantsMenaceToArtifactCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1 := trekB36Me(g), trekB36Opp(g, 1)
	integrant := b27Push(g, me.ID, "Extraction Integrant", "Artifact Creature — Borg", trekB36IntegrantOracle, "{1}{B}", 1, 1, "B")
	drone := b12Creature(g, me.ID, "Drone", "Artifact Creature — Borg", 2, 2)
	ox := b12Creature(g, me.ID, "Ox", "Creature — Ox", 2, 2)
	theirs := b12Creature(g, opp1.ID, "Their Golem", "Artifact Creature — Golem", 2, 2)
	for id, want := range map[uuid.UUID]bool{integrant: true, drone: true, ox: false, theirs: false} {
		c := tmtB1Card(t, g, id)
		if got := game.HasKeyword(&c, "menace"); got != want {
			t.Errorf("%s menace = %v, want %v", c.Name, got, want)
		}
	}
}

func TestTrekB36IntegrantShrinksEachBlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1 := trekB36Me(g), trekB36Opp(g, 1)
	b27Push(g, me.ID, "Extraction Integrant", "Artifact Creature — Borg", trekB36IntegrantOracle, "{1}{B}", 1, 1, "B")
	drone := b12Creature(g, me.ID, "Drone", "Artifact Creature — Borg", 4, 4)
	a := b12Creature(g, opp1.ID, "Blocker A", "Creature — Bear", 3, 3)
	b := b12Creature(g, opp1.ID, "Blocker B", "Creature — Bear", 3, 3)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(drone, opp1.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	// Menace: one blocker alone is refused, two together are legal.
	if err := g.DeclareBlocker(a, drone); err == nil {
		t.Fatal("a granted-menace attacker must not be blocked by one creature")
	}
	if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: a, Attacker: drone}, {Blocker: b, Attacker: drone}}); err != nil {
		t.Fatalf("DeclareBlockers: %v", err)
	}
	lockInBlocks(t, g)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if c := tmtB1Card(t, g, id); c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
			t.Errorf("%s is %d/%d, want 2/2 after blocking", c.Name, c.CurrentPower(), c.CurrentToughness())
		}
	}
}

// --- Kurn, Son of Mogh ------------------------------------------------------------

func TestTrekB36KurnDrawsAndLosesPerOpponentAttacked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1, opp2 := trekB36Me(g), trekB36Opp(g, 1), trekB36Opp(g, 2)
	kurn := b27Push(g, me.ID, "Kurn, Son of Mogh", "Legendary Creature — Klingon Warrior Officer", trekB36KurnOracle, "{2}{B}", 2, 3, "B")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	hand, life := me.Hand.Size(), me.Life
	if err := g.DeclareAttacker(kurn, opp1.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(bear, opp2.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	if triggerOnStack(g, kurn) == nil {
		t.Fatal("Kurn attacking should trigger battlefront")
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 || me.Life != life-2 {
		t.Errorf("hand %d→%d, life %d→%d; want two drawn and 2 lost for two opponents", hand, me.Hand.Size(), life, me.Life)
	}
	c := tmtB1Card(t, g, kurn)
	if !game.HasKeyword(&c, "deathtouch") {
		t.Error("Kurn has deathtouch")
	}
}

func TestTrekB36KurnCountsEachOpponentOnce(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1 := trekB36Me(g), trekB36Opp(g, 1)
	kurn := b27Push(g, me.ID, "Kurn, Son of Mogh", "Legendary Creature — Klingon Warrior Officer", trekB36KurnOracle, "{2}{B}", 2, 3, "B")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	hand, life := me.Hand.Size(), me.Life
	for _, id := range []uuid.UUID{kurn, bear} {
		if err := g.DeclareAttacker(id, opp1.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Errorf("hand %d→%d, life %d→%d; want one drawn and 1 lost for one opponent", hand, me.Hand.Size(), life, me.Life)
	}
}

// --- The Promise of Kahless -----------------------------------------------------

func TestTrekB36PromiseOfKahlessReturnsOtherKlingonsAndWarriors(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1 := trekB36Me(g), trekB36Opp(g, 1)
	worf := b12Creature(g, me.ID, "Klingon Sac", "Creature — Klingon", 2, 2)
	klingon := trekB36GraveyardCreature(g, me, "Dead Klingon", "Creature — Klingon Soldier")
	warrior := trekB36GraveyardCreature(g, me, "Dead Warrior", "Creature — Human Warrior")
	human := trekB36GraveyardCreature(g, me, "Dead Human", "Creature — Human")
	theirs := trekB36GraveyardCreature(g, opp1, "Their Klingon", "Creature — Klingon")

	castCatalogSpell(t, g, "The Promise of Kahless", "Sorcery", trekB36PromiseOracle, nil)
	passPriorityAroundTable(t, g)
	prompt := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil || prompt.ChooseMin != 0 || !slices.Contains(prompt.ChooseCards, worf) {
		t.Fatal("the sacrifice is a may over your Klingons and Warriors")
	}
	answerOwnPermanents(t, g, me.ID, worf)
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{klingon, warrior} {
		if c, ok := battlefieldCard(g, id); !ok || c.Controller != me.ID {
			t.Errorf("%s should be back on the battlefield under its owner", id)
		}
	}
	if !inZone(me.Graveyard, worf) {
		t.Error("the sacrificed Klingon is not \"other\" and stays in the graveyard")
	}
	if !inZone(me.Graveyard, human) {
		t.Error("a creature card that is neither Klingon nor Warrior stays")
	}
	if !inZone(opp1.Graveyard, theirs) {
		t.Error("an opponent's graveyard is not yours")
	}
}

func TestTrekB36PromiseOfKahlessDeclinedReturnsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := trekB36Me(g)
	worf := b12Creature(g, me.ID, "Klingon", "Creature — Klingon", 2, 2)
	ox := b12Creature(g, me.ID, "Ox", "Creature — Ox", 2, 2)
	dead := trekB36GraveyardCreature(g, me, "Dead Warrior", "Creature — Human Warrior")
	castCatalogSpell(t, g, "The Promise of Kahless", "Sorcery", trekB36PromiseOracle, nil)
	passPriorityAroundTable(t, g)
	prompt := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil || slices.Contains(prompt.ChooseCards, ox) {
		t.Fatal("only a Klingon or Warrior may be sacrificed")
	}
	answerOwnPermanents(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(worf) || !inZone(me.Graveyard, dead) {
		t.Error("declining sacrifices nothing and returns nothing")
	}
}

// --- Reclamation Specialist -------------------------------------------------------

func TestTrekB36ReclamationSpecialistSacrificesAnArtifactCreatureToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := trekB36Me(g)
	spec := b27Push(g, me.ID, "Reclamation Specialist", "Artifact Creature — Borg", trekB36ReclamationOracle, "{2}{B}", 2, 2, "B")
	drone := b12Creature(g, me.ID, "Drone", "Artifact Creature — Borg", 1, 1)
	ox := b12Creature(g, me.ID, "Ox", "Creature — Ox", 2, 2)
	advanceToMain(t, g)

	floatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, spec, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{ox}, Strict: true}); err == nil {
		t.Fatal("a nonartifact creature cannot pay the cost")
	}
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, spec, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{drone}, Strict: true})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(drone) {
		t.Error("the artifact creature was not sacrificed")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d→%d, want one drawn", hand, me.Hand.Size())
	}
	if err := g.ActivateCatalogAbility(me.ID, spec, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{spec}, Strict: true}); err == nil {
		t.Error("with no mana the ability cannot be activated")
	}
}

// --- The Beacon of Kahless ----------------------------------------------------------

func TestTrekB36BeaconAddsOneManaOfAnyColor(t *testing.T) {
	g := newCatalogGame(t)
	me := trekB36Me(g)
	beacon := b27Push(g, me.ID, "The Beacon of Kahless", "Legendary Artifact", trekB36BeaconOracle, "{2}{R}", 0, 0, "R")
	if err := g.ActivateManaAbility(me.ID, beacon, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := pendingOfKind(g, game.PendingChoiceMana)
	if pick == nil {
		t.Fatal("no colour pick")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "U"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := poolColors(me); !slices.Equal(got, []string{"U"}) {
		t.Errorf("pool = %v, want one {U}", got)
	}
}

func TestTrekB36BeaconBattlefrontAddsOnePerOpponentAttacked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1, opp2 := trekB36Me(g), trekB36Opp(g, 1), trekB36Opp(g, 2)
	beacon := b27Push(g, me.ID, "The Beacon of Kahless", "Legendary Artifact", trekB36BeaconOracle, "{2}{R}", 0, 0, "R")
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, opp1.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(b, opp2.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	if err := g.ActivateManaAbility(me.ID, beacon, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := pendingOfKind(g, game.PendingChoiceMana)
	if pick == nil {
		t.Fatal("no colour pick")
	}
	for _, c := range game.AllColors {
		if pick.ManaAmounts[c] != 2 {
			t.Errorf("amounts = %v, want 2 of every colour", pick.ManaAmounts)
			break
		}
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := poolColors(me); !slices.Equal(got, []string{"R", "R"}) {
		t.Errorf("pool = %v, want two {R}", got)
	}
}

func TestTrekB36BeaconBattlefrontAddsNothingBeforeAnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me := trekB36Me(g)
	beacon := b27Push(g, me.ID, "The Beacon of Kahless", "Legendary Artifact", trekB36BeaconOracle, "{2}{R}", 0, 0, "R")
	advanceToMain(t, g)
	_ = g.ActivateManaAbility(me.ID, beacon, 1, game.ManaAbilityParams{})
	if pick := pendingOfKind(g, game.PendingChoiceMana); pick != nil {
		t.Fatal("with no opponent attacked there is nothing to choose")
	}
	if got := poolColors(me); len(got) != 0 {
		t.Errorf("pool = %v, want empty", got)
	}
}

// --- Ferocious Colonist -----------------------------------------------------------

func TestTrekB36ColonistMeleeAndDoubleStrike(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1, opp2 := trekB36Me(g), trekB36Opp(g, 1), trekB36Opp(g, 2)
	colonist := b27Push(g, me.ID, "Ferocious Colonist", "Creature — Klingon Citizen", trekB36ColonistOracle, "{1}{R}", 0, 1, "R")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	c := tmtB1Card(t, g, colonist)
	if !game.HasKeyword(&c, "double strike") || !game.HasKeyword(&c, game.KeywordMelee) {
		t.Fatal("the Colonist has double strike and melee")
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(colonist, opp1.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(bear, opp2.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if c := tmtB1Card(t, g, colonist); c.CurrentPower() != 2 || c.CurrentToughness() != 3 {
		t.Errorf("Colonist is %d/%d, want 2/3 for two opponents attacked", c.CurrentPower(), c.CurrentToughness())
	}
}
