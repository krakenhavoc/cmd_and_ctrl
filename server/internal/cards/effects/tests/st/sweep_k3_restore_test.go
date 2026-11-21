package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// k3InExileWithNoRecord reports that a spell went to exile and took its
// stack record with it.
func k3InExileWithNoRecord(g *game.Game, id uuid.UUID) bool {
	_, onStack := g.StackMeta[id]
	return g.Exile.Contains(id) && !onStack && !g.Stack.Contains(id)
}

// --- Mindbreak Trap ----------------------------------------------------------

// Two target spells, the first a Needle: its exile asks for the order, the
// restored table still has both on the stack, and the answer exiles the
// Needle and then the second spell, which rode the tail's Carry.
func TestAMindbreakTrapRunIsARestorePointAtItsFirstExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	needle := effectstest.CastCatalogSpell(t, g, "Needle", "Instant", "", nil)
	other := effectstest.CastCatalogSpell(t, g, "Spell B", "Instant", "", nil)
	effectstest.SeedExileOrderProbes(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Mindbreak Trap", "Instant — Trap", effectstest.SweepFMindbreakOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: needle}, {Kind: game.TargetCard, ID: other}})
	effectstest.PassPriorityAroundTable(t, g)

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	if !g.Stack.Contains(needle) || !g.Stack.Contains(other) {
		t.Fatal("a target left the stack before the restored answer")
	}
	effectstest.AnswerExileOrder(t, g)
	if !k3InExileWithNoRecord(g, needle) {
		t.Error("the Needle was not exiled after the restored answer")
	}
	if !k3InExileWithNoRecord(g, other) {
		t.Error("the second target, the rest of the run, was not exiled")
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompts open after the run finished", len(g.PendingChoices))
	}
}

// --- Aven Interrupter --------------------------------------------------------

// Aven Interrupter over a Needle: the exile's ordering prompt is a restore
// point, and the restored answer exiles the spell and plots it.
func TestAnAvenInterrupterExileIsARestorePointAndStillPlots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	needle := effectstest.CastCatalogSpell(t, g, "Needle", "Sorcery", "", nil)
	effectstest.SeedExileOrderProbes(g, me.ID)
	aven := effectstest.CastCatalogSpell(t, g, "Aven Interrupter", "Creature — Bird Rogue", effectstest.AvenInterrupterOracle, nil)
	effectstest.PassUntilOnBattlefield(t, g, aven)
	effectstest.PickCard(t, g, me.ID, needle)
	effectstest.PassPriorityAroundTable(t, g)

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	if !g.Stack.Contains(needle) {
		t.Fatal("the spell left the stack before the restored answer")
	}
	effectstest.AnswerExileOrder(t, g)
	if !k3InExileWithNoRecord(g, needle) {
		t.Fatal("the spell was not exiled after the restored answer")
	}
	perm := g.CastPermissionOnCardByIDForEffect(needle)
	if !perm.Granted() || perm.Player != me.ID || perm.Timing != game.TimingPlot {
		t.Fatalf("plot permission = %+v, want the owner's TimingPlot grant off the disk", perm)
	}
}

// --- Abdel Adrian, Gorion's Ward ---------------------------------------------

// Abdel Adrian over a Needle and a Bear: the Needle's exile asks for the
// order, the restored answer exiles it and then the Bear (the picks left
// ride the tail's Carry), makes a Soldier for each, and both come back when
// Abdel leaves (the record rode the keyed exile).
func TestAnAbdelAdrianExileIsARestorePointAndStillCounts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	needle := effectstest.PushVanillaCreature(g, me.ID, "Needle", 1, 1)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.SeedExileOrderProbes(g, me.ID)
	abdel := effectstest.CastCatalogSpell(t, g, "Abdel Adrian, Gorion's Ward", "Legendary Creature — Human Warrior", effectstest.McAbdelOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no choose-cards prompt")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{needle, bear}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	me = effectstest.SeatByID(g, me.ID)
	if !g.Battlefield.Contains(needle) || !g.Battlefield.Contains(bear) {
		t.Fatal("a pick left the battlefield before the restored answer")
	}
	effectstest.AnswerExileOrder(t, g)
	if !g.Exile.Contains(needle) || !g.Exile.Contains(bear) {
		t.Fatalf("exiled Needle=%v Bear=%v, want both", g.Exile.Contains(needle), g.Exile.Contains(bear))
	}
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Soldier"); n != 2 {
		t.Errorf("Soldiers = %d, want 2", n)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(abdel) })
	g.RunStateChecksForTest()
	for _, name := range []string{"Needle", "Bear"} {
		if _, ok := effectstest.E2BattlefieldNamed(g, name); !ok {
			t.Errorf("%s did not return when Abdel left", name)
		}
	}
}

// --- Screaming Nemesis -------------------------------------------------------

// Screaming Nemesis dealt 2: its reflected 2 at an opponent asks for the
// order of two damage effects, the damage frame is written to the restore
// point, and the restored answer deals the 2 and bars the opponent from
// gaining life, the keyed tail off the disk.
func TestAScreamingNemesisDamageOrderIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nemesis := effectstest.PushPermanentForTest(g, me.ID, "Screaming Nemesis", effectstest.ScreamingNemesisOracle, "Creature — Spirit")
	effectstest.SeedOrderProbes(g, opp.ID, effectstest.NemesisDamageOrderProbeOracles, "Damage Order Probe")
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: uuid.Nil, Target: nemesis, Amount: 2})
	})
	g.RunStateChecksForTest()
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	life := opp.Life
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Fatalf("the damage landed before the ordering prompt was answered: %d → %d", life, opp.Life)
	}

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeDamage)
	effectstest.AnswerExileOrder(t, g)
	if got := effectstest.LifeTotalOf(g, opp.ID); got != life-2 {
		t.Errorf("opponent at %d after the restored answer, want %d", got, life-2)
	}
	if effectstest.Gains(t, g, opp.ID) {
		t.Error("the damaged player gained life: the keyed tail did not run off the disk")
	}
	if !effectstest.Gains(t, g, me.ID) {
		t.Error("the Nemesis's controller can't gain life")
	}
}

// A prevented reflection is told it dealt nothing: no "can't gain life".
func TestAScreamingNemesisKeyedTailHearsZeroWhenNothingIsDealt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		ctx := effects.KeyedContext(g, me.ID, uuid.Nil)
		tail := ctx.Tail(effects.B35NemesisCantGainLife)
		tail.Carry = []uuid.UUID{opp.ID}
		if err := g.DealDamageToPlayerKeyedForEffect(uuid.Nil, opp.ID, 0, tail); err != nil {
			t.Fatalf("DealDamageToPlayerKeyedForEffect: %v", err)
		}
	})
	if !effectstest.Gains(t, g, opp.ID) {
		t.Error("zero damage barred the player from gaining life")
	}
}
