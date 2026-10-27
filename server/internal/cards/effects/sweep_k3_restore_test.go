package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sweep_k3_restore_test.go — #4253: three keyed engine forms, each moved
// card paused on a CR 616 ordering prompt, restored STRICTLY (the boot
// path) and answered on the restored table:
//
//   - "exile target spell" with its "then" as data (ExileTargetSpell.ThenTail,
//     Game.ExileSpellKeyedForEffect): Mindbreak Trap's run over its targets
//     and Aven Interrupter's plot;
//   - ExileUntil.ThenTail, riding the record of the return: Abdel Adrian's
//     count of the permanents exiled this way;
//   - damage to a player with its "then" as data, the damage event's frame
//     written to the restore point (Game.DealDamageToPlayerKeyedForEffect):
//     Screaming Nemesis's "can't gain life".

// nemesisDamageOrderProbeOracles are two test cards, each "if a source named
// Screaming Nemesis would deal damage to a player, it deals that much
// instead": two declared effects, so the Nemesis's damage to a player asks
// for their order (CR 616.1). Catalog rows, so a restored table gathers
// them again.
var nemesisDamageOrderProbeOracles = []string{"test-4253-k3-damage-order-a", "test-4253-k3-damage-order-b"}

func init() {
	for i, oracle := range nemesisDamageOrderProbeOracles {
		Register(Spec{
			OracleID: oracle,
			Name:     "Damage Order Probe " + string(rune('A'+i)),
			Replacements: []game.ReplacementEffect{{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
					if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 {
						return false
					}
					if _, isCard := g.LookupCardForEffect(ev.DamageTarget); isCard {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.DamageSource)
					return ok && c.Name == "Screaming Nemesis"
				},
				Replace: func(*game.ReplacementEvent, *game.Game, *game.Card) error { return nil },
				Label:   "Damage Order Probe " + string(rune('A'+i)),
			}},
		})
	}
}

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
	g := newCatalogGame(t)
	me := g.Seats[0]
	needle := castCatalogSpell(t, g, "Needle", "Instant", "", nil)
	other := castCatalogSpell(t, g, "Spell B", "Instant", "", nil)
	seedExileOrderProbes(g, me.ID)
	castCatalogSpell(t, g, "Mindbreak Trap", "Instant — Trap", sweepFMindbreakOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: needle}, {Kind: game.TargetCard, ID: other}})
	passPriorityAroundTable(t, g)

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	if !g.Stack.Contains(needle) || !g.Stack.Contains(other) {
		t.Fatal("a target left the stack before the restored answer")
	}
	answerExileOrder(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	needle := castCatalogSpell(t, g, "Needle", "Sorcery", "", nil)
	seedExileOrderProbes(g, me.ID)
	aven := castCatalogSpell(t, g, "Aven Interrupter", "Creature — Bird Rogue", avenInterrupterOracle, nil)
	passUntilOnBattlefield(t, g, aven)
	pickCard(t, g, me.ID, needle)
	passPriorityAroundTable(t, g)

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	if !g.Stack.Contains(needle) {
		t.Fatal("the spell left the stack before the restored answer")
	}
	answerExileOrder(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	needle := pushVanillaCreature(g, me.ID, "Needle", 1, 1)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	seedExileOrderProbes(g, me.ID)
	abdel := castCatalogSpell(t, g, "Abdel Adrian, Gorion's Ward", "Legendary Creature — Human Warrior", mcAbdelOracle, nil)
	passPriorityAroundTable(t, g)
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no choose-cards prompt")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{needle, bear}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	me = seatByID(g, me.ID)
	if !g.Battlefield.Contains(needle) || !g.Battlefield.Contains(bear) {
		t.Fatal("a pick left the battlefield before the restored answer")
	}
	answerExileOrder(t, g)
	if !g.Exile.Contains(needle) || !g.Exile.Contains(bear) {
		t.Fatalf("exiled Needle=%v Bear=%v, want both", g.Exile.Contains(needle), g.Exile.Contains(bear))
	}
	if n := countBattlefieldNamed(g, me.ID, "Soldier"); n != 2 {
		t.Errorf("Soldiers = %d, want 2", n)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(abdel) })
	g.RunStateChecksForTest()
	for _, name := range []string{"Needle", "Bear"} {
		if _, ok := e2BattlefieldNamed(g, name); !ok {
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nemesis := pushPermanentForTest(g, me.ID, "Screaming Nemesis", screamingNemesisOracle, "Creature — Spirit")
	seedOrderProbes(g, opp.ID, nemesisDamageOrderProbeOracles, "Damage Order Probe")
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: uuid.Nil, Target: nemesis, Amount: 2})
	})
	g.RunStateChecksForTest()
	answerPickTargetPlayer(t, g, opp.ID)
	life := opp.Life
	passPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Fatalf("the damage landed before the ordering prompt was answered: %d → %d", life, opp.Life)
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeDamage)
	answerExileOrder(t, g)
	if got := lifeTotalOf(g, opp.ID); got != life-2 {
		t.Errorf("opponent at %d after the restored answer, want %d", got, life-2)
	}
	if gains(t, g, opp.ID) {
		t.Error("the damaged player gained life: the keyed tail did not run off the disk")
	}
	if !gains(t, g, me.ID) {
		t.Error("the Nemesis's controller can't gain life")
	}
}

// A prevented reflection is told it dealt nothing: no "can't gain life".
func TestAScreamingNemesisKeyedTailHearsZeroWhenNothingIsDealt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		ctx := keyedContext(g, me.ID, uuid.Nil)
		tail := ctx.Tail(b35NemesisCantGainLife)
		tail.Carry = []uuid.UUID{opp.ID}
		if err := g.DealDamageToPlayerKeyedForEffect(uuid.Nil, opp.ID, 0, tail); err != nil {
			t.Fatalf("DealDamageToPlayerKeyedForEffect: %v", err)
		}
	})
	if !gains(t, g, opp.ID) {
		t.Error("zero damage barred the player from gaining life")
	}
}
