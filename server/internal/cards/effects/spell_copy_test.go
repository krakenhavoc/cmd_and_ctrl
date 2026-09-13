package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spell_copy_test.go — CR 706.10. The three facts the copy has to
// get right are all observable on one board: the copy runs the
// spell's effect, the copy's targets can differ from the
// original's, and the copy leaves NOTHING behind when it is done.
//
// The last one is the test that would have caught the obvious first
// draft. Routing a resolved copy through the ordinary instant exit
// is a one-word mistake and produces a graveyard with two Lightning
// Bolts in it — which nothing in the damage assertions would
// notice, and which Tarmogoyf, Regrowth and flashback all would.

const (
	reverberateOracle        = "a1f55890-31c5-4ed4-a2cd-7a4a9f05f8ca"
	twincastOracle           = "8f878efc-850f-43d2-a6fe-5ea8d1dd5afb"
	increasingVengeanceOracl = "a5ca7bd9-0964-405f-adb9-7c27153595e6"
)

func lifeOf(g *game.Game, playerID uuid.UUID) int {
	for _, p := range g.Seats {
		if p.ID == playerID {
			return p.Life
		}
	}
	return -1
}

// graveyardSize counts a seat's graveyard. A copy that "ceased to
// exist" correctly leaves this number exactly where a copy that
// went to the graveyard would not.
func graveyardSize(g *game.Game, playerID uuid.UUID) int {
	for _, p := range g.Seats {
		if p.ID == playerID {
			return p.Graveyard.Size()
		}
	}
	return -1
}

// Reverberate on your own Lightning Bolt, re-aimed at a different
// opponent. Both copies deal their 3; the copy is not a card and so
// leaves no trace.
func TestReverberateCopiesAndRetargets(t *testing.T) {
	g := newCatalogGame(t)
	me, victimA, victimB := g.Seats[0].ID, g.Seats[1].ID, g.Seats[2].ID

	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victimA}})
	castCatalogSpell(t, g, "Reverberate", "Instant", reverberateOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})

	// Reverberate is on top and resolves first, which opens the CR
	// 706.10c prompt before anything reaches the stack.
	for i := 0; i < 8 && latestPickTarget(g, me) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	prompt := latestPickTarget(g, me)
	if prompt == nil {
		t.Fatal("no re-target prompt for the copy")
	}
	if !hasID(prompt.PickTargetPlayers, victimB) {
		t.Fatalf("the copy's legal set should offer every player: %v", prompt.PickTargetPlayers)
	}
	if err := g.ResolvePickTarget(prompt.ID, me,
		game.TargetRef{Kind: game.TargetPlayer, ID: victimB}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := lifeOf(g, victimA); got != 37 {
		t.Errorf("original target life = %d, want 37", got)
	}
	if got := lifeOf(g, victimB); got != 37 {
		t.Errorf("re-targeted copy life = %d, want 37", got)
	}
	// Lightning Bolt and Reverberate — and nothing else. A copy that
	// went to a graveyard would make this 3.
	if got := graveyardSize(g, me); got != 2 {
		t.Errorf("graveyard = %d cards, want 2 (a copy is not a card, CR 706.10)", got)
	}
}

// The copy's controller is the player who made it, not the
// controller of the spell copied. Reverberate on an OPPONENT's Bolt
// gives you the second one, aimed wherever you like.
func TestReverberateOnAnOpponentsSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID

	// The active seat is me; seed the opponent's Bolt by hand so it
	// is genuinely theirs.
	bolt := uuid.New()
	oppSeat := g.Seats[1]
	oppSeat.Hand.PushTop(game.Card{
		InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant",
		OracleID: lightningBoltOracle, Owner: opp, Controller: opp,
	})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	// Active seat passes so the opponent holds priority and can cast.
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if err := g.CastSpell(opp, bolt, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me}},
	}); err != nil {
		t.Fatalf("opponent CastSpell: %v", err)
	}

	rev := uuid.New()
	g.Seats[0].Hand.PushTop(game.Card{
		InstanceID: rev, Name: "Reverberate", TypeLine: "Instant",
		OracleID: reverberateOracle, Owner: me, Controller: me,
	})
	if err := g.CastSpell(me, rev, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}},
	}); err != nil {
		t.Fatalf("CastSpell Reverberate: %v", err)
	}

	for i := 0; i < 8 && latestPickTarget(g, me) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	prompt := latestPickTarget(g, me)
	if prompt == nil {
		t.Fatal("the COPY's controller chooses its targets — prompt should address me")
	}
	if err := g.ResolvePickTarget(prompt.ID, me,
		game.TargetRef{Kind: game.TargetPlayer, ID: opp}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := lifeOf(g, opp); got != 37 {
		t.Errorf("the copy you aimed back at its owner: life = %d, want 37", got)
	}
	if got := lifeOf(g, me); got != 37 {
		t.Errorf("the original still resolves: life = %d, want 37", got)
	}
}

// Twincast is Reverberate in blue and shares the primitive; the
// test that earns its keep is the untouched-targets path — decline
// the change by re-picking what was already there (CR 706.10c).
func TestTwincastKeepingTheSameTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0].ID, g.Seats[1].ID

	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim}})
	castCatalogSpell(t, g, "Twincast", "Instant", twincastOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})

	for i := 0; i < 8 && latestPickTarget(g, me) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	prompt := latestPickTarget(g, me)
	if prompt == nil {
		t.Fatal("no re-target prompt")
	}
	if err := g.ResolvePickTarget(prompt.ID, me,
		game.TargetRef{Kind: game.TargetPlayer, ID: victim}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := lifeOf(g, victim); got != 34 {
		t.Errorf("both bolts at one head: life = %d, want 34", got)
	}
}

// Increasing Vengeance's "you control" clause is the narrowing that
// distinguishes it. An opponent's spell must not be a legal target.
func TestIncreasingVengeanceOnlyCopiesYourOwnSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID

	bolt := uuid.New()
	g.Seats[1].Hand.PushTop(game.Card{
		InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant",
		OracleID: lightningBoltOracle, Owner: opp, Controller: opp,
	})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if err := g.CastSpell(opp, bolt, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me}},
	}); err != nil {
		t.Fatalf("opponent CastSpell: %v", err)
	}

	iv := uuid.New()
	g.Seats[0].Hand.PushTop(game.Card{
		InstanceID: iv, Name: "Increasing Vengeance", TypeLine: "Instant",
		OracleID: increasingVengeanceOracl, Owner: me, Controller: me,
	})
	err := g.CastSpell(me, iv, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}},
	})
	if err != game.ErrIllegalTarget {
		t.Fatalf(`"you control" must reject an opponent's spell: got %v`, err)
	}
}

// The positive half of the same clause, and the reason it is a
// separate test: a rejection assertion alone would pass for the
// wrong reason if `Controller` were never populated on a card sitting
// on the stack — the card would be uncastable at anything, and
// nothing above would notice.
func TestIncreasingVengeanceCopiesYourOwnSpellTwice(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0].ID, g.Seats[1].ID

	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim}})
	castCatalogSpell(t, g, "Increasing Vengeance", "Instant", increasingVengeanceOracl,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})

	for i := 0; i < 8 && latestPickTarget(g, me) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	prompt := latestPickTarget(g, me)
	if prompt == nil {
		t.Fatal("no re-target prompt — the copy was never offered one")
	}
	if err := g.ResolvePickTarget(prompt.ID, me,
		game.TargetRef{Kind: game.TargetPlayer, ID: victim}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)

	// Hard-cast (not from a graveyard), so exactly one copy: 3 from
	// the copy plus 3 from the original.
	if got := lifeOf(g, victim); got != 34 {
		t.Errorf("life = %d, want 34 (one copy + the original)", got)
	}
	// Two cards in the graveyard — the Bolt and the Vengeance. The
	// copy is not a card.
	if got := graveyardSize(g, me); got != 2 {
		t.Errorf("graveyard = %d, want 2", got)
	}
}

// "If this spell was cast from a graveyard, copy that spell TWICE
// instead." Two copies is not one copy resolving twice: each is
// created separately and each is offered its own CR 706.10c target
// choice, so the assertion is two prompts and three Bolts' worth of
// damage.
//
// The branch reads CastFromZone, not the alternative cost that was
// paid, because that is what the card says — so this test drives the
// S29 flashback path end to end rather than stubbing the field.
func TestIncreasingVengeanceFlashedBackCopiesTwice(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0].ID, g.Seats[1].ID

	iv := seedGraveyardCard(t, g, "Increasing Vengeance", "Instant", increasingVengeanceOracl)
	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim}})
	if err := g.CastSpell(me, iv, game.CastSpellParams{
		FromZone:        "graveyard",
		AlternativeCost: "flashback",
		Targets:         []game.TargetRef{{Kind: game.TargetCard, ID: bolt}},
	}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}

	// Two independent re-target prompts, answered one at a time.
	for answered := 0; answered < 2; answered++ {
		for i := 0; i < 8 && latestPickTarget(g, me) == nil; i++ {
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
		}
		prompt := latestPickTarget(g, me)
		if prompt == nil {
			t.Fatalf("copy %d of 2 was never offered a target choice", answered+1)
		}
		if err := g.ResolvePickTarget(prompt.ID, me,
			game.TargetRef{Kind: game.TargetPlayer, ID: victim}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
	}
	passPriorityAroundTable(t, g)

	if got := lifeOf(g, victim); got != 31 {
		t.Errorf("life = %d, want 31 (two copies + the original)", got)
	}
	// CR 702.34a: the flashed-back card is exiled, not returned. And
	// neither copy is anywhere at all.
	if !g.Exile.Contains(iv) {
		t.Error("a flashed-back Increasing Vengeance must be exiled")
	}
	if got := graveyardSize(g, me); got != 1 {
		t.Errorf("graveyard = %d, want 1 (the Bolt alone)", got)
	}
}
