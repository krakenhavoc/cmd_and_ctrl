package testsu

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestUntapHoldCardsAreRegistered(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		full         bool
	}{
		{"Ty Lee, Chi Blocker", effectstest.OracleTyLee, true},
		{"Dungeon Geists", effectstest.OracleDungeonGeists, true},
		{"Tidebinder Mage", effectstest.OracleTidebinderMage, true},
		{"Rust Tick", effectstest.OracleRustTick, true},
		{"Amber Prison", effectstest.OracleAmberPrison, true},
	} {
		spec, ok := effects.Lookup(tc.oracle)
		if !ok || spec.Name != tc.name {
			t.Fatalf("Lookup(%s) = %q, %v", tc.oracle, spec.Name, ok)
		}
		if (spec.Completeness == effects.CompletenessFull) != tc.full {
			t.Errorf("%s completeness = %v, want full %v", tc.name, spec.Completeness, tc.full)
		}
	}
	// Ty Lee's last gap was prowess, closed by #706: no caveat left,
	// and both her printed keywords declared.
	spec, _ := effects.Lookup(effectstest.OracleTyLee)
	if len(spec.Caveats) != 0 || !slices.Contains(spec.PrintedKeywords, "flash") ||
		!slices.Contains(spec.PrintedKeywords, "prowess") {
		t.Errorf("Ty Lee caveats %q keywords %q, want no caveat and flash + prowess", spec.Caveats, spec.PrintedKeywords)
	}
}

// Ty Lee: the held creature sits out every one of its controller's
// untap steps while Ty Lee stays, and untaps as normal once Ty Lee is
// gone.
func TestTyLeeHoldsItsTargetForAsLongAsYouControlTyLee(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "Target")
	tyLee := effectstest.CastHoldCreature(t, g, "Ty Lee, Chi Blocker", "Legendary Creature — Human Performer Ally", effectstest.OracleTyLee, 2, 1)
	// "Up to one target creature" offers Ty Lee herself too.
	if opts := effectstest.PickTargetPrompt(t, g).PickTargetCards; !slices.Contains(opts, tyLee) || !slices.Contains(opts, target) {
		t.Fatalf("Ty Lee's target options = %v, want both creatures", opts)
	}
	effectstest.AnswerPickTarget(t, g, target)
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.TappedForTest(t, g, target) || effectstest.HoldCount(t, g, target) != 1 {
		t.Fatalf("after Ty Lee's trigger: tapped %v holds %d, want tapped with one hold", effectstest.TappedForTest(t, g, target), effectstest.HoldCount(t, g, target))
	}
	for round := 0; round < 2; round++ {
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		if !effectstest.TappedForTest(t, g, target) {
			t.Fatalf("round %d: the held creature untapped while Ty Lee is under your control", round+1)
		}
		effectstest.AdvanceToUpkeepOf(t, g, 0)
	}

	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, tyLee); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.TappedForTest(t, g, target) {
		t.Error("the creature stayed tapped after Ty Lee left the battlefield")
	}
	if effectstest.HoldCount(t, g, target) != 0 {
		t.Error("the expired hold was not dropped")
	}
}

// TestTyLeeETBTriggerFiresOnCastFromHand is the regression guard for
// #339 ("[in-app] Ty Lee ETB Not Firing"). The reported bug was a
// coverage gap, not an engine defect: Ty Lee had no Spec at all, so
// harvestFromZone (server/internal/game/triggers.go) queued nothing
// on entry — see the issue's 2026-09-14 triage comment. Fixed by
// #1351 (#1313), which registered the card; the extended hold
// behaviour (several untap steps, releasing when Ty Lee leaves) is
// covered above by TestTyLeeHoldsItsTargetForAsLongAsYouControlTyLee.
// This test pins the original complaint on its own: casting Ty Lee
// from hand and resolving her the way a player would must put her
// ETB trigger on the stack and resolve it per the printed text —
// tapping the chosen target and starting the "for as long as you
// control ~" hold (#1313).
func TestTyLeeETBTriggerFiresOnCastFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	target := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "Target")

	effectstest.CastHoldCreature(t, g, "Ty Lee, Chi Blocker", "Legendary Creature — Human Performer Ally", effectstest.OracleTyLee, 2, 1)

	// The ETB trigger fired: there is a target prompt open rather
	// than a vanilla body sitting on the battlefield with nothing
	// queued.
	effectstest.AnswerPickTarget(t, g, target)
	effectstest.PassPriorityAroundTable(t, g)

	// It resolved per the printed text: the target is tapped, and
	// the untap-lock hold (#1313) has started.
	if !effectstest.TappedForTest(t, g, target) {
		t.Fatal("Ty Lee's ETB trigger did not tap its target")
	}
	if effectstest.HoldCount(t, g, target) != 1 {
		t.Fatal("Ty Lee's ETB trigger did not start the untap-lock hold (#1313)")
	}
}

// Dungeon Geists ruling (2019-07-12): "If Dungeon Geists leaves the
// battlefield before its triggered ability has resolved, the target
// creature will be tapped, but it will be able to untap as normal."
func TestDungeonGeistsLeavingInResponseTapsWithoutAHold(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "Target")
	geists := effectstest.CastHoldCreature(t, g, "Dungeon Geists", "Creature — Spirit", effectstest.OracleDungeonGeists, 3, 3)
	if opts := effectstest.PickTargetPrompt(t, g).PickTargetCards; slices.Contains(opts, geists) {
		t.Fatal("Dungeon Geists may target only a creature an opponent controls")
	}
	effectstest.AnswerPickTarget(t, g, target)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, geists); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedForTest(t, g, target) {
		t.Fatal("the trigger did not tap its target")
	}
	if effectstest.HoldCount(t, g, target) != 0 {
		t.Fatal("a hold was recorded although its duration never started (CR 611.2b)")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.TappedForTest(t, g, target) {
		t.Error("the creature did not untap as normal")
	}
}

// Dungeon Geists ruling (2019-07-12): "If another player gains control
// of Dungeon Geists, its effect expires. It won't keep the creature
// from untapping anymore, even if you later regain control."
func TestDungeonGeistsHoldEndsWhenControlIsLostEvenIfRegained(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	target := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "Target")
	geists := effectstest.CastHoldCreature(t, g, "Dungeon Geists", "Creature — Spirit", effectstest.OracleDungeonGeists, 3, 3)
	effectstest.AnswerPickTarget(t, g, target)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HoldCount(t, g, target) != 1 {
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
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if c, _ := effectstest.BattlefieldCard(g, geists); c.Controller != g.Seats[0].ID {
		t.Fatalf("control of Dungeon Geists did not revert at cleanup")
	}
	if effectstest.TappedForTest(t, g, target) {
		t.Error("the hold survived losing control of Dungeon Geists")
	}
}

// Tidebinder Mage: the colour-narrowed target clause, and the same hold.
func TestTidebinderMageHoldsOnlyARedOrGreenOpposingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	red := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "Red")
	blue := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "Blue")
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
	effectstest.CastHoldCreature(t, g, "Tidebinder Mage", "Creature — Merfolk Wizard", effectstest.OracleTidebinderMage, 2, 2)
	opts := effectstest.PickTargetPrompt(t, g).PickTargetCards
	if !slices.Contains(opts, red) || slices.Contains(opts, blue) {
		t.Fatalf("Tidebinder Mage's options = %v, want the red creature and not the blue one", opts)
	}
	effectstest.AnswerPickTarget(t, g, red)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if !effectstest.TappedForTest(t, g, red) {
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
		{"Rust Tick", effectstest.OracleRustTick, "Artifact Creature — Insect", "Artifact", 1},
		{"Amber Prison", effectstest.OracleAmberPrison, "Artifact", "Basic Land — Forest", 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			source := effectstest.PushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
			target := effectstest.PushPermanentForTest(g, opp.ID, "Target", "", tc.targetType)
			effectstest.AdvanceToMain(t, g)
			for i := 0; i < tc.mana; i++ {
				effectstest.B06AddMana(me, "C")
			}
			effectstest.B16Activate(t, g, me.ID, source, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(target)})
			if !effectstest.TappedForTest(t, g, source) || !effectstest.TappedForTest(t, g, target) || effectstest.HoldCount(t, g, target) != 1 {
				t.Fatalf("after activation: source tapped %v, target tapped %v, holds %d",
					effectstest.TappedForTest(t, g, source), effectstest.TappedForTest(t, g, target), effectstest.HoldCount(t, g, target))
			}

			// Decline to untap the source: the target stays locked.
			choice := effectstest.AdvanceToUntapChoiceOf(t, g, 0)
			if err := g.ResolveUntapChoice(choice.ID, me.ID, nil); err != nil {
				t.Fatalf("decline: %v", err)
			}
			effectstest.AdvanceToUpkeepOf(t, g, 1)
			if !effectstest.TappedForTest(t, g, target) {
				t.Fatal("the target untapped while the source stayed tapped")
			}

			// Untap the source: the hold is over.
			choice = effectstest.AdvanceToUntapChoiceOf(t, g, 0)
			if err := g.ResolveUntapChoice(choice.ID, me.ID, []uuid.UUID{source}); err != nil {
				t.Fatalf("accept: %v", err)
			}
			effectstest.AdvanceToUpkeepOf(t, g, 1)
			if effectstest.TappedForTest(t, g, target) {
				t.Error("the target stayed tapped after the source untapped")
			}
		})
	}
}
