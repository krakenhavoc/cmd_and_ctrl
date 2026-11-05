package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b18_a_cards_test.go — Star Trek, batch 18 (#4973), part A. Each
// test plays the real catalog entry and asserts what its text says.

const (
	trekB18ASurveyorOracle = "1b248ef5-ec57-481d-b47a-357d3867054f"
	trekB18ASecurityOracle = "896077f9-b26e-4964-9279-03ee284a3e89"
	trekB18AAnalystOracle  = "bca4d45d-cae4-46f3-b5ac-22fbac59fe56"
	trekB18AWranglerOracle = "a066eeb1-61fc-448b-bbab-650d672652ad"
	trekB18AIdentityOracle = "8cb700bb-daa5-4a26-bb84-e22bc11bdaff"
)

func TestTrekB18AEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB18ASurveyorOracle: "Silicate Surveyor",
		trekB18ASecurityOracle: "Starfleet Security",
		trekB18AAnalystOracle:  "Tal Shiar Analyst",
		trekB18AWranglerOracle: "Targ Wrangler",
		trekB18AIdentityOracle: "Test of Identity",
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

// --- fixtures ------------------------------------------------------------------

// trekB18AClues counts the Clue tokens p controls.
func trekB18AClues(g *game.Game, p uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == p && c.Name == "Clue" && c.HasSubtype("Clue") {
				n++
			}
		}
	})
	return n
}

// trekB18AAttack declares the given creatures as attackers against
// defender and settles the triggers the declaration makes.
func trekB18AAttack(t *testing.T, g *game.Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range attackers {
		if err := g.DeclareAttacker(id, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
}

// trekB18ACastIdentity casts Test of Identity from the active seat's hand
// with two Mountains to pay for it, strictly, and returns the cast error.
func trekB18ACastIdentity(t *testing.T, g *game.Game, modes []int, targets []game.TargetRef) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
	pushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Test of Identity", TypeLine: "Instant", ManaCost: "{1}{R}",
		OracleID: trekB18AIdentityOracle, Owner: me.ID, Controller: me.ID,
	})
	return g.CastSpell(me.ID, id, game.CastSpellParams{Modes: modes, Targets: targets, Strict: true, AutoTap: true})
}

// --- Silicate Surveyor -----------------------------------------------------------

func TestTrekB18SilicateSurveyorExilesTheTopCardToPlayAndHasReach(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
	)
	surveyor := castHoldCreature(t, g, "Silicate Surveyor", "Creature — Tholian Scout", trekB18ASurveyorOracle, 4, 4)
	if !g.Exile.Contains(ids[0]) {
		t.Fatal("the top card was not exiled")
	}
	if !me.Library.Contains(ids[1]) {
		t.Error("the second card left the library")
	}
	if !hasEffectiveKeyword(t, g, surveyor, "reach") {
		t.Error("Silicate Surveyor has no reach")
	}
	if err := g.CastSpell(me.ID, ids[0], game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("the exiled card is not playable this turn: %v", err)
	}
}

// --- Starfleet Security ----------------------------------------------------------

func TestTrekB18StarfleetSecurityInvestigatesWhenItDiesDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	security := b12Push(g, me.ID, "Starfleet Security", "Creature — Human Officer", trekB18ASecurityOracle, 1, 1)
	if !hasEffectiveKeyword(t, g, security, "haste") {
		t.Error("Starfleet Security has no haste")
	}
	b15Destroy(t, g, security)
	if n := trekB18AClues(g, me.ID); n != 1 {
		t.Errorf("%d Clues after it died during its controller's turn, want 1", n)
	}
}

func TestTrekB18StarfleetSecurityDoesNothingWhenItDiesDuringAnotherTurn(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	security := b12Push(g, opp.ID, "Starfleet Security", "Creature — Human Officer", trekB18ASecurityOracle, 1, 1)
	b15Destroy(t, g, security)
	if g.Battlefield.Contains(security) {
		t.Fatal("Starfleet Security did not die")
	}
	for _, p := range g.Seats {
		if n := trekB18AClues(g, p.ID); n != 0 {
			t.Errorf("seat %d has %d Clues: it died outside its controller's turn", p.Seat, n)
		}
	}
}

func TestTrekB18StarfleetSecurityLeavingWithoutDyingDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	security := b12Push(g, me.ID, "Starfleet Security", "Creature — Human Officer", trekB18ASecurityOracle, 1, 1)
	g.WithWriteLock(func() {
		_ = ExileTarget{Target: security}.Apply(NewContext(g, &game.StackItem{Controller: me.ID}))
	})
	passPriorityAroundTable(t, g)
	if n := trekB18AClues(g, me.ID); n != 0 {
		t.Errorf("%d Clues after an exile, want 0: exile is not dying", n)
	}
}

// --- Tal Shiar Analyst -----------------------------------------------------------

func TestTrekB18TalShiarAnalystPingsAndUntapsOnNoncreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	analyst := b12Push(g, me.ID, "Tal Shiar Analyst", "Creature — Romulan Scientist", trekB18AAnalystOracle, 3, 2)
	advanceToMain(t, g)
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	if err := g.ActivateCatalogAbility(me.ID, analyst, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := before[p.ID] - 1
		if p.ID == me.ID {
			want = before[p.ID]
		}
		if p.Life != want {
			t.Errorf("seat %d life %d → %d, want %d", p.Seat, before[p.ID], p.Life, want)
		}
	}
	if !b13Tapped(t, g, analyst) {
		t.Fatal("the ping taps the Analyst")
	}
	castCatalogSpell(t, g, "Mind Stone", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if b13Tapped(t, g, analyst) {
		t.Error("a noncreature (artifact) spell did not untap the Analyst")
	}
	if err := g.ActivateCatalogAbility(me.ID, analyst, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("second activation: %v", err)
	}
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if !b13Tapped(t, g, analyst) {
		t.Error("a creature spell untapped the Analyst")
	}
}

func TestTrekB18TalShiarAnalystIgnoresAnOpponentsSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	analyst := b12Push(g, me.ID, "Tal Shiar Analyst", "Creature — Romulan Scientist", trekB18AAnalystOracle, 3, 2)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, analyst, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	batch01OpponentCasts(t, g, opp, "Opt", "", "", nil)
	passPriorityAroundTable(t, g)
	if !b13Tapped(t, g, analyst) {
		t.Error("an opponent's noncreature spell untapped the Analyst")
	}
}

func TestTrekB18TalShiarAnalystCantPingWhileSummoningSick(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	analyst := pushCatalogPermanent(g, me.ID, "Tal Shiar Analyst", "Creature — Romulan Scientist", trekB18AAnalystOracle, true)
	advanceToMain(t, g)
	life := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].Life
	if err := g.ActivateCatalogAbility(me.ID, analyst, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("a summoning-sick Analyst activated its {T} ability")
	}
	passPriorityAroundTable(t, g)
	if got := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].Life; got != life {
		t.Errorf("opponent life %d → %d after a refused activation", life, got)
	}
}

// --- Targ Wrangler ---------------------------------------------------------------

// trekB18AWranglerBoard seeds a Targ Wrangler, a nonland artifact and a
// land for the active seat.
func trekB18AWranglerBoard(g *game.Game) (me *game.Player, wrangler, relic, land uuid.UUID) {
	me = g.Seats[g.Turn.ActiveSeat]
	wrangler = b12Push(g, me.ID, "Targ Wrangler", "Creature — Klingon Warrior", trekB18AWranglerOracle, 5, 3)
	relic = b12Permanent(g, me.ID, "Relic", "Artifact")
	land = pushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
	return me, wrangler, relic, land
}

func TestTrekB18TargWranglerMakesYouSacrificeANonlandPermanentAfterAttacking(t *testing.T) {
	g := newCatalogGame(t)
	me, wrangler, relic, land := trekB18AWranglerBoard(g)
	if !hasEffectiveKeyword(t, g, wrangler, "menace") {
		t.Error("Targ Wrangler has no menace")
	}
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	trekB18AAttack(t, g, opp.ID, wrangler)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	prompt := sacrificeChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no sacrifice prompt at the end step after attacking")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range prompt.SacrificeOptions {
		offered[id] = true
	}
	if !offered[wrangler] || !offered[relic] || offered[land] {
		t.Errorf("offered %v: want the Wrangler and the artifact, not the land", prompt.SacrificeOptions)
	}
	answerSacrifice(t, g, me.ID, relic)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) {
		t.Error("the chosen artifact was not sacrificed")
	}
	if !g.Battlefield.Contains(wrangler) || !g.Battlefield.Contains(land) {
		t.Error("something other than the chosen permanent left")
	}
}

func TestTrekB18TargWranglerMaySacrificeItself(t *testing.T) {
	g := newCatalogGame(t)
	me, wrangler, relic, _ := trekB18AWranglerBoard(g)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	trekB18AAttack(t, g, opp.ID, wrangler)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, me.ID, wrangler)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wrangler) || !me.Graveyard.Contains(wrangler) {
		t.Error("the Wrangler was not sacrificed")
	}
	if !g.Battlefield.Contains(relic) {
		t.Error("the artifact left although the Wrangler was chosen")
	}
}

func TestTrekB18TargWranglerDoesNothingIfItDidNotAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, wrangler, relic, _ := trekB18AWranglerBoard(g)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if p := sacrificeChoiceFor(g, me.ID); p != nil {
		t.Fatal("a sacrifice was asked although the Wrangler did not attack")
	}
	if !g.Battlefield.Contains(wrangler) || !g.Battlefield.Contains(relic) {
		t.Error("a permanent left although the Wrangler did not attack")
	}
}

func TestTrekB18TargWranglerSacrificeStandsAfterItLeft(t *testing.T) {
	// The intervening "if" is checked again as the trigger resolves,
	// against the object as it last existed (CR 603.4, 608.2h): a
	// Wrangler killed in response to its own trigger still attacked this
	// turn, so its controller still sacrifices.
	g := newCatalogGame(t)
	me, wrangler, relic, _ := trekB18AWranglerBoard(g)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	trekB18AAttack(t, g, opp.ID, wrangler)
	advanceTo(t, g, game.StepEnd)
	// Kill it in response to its own trigger: the trigger still resolves.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(wrangler) })
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, me.ID, relic)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) {
		t.Error("the trigger did not resolve after its source left: the Wrangler still attacked this turn")
	}
}

// --- Test of Identity ------------------------------------------------------------

func TestTrekB18TestOfIdentityPumpsAKlingonWithDoubleStrike(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	klingon := b12Creature(g, me.ID, "Warrior", "Creature — Klingon Warrior", 2, 2)
	if err := trekB18ACastIdentity(t, g, []int{0}, cardRefs(klingon)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, klingon); p != 3 {
		t.Errorf("power %d, want 3", p)
	}
	if !hasEffectiveKeyword(t, g, klingon, "double strike") {
		t.Error("the Klingon did not gain double strike")
	}
}

func TestTrekB18TestOfIdentityCountsAChangelingAsAKlingon(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	shifter := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Shifter", TypeLine: "Creature — Shapeshifter",
		Keywords: []string{game.KeywordChangeling}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	if err := trekB18ACastIdentity(t, g, []int{0}, cardRefs(shifter)); err != nil {
		t.Fatalf("a changeling is a Klingon: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, shifter, "double strike") {
		t.Error("the changeling did not gain double strike")
	}
}

func TestTrekB18TestOfIdentityRefusesANonKlingon(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	human := b12Creature(g, me.ID, "Ensign", "Creature — Human Officer", 2, 2)
	if err := trekB18ACastIdentity(t, g, []int{0}, cardRefs(human)); err == nil {
		t.Fatal("Test of Identity targeted a non-Klingon")
	}
}

func TestTrekB18TestOfIdentityDiscardsToDrawTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	if err := trekB18ACastIdentity(t, g, []int{1}, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	hand, library := me.Hand.Size(), me.Library.Size()
	passPriorityAroundTable(t, g)
	prompt := discardChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no discard prompt")
	}
	if prompt.ChooseMin != 1 || prompt.ChooseMax != 1 {
		t.Errorf("discard bounds [%d,%d], want exactly one: the discard is mandatory", prompt.ChooseMin, prompt.ChooseMax)
	}
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d, want one discarded and two drawn", hand, got)
	}
	if got := library - me.Library.Size(); got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
}

func TestTrekB18TestOfIdentityDrawsNothingWithAnEmptyHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	me.Hand.Cards = nil
	if err := trekB18ACastIdentity(t, g, []int{1}, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	library := me.Library.Size()
	passPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != library {
		t.Errorf("drew %d cards with nothing to discard, want 0", library-got)
	}
	if me.Hand.Size() != 0 {
		t.Errorf("hand has %d cards, want 0", me.Hand.Size())
	}
}
