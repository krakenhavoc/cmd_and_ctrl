package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// untap_hold_cards_test.go — #1313's proof cards: untap holds with a
// CR 611.2b "for as long as" duration (ADR 0058's 2026-09-23
// amendment).

const (
	oracleTyLee          = "081ad4e3-cda3-41cc-890f-412611dc9ea0"
	oracleDungeonGeists  = "ab5ebae2-cd77-4a7d-a93b-8042cd486429"
	oracleTidebinderMage = "f881378b-b539-4ea8-981d-e01ee82af105"
)

// castHoldCreature casts a creature spell with real P/T from the active
// player's hand and resolves it, leaving its ETB trigger waiting on its
// target prompt. castCatalogSpell's P/T-less creature would die to
// CR 704.5f before the "for as long as you control ~" hold could start.
func castHoldCreature(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	return id
}

// pickTargetPrompt returns the open pick_target prompt, failing the
// test if there is none.
func pickTargetPrompt(t *testing.T, g *game.Game) *game.PendingChoice {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			return c
		}
	}
	t.Fatal("no pick_target prompt is open")
	return nil
}

func answerPickTarget(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: target}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}

func holdCount(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("battlefield card %s is missing", id)
	}
	n := 0
	for _, skip := range c.NextUntapSkips {
		if skip.While != nil {
			n++
		}
	}
	return n
}

func TestUntapHoldCardsAreRegistered(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		full         bool
	}{
		{"Ty Lee, Chi Blocker", oracleTyLee, false},
		{"Dungeon Geists", oracleDungeonGeists, true},
		{"Tidebinder Mage", oracleTidebinderMage, true},
		{"Rust Tick", oracleRustTick, true},
		{"Amber Prison", oracleAmberPrison, true},
	} {
		spec, ok := Lookup(tc.oracle)
		if !ok || spec.Name != tc.name {
			t.Fatalf("Lookup(%s) = %q, %v", tc.oracle, spec.Name, ok)
		}
		if (spec.Completeness == CompletenessFull) != tc.full {
			t.Errorf("%s completeness = %v, want full %v", tc.name, spec.Completeness, tc.full)
		}
	}
	// Ty Lee's only remaining gap is prowess (#706).
	spec, _ := Lookup(oracleTyLee)
	if len(spec.Caveats) != 1 || !slices.Contains(spec.PrintedKeywords, "flash") {
		t.Errorf("Ty Lee caveats %q keywords %q, want the prowess caveat and flash", spec.Caveats, spec.PrintedKeywords)
	}
}

// Ty Lee: the held creature sits out every one of its controller's
// untap steps while Ty Lee stays, and untaps as normal once Ty Lee is
// gone.
func TestTyLeeHoldsItsTargetForAsLongAsYouControlTyLee(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := pushCreatureToBattlefieldForTest(g, opp.ID, "Target")
	tyLee := castHoldCreature(t, g, "Ty Lee, Chi Blocker", "Legendary Creature — Human Performer Ally", oracleTyLee, 2, 1)
	// "Up to one target creature" offers Ty Lee herself too.
	if opts := pickTargetPrompt(t, g).PickTargetCards; !slices.Contains(opts, tyLee) || !slices.Contains(opts, target) {
		t.Fatalf("Ty Lee's target options = %v, want both creatures", opts)
	}
	answerPickTarget(t, g, target)
	passPriorityAroundTable(t, g)

	if !tappedForTest(t, g, target) || holdCount(t, g, target) != 1 {
		t.Fatalf("after Ty Lee's trigger: tapped %v holds %d, want tapped with one hold", tappedForTest(t, g, target), holdCount(t, g, target))
	}
	for round := 0; round < 2; round++ {
		advanceToUpkeepOf(t, g, 1)
		if !tappedForTest(t, g, target) {
			t.Fatalf("round %d: the held creature untapped while Ty Lee is under your control", round+1)
		}
		advanceToUpkeepOf(t, g, 0)
	}

	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, tyLee); err != nil {
		t.Fatal(err)
	}
	advanceToUpkeepOf(t, g, 1)
	if tappedForTest(t, g, target) {
		t.Error("the creature stayed tapped after Ty Lee left the battlefield")
	}
	if holdCount(t, g, target) != 0 {
		t.Error("the expired hold was not dropped")
	}
}

// Dungeon Geists ruling (2019-07-12): "If Dungeon Geists leaves the
// battlefield before its triggered ability has resolved, the target
// creature will be tapped, but it will be able to untap as normal."
func TestDungeonGeistsLeavingInResponseTapsWithoutAHold(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := pushCreatureToBattlefieldForTest(g, opp.ID, "Target")
	geists := castHoldCreature(t, g, "Dungeon Geists", "Creature — Spirit", oracleDungeonGeists, 3, 3)
	if opts := pickTargetPrompt(t, g).PickTargetCards; slices.Contains(opts, geists) {
		t.Fatal("Dungeon Geists may target only a creature an opponent controls")
	}
	answerPickTarget(t, g, target)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, geists); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if !tappedForTest(t, g, target) {
		t.Fatal("the trigger did not tap its target")
	}
	if holdCount(t, g, target) != 0 {
		t.Fatal("a hold was recorded although its duration never started (CR 611.2b)")
	}
	advanceToUpkeepOf(t, g, 1)
	if tappedForTest(t, g, target) {
		t.Error("the creature did not untap as normal")
	}
}

// Dungeon Geists ruling (2019-07-12): "If another player gains control
// of Dungeon Geists, its effect expires. It won't keep the creature
// from untapping anymore, even if you later regain control."
func TestDungeonGeistsHoldEndsWhenControlIsLostEvenIfRegained(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	target := pushCreatureToBattlefieldForTest(g, opp.ID, "Target")
	geists := castHoldCreature(t, g, "Dungeon Geists", "Creature — Spirit", oracleDungeonGeists, 3, 3)
	answerPickTarget(t, g, target)
	passPriorityAroundTable(t, g)
	if holdCount(t, g, target) != 1 {
		t.Fatal("no hold recorded")
	}
	// A Threaten: the opponent controls Dungeon Geists until end of
	// turn, and gets nothing else out of it.
	g.WithWriteLock(func() {
		if !g.GainControlForEffect(geists, geists, opp.ID, g.UntilEndOfTurnDuration(), "test threaten") {
			t.Fatal("control change refused")
		}
		g.RecomputeLayersIfStaleLocked()
	})
	advanceToUpkeepOf(t, g, 1)
	if c, _ := battlefieldCard(g, geists); c.Controller != g.Seats[0].ID {
		t.Fatalf("control of Dungeon Geists did not revert at cleanup")
	}
	if tappedForTest(t, g, target) {
		t.Error("the hold survived losing control of Dungeon Geists")
	}
}

// Tidebinder Mage: the colour-narrowed target clause, and the same hold.
func TestTidebinderMageHoldsOnlyARedOrGreenOpposingCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	red := pushCreatureToBattlefieldForTest(g, opp.ID, "Red")
	blue := pushCreatureToBattlefieldForTest(g, opp.ID, "Blue")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case red:
				g.Battlefield.Cards[i].Colors = []string{"R"}
			case blue:
				g.Battlefield.Cards[i].Colors = []string{"U"}
			}
		}
	})
	castHoldCreature(t, g, "Tidebinder Mage", "Creature — Merfolk Wizard", oracleTidebinderMage, 2, 2)
	opts := pickTargetPrompt(t, g).PickTargetCards
	if !slices.Contains(opts, red) || slices.Contains(opts, blue) {
		t.Fatalf("Tidebinder Mage's options = %v, want the red creature and not the blue one", opts)
	}
	answerPickTarget(t, g, red)
	passPriorityAroundTable(t, g)
	advanceToUpkeepOf(t, g, 1)
	if !tappedForTest(t, g, red) {
		t.Error("the red creature untapped while you control Tidebinder Mage")
	}
}

// Rust Tick and Amber Prison: the activated ability's hold lasts while
// the source stays tapped — which the "you may choose not to untap"
// clause lets its controller decide each turn.
func TestRemainsTappedHoldsFollowTheSourcesOptOut(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine, targetType string
		mana                               int
	}{
		{"Rust Tick", oracleRustTick, "Artifact Creature — Insect", "Artifact", 1},
		{"Amber Prison", oracleAmberPrison, "Artifact", "Basic Land — Forest", 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			source := pushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
			target := pushPermanentForTest(g, opp.ID, "Target", "", tc.targetType)
			advanceToMain(t, g)
			for i := 0; i < tc.mana; i++ {
				b06AddMana(me, "C")
			}
			b16Activate(t, g, me.ID, source, 0, game.ActivateAbilityParams{Targets: cardRefs(target)})
			if !tappedForTest(t, g, source) || !tappedForTest(t, g, target) || holdCount(t, g, target) != 1 {
				t.Fatalf("after activation: source tapped %v, target tapped %v, holds %d",
					tappedForTest(t, g, source), tappedForTest(t, g, target), holdCount(t, g, target))
			}

			// Decline to untap the source: the target stays locked.
			choice := advanceToUntapChoiceOf(t, g, 0)
			if err := g.ResolveUntapChoice(choice.ID, me.ID, nil); err != nil {
				t.Fatalf("decline: %v", err)
			}
			advanceToUpkeepOf(t, g, 1)
			if !tappedForTest(t, g, target) {
				t.Fatal("the target untapped while the source stayed tapped")
			}

			// Untap the source: the hold is over.
			choice = advanceToUntapChoiceOf(t, g, 0)
			if err := g.ResolveUntapChoice(choice.ID, me.ID, []uuid.UUID{source}); err != nil {
				t.Fatalf("accept: %v", err)
			}
			advanceToUpkeepOf(t, g, 1)
			if tappedForTest(t, g, target) {
				t.Error("the target stayed tapped after the source untapped")
			}
		})
	}
}
