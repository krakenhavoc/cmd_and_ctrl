package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// slc_slz_mps_singles_test.go — Beta Ray Bill and Doctor Strange, Master
// of Magic (Secret Lair Countdown, #5463), Clown Car (The Zeta Set, #5461)
// and Paradox Engine (Kaladesh Inventions, #5538), driven through the
// engine.

const (
	s3BetaRayBillOracle   = "ebb58466-dca1-4a5a-83fe-f48d75063cbe"
	s3DoctorStrangeOracle = "554ac5df-b49c-4041-963d-408160f59fc1"
	s3ClownCarOracle      = "e0a6947f-3d52-45e0-8fd7-0bc07cc8c784"
	s3ParadoxEngineOracle = "d38937a6-ab1c-4de8-a278-d81a6ce0d921"
)

func TestS3SinglesAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s3BetaRayBillOracle:   "Beta Ray Bill",
		s3DoctorStrangeOracle: "Doctor Strange, Master of Magic",
		s3ClownCarOracle:      "Clown Car",
		s3ParadoxEngineOracle: "Paradox Engine",
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

// --- Beta Ray Bill ------------------------------------------------------------

// s3BillTable is Beta Ray Bill under the active player, in their main
// phase.
func s3BillTable(t *testing.T) (valiantTable, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	seat := g.Turn.ActiveSeat
	tb := valiantTable{g: g, seat: seat, me: g.Seats[seat], them: g.Seats[(seat+1)%len(g.Seats)]}
	tb.hero = b12Push(g, tb.me.ID, "Beta Ray Bill", "Legendary Creature — Alien Warrior Hero", s3BetaRayBillOracle, 2, 3)
	return tb, tb.hero
}

// Valiant with a God you control: Bill and the chosen God each get a
// +1/+1 counter. An opponent's God is not a legal pick, and nothing else
// gets a counter.
func TestS3BetaRayBillPutsACounterOnHimAndYourGod(t *testing.T) {
	tb, bill := s3BillTable(t)
	god := b12Creature(tb.g, tb.me.ID, "Test God", "Legendary Enchantment Creature — God", 4, 4)
	theirGod := b12Creature(tb.g, tb.them.ID, "Their God", "Legendary Enchantment Creature — God", 4, 4)
	bear := b12Creature(tb.g, tb.me.ID, "Bear", "Creature — Bear", 2, 2)
	valiantTarget(t, tb.g, tb.me, bill)
	ch := pickTargetPrompt(t, tb.g)
	if err := tb.g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: theirGod}); err == nil {
		t.Fatal("an opponent's God was accepted as \"target God you control\"")
	}
	if err := tb.g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
		t.Fatal("a non-God was accepted as \"target God you control\"")
	}
	answerPickTarget(t, tb.g, god)
	passPriorityAroundTable(t, tb.g)
	if n := plusOneCounters(tb.g, bill); n != 1 {
		t.Errorf("Beta Ray Bill has %d +1/+1 counters, want 1", n)
	}
	if n := plusOneCounters(tb.g, god); n != 1 {
		t.Errorf("your God has %d +1/+1 counters, want 1", n)
	}
	if n := plusOneCounters(tb.g, theirGod) + plusOneCounters(tb.g, bear); n != 0 {
		t.Errorf("%d counters went on permanents that were not chosen", n)
	}
	// Only the first time each turn.
	valiantTarget(t, tb.g, tb.me, bill)
	if n := valiantPickTargetPrompts(tb.g) + valiantTriggersFrom(tb.g, bill); n != 0 {
		t.Fatalf("the second targeting this turn triggered valiant %d times", n)
	}
	passPriorityAroundTable(t, tb.g)
	if n := plusOneCounters(tb.g, bill); n != 1 {
		t.Errorf("Beta Ray Bill has %d +1/+1 counters after a second targeting this turn, want 1", n)
	}
}

// "Up to one": with no God, the trigger still resolves and Bill gets his
// counter.
func TestS3BetaRayBillWithNoGodStillGetsHisCounter(t *testing.T) {
	tb, bill := s3BillTable(t)
	valiantTarget(t, tb.g, tb.me, bill)
	if n := valiantPickTargetPrompts(tb.g); n != 0 {
		// A prompt with no candidates may still ask; answer it with none.
		ch := pickTargetPrompt(t, tb.g)
		if err := tb.g.ResolvePickTargets(ch.ID, ch.Chooser, nil); err != nil {
			t.Fatalf("choose no God: %v", err)
		}
	}
	passPriorityAroundTable(t, tb.g)
	if n := plusOneCounters(tb.g, bill); n != 1 {
		t.Errorf("Beta Ray Bill has %d +1/+1 counters, want 1", n)
	}
}

// Choosing no God when one is there: Bill alone gets the counter.
func TestS3BetaRayBillMayChooseNoGod(t *testing.T) {
	tb, bill := s3BillTable(t)
	god := b12Creature(tb.g, tb.me.ID, "Test God", "Legendary Enchantment Creature — God", 4, 4)
	valiantTarget(t, tb.g, tb.me, bill)
	ch := pickTargetPrompt(t, tb.g)
	if err := tb.g.ResolvePickTargets(ch.ID, ch.Chooser, nil); err != nil {
		t.Fatalf("choose no God: %v", err)
	}
	passPriorityAroundTable(t, tb.g)
	if n := plusOneCounters(tb.g, bill); n != 1 {
		t.Errorf("Beta Ray Bill has %d +1/+1 counters, want 1", n)
	}
	if n := plusOneCounters(tb.g, god); n != 0 {
		t.Errorf("the God nobody chose has %d +1/+1 counters", n)
	}
}

// An opponent targeting Bill does not trigger valiant.
func TestS3BetaRayBillIgnoresAnOpponentsSpell(t *testing.T) {
	tb, bill := s3BillTable(t)
	valiantTarget(t, tb.g, tb.them, bill)
	if n := valiantPickTargetPrompts(tb.g) + valiantTriggersFrom(tb.g, bill); n != 0 {
		t.Fatalf("an opponent's spell triggered valiant %d times", n)
	}
	passPriorityAroundTable(t, tb.g)
	if n := plusOneCounters(tb.g, bill); n != 0 {
		t.Errorf("Beta Ray Bill has %d +1/+1 counters from an opponent's spell", n)
	}
}

// --- Doctor Strange, Master of Magic -----------------------------------------

func TestS3DoctorStrangeDrawsOnlyAfterANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	strange := b21Push(g, me.ID, "Doctor Strange, Master of Magic", "Legendary Creature — Human Sorcerer Hero", s3DoctorStrangeOracle, 2, 3)
	advanceToMain(t, g)
	for _, kw := range []string{"flying", "vigilance", "prowess"} {
		if !hasEffectiveKeyword(t, g, strange, kw) {
			t.Errorf("no %s", kw)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, strange, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the draw activated before any spell was cast")
	}
	// A creature spell does not turn it on.
	msc17Play(t, g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if err := g.ActivateCatalogAbility(me.ID, strange, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the draw activated after only a creature spell")
	}
	castCatalogSpell(t, g, "Filler", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	// Prowess: the instant pumped him.
	if p := cardByID(g, strange).CurrentPower(); p != 3 {
		t.Errorf("power %d after a noncreature spell, want 3 (prowess)", p)
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, strange, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate after an instant: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, got)
	}
	if !tappedNow(g, strange) {
		t.Error("Doctor Strange did not tap for the draw")
	}
	if err := g.ActivateCatalogAbility(me.ID, strange, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("Doctor Strange activated again while tapped")
	}
}

// --- Clown Car ----------------------------------------------------------------

// s3CastClownCar casts Clown Car for X from the active player's hand,
// in their main phase, and does not resolve it.
func s3CastClownCar(t *testing.T, g *game.Game, x int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Clown Car", TypeLine: "Artifact — Vehicle", OracleID: s3ClownCarOracle, ManaCost: "{X}",
		Power: 1, Toughness: 1, Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{XValue: x}); err != nil {
		t.Fatalf("CastSpell Clown Car: %v", err)
	}
	return id
}

// s3ClownRobots counts the Clown Robot tokens `owner` controls, and fails
// on one that is not a 1/1 white artifact creature token.
func s3ClownRobots(t *testing.T, g *game.Game, owner uuid.UUID) int {
	t.Helper()
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller != owner || c.Name != "Clown Robot" {
			continue
		}
		n++
		if !c.IsToken() || !c.IsCreature() || !c.HasCardType("artifact") || !c.HasColor("W") ||
			c.CurrentPower() != 1 || c.CurrentToughness() != 1 || !c.HasSubtypeWord("Clown") || !c.HasSubtypeWord("Robot") {
			t.Errorf("Clown Robot %+v is not a 1/1 white Clown Robot artifact creature token", c)
		}
	}
	return n
}

// "Roll X six-sided dice. For each odd result, create a 1/1 white Clown
// Robot artifact creature token. For each even result, put a +1/+1
// counter on this Vehicle." The results are random, so each X is cast in
// a fresh game and the invariants are checked against the dice the game
// announced: X d6 from the Car, one token per odd result, one counter per
// even one, in one batch. Over the loop both kinds of result turn up.
func TestS3ClownCarRollsXD6IntoRobotsAndCounters(t *testing.T) {
	sawOdd, sawEven := false, false
	for x := 1; x <= 8; x++ {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		from := len(g.Events)
		car := s3CastClownCar(t, g, x)
		if got := s5682Rolls(t, g, from, me.ID, car, 6); len(got) != 0 {
			t.Fatalf("X=%d: rolled %v while the spell was on the stack", x, got)
		}
		passPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(car) {
			t.Fatalf("X=%d: Clown Car is not on the battlefield", x)
		}
		rolls := s5682Rolls(t, g, from, me.ID, car, 6)
		if len(rolls) != x {
			t.Fatalf("X=%d: rolled %v, want %d d6", x, rolls, x)
		}
		odd, even := 0, 0
		for _, r := range rolls {
			if r%2 == 1 {
				odd++
			} else {
				even++
			}
		}
		sawOdd = sawOdd || odd > 0
		sawEven = sawEven || even > 0
		robots := s3ClownRobots(t, g, me.ID)
		counters := plusOneCounters(g, car)
		if robots != odd || counters != even {
			t.Errorf("X=%d: rolled %v, got %d Clown Robots and %d counters; want %d and %d", x, rolls, robots, counters, odd, even)
		}
		if robots+counters != x {
			t.Errorf("X=%d: %d Robots plus %d counters, want X in all", x, robots, counters)
		}
		batches := map[uint64]bool{}
		g.WithWriteLock(func() {
			for _, ev := range g.Events[from:] {
				if ev.Kind == game.EventRollDie {
					batches[ev.BatchSeq] = true
				}
			}
		})
		if len(batches) != 1 {
			t.Errorf("X=%d: the dice were announced in %d batches, want one", x, len(batches))
		}
	}
	if !sawOdd || !sawEven {
		t.Errorf("over X=1..8, saw an odd result %v and an even result %v; want both", sawOdd, sawEven)
	}
}

// Cast for X=0 it rolls nothing: a plain 1/1 Vehicle with no counters and
// no Robots.
func TestS3ClownCarCastForZeroRollsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	from := len(g.Events)
	car := s3CastClownCar(t, g, 0)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(car) {
		t.Fatal("Clown Car is not on the battlefield")
	}
	if got := s5682Rolls(t, g, from, me.ID, car, 6); len(got) != 0 {
		t.Errorf("rolled %v for X=0", got)
	}
	if n := s3ClownRobots(t, g, me.ID); n != 0 {
		t.Errorf("%d Clown Robots for X=0", n)
	}
	if n := plusOneCounters(g, car); n != 0 {
		t.Errorf("%d counters for X=0", n)
	}
}

// Crew 2: tapping a 2-power creature makes the Car an artifact creature.
func TestS3ClownCarCrewTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	car := s3CastClownCar(t, g, 0)
	passPriorityAroundTable(t, g)
	small := b12Creature(g, me.ID, "Squire", "Creature — Human", 1, 1)
	big := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if cardByID(g, car).IsCreature() {
		t.Fatal("Clown Car is a creature before it is crewed")
	}
	if err := g.ActivateCatalogAbility(me.ID, car, 0, game.ActivateAbilityParams{Strict: true, CrewIDs: []uuid.UUID{small}}); err == nil {
		t.Fatal("crewed with total power 1")
	}
	if err := g.ActivateCatalogAbility(me.ID, car, 0, game.ActivateAbilityParams{Strict: true, CrewIDs: []uuid.UUID{big}}); err != nil {
		t.Fatalf("crew with power 2: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !cardByID(g, car).IsCreature() {
		t.Error("Clown Car is not a creature after it was crewed")
	}
	if !tappedNow(g, big) {
		t.Error("the crew did not tap")
	}
}

// --- Paradox Engine ---------------------------------------------------------

func s3HasTriggerFrom(g *game.Game, source uuid.UUID) bool {
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered && item.SourceCardID == source {
			return true
		}
	}
	return false
}

// "Whenever you cast a spell, untap all nonland permanents you control."
// The trigger goes on the stack above the spell and resolves while the
// spell is still there; it untaps the controller's nonland permanents,
// the Engine included, and leaves lands and other players' permanents
// tapped.
func TestS3ParadoxEngineUntapsYourNonlandPermanentsOnCast(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	them := opponentOf(g)
	engine := b21Push(g, me.ID, "Paradox Engine", "Legendary Artifact", s3ParadoxEngineOracle, 0, 0)
	rock := b12Permanent(g, me.ID, "Mind Stone", "Artifact")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	land := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := b12Creature(g, them.ID, "Their Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{engine, rock, bear, land, theirs} {
		tapCard(g, id)
	}
	spell := castCatalogSpell(t, g, "Filler", "Instant", "", nil)
	if !s3HasTriggerFrom(g, engine) {
		t.Fatal("casting a spell put no Paradox Engine trigger on the stack")
	}
	if !tappedNow(g, bear) {
		t.Fatal("the untap happened before the trigger resolved")
	}
	resolveTopOfStack(t, g)
	if !g.Stack.Contains(spell) {
		t.Fatal("the spell left the stack before the trigger resolved")
	}
	for name, id := range map[string]uuid.UUID{"Paradox Engine": engine, "your artifact": rock, "your creature": bear} {
		if tappedNow(g, id) {
			t.Errorf("%s is still tapped", name)
		}
	}
	if !tappedNow(g, land) {
		t.Error("your land was untapped")
	}
	if !tappedNow(g, theirs) {
		t.Error("an opponent's creature was untapped")
	}
	passPriorityAroundTable(t, g)
}

// An opponent's spell does not trigger it.
func TestS3ParadoxEngineIgnoresAnOpponentsSpell(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	them := opponentOf(g)
	engine := b21Push(g, me.ID, "Paradox Engine", "Legendary Artifact", s3ParadoxEngineOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	tapCard(g, bear)
	valiantTarget(t, g, them)
	if s3HasTriggerFrom(g, engine) {
		t.Fatal("an opponent's spell triggered Paradox Engine")
	}
	passPriorityAroundTable(t, g)
	if !tappedNow(g, bear) {
		t.Error("an opponent's spell untapped your creature")
	}
}
