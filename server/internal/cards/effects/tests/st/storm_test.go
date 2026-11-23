package testsst

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- the count ------------------------------------------------------

// TestGrapeshotCopiesForEachSpellCastBeforeIt — the headline. Two
// spells, then Grapeshot: storm count 2, three resolutions, three
// damage.
func TestGrapeshotCopiesForEachSpellCastBeforeIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	before := effectstest.LifeOf(g, opp)

	effectstest.StormFiller(t, g)
	effectstest.StormFiller(t, g)
	grapeshot := effectstest.CastCatalogSpell(t, g, "Grapeshot", "Sorcery", effectstest.GrapeshotOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp}})

	if effectstest.TriggerOnStack(g, grapeshot) == nil {
		t.Fatal("no storm trigger went on the stack with the spell")
	}
	effectstest.AnswerCopyPrompts(t, g, me, opp, 2)

	if got := effectstest.LifeOf(g, opp); got != before-3 {
		t.Errorf("life %d → %d, want -3 (the spell plus two copies)", before, got)
	}
	// Two fillers and the Grapeshot, and nothing else: a copy is not
	// a card and leaves no graveyard trace (CR 707.10).
	if got := effectstest.GraveyardSize(g, me); got != 3 {
		t.Errorf("graveyard = %d cards, want 3 (a copy is not a card)", got)
	}
}

// TestTheTurnsFirstStormSpellCopiesNothing — storm count 0 is an
// ordinary resolution, not an error, and the trigger still goes on
// the stack and still resolves.
func TestTheTurnsFirstStormSpellCopiesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	before := effectstest.LifeOf(g, opp)

	grapeshot := effectstest.CastCatalogSpell(t, g, "Grapeshot", "Sorcery", effectstest.GrapeshotOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp}})
	if effectstest.TriggerOnStack(g, grapeshot) == nil {
		t.Fatal("the trigger fires even at a count of zero")
	}
	effectstest.AnswerCopyPrompts(t, g, me, opp, 0)

	if got := effectstest.LifeOf(g, opp); got != before-1 {
		t.Errorf("life %d → %d, want -1 (no copies)", before, got)
	}
}

// TestBrainFreezeCountsAnOpponentsSpells — CR 702.40a's "each other
// spell", which is the whole reason the count is table-wide. Brain
// Freeze is cast in response to somebody else's turn and counts THEIR
// spells; a per-player count makes the card a blank.
func TestBrainFreezeCountsAnOpponentsSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMainOf(t, g, 1)

	// Two spells by the opponent, on the opponent's turn.
	effectstest.CastFromSeat(t, g, opp, "Filler A", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastFromSeat(t, g, opp, "Filler B", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)

	libraryBefore := opp.Library.Size()
	effectstest.CastFromSeat(t, g, me, "Brain Freeze", "Instant", effectstest.BrainFreezeOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.AnswerCopyPrompts(t, g, me.ID, opp.ID, 2)

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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	before := effectstest.LifeOf(g, opp)

	effectstest.StormFiller(t, g)
	effectstest.CastCatalogSpell(t, g, "Grapeshot", "Sorcery", effectstest.GrapeshotOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp}})
	effectstest.AnswerCopyPrompts(t, g, me, opp, 1) // count 1: the filler
	effectstest.CastCatalogSpell(t, g, "Grapeshot", "Sorcery", effectstest.GrapeshotOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp}})
	effectstest.AnswerCopyPrompts(t, g, me, opp, 2) // count 2: the filler and the first Grapeshot

	// (1 + 1) + (1 + 2) = 5. Counting the first Grapeshot's copy
	// would make it 6.
	if got := effectstest.LifeOf(g, opp); got != before-5 {
		t.Errorf("life %d → %d, want -5", before, got)
	}
}

// TestStormCountIsTakenAtResolution — a spell cast in RESPONSE to the
// storm trigger was cast AFTER the storm spell, so it does not count.
// This is the assertion behind ADR 0086 Decision 2: the count is read
// as the trigger resolves, and reading it there is still right
// because the cast order cannot be rewritten behind it.
func TestStormCountIsTakenAtResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := effectstest.LifeOf(g, opp.ID)

	effectstest.StormFiller(t, g)
	effectstest.CastCatalogSpell(t, g, "Grapeshot", "Sorcery", effectstest.GrapeshotOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	// The opponent responds to the storm trigger with two spells of
	// their own. Both resolve before the trigger does.
	effectstest.CastFromSeat(t, g, opp, "Response A", "Instant", "", nil)
	effectstest.CastFromSeat(t, g, opp, "Response B", "Instant", "", nil)

	effectstest.AnswerCopyPrompts(t, g, me.ID, opp.ID, 1)

	// Count 1, not 3: only the filler was cast BEFORE the Grapeshot.
	if got := effectstest.LifeOf(g, opp.ID); got != before-2 {
		t.Errorf("life %d → %d, want -2 (one copy — the responses came after)", before, got)
	}
}

// --- the copies are separate objects --------------------------------

// TestGrapeshotCopiesMayChooseNewTargets — CR 702.40a's second
// sentence, and the reason Grapeshot kills a board rather than a
// player. Each copy is offered its own choice and each choice sticks.
func TestGrapeshotCopiesMayChooseNewTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b := g.Seats[0].ID, g.Seats[1].ID, g.Seats[2].ID
	lifeA, lifeB := effectstest.LifeOf(g, a), effectstest.LifeOf(g, b)

	effectstest.StormFiller(t, g)
	effectstest.CastCatalogSpell(t, g, "Grapeshot", "Sorcery", effectstest.GrapeshotOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: a}})

	// One copy, re-aimed at the other opponent.
	for i := 0; i < 8 && effectstest.LatestPickTarget(g, me) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	prompt := effectstest.LatestPickTarget(g, me)
	if prompt == nil {
		t.Fatal("the copy was never offered new targets")
	}
	if !effectstest.HasID(prompt.PickTargetPlayers, b) {
		t.Fatalf("the copy's legal set should offer every player: %v", prompt.PickTargetPlayers)
	}
	if err := g.ResolvePickTarget(prompt.ID, me, game.TargetRef{Kind: game.TargetPlayer, ID: b}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.LifeOf(g, a); got != lifeA-1 {
		t.Errorf("the original's target: %d → %d, want -1", lifeA, got)
	}
	if got := effectstest.LifeOf(g, b); got != lifeB-1 {
		t.Errorf("the copy's new target: %d → %d, want -1", lifeB, got)
	}
}

// TestEmptyTheWarrensAsksNothing — the storm card with no target
// clause. Its reminder text has no "you may choose new targets"
// because there is nothing to choose, and the card file says nothing
// about it: the engine skips the CR 707.10c offer for a copy whose
// original named no target.
func TestEmptyTheWarrensAsksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID

	effectstest.StormFiller(t, g)
	effectstest.StormFiller(t, g)
	effectstest.CastCatalogSpell(t, g, "Empty the Warrens", "Sorcery", effectstest.EmptyTheWarrensOracle, nil)
	effectstest.AnswerCopyPrompts(t, g, me, me, 0)

	// Three resolutions × two Goblins.
	if got := effectstest.CountOnBattlefield(g, "Goblin", me); got != 6 {
		t.Errorf("%d Goblins, want 6 (the spell plus two copies, two each)", got)
	}
}

// TestFlusterstormTaxesEachCopySeparately — the row's named card.
// Each copy is its own "unless its controller pays {1}", so a storm
// count of one is two separate taxes and paying one does not pay the
// other.
func TestFlusterstormTaxesEachCopySeparately(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMainOf(t, g, 1)

	effectstest.CastFromSeat(t, g, opp, "Filler", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	victim := effectstest.CastFromSeat(t, g, opp, "Divination", "Sorcery", "", nil)
	effectstest.CastFromSeat(t, g, me, "Flusterstorm", "Instant", effectstest.FlusterstormOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})

	// Storm count 2 (the filler and the Divination), so three
	// Flusterstorms resolve and each asks its own question. The
	// victim pays the first two and runs out on the third — which is
	// the card: one {1} does not buy safety from the copies.
	for i := 0; i < 2; i++ {
		opp.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
	asks := 0
	for i := 0; i < 64 && !effectstest.StackFullyEmpty(g); i++ {
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			// The copies keep the original's target: every one of
			// them is about the same spell.
			if err := g.ResolvePickTarget(p.ID, me.ID,
				game.TargetRef{Kind: game.TargetCard, ID: victim}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if effectstest.HasPayUnlessFor(g, opp.ID) {
			asks++
			effectstest.AnswerPayUnless(t, g, opp.ID, asks <= 2)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	mine, theirs := effectstest.LifeOf(g, me), effectstest.LifeOf(g, opp)

	effectstest.StormFiller(t, g)
	effectstest.StormFiller(t, g)
	effectstest.CastCatalogSpell(t, g, "Tendrils of Agony", "Sorcery", effectstest.TendrilsOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp}})
	effectstest.AnswerCopyPrompts(t, g, me, opp, 2)

	if got := effectstest.LifeOf(g, opp); got != theirs-6 {
		t.Errorf("their life %d → %d, want -6", theirs, got)
	}
	if got := effectstest.LifeOf(g, me); got != mine+6 {
		t.Errorf("my life %d → %d, want +6", mine, got)
	}
}

// --- the awkward corners --------------------------------------------

// TestACounteredStormSpellKeepsItsTriggerOnTheStack — CR 113.7a: once
// triggered, an ability exists on the stack independently of its
// source, so countering the spell does not counter the trigger.
//
// And the trigger still COPIES it (#1255, CR 608.2h). The storm
// ability names the spell and does not target it, so when the spell
// has left the stack the copies are made from its last-known
// information. Until #1255 the shared copy path answered
// ErrCardNotFound here and the trigger made nothing; the original
// Grapeshot deals no damage (it was countered), the one copy deals 1.
func TestACounteredStormSpellKeepsItsTriggerOnTheStack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	before := effectstest.LifeOf(g, opp)

	effectstest.StormFiller(t, g)
	grapeshot := effectstest.CastCatalogSpell(t, g, "Grapeshot", "Sorcery", effectstest.GrapeshotOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp}})
	trigger := effectstest.TriggerOnStack(g, grapeshot)
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
	if effectstest.TriggerOnStack(g, grapeshot) == nil {
		t.Error("countering the spell took its storm trigger with it (CR 113.7a)")
	}

	effectstest.AnswerCopyPrompts(t, g, me, opp, 1)
	if got := effectstest.LifeOf(g, opp); got != before-1 {
		t.Errorf("life %d → %d, want -1: the countered Grapeshot deals nothing, its one copy deals 1 (CR 608.2h)", before, got)
	}
	// The copy is not a card, and the countered original is in the
	// graveyard once: the filler and the Grapeshot.
	if got := effectstest.GraveyardSize(g, me); got != 2 {
		t.Errorf("graveyard = %d cards, want 2 (the filler and the countered Grapeshot)", got)
	}
}

// TestBrainFreezeCounteredInResponseStillMills — the same rule on the
// storm card a counterspell is most often pointed at. Two spells
// before it, so two copies, and each mills three: the countered
// original mills nothing.
func TestBrainFreezeCounteredInResponseStillMills(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	effectstest.StormFiller(t, g)
	effectstest.StormFiller(t, g)
	libraryBefore := opp.Library.Size()
	freeze := effectstest.CastCatalogSpell(t, g, "Brain Freeze", "Instant", effectstest.BrainFreezeOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if effectstest.TriggerOnStack(g, freeze) == nil {
		t.Fatal("no storm trigger")
	}
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(freeze); err != nil {
			t.Fatalf("CounterTargetForEffect: %v", err)
		}
	})

	effectstest.AnswerCopyPrompts(t, g, me.ID, opp.ID, 2)
	if got := libraryBefore - opp.Library.Size(); got != 6 {
		t.Errorf("milled %d cards, want 6 (two copies of three; the countered original mills nothing)", got)
	}
}

// TestUndoAcrossTheStormCopies — an undo taken before the trigger
// resolves rewinds every copy it made, and the table replays to the
// same place.
func TestUndoAcrossTheStormCopies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID

	effectstest.StormFiller(t, g)
	effectstest.StormFiller(t, g)
	effectstest.CastCatalogSpell(t, g, "Grapeshot", "Sorcery", effectstest.GrapeshotOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp}})

	atCast := effectstest.LifeOf(g, opp)
	snap := g.Clone()

	effectstest.AnswerCopyPrompts(t, g, me, opp, 2)
	if got := effectstest.LifeOf(g, opp); got != atCast-3 {
		t.Fatalf("pre-undo: life %d → %d, want -3", atCast, got)
	}

	g.WithWriteLock(func() { g.RestoreFrom(snap) })

	if got := effectstest.LifeOf(g, opp); got != atCast {
		t.Errorf("undo did not rewind the copies' damage: life %d, want %d", got, atCast)
	}
	if effectstest.StackFullyEmpty(g) {
		t.Fatal("undo did not restore the spell and its trigger to the stack")
	}
	if n := len(g.PendingChoices); n != 0 {
		t.Errorf("undo left %d prompts open, want 0", n)
	}
	// And it replays to the same place, with the same count — the
	// cast order rewound with everything else, so the copies are not
	// re-counted.
	effectstest.AnswerCopyPrompts(t, g, me, opp, 2)
	if got := effectstest.LifeOf(g, opp); got != atCast-3 {
		t.Errorf("replay after undo: life %d → %d, want -3", atCast, got)
	}
}

// TestEveryStormCardDeclaresTheKeywordAndNothingElse — the point of
// #1238. A card with printed storm is its printed effect plus one
// line; if a storm card ever needs bespoke trigger code again, this
// is where it shows up.
func TestEveryStormCardDeclaresTheKeywordAndNothingElse(t *testing.T) {
	for _, oracle := range []string{
		effectstest.GrapeshotOracle, effectstest.BrainFreezeOracle, effectstest.EmptyTheWarrensOracle,
		effectstest.FlusterstormOracle, effectstest.TendrilsOracle,
	} {
		spec, ok := effects.Lookup(oracle)
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
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: Completeness = %v, Caveats = %v, want Full and none",
				spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}
