package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pay_action_restore_test.go — #4361 S8: a non-mana pay-unless's "yes"
// waits for its payment as data (game/pay_action_keyed.go), so a payment
// paused on a CR 616 ordering prompt is a restore point. Each test pays,
// stops on the ordering prompt the payment asks, takes a STRICT restore
// (the boot path), answers on the restored table, and checks the "yes"
// ran off the disk.

// payLinksUntilPaused answers the open any-player offer's links, `who`
// paying with `cards` and everyone else declining, until no link is open:
// the payment has started and stopped on a prompt of its own.
func payLinksUntilPaused(t *testing.T, g *game.Game, who uuid.UUID, cards ...uuid.UUID) {
	t.Helper()
	pay := appOnly(who, cards...)
	for i := 0; i < 8; i++ {
		link := appOpenLink(g)
		if link == nil {
			return
		}
		yes, ids := pay(link)
		if err := g.ResolvePayUnlessWithCards(link.ID, link.Chooser, yes, ids); err != nil {
			t.Fatalf("answer link %d: %v", i, err)
		}
	}
	t.Fatal("the offer never ended")
}

// paidBy reports whether the table logged an any-player payment by `who`.
func paidBy(g *game.Game, who uuid.UUID) bool {
	for _, ev := range g.Events {
		if ev.Kind == game.EventAnyPlayerPaid && ev.Actor == who {
			return true
		}
	}
	return false
}

// Browbeat: an opponent who holds the damage-order probes takes the 5, the
// damage asks them for the order, and the restored answer deals it and then
// runs the offer's "paid" (logged, and no one draws three).
func TestABrowbeatDamagePaymentOrderIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedDamageOrderProbes(g, opp.ID)
	life := opp.Life
	castCatalogSpell(t, g, "Browbeat", "Sorcery", appBrowbeat, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	passPriorityAroundTable(t, g)
	payLinksUntilPaused(t, g, opp.ID)
	if opp.Life != life || paidBy(g, opp.ID) {
		t.Fatal("the damage or the payment's \"yes\" happened before the ordering prompt was answered")
	}
	hand := me.Hand.Size()

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeDamage)
	answerExileOrder(t, g)
	passPriorityAroundTable(t, g)
	if got := lifeTotalOf(g, opp.ID); got != life-5 {
		t.Errorf("payer at %d after the restored answer, want %d", got, life-5)
	}
	if !paidBy(g, opp.ID) {
		t.Error("the offer's \"paid\" did not run after the restored answer: the payment's tail did not come off the disk")
	}
	if got := seatByID(g, me.ID).Hand.Size(); got != hand {
		t.Errorf("target's hand %d → %d: Browbeat's draw ran although a player took the damage", hand, got)
	}
}

// Carrion Rats: the defender exiles a Needle from their graveyard, the
// exile asks for the order, and the restored answer exiles it and then
// runs the offer's "if a player does" (the Rats assign no combat damage).
func TestACarrionRatsExilePaymentOrderIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rats := carrionPush(g, me, "Carrion Rats", carrionRatsOracle, 2, 1)
	needle := game.NewCard("Needle", opp.ID)
	needle.TypeLine = "Sorcery"
	opp.Graveyard.PushTop(needle)
	seedExileOrderProbes(g, opp.ID)
	carrionAttack(t, g, rats, opp)
	payLinksUntilPaused(t, g, opp.ID, needle.InstanceID)
	if g.Exile.Contains(needle.InstanceID) || g.PermanentAssignsNoCombatDamageForEffect(rats) {
		t.Fatal("the exile or the payment's \"yes\" happened before the ordering prompt was answered")
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	answerExileOrder(t, g)
	if !g.Exile.Contains(needle.InstanceID) {
		t.Error("the Needle was not exiled by the restored answer")
	}
	if !g.PermanentAssignsNoCombatDamageForEffect(rats) {
		t.Error("the Rats are not marked after the restored answer: the payment's tail did not come off the disk")
	}
}
