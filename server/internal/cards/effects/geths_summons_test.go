package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// geths_summons_test.go — #5860: the corrupted clauses exist only for an
// opponent who had three or more poison counters as the spell was cast
// (CR 601.2c), and that answer is fixed on the item (CR 601.2i): losing
// the poison in response keeps the target, gaining it gives none, and a
// restore mid-stack keeps the stamp.

const gethsSummonsOracle = "c345be5e-243d-44c8-a7c7-280bdb235223"

func TestGethsSummonsIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(gethsSummonsOracle)
	if !ok || spec.Name != "Geth's Summons" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registration: %+v", spec)
	}
	if n := spec.Targets.ClauseCount(); n != game.MaxPlayers {
		t.Fatalf("clauses = %d, want 1 + one per possible opponent", n)
	}
}

// gsTable is a four-seat game with the caster active, a creature card in
// the caster's graveyard and one in each opponent's, opponents listed in
// turn order after the caster.
type gsTable struct {
	g    *game.Game
	me   *game.Player
	opps []*game.Player
	mine uuid.UUID
	gy   []uuid.UUID // gy[i] is in opps[i]'s graveyard
}

func newGSTable(t *testing.T) gsTable {
	t.Helper()
	g := newCatalogGame(t)
	advanceTo(t, g, game.StepPrecombatMain)
	at := g.Turn.ActiveSeat
	tb := gsTable{g: g, me: g.Seats[at]}
	tb.mine = adGraveyardCreature(tb.me, "Mine")
	for n := 1; n < len(g.Seats); n++ {
		p := g.Seats[(at+n)%len(g.Seats)]
		tb.opps = append(tb.opps, p)
		tb.gy = append(tb.gy, adGraveyardCreature(p, "Theirs "+string(rune('0'+n))))
	}
	return tb
}

func (tb gsTable) poison(t *testing.T, opp, n int) {
	t.Helper()
	if err := tb.g.SetPoison(tb.opps[opp].ID, n); err != nil {
		t.Fatalf("SetPoison: %v", err)
	}
}

func (tb gsTable) cast(targets ...game.TargetRef) (uuid.UUID, error) {
	card := handCardForTest(tb.me, "Geth's Summons", "Sorcery", gethsSummonsOracle)
	tb.me.ManaPool.AddMana(append(cloudDeckMana("C", 2), cloudDeckMana("B", 2)...)...)
	return card, tb.g.CastSpell(tb.me.ID, card, game.CastSpellParams{Strict: true, Targets: targets})
}

func gsRef(id uuid.UUID, slot int) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id, Slot: slot}
}

func gsController(g *game.Game, id uuid.UUID) (uuid.UUID, bool) {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == id {
			return c.Controller, true
		}
	}
	return uuid.Nil, false
}

func TestGethsSummonsNoOpponentPoisonedReturnsOnlyYourOwn(t *testing.T) {
	tb := newGSTable(t)
	tb.poison(t, 0, 2) // two is not three
	if _, err := tb.cast(gsRef(tb.mine, 0), gsRef(tb.gy[0], 1)); err == nil {
		t.Fatal("an opponent with two poison counters offered a target")
	}
	if _, err := tb.cast(gsRef(tb.mine, 0)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, tb.g)
	if c, ok := gsController(tb.g, tb.mine); !ok || c != tb.me.ID {
		t.Fatalf("own creature: on battlefield %v, controller %v", ok, c)
	}
	for _, id := range tb.gy {
		if _, ok := gsController(tb.g, id); ok {
			t.Error("an opponent's creature entered")
		}
	}
}

func TestGethsSummonsOnePoisonedOpponentGivesOneTargetFromTheirGraveyardOnly(t *testing.T) {
	tb := newGSTable(t)
	tb.poison(t, 1, 3)
	// The first opponent is not poisoned: nothing of theirs is on offer.
	if _, err := tb.cast(gsRef(tb.gy[0], 1)); err == nil {
		t.Fatal("an unpoisoned opponent's graveyard was targetable")
	}
	// The poisoned opponent's clause cannot reach another graveyard.
	if _, err := tb.cast(gsRef(tb.gy[0], 2)); err == nil {
		t.Fatal("the second opponent's clause took the first opponent's card")
	}
	if _, err := tb.cast(gsRef(tb.mine, 0), gsRef(tb.gy[1], 2)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, tb.g)
	if c, ok := gsController(tb.g, tb.gy[1]); !ok || c != tb.me.ID {
		t.Fatalf("stolen creature: on battlefield %v, controller %v; want under the caster's control", ok, c)
	}
	if c, ok := gsController(tb.g, tb.mine); !ok || c != tb.me.ID {
		t.Fatalf("own creature: on battlefield %v, controller %v", ok, c)
	}
	if _, ok := gsController(tb.g, tb.gy[0]); ok {
		t.Error("the unpoisoned opponent's creature entered")
	}
}

func TestGethsSummonsTwoPoisonedOpponentsEachGiveOne(t *testing.T) {
	tb := newGSTable(t)
	tb.poison(t, 0, 3)
	tb.poison(t, 2, 7)
	// Two from one opponent's graveyard is refused: each clause is up to one.
	extra := adGraveyardCreature(tb.opps[0], "Theirs extra")
	if _, err := tb.cast(gsRef(tb.gy[0], 1), gsRef(extra, 1)); err == nil {
		t.Fatal("two targets in one opponent's clause")
	}
	if _, err := tb.cast(gsRef(tb.gy[0], 1), gsRef(tb.gy[2], 3)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, tb.g)
	for _, id := range []uuid.UUID{tb.gy[0], tb.gy[2]} {
		if c, ok := gsController(tb.g, id); !ok || c != tb.me.ID {
			t.Errorf("%v: on battlefield %v, controller %v", id, ok, c)
		}
	}
	if _, ok := gsController(tb.g, tb.mine); ok {
		t.Error("the caster's own creature entered without being targeted")
	}
}

func TestGethsSummonsKeepsATargetWhoseOwnerLostPoisonAfterTheCast(t *testing.T) {
	tb := newGSTable(t)
	tb.poison(t, 0, 3)
	if _, err := tb.cast(gsRef(tb.gy[0], 1)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	// In response the opponent loses their poison, and another gains it:
	// neither changes what was fixed as the spell was cast.
	tb.poison(t, 0, 0)
	tb.poison(t, 1, 5)
	passPriorityAroundTable(t, tb.g)
	if c, ok := gsController(tb.g, tb.gy[0]); !ok || c != tb.me.ID {
		t.Fatalf("target of a de-poisoned opponent: on battlefield %v, controller %v; want it returned", ok, c)
	}
}

func TestGethsSummonsGainingPoisonAfterTheCastOffersNoNewTarget(t *testing.T) {
	tb := newGSTable(t)
	tb.poison(t, 0, 3)
	spell, err := tb.cast(gsRef(tb.gy[0], 1))
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	tb.poison(t, 1, 5)
	// The second opponent's clause is still closed on the item: their card
	// is not legal for it, judged as the re-check and a retarget judge it.
	var legal bool
	tb.g.ReadSnapshot(func() {
		legal = tb.g.TargetStillLegalForEffect(tb.g.StackItemForEffect(spell), gsRef(tb.gy[1], 2))
	})
	if legal {
		t.Fatal("an opponent poisoned after the cast opened their clause")
	}
}

func TestGethsSummonsCastStampSurvivesARestoreMidStack(t *testing.T) {
	tb := newGSTable(t)
	tb.poison(t, 0, 3)
	if _, err := tb.cast(gsRef(tb.gy[0], 1)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	blob, err := json.Marshal(tb.g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var round game.GameSnapshot
	if err := json.Unmarshal(blob, &round); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	restored, err := round.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	// After the restore the opponent loses the poison; the target holds.
	if err := restored.SetPoison(tb.opps[0].ID, 0); err != nil {
		t.Fatalf("SetPoison: %v", err)
	}
	passPriorityAroundTable(t, restored)
	if c, ok := gsController(restored, tb.gy[0]); !ok || c != tb.me.ID {
		t.Fatalf("restored game: on battlefield %v, controller %v; want the target returned under the caster's control", ok, c)
	}
}
