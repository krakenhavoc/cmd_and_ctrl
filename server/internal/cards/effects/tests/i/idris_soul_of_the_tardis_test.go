package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	idrisOracle   = "abb5210f-597c-450d-9239-6253a9136680"
	idrisTypeLine = "Legendary Creature — Human Incarnation"
)

func TestIdrisIsFullWithNoCaveats(t *testing.T) {
	spec, ok := effects.Lookup(idrisOracle)
	if !ok || spec.Name != "Idris, Soul of the TARDIS" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// idrisEnters casts Idris and answers its imprint with `pick`.
func idrisEnters(t *testing.T, g *game.Game, me *game.Player, pick uuid.UUID) uuid.UUID {
	t.Helper()
	idris := effectstest.RfdCast(t, g, "Idris, Soul of the TARDIS", idrisTypeLine, idrisOracle, 3, 3)
	// Soul Warden triggers on Idris's arrival too: order the two triggers.
	effectstest.SettleOrderingTriggers(t, g)
	// A lone candidate may be taken without a prompt.
	effectstest.AnswerPermanentPicks(t, g, me.ID, map[uuid.UUID][]uuid.UUID{me.ID: {pick}})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(pick) {
		t.Fatal("Idris did not exile the artifact")
	}
	return idris
}

// TestIdrisBorrowsTheExiledArtifactsTriggerAndGetsItsManaValue: the imprint
// exile is linked (CR 607.2a), Idris has the exiled card's triggered ability
// (ADR 0155's 2026-11-05 amendment) and +X/+X for its mana value; when Idris
// leaves, the card returns (CR 610.3) and Vanishing 3 put three time
// counters on Idris as it entered (CR 702.63a).
func TestIdrisBorrowsTheExiledArtifactsTriggerAndGetsItsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	// A Soul Warden printed as an artifact creature: its catalog row is the
	// lent trigger, and {W} is a mana value of 1.
	warden := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Soul Warden", TypeLine: "Artifact Creature — Human Cleric", ManaCost: "{W}",
		OracleID: effectstest.B04SoulWardenOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	idris := idrisEnters(t, g, me, warden)

	c := effectstest.ElExiledCard(t, g, warden)
	if c.ExiledWith != effectstest.ElRef(t, g, idris) {
		t.Fatalf("the artifact is exiled linked to %+v, want this Idris object", c.ExiledWith)
	}
	v := effectstest.GaLayered(t, g, idris)
	if v.CurrentPower() != 4 || v.CurrentToughness() != 4 {
		t.Errorf("Idris is %d/%d, want 4/4 (+1/+1 for the exiled card's mana value)", v.CurrentPower(), v.CurrentToughness())
	}
	if got := v.Counters[game.CounterTime]; got != 3 {
		t.Errorf("Idris entered with %d time counters, want 3", got)
	}

	life := me.Life
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, game.Card{Name: "Idris Test Bear", TypeLine: "Token Creature — Bear", Power: 2, Toughness: 2}, 1)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Fatalf("life %d → %d, want the lent \"whenever another creature enters, you gain 1 life\"", life, me.Life)
	}

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(idris); err != nil {
			t.Fatalf("destroy Idris: %v", err)
		}
	})
	g.RunStateChecksForTest()
	effectstest.SettleOrderingTriggers(t, g)
	if g.Exile.Contains(warden) {
		t.Error("the exiled card is still in exile after Idris left")
	}
	back := 0
	for _, b := range g.Battlefield.Cards {
		if b.Name == "Soul Warden" && b.Controller == me.ID {
			back++
		}
	}
	if back != 1 {
		t.Errorf("%d Soul Wardens on the battlefield after Idris left, want the exiled card back", back)
	}
}

// TestIdrisBorrowsAnActivatedAbilityAndOnlyFromAnArtifact: a non-artifact is
// not a candidate, and the exiled artifact's mana ability is lent.
func TestIdrisBorrowsAnActivatedAbilityAndOnlyFromAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	elves := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Llanowar Elves", TypeLine: "Artifact Creature — Elf Druid", ManaCost: "{G}",
		OracleID: effectstest.BlLlanowarOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	idris := effectstest.RfdCast(t, g, "Idris, Soul of the TARDIS", idrisTypeLine, idrisOracle, 3, 3)
	var ask *game.PendingChoice
	g.ReadSnapshot(func() {
		for _, pc := range g.PendingChoices {
			if pc != nil && pc.Kind == game.PendingChoiceOwnPermanents {
				ask = pc
			}
		}
	})
	if ask == nil {
		t.Fatal("Idris asked for no artifact to exile")
	}
	if err := g.ResolveOwnPermanents(ask.ID, me.ID, []uuid.UUID{bear}); err == nil {
		t.Error("a non-artifact creature was accepted")
	}
	if effectstest.E2Zone(g, bear) == game.ZoneExile {
		t.Fatal("a non-artifact creature was exiled")
	}
	effectstest.AnswerPermanentPicks(t, g, me.ID, map[uuid.UUID][]uuid.UUID{me.ID: {elves}})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(elves) {
		t.Fatal("the artifact was not exiled")
	}
	if idx, _ := effectstest.BlBorrowedMana(t, g, idris); len(idx) != 1 {
		t.Fatalf("Idris has %d lent mana rows, want the exiled card's {T}: Add {G}", len(idx))
	}
}
