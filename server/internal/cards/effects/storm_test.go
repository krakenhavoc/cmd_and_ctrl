package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// storm_test.go — CR 702.40 end to end, on the five cards that print
// it. The engine half (TurnTally.Casts, the count, the announcement)
// is pinned in game/storm_test.go; this file is about what a player
// sees.
//
// Storm is the keyword where the obvious test is the wrong one. "N
// copies were created" is easy to assert and tells you almost
// nothing: the mistakes that matter are all about WHICH spells the N
// counted (an opponent's, a countered one, the storm spell itself, a
// copy) and about each copy being its own object with its own
// target and its own resolution. So every test below is either about
// the count's membership or about the copies being separate.

const (
	grapeshotOracle       = "ebd2d760-5ad8-4027-b124-171822f3edfe"
	brainFreezeOracle     = "464c0150-3dbc-403b-9ada-fef25ab1f29d"
	emptyTheWarrensOracle = "3a59c882-8bb8-49ba-862f-125020dd5bec"
	flusterstormOracle    = "86bf58f2-7f25-4e10-b797-25e0e8e67769"
	tendrilsOracle        = "78db3190-59bc-4033-95bc-32d8ed872b6a"
)

// stormFiller casts a plain non-catalog instant, which is what a
// storm count is made of: a spell that was cast and did nothing else
// worth asserting.
func stormFiller(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := castCatalogSpell(t, g, "Filler", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	return id
}

// castFromSeat puts a card in `seat`'s hand and casts it from there,
// whoever's turn it is. castCatalogSpell always casts from the ACTIVE
// seat, and half of storm is about the spells somebody ELSE cast.
func castFromSeat(t *testing.T, g *game.Game, seat *game.Player, name, typeLine, oracle string, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	id := uuid.New()
	seat.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: seat.ID, Controller: seat.ID,
	})
	if err := g.CastSpell(seat.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("cast %s from seat %s: %v", name, seat.ID, err)
	}
	return id
}

// answerCopyPrompts settles a storm resolution: pass priority until a
// re-target prompt opens, answer it with `target`, repeat, and stop
// when the stack is empty. `wantPrompts` is asserted, because "how
// many separate choices did the player get" IS the per-copy shape
// (ADR 0086 Decision 4) and a batched prompt would pass every
// life-total assertion in this file.
func answerCopyPrompts(t *testing.T, g *game.Game, chooser, target uuid.UUID, wantPrompts int) {
	t.Helper()
	seen := 0
	for i := 0; i < 64; i++ {
		if stackFullyEmpty(g) && latestPickTarget(g, chooser) == nil {
			break
		}
		if p := latestPickTarget(g, chooser); p != nil {
			seen++
			if err := g.ResolvePickTarget(p.ID, chooser,
				game.TargetRef{Kind: game.TargetPlayer, ID: target}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if seen != wantPrompts {
		t.Errorf("the copies opened %d re-target prompts, want %d (one per copy, CR 707.10c)", seen, wantPrompts)
	}
	if !stackFullyEmpty(g) {
		t.Fatalf("the stack never emptied: %d cards, %d items, %d pending triggers",
			g.Stack.Size(), len(g.StackMeta), len(g.PendingTriggers))
	}
}

// --- the count ------------------------------------------------------

// TestGrapeshotCopiesForEachSpellCastBeforeIt — the headline. Two
// spells, then Grapeshot: storm count 2, three resolutions, three
// damage.
func TestGrapeshotCopiesForEachSpellCastBeforeIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	before := lifeOf(g, opp)

	stormFiller(t, g)
	stormFiller(t, g)
	grapeshot := castCatalogSpell(t, g, "Grapeshot", "Sorcery", grapeshotOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp}})

	if triggerOnStack(g, grapeshot) == nil {
		t.Fatal("no storm trigger went on the stack with the spell")
	}
	answerCopyPrompts(t, g, me, opp, 2)

	if got := lifeOf(g, opp); got != before-3 {
		t.Errorf("life %d → %d, want -3 (the spell plus two copies)", before, got)
	}
	// Two fillers and the Grapeshot, and nothing else: a copy is not
	// a card and leaves no graveyard trace (CR 707.10).
	if got := graveyardSize(g, me); got != 3 {
		t.Errorf("graveyard = %d cards, want 3 (a copy is not a card)", got)
	}
}

// TestTheTurnsFirstStormSpellCopiesNothing — storm count 0 is an
// ordinary resolution, not an error, and the trigger still goes on
// the stack and still resolves.
func TestTheTurnsFirstStormSpellCopiesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	before := lifeOf(g, opp)

	grapeshot := castCatalogSpell(t, g, "Grapeshot", "Sorcery", grapeshotOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp}})
	if triggerOnStack(g, grapeshot) == nil {
		t.Fatal("the trigger fires even at a count of zero")
	}
	answerCopyPrompts(t, g, me, opp, 0)

	if got := lifeOf(g, opp); got != before-1 {
		t.Errorf("life %d → %d, want -1 (no copies)", before, got)
	}
}

// TestBrainFreezeCountsAnOpponentsSpells — CR 702.40a's "each other
// spell", which is the whole reason the count is table-wide. Brain
// Freeze is cast in response to somebody else's turn and counts THEIR
// spells; a per-player count makes the card a blank.
func TestBrainFreezeCountsAnOpponentsSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMainOf(t, g, 1)

	// Two spells by the opponent, on the opponent's turn.
	castFromSeat(t, g, opp, "Filler A", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	castFromSeat(t, g, opp, "Filler B", "Instant", "", nil)
	passPriorityAroundTable(t, g)

	libraryBefore := opp.Library.Size()
	castFromSeat(t, g, me, "Brain Freeze", "Instant", brainFreezeOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	answerCopyPrompts(t, g, me.ID, opp.ID, 2)

	// Three resolutions × three cards.
	if got := libraryBefore - opp.Library.Size(); got != 9 {
		t.Errorf("milled %d cards, want 9 (three mills of three)", got)
	}
}

// TestStormDoesNotCountItsOwnCopies — a copy is created, not cast
// (CR 707.10), so it never joins the count. Two storm spells in a
// turn is the shape that exposes it: if the first one's copies
// counted, the second would be off by two.
func TestStormDoesNotCountItsOwnCopies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	before := lifeOf(g, opp)

	stormFiller(t, g)
	castCatalogSpell(t, g, "Grapeshot", "Sorcery", grapeshotOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp}})
	answerCopyPrompts(t, g, me, opp, 1) // count 1: the filler
	castCatalogSpell(t, g, "Grapeshot", "Sorcery", grapeshotOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp}})
	answerCopyPrompts(t, g, me, opp, 2) // count 2: the filler and the first Grapeshot

	// (1 + 1) + (1 + 2) = 5. Counting the first Grapeshot's copy
	// would make it 6.
	if got := lifeOf(g, opp); got != before-5 {
		t.Errorf("life %d → %d, want -5", before, got)
	}
}

// TestStormCountIsTakenAtResolution — a spell cast in RESPONSE to the
// storm trigger was cast AFTER the storm spell, so it does not count.
// This is the assertion behind ADR 0086 Decision 2: the count is read
// as the trigger resolves, and reading it there is still right
// because the cast order cannot be rewritten behind it.
func TestStormCountIsTakenAtResolution(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := lifeOf(g, opp.ID)

	stormFiller(t, g)
	castCatalogSpell(t, g, "Grapeshot", "Sorcery", grapeshotOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	// The opponent responds to the storm trigger with two spells of
	// their own. Both resolve before the trigger does.
	castFromSeat(t, g, opp, "Response A", "Instant", "", nil)
	castFromSeat(t, g, opp, "Response B", "Instant", "", nil)

	answerCopyPrompts(t, g, me.ID, opp.ID, 1)

	// Count 1, not 3: only the filler was cast BEFORE the Grapeshot.
	if got := lifeOf(g, opp.ID); got != before-2 {
		t.Errorf("life %d → %d, want -2 (one copy — the responses came after)", before, got)
	}
}

// --- the copies are separate objects --------------------------------

// TestGrapeshotCopiesMayChooseNewTargets — CR 702.40a's second
// sentence, and the reason Grapeshot kills a board rather than a
// player. Each copy is offered its own choice and each choice sticks.
func TestGrapeshotCopiesMayChooseNewTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b := g.Seats[0].ID, g.Seats[1].ID, g.Seats[2].ID
	lifeA, lifeB := lifeOf(g, a), lifeOf(g, b)

	stormFiller(t, g)
	castCatalogSpell(t, g, "Grapeshot", "Sorcery", grapeshotOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: a}})

	// One copy, re-aimed at the other opponent.
	for i := 0; i < 8 && latestPickTarget(g, me) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	prompt := latestPickTarget(g, me)
	if prompt == nil {
		t.Fatal("the copy was never offered new targets")
	}
	if !hasID(prompt.PickTargetPlayers, b) {
		t.Fatalf("the copy's legal set should offer every player: %v", prompt.PickTargetPlayers)
	}
	if err := g.ResolvePickTarget(prompt.ID, me, game.TargetRef{Kind: game.TargetPlayer, ID: b}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := lifeOf(g, a); got != lifeA-1 {
		t.Errorf("the original's target: %d → %d, want -1", lifeA, got)
	}
	if got := lifeOf(g, b); got != lifeB-1 {
		t.Errorf("the copy's new target: %d → %d, want -1", lifeB, got)
	}
}

// TestEmptyTheWarrensAsksNothing — the storm card with no target
// clause. Its reminder text has no "you may choose new targets"
// because there is nothing to choose, and the card file says nothing
// about it: the engine skips the CR 707.10c offer for a copy whose
// original named no target.
func TestEmptyTheWarrensAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0].ID

	stormFiller(t, g)
	stormFiller(t, g)
	castCatalogSpell(t, g, "Empty the Warrens", "Sorcery", emptyTheWarrensOracle, nil)
	answerCopyPrompts(t, g, me, me, 0)

	// Three resolutions × two Goblins.
	if got := countOnBattlefield(g, "Goblin", me); got != 6 {
		t.Errorf("%d Goblins, want 6 (the spell plus two copies, two each)", got)
	}
}

// TestFlusterstormTaxesEachCopySeparately — the row's named card.
// Each copy is its own "unless its controller pays {1}", so a storm
// count of one is two separate taxes and paying one does not pay the
// other.
func TestFlusterstormTaxesEachCopySeparately(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMainOf(t, g, 1)

	castFromSeat(t, g, opp, "Filler", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	victim := castFromSeat(t, g, opp, "Divination", "Sorcery", "", nil)
	castFromSeat(t, g, me, "Flusterstorm", "Instant", flusterstormOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})

	// Storm count 2 (the filler and the Divination), so three
	// Flusterstorms resolve and each asks its own question. The
	// victim pays the first two and runs out on the third — which is
	// the card: one {1} does not buy safety from the copies.
	for i := 0; i < 2; i++ {
		opp.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
	asks := 0
	for i := 0; i < 64 && !stackFullyEmpty(g); i++ {
		if p := latestPickTarget(g, me.ID); p != nil {
			// The copies keep the original's target: every one of
			// them is about the same spell.
			if err := g.ResolvePickTarget(p.ID, me.ID,
				game.TargetRef{Kind: game.TargetCard, ID: victim}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if hasPayUnlessFor(g, opp.ID) {
			asks++
			answerPayUnless(t, g, opp.ID, asks <= 2)
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if asks != 3 {
		t.Errorf("the victim was asked %d times, want 3 (the spell plus two copies, each its own tax)", asks)
	}
	if g.Stack.Contains(victim) {
		t.Error("the third, unpaid Flusterstorm did not counter the spell")
	}
}

// TestTendrilsOfAgonyDrainsOncePerCopy — two life each way per
// resolution, and "you gain" is the copy's controller.
func TestTendrilsOfAgonyDrainsOncePerCopy(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	mine, theirs := lifeOf(g, me), lifeOf(g, opp)

	stormFiller(t, g)
	stormFiller(t, g)
	castCatalogSpell(t, g, "Tendrils of Agony", "Sorcery", tendrilsOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp}})
	answerCopyPrompts(t, g, me, opp, 2)

	if got := lifeOf(g, opp); got != theirs-6 {
		t.Errorf("their life %d → %d, want -6", theirs, got)
	}
	if got := lifeOf(g, me); got != mine+6 {
		t.Errorf("my life %d → %d, want +6", mine, got)
	}
}

// --- the awkward corners --------------------------------------------

// TestACounteredStormSpellKeepsItsTriggerOnTheStack — CR 113.7a: once
// triggered, an ability exists on the stack independently of its
// source, so countering the spell does not counter the trigger.
//
// The copies are the OPEN half (ADR 0086 Decision 6): the shared copy
// path answers ErrCardNotFound for a spell that has left the stack
// and CopySpell treats that as "the copy effect did nothing", which
// CR 608.2h's last-known-information reading says is wrong. It is not
// fixed here because the fix must not reach Reverberate, which
// TARGETS the spell and is correctly countered by game rules when
// that target is gone. This test pins what the engine does do, so the
// day that changes it changes here too.
func TestACounteredStormSpellKeepsItsTriggerOnTheStack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	before := lifeOf(g, opp)

	stormFiller(t, g)
	grapeshot := castCatalogSpell(t, g, "Grapeshot", "Sorcery", grapeshotOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp}})
	trigger := triggerOnStack(g, grapeshot)
	if trigger == nil {
		t.Fatal("no storm trigger")
	}

	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(grapeshot); err != nil {
			t.Errorf("CounterTargetForEffect: %v", err)
		}
	})

	if g.Stack.Contains(grapeshot) {
		t.Fatal("the spell was not countered — the fixture is wrong")
	}
	if triggerOnStack(g, grapeshot) == nil {
		t.Error("countering the spell took its storm trigger with it (CR 113.7a)")
	}

	answerCopyPrompts(t, g, me, opp, 0)
	if got := lifeOf(g, opp); got != before {
		t.Errorf("life %d → %d: a countered Grapeshot deals no damage", before, got)
	}
}

// TestUndoAcrossTheStormCopies — an undo taken before the trigger
// resolves rewinds every copy it made, and the table replays to the
// same place.
func TestUndoAcrossTheStormCopies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID

	stormFiller(t, g)
	stormFiller(t, g)
	castCatalogSpell(t, g, "Grapeshot", "Sorcery", grapeshotOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp}})

	atCast := lifeOf(g, opp)
	snap := g.Clone()

	answerCopyPrompts(t, g, me, opp, 2)
	if got := lifeOf(g, opp); got != atCast-3 {
		t.Fatalf("pre-undo: life %d → %d, want -3", atCast, got)
	}

	g.WithWriteLock(func() { g.RestoreFrom(snap) })

	if got := lifeOf(g, opp); got != atCast {
		t.Errorf("undo did not rewind the copies' damage: life %d, want %d", got, atCast)
	}
	if stackFullyEmpty(g) {
		t.Fatal("undo did not restore the spell and its trigger to the stack")
	}
	if n := len(g.PendingChoices); n != 0 {
		t.Errorf("undo left %d prompts open, want 0", n)
	}
	// And it replays to the same place, with the same count — the
	// cast order rewound with everything else, so the copies are not
	// re-counted.
	answerCopyPrompts(t, g, me, opp, 2)
	if got := lifeOf(g, opp); got != atCast-3 {
		t.Errorf("replay after undo: life %d → %d, want -3", atCast, got)
	}
}

// TestEveryStormCardDeclaresTheKeywordAndNothingElse — the point of
// #1238. A card with printed storm is its printed effect plus one
// line; if a storm card ever needs bespoke trigger code again, this
// is where it shows up.
func TestEveryStormCardDeclaresTheKeywordAndNothingElse(t *testing.T) {
	for _, oracle := range []string{
		grapeshotOracle, brainFreezeOracle, emptyTheWarrensOracle,
		flusterstormOracle, tendrilsOracle,
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Fatalf("%s is not registered", oracle)
		}
		if len(spec.Triggered) != 1 {
			t.Errorf("%s declares %d triggered abilities, want exactly one (storm)", spec.Name, len(spec.Triggered))
			continue
		}
		got := spec.Triggered[0]
		if !got.FromStack || len(got.Watches) != 1 || got.Watches[0] != game.EventCast {
			t.Errorf("%s's trigger is not the storm shape: FromStack=%v watches=%v",
				spec.Name, got.FromStack, got.Watches)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: Completeness = %v, Caveats = %v, want Full and none",
				spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}
