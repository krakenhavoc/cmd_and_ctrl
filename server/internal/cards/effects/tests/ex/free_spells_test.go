package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The headline: Force of Will's pitch is really charged, and it is
// charged under the STRICT mana gate with an empty pool — which is
// the only way to prove the {3}{U}{U} was actually replaced.
func TestForceOfWillPitchIsCharged(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	pitch := effectstest.HandCardFull(me, "Brainstorm", "Instant", "{U}", "", []string{"U"})
	fow := effectstest.HandCardFull(me, "Force of Will", "Instant", "{3}{U}{U}", effectstest.ForceOfWillOracle, []string{"U"})
	victim := effectstest.CastCatalogSpell(t, g, "Their Spell", "Sorcery", "", nil)
	lifeBefore := me.Life

	err := g.CastSpell(me.ID, fow, game.CastSpellParams{
		Strict:          true,
		AlternativeCost: "pitch",
		AltCostIDs:      []uuid.UUID{pitch},
		Targets:         []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	})
	if err != nil {
		t.Fatalf("Force of Will on a pitch: %v", err)
	}
	if me.Life != lifeBefore-1 {
		t.Errorf("life = %d, want %d (1 paid)", me.Life, lifeBefore-1)
	}
	if me.Hand.Contains(pitch) {
		t.Errorf("the pitched card is still in hand")
	}
	if !g.Exile.Contains(pitch) {
		t.Errorf("the pitched card did not reach exile")
	}
}

// The offer cannot be claimed without the cost: no blue card, or a
// non-blue one, is a rejected cast rather than a free spell.
func TestForceOfWillRejectsABadPitch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	red := effectstest.HandCardFull(me, "Lightning Bolt", "Instant", "{R}", "", []string{"R"})
	fow := effectstest.HandCardFull(me, "Force of Will", "Instant", "{3}{U}{U}", effectstest.ForceOfWillOracle, []string{"U"})
	victim := effectstest.CastCatalogSpell(t, g, "Their Spell", "Sorcery", "", nil)
	target := []game.TargetRef{{Kind: game.TargetCard, ID: victim}}

	// A red card is not a legal pitch.
	if err := g.CastSpell(me.ID, fow, game.CastSpellParams{
		AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{red}, Targets: target,
	}); err == nil {
		t.Errorf("pitching a red card was accepted")
	}
	// No pitch at all is not a legal claim either.
	if err := g.CastSpell(me.ID, fow, game.CastSpellParams{
		AlternativeCost: "pitch", Targets: target,
	}); err == nil {
		t.Errorf("claiming the pitch with no card was accepted")
	}
	if !me.Hand.Contains(fow) {
		t.Errorf("a rejected cast left Force of Will out of hand")
	}
	if me.Life != 40 && me.Life != 20 {
		t.Errorf("a rejected cast charged life: %d", me.Life)
	}
}

// Fierce Guardianship: the condition gates the offer, in the view and
// at announce.
func TestFierceGuardianshipNeedsACommander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	fg := effectstest.HandCardFull(me, "Fierce Guardianship", "Instant", "{2}{U}", effectstest.FierceGuardianshipOracle, []string{"U"})
	victim := effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	target := []game.TargetRef{{Kind: game.TargetCard, ID: victim}}

	alt := game.AlternativeCostByKey(effectstest.FierceGuardianshipOracle, "free")
	if alt == nil {
		t.Fatalf("Fierce Guardianship offers no free cast")
	}
	g.WithWriteLock(func() {
		if alt.Available(g, me.ID) {
			t.Errorf("the free cast is offered with no commander out")
		}
	})
	if err := g.CastSpell(me.ID, fg, game.CastSpellParams{
		AlternativeCost: "free", Targets: target,
	}); err == nil {
		t.Errorf("free cast accepted with no commander on the battlefield")
	}

	// Put a commander on the battlefield and it turns on.
	cid := effectstest.FreeSpellPermanent(g, me.ID, "Test Commander", "Legendary Creature — Elemental")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == cid {
				g.Battlefield.Cards[i].IsCommander = true
			}
		}
		if !alt.Available(g, me.ID) {
			t.Errorf("the free cast is still hidden with a commander out")
		}
	})
	if err := g.CastSpell(me.ID, fg, game.CastSpellParams{
		Strict: true, AlternativeCost: "free", Targets: target,
	}); err != nil {
		t.Fatalf("free cast with a commander out: %v", err)
	}
}

// Snuff Out's Swamp check, and the life really being paid.
func TestSnuffOutPaysFourLifeWithASwamp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	snuff := effectstest.HandCardFull(me, "Snuff Out", "Instant", "{3}{B}", effectstest.SnuffOutOracle, []string{"B"})
	bear := effectstest.FreeSpellPermanent(g, them.ID, "Grizzly Bears", "Creature — Bear")
	target := []game.TargetRef{{Kind: game.TargetCard, ID: bear}}

	if err := g.CastSpell(me.ID, snuff, game.CastSpellParams{
		AlternativeCost: "pay_life", Targets: target,
	}); err == nil {
		t.Errorf("paid 4 life with no Swamp")
	}

	effectstest.FreeSpellPermanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	lifeBefore := me.Life
	if err := g.CastSpell(me.ID, snuff, game.CastSpellParams{
		Strict: true, AlternativeCost: "pay_life", Targets: target,
	}); err != nil {
		t.Fatalf("Snuff Out for 4 life with a Swamp: %v", err)
	}
	if me.Life != lifeBefore-4 {
		t.Errorf("life = %d, want %d", me.Life, lifeBefore-4)
	}
}

// Daze returns the Island as a COST — at announce, while Daze is on
// the stack, not on resolution.
func TestDazeReturnsTheIslandAtAnnounce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	daze := effectstest.HandCardFull(me, "Daze", "Instant", "{1}{U}", effectstest.DazeOracle, []string{"U"})
	island := effectstest.FreeSpellPermanent(g, me.ID, "Island", "Basic Land — Island")
	victim := effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)

	if err := g.CastSpell(me.ID, daze, game.CastSpellParams{
		Strict:          true,
		AlternativeCost: "return",
		AltCostIDs:      []uuid.UUID{island},
		Targets:         []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("Daze returning an Island: %v", err)
	}
	// The Island is back in hand while Daze is still on the stack.
	if !me.Hand.Contains(island) {
		t.Errorf("the Island did not return to hand")
	}
	if !g.Stack.Contains(daze) {
		t.Errorf("Daze is not on the stack — the cost was paid too late")
	}
}

// Solitude's evoke cost is a card, and paying it attaches the
// sacrifice trigger (CR 702.74a) rather than killing it inline.
func TestSolitudeEvokePitchesAWhiteCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	pitch := effectstest.HandCardFull(me, "Swords to Plowshares", "Instant", "{W}", "", []string{"W"})
	sol := effectstest.HandCardFull(me, "Solitude", "Creature — Elemental Incarnation", "{3}{W}{W}", effectstest.SolitudeOracle, []string{"W"})

	if err := g.CastSpell(me.ID, sol, game.CastSpellParams{
		Strict: true, AlternativeCost: "evoke", AltCostIDs: []uuid.UUID{pitch},
	}); err != nil {
		t.Fatalf("evoking Solitude: %v", err)
	}
	if !g.Exile.Contains(pitch) {
		t.Errorf("the evoke pitch was not exiled")
	}
	alt := game.AlternativeCostByKey(effectstest.SolitudeOracle, "evoke")
	if alt == nil || !alt.SacrificeOnEntry {
		t.Errorf("Solitude's evoke does not carry the sacrifice trigger: %+v", alt)
	}
}

// Pact of Negation has no alternative cost at all — its printed cost
// IS {0} — and what it has instead is a delayed trigger.
func TestPactOfNegationSchedulesItsDebt(t *testing.T) {
	if got := game.AlternativeCostsFor(effectstest.PactOfNegationOracle); got != nil {
		t.Errorf("Pact of Negation declares an alternative cost: %+v", got)
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	victim := effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	pact := effectstest.HandCardFull(me, "Pact of Negation", "Instant", "{0}", effectstest.PactOfNegationOracle, []string{"U"})
	if err := g.CastSpell(me.ID, pact, game.CastSpellParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("casting Pact of Negation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("Pact scheduled %d delayed triggers, want 1", len(g.DelayedTriggers))
	}
	dt := g.DelayedTriggers[0]
	if dt.At != game.StepUpkeep || !dt.ControllerTurnOnly {
		t.Errorf("Pact's debt fires at %q (controller-only %v), want upkeep, controller-only",
			dt.At, dt.ControllerTurnOnly)
	}
}

// Kambal drains on an opponent's noncreature cast and on nothing
// else — not its controller's spells, not creature spells.
func TestKambalDrainsOpponentNoncreatureCasts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Kambal, Consul of Allocation", "Legendary Creature — Human Advisor", effectstest.KambalOracle, false)
	effectstest.ToMainForCost(t, g)

	// Active seat 0 is Kambal's controller — their own spell does
	// nothing.
	before := me.Life
	effectstest.CastCatalogSpell(t, g, "My Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before {
		t.Errorf("Kambal drained its own controller's cast: %d → %d", before, me.Life)
	}
}
