package effects

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// the_ring_test.go — ADR 0114 PR 2: the tempt, the Ring emblem's first
// line and the Ring-bearer (CR 701.54). No printed card tempts yet (PR
// 3 lands the first ones together with the emblem's other three
// lines), so these drive the keyword action directly and through
// test-only sources: a permanent with "{0}: The Ring tempts you" and
// watchers built from the trigger constructors.

const (
	ringTempterOracle = "test-ring-tempter"
	ringWatcherOracle = "test-ring-watcher"
)

// ringTempt has the Ring tempt `player` directly, as a resolving effect
// would, and reports what the continuation was handed (uuid.Nil until
// it runs — it does not run while a ring_bearer prompt is open).
func ringTempt(t *testing.T, g *game.Game, player uuid.UUID) *uuid.UUID {
	t.Helper()
	got := new(uuid.UUID)
	*got = uuid.Max
	g.WithWriteLock(func() {
		if err := g.RingTemptsForEffect(player, uuid.Nil, func(_ *game.Game, rb uuid.UUID) error {
			*got = rb
			return nil
		}); err != nil {
			t.Fatalf("RingTemptsForEffect: %v", err)
		}
	})
	return got
}

// ringPrompt is the open ring_bearer prompt `player` owes, or nil.
func ringPrompt(g *game.Game, player uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceRingBearer && c.Chooser == player {
			return c
		}
	}
	return nil
}

// answerRing answers `player`'s open ring_bearer prompt with `pick`.
func answerRing(t *testing.T, g *game.Game, player, pick uuid.UUID) {
	t.Helper()
	c := ringPrompt(g, player)
	if c == nil {
		t.Fatal("no ring_bearer prompt is open")
	}
	if err := g.ResolveRingBearer(c.ID, player, []uuid.UUID{pick}); err != nil {
		t.Fatalf("ResolveRingBearer: %v", err)
	}
}

// ringBearerOf reads `player`'s Ring-bearer under the lock.
func ringBearerOf(g *game.Game, player uuid.UUID) uuid.UUID {
	var id uuid.UUID
	g.ReadSnapshot(func() { id = game.RingBearerOf(g, player) })
	return id
}

func ringCount(g *game.Game, player uuid.UUID) int {
	var n int
	g.ReadSnapshot(func() { n = game.RingTemptCount(g, player) })
	return n
}

// ringEmblems is how many Ring emblems `p` has.
func ringEmblems(p *game.Player) int {
	n := 0
	if p.Emblems == nil {
		return 0
	}
	for _, c := range p.Emblems.Cards {
		if c.IsRingEmblem() {
			n++
		}
	}
	return n
}

func ringBFCard(t *testing.T, g *game.Game, id uuid.UUID) *game.Card {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	t.Fatalf("%s is not on the battlefield", id)
	return nil
}

func ringEvents(g *game.Game) []game.Event { return eventsOfKind(g, game.EventRingTempted) }

// --- §1: the emblem, the count, the event --------------------------

// The emblem is created once, at the first temptation, BEFORE the
// choice (CR 701.54c): it is already in the command zone while the
// prompt is open.
func TestTheRingEmblemIsCreatedOnceBeforeTheChoice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)

	got := ringTempt(t, g, me.ID)
	if ringPrompt(g, me.ID) == nil {
		t.Fatal("two creatures and no ring_bearer prompt")
	}
	if ringEmblems(me) != 1 || ringCount(g, me.ID) != 1 {
		t.Fatalf("while the prompt is open: %d emblems at count %d, want one at 1", ringEmblems(me), ringCount(g, me.ID))
	}
	if *got != uuid.Max {
		t.Fatal("the rest of the sentence ran before the choice")
	}
	if len(ringEvents(g)) != 0 {
		t.Fatal("the tempt was announced before the choice")
	}
	answerRing(t, g, me.ID, a)
	if *got != a {
		t.Fatalf("the continuation was handed %v, want the chosen creature", *got)
	}

	ringTempt(t, g, me.ID)
	answerRing(t, g, me.ID, a)
	if ringEmblems(me) != 1 {
		t.Fatalf("%d Ring emblems after two temptations; each player has one", ringEmblems(me))
	}
	if ringCount(g, me.ID) != 2 {
		t.Fatalf("count %d, want 2", ringCount(g, me.ID))
	}
}

// The count rises on every temptation, including one with no creature
// to choose, and the event fires either way (CR 701.54d) with the new
// count and the creature chosen (or none).
func TestTheRingTemptsEvenWithNoCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]

	got := ringTempt(t, g, me.ID)
	if *got != uuid.Nil {
		t.Fatalf("with no creature the continuation got %v, want uuid.Nil", *got)
	}
	if ringPrompt(g, me.ID) != nil {
		t.Fatal("a prompt with no candidates")
	}
	evs := ringEvents(g)
	if len(evs) != 1 || evs[0].Actor != me.ID || evs[0].CardID != uuid.Nil || evs[0].Amount != 1 {
		t.Fatalf("events %+v, want one for the player, no creature, count 1", evs)
	}

	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	ringTempt(t, g, me.ID)
	evs = ringEvents(g)
	if len(evs) != 2 || evs[1].CardID != bear || evs[1].Amount != 2 || evs[1].Label != game.RingTemptedForced {
		t.Fatalf("second event %+v, want the only creature, count 2, forced", evs[len(evs)-1])
	}
	if ringBearerOf(g, me.ID) != bear {
		t.Fatal("the only creature is not the Ring-bearer")
	}
}

// Owner decision 2: one creature is chosen automatically, and the log
// says so. Two are asked. None says so too.
func TestTheRingLogLines(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ringTempt(t, g, me.ID)
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	ringTempt(t, g, me.ID)
	giant := b12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	ringTempt(t, g, me.ID)
	answerRing(t, g, me.ID, giant)
	_ = bear

	var lines []string
	for _, e := range protocol.ViewOfGame(g).Log {
		if e.Kind == protocol.LogRingTempted {
			lines = append(lines, e.Text)
		}
	}
	want := []string{
		"The Ring tempts A (1) — A controls no creature",
		"The Ring tempts A (2) — A's only creature, Grizzly Bears, becomes their Ring-bearer",
		"The Ring tempts A (3) — A chooses Hill Giant as their Ring-bearer",
	}
	if strings.Join(lines, "\n") != strings.Join(want, "\n") {
		t.Fatalf("log lines:\n%s\nwant:\n%s", strings.Join(lines, "\n"), strings.Join(want, "\n"))
	}
}

// Re-choosing the creature that already is the Ring-bearer still counts
// as choosing it (2023-06-16 ruling): the event fires again with it.
func TestTheRingRechoosingTheSameBearerFiresAgain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	ringTempt(t, g, me.ID)
	answerRing(t, g, me.ID, a)
	ringTempt(t, g, me.ID)
	answerRing(t, g, me.ID, a)
	evs := ringEvents(g)
	if len(evs) != 2 || evs[0].CardID != a || evs[1].CardID != a {
		t.Fatalf("events %+v, want two naming the same creature", evs)
	}
	if ringBearerOf(g, me.ID) != a {
		t.Fatal("the creature stopped being the Ring-bearer")
	}
}

// --- §3: the designation ------------------------------------------

// A second Ring-bearer clears the first: one per player.
func TestTheRingASecondBearerClearsTheFirst(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	ringTempt(t, g, me.ID)
	answerRing(t, g, me.ID, a)
	ringTempt(t, g, me.ID)
	answerRing(t, g, me.ID, b)
	if ringBFCard(t, g, a).RingBearer {
		t.Error("the first Ring-bearer kept the designation")
	}
	if ringBearerOf(g, me.ID) != b {
		t.Error("the second creature is not the Ring-bearer")
	}
}

// CR 701.54a: "until … another player gains control of it". The thief's
// own later temptation can choose it.
func TestTheRingAControlChangeEndsTheDesignation(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	ringTempt(t, g, me.ID)
	if ringBearerOf(g, me.ID) != bear {
		t.Fatal("setup: the only creature is not the Ring-bearer")
	}

	g.WithWriteLock(func() {
		if err := (GainControl{Target: bear, Controller: opp.ID}).Apply(ctxFor(g, &game.StackItem{Controller: opp.ID})); err != nil {
			t.Fatalf("GainControl: %v", err)
		}
		g.RecomputeLayersIfStaleLocked()
	})
	if c := ringBFCard(t, g, bear); c.Controller != opp.ID || c.RingBearer {
		t.Fatalf("after the theft: controller %v, ring-bearer %v — want the thief's and no designation", c.Controller, c.RingBearer)
	}
	if ringBearerOf(g, me.ID) != uuid.Nil || ringBearerOf(g, opp.ID) != uuid.Nil {
		t.Fatal("somebody still has a Ring-bearer")
	}
	if c := ringBFCard(t, g, bear); c.IsLegendary() {
		t.Error("the stolen creature is still legendary")
	}

	ringTempt(t, g, opp.ID)
	if ringBearerOf(g, opp.ID) != bear {
		t.Fatal("the thief's own temptation did not choose the stolen creature")
	}
}

// CR 400.7: a bounced or flickered Ring-bearer is a new object with no
// designation.
func TestTheRingBounceAndFlickerEndTheDesignation(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	ringTempt(t, g, me.ID)

	g.WithWriteLock(func() {
		if err := (Flicker{Target: bear}).Apply(ctxFor(g, &game.StackItem{Controller: me.ID, SourceCardID: uuid.New()})); err != nil {
			t.Fatalf("Flicker: %v", err)
		}
	})
	if ringBearerOf(g, me.ID) != uuid.Nil {
		t.Fatal("a flickered Ring-bearer came back still carrying the Ring")
	}

	giant := b12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	// Two creatures now (the flickered Bears and the Giant).
	ringTempt(t, g, me.ID)
	answerRing(t, g, me.ID, giant)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, giant); err != nil {
		t.Fatalf("bounce: %v", err)
	}
	if ringBearerOf(g, me.ID) != uuid.Nil {
		t.Fatal("a bounced Ring-bearer left the designation behind")
	}
	for _, c := range me.Hand.Cards {
		if c.RingBearer {
			t.Fatalf("%s carries the designation in hand", c.Name)
		}
	}
}

// CR 702.26b / 702.26d: a phased-out Ring-bearer is nobody's, keeps the
// designation, and is back when it phases in. A new choice while it is
// phased out does clear it.
func TestTheRingPhasedOutBearerIsNobodysAndComesBack(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	ringTempt(t, g, opp.ID)

	g.WithWriteLock(func() {
		if err := g.PhaseOutForEffect(uuid.Nil, bear); err != nil {
			t.Fatalf("phase out: %v", err)
		}
	})
	if ringBearerOf(g, opp.ID) != uuid.Nil {
		t.Fatal("a phased-out permanent is somebody's Ring-bearer")
	}
	passTurnsTo(t, g, 1)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("setup: the Bears did not phase in")
	}
	if ringBearerOf(g, opp.ID) != bear {
		t.Fatal("the Ring-bearer did not come back with the designation")
	}

	// Phase it out again, and tempt with another creature out: the old
	// bearer stops being one where it is.
	g.WithWriteLock(func() {
		if err := g.PhaseOutForEffect(uuid.Nil, bear); err != nil {
			t.Fatalf("phase out: %v", err)
		}
	})
	giant := b12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	ringTempt(t, g, opp.ID)
	if ringBearerOf(g, opp.ID) != giant {
		t.Fatal("the only phased-in creature was not chosen")
	}
	passTurnsTo(t, g, 1)
	if ringBearerOf(g, opp.ID) != giant {
		t.Fatal("the old Ring-bearer phased in still carrying the designation")
	}
}

// CR 701.54b: being a Ring-bearer is not copiable. A Clone of one is
// neither a Ring-bearer nor legendary.
func TestTheRingACloneOfTheBearerIsNotOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	ringTempt(t, g, me.ID)

	clone := castCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", oracleClone, nil)
	resolveWithCopyChoice(t, g, bear)
	var cloneID uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID != bear && c.Name == "Grizzly Bears" {
			cloneID = c.InstanceID
		}
	}
	if cloneID == uuid.Nil {
		t.Fatalf("setup: Clone %s did not enter as a copy of the Bears", clone)
	}
	c := ringBFCard(t, g, cloneID)
	if c.RingBearer {
		t.Error("the Clone copied the designation")
	}
	if c.IsLegendary() {
		t.Error("the Clone copied the Ring's legendary, which is a layer 4 effect, not a copiable value")
	}
	if ringBearerOf(g, me.ID) != bear {
		t.Error("the original stopped being the Ring-bearer")
	}
}

// --- the emblem's first line: legendary ---------------------------

// "Your Ring-bearer is legendary": a layer 4 effect, so the legend rule
// sees it. A same-named legendary under the same controller goes to the
// legend rule (CR 704.5j) — the printed outcome.
func TestTheRingMakesTheBearerLegendaryAndTheLegendRuleSeesIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := b12Creature(g, me.ID, "Nazgûl", "Creature — Zombie Wraith Knight", 1, 2)
	b := b12Creature(g, me.ID, "Nazgûl", "Creature — Zombie Wraith Knight", 1, 2)
	if ringBFCard(t, g, a).IsLegendary() {
		t.Fatal("setup: the creature is legendary before the Ring")
	}
	ringTempt(t, g, me.ID)
	answerRing(t, g, me.ID, a)
	if !ringBFCard(t, g, a).IsLegendary() {
		t.Fatal("the Ring-bearer is not legendary")
	}
	if ringBFCard(t, g, b).IsLegendary() {
		t.Fatal("the other creature became legendary too")
	}
	g.RunStateChecksForTest()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceLegendRule {
			t.Fatal("one legendary Nazgûl and one non-legendary one went to the legend rule")
		}
	}

	// Now a legendary one by print: the bearer and it share a name.
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	legend := pushBattlefieldCardWithTimestamp(g2, game.Card{
		InstanceID: uuid.New(), Name: "Gollum", TypeLine: "Legendary Creature — Halfling Horror",
		Power: 1, Toughness: 1, Owner: me2.ID, Controller: me2.ID,
	})
	plain := b12Creature(g2, me2.ID, "Gollum", "Creature — Halfling Horror", 1, 1)
	ringTempt(t, g2, me2.ID)
	answerRing(t, g2, me2.ID, plain)
	g2.RunStateChecksForTest()
	found := false
	for _, c := range g2.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceLegendRule {
			found = true
		}
	}
	if !found {
		t.Fatalf("two legendary Gollums under one controller (%s, %s) and no legend rule", legend, plain)
	}
}

// --- §5: the evasion ----------------------------------------------

// "can't be blocked by creatures with greater power": refused by the
// verb, withheld by the enumerator, equal power allowed, and the
// refusal names the Ring.
func TestTheRingBearerCantBeBlockedByGreaterPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Runeclaw Bear", "Creature — Bear", 2, 2)
	ringTempt(t, g, me.ID)
	answerRing(t, g, me.ID, bear)

	giant := b12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	twin := b12Creature(g, opp.ID, "Balduvian Bears", "Creature — Bear", 2, 2)
	brAttack(t, g, bear, other)

	offered := brOffered(t, g, opp.ID)
	if brOffers(offered, giant, bear) {
		t.Error("the enumerator offers a greater-power blocker on the Ring-bearer")
	}
	if !brOffers(offered, twin, bear) {
		t.Error("the enumerator withholds an equal-power blocker")
	}
	if !brOffers(offered, giant, other) {
		t.Error("the rule bound a creature that is not the Ring-bearer")
	}

	refused := brRefusal(t, g.DeclareBlocker(giant, bear), game.BlockReasonCantBeBlockedBy)
	var emblem uuid.UUID
	for _, c := range me.Emblems.Cards {
		if c.IsRingEmblem() {
			emblem = c.InstanceID
		}
	}
	if refused.Source != emblem {
		t.Errorf("the refusal's source is %s, want the Ring emblem %s", refused.Source, emblem)
	}
	if got := refused.Sentence(uuid.Nil); got != "Grizzly Bears can't be blocked by creatures with greater power (The Ring)." {
		t.Errorf("sentence %q", got)
	}
	if err := g.DeclareBlocker(twin, bear); err != nil {
		t.Fatalf("an equal-power creature blocks the Ring-bearer: %v", err)
	}
	if err := g.DeclareBlocker(giant, other); err != nil {
		t.Fatalf("the Giant blocks the other attacker: %v", err)
	}
}

// Each player's Ring binds their own Ring-bearer, and Locke, Treasure
// Hunter keeps its own evasion on the shared predicate.
func TestTheRingBindsOnlyItsOwnersBearerAndLockeStillEvades(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	locke := b12Push(g, me.ID, "Locke, Treasure Hunter", "Legendary Creature — Human Rogue", "a800bec4-bacc-43f9-a773-dd795959935b", 2, 3)
	// The opponent has a Ring and a Ring-bearer; mine has neither.
	b12Creature(g, opp.ID, "Elvish Mystic", "Creature — Elf", 1, 1)
	ringTempt(t, g, opp.ID)
	giant := b12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	brAttack(t, g, bear, locke)

	offered := brOffered(t, g, opp.ID)
	if !brOffers(offered, giant, bear) {
		t.Error("the opponent's Ring bound MY creature")
	}
	if brOffers(offered, giant, locke) {
		t.Error("Locke lost its own evasion")
	}
}

// Four players, four emblems, four Ring-bearers.
func TestTheRingEachPlayerHasTheirOwn(t *testing.T) {
	g := newCatalogGame(t)
	bearers := map[uuid.UUID]uuid.UUID{}
	for _, p := range g.Seats {
		bearers[p.ID] = b12Creature(g, p.ID, "Bear of "+p.Name, "Creature — Bear", 2, 2)
		ringTempt(t, g, p.ID)
	}
	ringTempt(t, g, g.Seats[2].ID)
	for i, p := range g.Seats {
		if ringEmblems(p) != 1 {
			t.Errorf("seat %d has %d Ring emblems", i, ringEmblems(p))
		}
		if got := ringBearerOf(g, p.ID); got != bearers[p.ID] {
			t.Errorf("seat %d's Ring-bearer is %v, want their own Bear", i, got)
		}
		want := 1
		if i == 2 {
			want = 2
		}
		if got := ringCount(g, p.ID); got != want {
			t.Errorf("seat %d's count is %d, want %d", i, got, want)
		}
		if !ringBFCard(t, g, bearers[p.ID]).IsLegendary() {
			t.Errorf("seat %d's Ring-bearer is not legendary", i)
		}
	}
}

// CR 800.4a: a player who leaves takes their emblem, and their open
// ring_bearer prompt goes with them. The tempting object was theirs, so
// its resolution ends with them (CR 800.4a: their objects on the stack
// cease to exist) — the frame is abandoned, not run on their behalf, and
// nothing is announced (CR 800.4d) — and the table is not left waiting
// on anything.
func TestTheRingLeavingTheGameDropsThePromptAndTakesTheEmblem(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	b12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	got := ringTempt(t, g, opp.ID)
	if ringPrompt(g, opp.ID) == nil {
		t.Fatal("setup: no prompt")
	}
	if err := g.Concede(opp.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if ringPrompt(g, opp.ID) != nil {
		t.Fatal("the departed player's prompt is still open")
	}
	if *got != uuid.Max {
		t.Fatalf("the departed player's frame ran with %v; it is abandoned with them", *got)
	}
	if ringEmblems(opp) != 0 {
		t.Fatal("the departed player kept their Ring")
	}
	if len(ringEvents(g)) != 0 {
		t.Fatal("a temptation was announced for a player who has left")
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompts left open after the departure", len(g.PendingChoices))
	}
}

// The prompt's dropDefault (ADR 0114 §4): when every candidate leaves
// the battlefield while it is open, the prompt is withdrawn and the
// tempt completes with nothing chosen (CR 701.54d) — the rest of the
// sentence runs and the temptation is announced.
func TestTheRingPromptWithNoCandidateLeftCompletesTheTempt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	got := ringTempt(t, g, me.ID)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, id); err != nil {
			t.Fatalf("move: %v", err)
		}
	}
	if ringPrompt(g, me.ID) != nil {
		t.Fatal("the prompt outlived its last candidate")
	}
	if *got != uuid.Nil {
		t.Fatalf("the continuation got %v, want it run with nothing chosen", *got)
	}
	evs := ringEvents(g)
	if len(evs) != 1 || evs[0].CardID != uuid.Nil || evs[0].Amount != 1 {
		t.Fatalf("events %+v, want one temptation with no creature", evs)
	}
}

// --- §1's card side: the primitive and the trigger constructors ---

// Through a test source's ability on the stack: TheRingTemptsYou, and
// the three trigger shapes. "Whenever the Ring tempts you" fires with
// no creature; "whenever you choose a creature as your Ring-bearer"
// does not, and does fire on re-choosing; the intervening "if you chose
// a creature other than ~" fires only for another creature.
func TestTheRingTriggerConstructors(t *testing.T) {
	var tempted, chose, other int
	registerForTest(t, Spec{
		OracleID: ringTempterOracle, Name: "Test Ring Tempter",
		Activated: []ActivatedAbility{{
			Label:  "The Ring tempts you.",
			Effect: Do(TheRingTemptsYou{}),
		}},
	})
	registerForTest(t, Spec{
		OracleID: ringWatcherOracle, Name: "Test Ring Watcher",
		Triggered: []game.TriggeredAbility{
			WheneverTheRingTemptsYou("Test Ring Watcher — tempted", func(*game.Game, *game.StackItem) error { tempted++; return nil }),
			WheneverYouChooseARingBearer("Test Ring Watcher — chose", func(*game.Game, *game.StackItem) error { chose++; return nil }),
			IfYouChoseAnotherRingBearer(WheneverTheRingTemptsYou("Test Ring Watcher — another", func(*game.Game, *game.StackItem) error { other++; return nil })),
		},
	})
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	src := pushCatalogPermanent(g, me.ID, "Test Ring Tempter", "Enchantment", ringTempterOracle, false)
	watcher := pushCatalogPermanent(g, me.ID, "Test Ring Watcher", "Enchantment", ringWatcherOracle, false)
	activate := func() {
		t.Helper()
		if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		// Settle: the watcher's triggers fire together, so their
		// controller orders them (CR 603.3b) in the offered order.
		for i := 0; i < 8; i++ {
			passPriorityAroundTable(t, g)
			var order *game.PendingChoice
			for _, c := range g.PendingChoices {
				if c != nil && c.Kind == game.PendingChoiceTriggerOrder {
					order = c
				}
			}
			if order == nil {
				break
			}
			if err := g.ResolveTriggerOrder(order.ID, order.Chooser, append([]uuid.UUID(nil), order.TriggerOrderIDs...)); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
		}
	}

	activate() // no creature
	if tempted != 1 || chose != 0 || other != 0 {
		t.Fatalf("no creature: tempted %d chose %d other %d, want 1 0 0", tempted, chose, other)
	}

	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	activate() // the only creature: chosen for me
	if tempted != 2 || chose != 1 || other != 1 {
		t.Fatalf("one creature: tempted %d chose %d other %d, want 2 1 1", tempted, chose, other)
	}
	activate() // re-chosen
	if chose != 2 {
		t.Fatalf("re-choosing the same creature: chose %d, want 2", chose)
	}
	if ringBearerOf(g, me.ID) != bear || ringCount(g, me.ID) != 3 {
		t.Fatalf("bearer %v count %d", ringBearerOf(g, me.ID), ringCount(g, me.ID))
	}
	_ = watcher
}

// --- §9: the wire --------------------------------------------------

func TestTheRingOnTheWire(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	giant := b12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	ringTempt(t, g, me.ID)

	view := protocol.ViewOfGameFor(g, me.ID.String())
	var pc *protocol.PendingChoiceView
	for i := range view.PendingChoices {
		if view.PendingChoices[i].Kind == string(game.PendingChoiceRingBearer) {
			pc = &view.PendingChoices[i]
		}
	}
	if pc == nil {
		t.Fatal("no ring_bearer entry in pending_choices")
	}
	if pc.Reason != "choose your Ring-bearer" || pc.ChooseMin != 1 || pc.ChooseMax != 1 || len(pc.Options) != 2 {
		t.Fatalf("prompt %q min %d max %d options %d", pc.Reason, pc.ChooseMin, pc.ChooseMax, len(pc.Options))
	}
	// The enumerator offers one move per creature, labelled.
	var labels []string
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Kind == legal.KindChoice {
			labels = append(labels, m.Label)
		}
	}
	if len(labels) != 2 || !strings.HasPrefix(labels[0], "choose your Ring-bearer: choose Ring-bearer ") {
		t.Fatalf("moves %q", labels)
	}
	answerRing(t, g, me.ID, giant)

	raw, err := json.Marshal(protocol.ViewOfGame(g))
	if err != nil {
		t.Fatal(err)
	}
	var wire struct {
		Seats []struct {
			Emblems []struct {
				Label string `json:"label"`
				Text  string `json:"text"`
				Level int    `json:"level"`
				Lines []struct {
					Text string `json:"text"`
					At   int    `json:"at"`
				} `json:"lines"`
			} `json:"emblems"`
		} `json:"seats"`
		Battlefield struct {
			Cards []struct {
				InstanceID string `json:"instance_id"`
				RingBearer bool   `json:"ring_bearer"`
				TypeLine   string `json:"type_line"`
			} `json:"cards"`
		} `json:"battlefield"`
	}
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatal(err)
	}
	em := wire.Seats[0].Emblems
	if len(em) != 1 || em[0].Label != "The Ring" || em[0].Level != 1 || em[0].Text != theRingLine1 ||
		len(em[0].Lines) != 1 || em[0].Lines[0].At != 1 {
		t.Fatalf("emblems %+v", em)
	}
	for _, c := range wire.Battlefield.Cards {
		switch c.InstanceID {
		case giant.String():
			if !c.RingBearer {
				t.Error("the Ring-bearer is not marked ring_bearer")
			}
		case bear.String():
			if c.RingBearer {
				t.Error("the other creature is marked ring_bearer")
			}
		}
	}
}

// --- §8: persistence ------------------------------------------------

// The emblem's count, the designation and the legendary grant survive a
// capture → JSON → restore, and the restored table tempts on from the
// count it had.
func TestTheRingSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	ringTempt(t, g, me.ID)
	ringTempt(t, g, me.ID)

	r := restoreRoundTrip(t, g, true)
	rme := r.Seats[0]
	if ringEmblems(rme) != 1 || ringCount(r, rme.ID) != 2 {
		t.Fatalf("restored: %d emblems at count %d, want one at 2", ringEmblems(rme), ringCount(r, rme.ID))
	}
	if ringBearerOf(r, rme.ID) != bear {
		t.Fatal("restored: the Ring-bearer was lost")
	}
	if !ringBFCard(t, r, bear).IsLegendary() {
		t.Fatal("restored: the Ring-bearer is not legendary — the emblem's static did not come back")
	}
	ringTempt(t, r, rme.ID)
	if ringCount(r, rme.ID) != 3 {
		t.Fatalf("restored table tempted to %d, want 3", ringCount(r, rme.ID))
	}
}

// A table sitting on an open ring_bearer prompt holds a Go continuation,
// so it is not a restore point for that moment (ADR 0114 §4: the
// existing posture for every resolution-time pick, counted under
// ChoiceResumeFrames). An undo snapshot (Clone) carries the prompt and
// answers it on the clone; once answered, the table restores.
func TestTheRingMidTemptPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	ringTempt(t, g, me.ID)

	snap := g.CaptureSnapshot()
	if snap.Restorable() || snap.Continuations.ChoiceResumeFrames != 1 {
		t.Fatalf("mid-tempt capture: restorable %v, census %+v — want it counted as one resume frame", snap.Restorable(), snap.Continuations)
	}

	// The undo path: a clone answers its own copy of the prompt.
	clone := g.Clone()
	answerRing(t, clone, clone.Seats[0].ID, b)
	if ringBearerOf(clone, me.ID) != b || ringCount(clone, me.ID) != 1 {
		t.Fatal("the clone did not finish the tempt")
	}
	if ringPrompt(g, me.ID) == nil || ringBearerOf(g, me.ID) != uuid.Nil {
		t.Fatal("answering the clone touched the original")
	}

	answerRing(t, g, me.ID, a)
	r := restoreRoundTrip(t, g, true)
	if ringBearerOf(r, me.ID) != a || ringCount(r, me.ID) != 1 {
		t.Fatalf("restored after the answer: bearer %v count %d", ringBearerOf(r, me.ID), ringCount(r, me.ID))
	}
}
