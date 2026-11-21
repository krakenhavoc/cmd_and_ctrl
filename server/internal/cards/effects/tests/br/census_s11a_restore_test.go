package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// census_s11a_restore_test.go — #4361 S11a: five engine route batches
// whose continuation was a closure (zoneRoute.then) now ride
// routeAllKeyedLocked with a registered tail. Each exiles a card named
// Needle, so the two exile-order probes ask for their order (CR 616.1);
// the table is restored STRICTLY (the boot path) while that prompt is open
// and the prompt is answered on the restored table:
//
//   - suspend (engine/suspend-exiled): the card lands with its time counters;
//   - plot (engine/plot-exiled): the card lands plotted;
//   - foretell (engine/foretold-exiled): the card lands foretold and castable
//     on a later turn for its foretell cost;
//   - "exile the top card … until you exile another" (Unstable Amulet,
//     engine/exile-until-another-granted): the card lands playable;
//   - Chandra, Torch of Defiance's paid cast "or else"
//     (engine/paid-cast-or-else-exiled): the offer opens over the card.

// s11aExiled is the card in exile with this ID, or nil.
func s11aExiled(g *game.Game, id uuid.UUID) *game.Card {
	for i := range g.Exile.Cards {
		if g.Exile.Cards[i].InstanceID == id {
			return &g.Exile.Cards[i]
		}
	}
	return nil
}

// s11aPausedOnTheOrder checks the special action or ability left the card
// where it was, waiting on the ordering prompt.
func s11aPausedOnTheOrder(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	if g.Exile.Contains(id) {
		t.Fatal("the card reached exile before the ordering prompt was answered")
	}
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceReplacementOrder {
		t.Fatalf("pending choices %+v, want the exile's ordering prompt", g.PendingChoices)
	}
}

func TestASuspendExileOrderIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	id := effectstest.HandCardForTest(me, "Needle", "Creature — Horror", effectstest.NghcNihilith)
	effectstest.SeedExileOrderProbes(g, me.ID)
	me.ManaPool.AddMana(game.ManaToken{Color: "B"}, game.ManaToken{Color: "C"})
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionSuspend, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("suspend: %v", err)
	}
	s11aPausedOnTheOrder(t, g, id)

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	effectstest.AnswerExileOrder(t, g)
	c := s11aExiled(g, id)
	if c == nil {
		t.Fatal("the suspended card is not in exile after the restored answer")
	}
	if got := c.Counters[game.CounterTime]; got != 7 {
		t.Errorf("time counters = %d, want Nihilith's 7", got)
	}
	if !game.CardIsSuspended(*c) {
		t.Error("the card is not suspended")
	}
}

func TestAPlotExileOrderIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.AdvanceToMainOf(t, g, seat)
	id := effectstest.PlotHandCard(me, "Needle", "Creature — Djinn", "{4}{U}", effectstest.DjinnOfFoolsFallOracle, 4, 3)
	effectstest.SeedExileOrderProbes(g, me.ID)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "U"})
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionPlot, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("plot: %v", err)
	}
	s11aPausedOnTheOrder(t, g, id)

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	effectstest.AnswerExileOrder(t, g)
	if !g.Exile.Contains(id) {
		t.Fatal("the plotted card is not in exile after the restored answer")
	}
	perm := g.CastPermissionOnCardByIDForEffect(id)
	if perm == nil || !perm.Granted() || perm.Player != me.ID || perm.Timing != game.TimingPlot {
		t.Fatalf("plot permission = %+v, want the owner's TimingPlot grant off the disk", perm)
	}
}

func TestAForetellExileOrderIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	id := effectstest.HandCardForTest(me, "Needle", "Instant", effectstest.SawItComingOracle)
	effectstest.SeedExileOrderProbes(g, me.ID)
	seq := g.Turn.Seq
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionForetell, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("foretell: %v", err)
	}
	s11aPausedOnTheOrder(t, g, id)
	if g.ForetoldThisTurn[me.ID] != 0 {
		t.Fatal("the card counted as foretold before it landed")
	}

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	effectstest.AnswerExileOrder(t, g)
	c := s11aExiled(g, id)
	if c == nil || !c.FaceDown {
		t.Fatalf("the foretold card after the restored answer = %+v, want it face down in exile", c)
	}
	if got := g.ForetoldThisTurn[me.ID]; got != 1 {
		t.Errorf("foretold this turn = %d, want 1", got)
	}
	perm := g.CastPermissionOnCardByIDForEffect(id)
	if perm == nil || !perm.Granted() || perm.Player != me.ID || perm.AltCostKey != game.AltCostKeyForetell ||
		perm.Cost != "{1}{U}" || perm.NotBeforeSeq != seq+1 || !perm.CastOnly {
		t.Fatalf("foretell permission = %+v, want the owner's {1}{U} cast from turn %d", perm, seq+1)
	}
}

func TestAnUnstableAmuletExileOrderIsARestorePoint(t *testing.T) {
	g, me, _, amulet := effectstest.EuaAmulet(t, 1)
	id := effectstest.EuaTop(g, me, "Needle", "Instant", "{R}")
	effectstest.SeedExileOrderProbes(g, me.ID)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, amulet).Tapped = false })
	if err := g.ActivateCatalogAbility(me.ID, amulet, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	s11aPausedOnTheOrder(t, g, id)

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	me = effectstest.SeatByID(g, me.ID)
	effectstest.AnswerExileOrder(t, g)
	if !g.Exile.Contains(id) {
		t.Fatal("the top card is not in exile after the restored answer")
	}
	if !effectstest.EuaPlayable(g, me, id) {
		t.Error("the exiled card is not playable: the keyed grant did not run off the disk")
	}
	perm := g.CastPermissionOnCardByIDForEffect(id)
	if perm == nil || perm.Duration.Kind != game.UntilSourceExilesAnother || perm.Duration.Source != amulet {
		t.Errorf("permission = %+v, want an until-another window naming the Amulet", perm)
	}
}

func TestAChandraTorchExileOrderIsARestorePoint(t *testing.T) {
	spell := uuid.New()
	g, me, chandra := chandraTorchTable(t, 4, &game.Card{InstanceID: spell, Name: "Needle", TypeLine: "Sorcery", ManaCost: "{2}{U}"})
	effectstest.SeedExileOrderProbes(g, me.ID)
	before := effectstest.OpponentLives(g, me.ID)
	effectstest.B16Activate(t, g, me.ID, chandra, 0, game.ActivateAbilityParams{})
	s11aPausedOnTheOrder(t, g, spell)

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	effectstest.AnswerExileOrder(t, g)
	if !g.Exile.Contains(spell) {
		t.Fatal("the top card is not in exile after the restored answer")
	}
	prompt := effectstest.KeywordCastPromptFor(g, me.ID)
	if prompt == nil || prompt.Kind != game.PendingChoiceResolutionCast || effectstest.KeywordCastCard(prompt) != spell ||
		prompt.MayCastKeyword != game.MayCastKeywordPaidOrElse {
		t.Fatalf("a paid_or_else resolution_cast over the exiled card: %+v", prompt)
	}
	if prompt.Source != chandra {
		t.Errorf("the offer's source = %s, want Chandra", prompt.Source)
	}
	effectstest.AssertOpponentsLost(t, g, me.ID, before, 0)
	effectstest.AnswerMayCastPrompt(t, g, me.ID, false)
	effectstest.AssertOpponentsLost(t, g, me.ID, before, 2)
}
