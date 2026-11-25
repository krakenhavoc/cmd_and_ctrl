package teststo

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tribute_test.go — tribute (CR 702.104, #6251) through the real entry
// door: a cast creature with "Tribute 3" and "When this creature enters,
// if tribute wasn't paid, you gain 4 life" (Snake of the Golden Grove's
// shape, on a test oracle).
//
//	702.104a "Tribute N" means "As this creature enters, choose an
//	         opponent. That player may put an additional N +1/+1
//	         counters on it as it enters."
//	702.104b "if tribute wasn't paid" is true if the opponent chosen
//	         didn't have the creature enter with the counters.

const testTributeOracle = "test-6251-tribute-snake"

func registerTributeSnake(t *testing.T) {
	t.Helper()
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID:     testTributeOracle,
		Name:         "Test Tribute Snake",
		Completeness: effects.CompletenessFull,
		Replacements: effects.Tribute("Test Tribute Snake", 3),
		Triggered: []game.TriggeredAbility{
			effects.WhenEntersIfTributeWasntPaid("Test Tribute Snake — you gain 4 life", effects.Do(effects.GainLife{Amount: 4})),
		},
	})
}

// castTributeSnake casts the snake from the active seat and lets it
// resolve, which stops on the tribute's first question.
func castTributeSnake(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := effectstest.CastCatalogSpell(t, g, "Test Tribute Snake", "Creature — Snake", testTributeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func openTributeChoice(t *testing.T, g *game.Game, kind game.PendingChoiceKind) *game.PendingChoice {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == kind {
			return c
		}
	}
	var got []game.PendingChoiceKind
	for _, c := range g.PendingChoices {
		got = append(got, c.Kind)
	}
	t.Fatalf("no open %s prompt; open: %v", kind, got)
	return nil
}

// chooseTributeOpponent answers the controller's "choose an opponent"
// with `want`.
func chooseTributeOpponent(t *testing.T, g *game.Game, want uuid.UUID) {
	t.Helper()
	c := openTributeChoice(t, g, game.PendingChoiceEntryController)
	for i, opt := range c.PickOptions {
		if opt.Player == want {
			if err := g.ResolveEntryController(c.ID, c.Chooser, i); err != nil {
				t.Fatalf("ResolveEntryController: %v", err)
			}
			return
		}
	}
	t.Fatalf("seat %s is not offered", want)
}

// CR 702.104a: the controller chooses among the opponents, the creature
// has not entered yet, and the CHOSEN OPPONENT is asked whether to pay.
// Paying puts the three counters on it as it enters, leaves it under its
// caster, and the "if tribute wasn't paid" ability does not trigger.
func TestTributePaidByTheChosenOpponent(t *testing.T) {
	registerTributeSnake(t)
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+2)%4]
	life := caster.Life
	id := castTributeSnake(t, g)

	c := openTributeChoice(t, g, game.PendingChoiceEntryController)
	if c.Chooser != caster.ID {
		t.Fatalf("the opponent is chosen by %s, want the caster", c.Chooser)
	}
	if c.ControlPurpose != game.ControlForTribute {
		t.Fatalf("purpose = %q, want tribute", c.ControlPurpose)
	}
	if len(c.PickOptions) != 3 {
		t.Fatalf("options = %d, want the three opponents", len(c.PickOptions))
	}
	if effects.OnBattlefield(g, id) {
		t.Fatal("the creature is on the battlefield before its tribute was settled (CR 614.12a)")
	}
	chooseTributeOpponent(t, g, opp.ID)

	pay := openTributeChoice(t, g, game.PendingChoiceOptionalReplacement)
	if pay.Chooser != opp.ID {
		t.Fatalf("the payment is asked of %s, want the chosen opponent", pay.Chooser)
	}
	if effects.OnBattlefield(g, id) {
		t.Fatal("the creature is on the battlefield before the opponent answered")
	}
	if err := g.ResolveOptionalReplacement(pay.ID, opp.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	card, ok := effectstest.BattlefieldCardByID(g, id)
	if !ok {
		t.Fatal("the creature never entered")
	}
	if card.Controller != caster.ID {
		t.Fatalf("controller = %s, want the caster: tribute gives nothing away", card.Controller)
	}
	if n := card.Counters[game.CounterPlusOne]; n != 3 {
		t.Fatalf("+1/+1 counters = %d, want 3 (tribute 3 paid)", n)
	}
	if !g.TributePaidFor(id) {
		t.Fatal("TributePaid = false after the opponent paid")
	}
	if caster.Life != life {
		t.Fatalf("life %d -> %d: the \"tribute wasn't paid\" ability resolved after it was paid", life, caster.Life)
	}
}

// CR 702.104b: declining leaves the creature without counters, and the
// ETB ability triggers and resolves.
func TestTributeDeclinedTriggersTheNotPaidAbility(t *testing.T) {
	registerTributeSnake(t)
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	life := caster.Life
	id := castTributeSnake(t, g)

	chooseTributeOpponent(t, g, opp.ID)
	pay := openTributeChoice(t, g, game.PendingChoiceOptionalReplacement)
	if err := g.ResolveOptionalReplacement(pay.ID, opp.ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	card, ok := effectstest.BattlefieldCardByID(g, id)
	if !ok {
		t.Fatal("the creature never entered")
	}
	if n := card.Counters[game.CounterPlusOne]; n != 0 {
		t.Fatalf("+1/+1 counters = %d, want none (tribute declined)", n)
	}
	if g.TributePaidFor(id) {
		t.Fatal("TributePaid = true after the opponent declined")
	}
	if caster.Life != life+4 {
		t.Fatalf("life = %d, want %d: the \"tribute wasn't paid\" ability did not resolve", caster.Life, life+4)
	}
}

// CR 603.4 / 608.2h: the intervening "if" is re-checked as the trigger
// resolves, from the creature's last-known information when it has left.
// A snake bounced in response still gains its controller the life.
func TestTributeNotPaidReadsLastKnownInformation(t *testing.T) {
	registerTributeSnake(t)
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+3)%4]
	life := caster.Life
	id := castTributeSnake(t, g)
	chooseTributeOpponent(t, g, opp.ID)
	pay := openTributeChoice(t, g, game.PendingChoiceOptionalReplacement)
	if err := g.ResolveOptionalReplacement(pay.ID, opp.ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	// The ETB trigger is pending or on the stack: bounce the snake first.
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneHand, Owner: caster.ID}, id); err != nil {
		t.Fatalf("bounce: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if caster.Life != life+4 {
		t.Fatalf("life = %d, want %d: the trigger lost its condition with its creature", caster.Life, life+4)
	}
}
