package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// carrionGraveyard puts n filler cards into p's graveyard.
func carrionGraveyard(p *game.Player, n int) []uuid.UUID {
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		c := game.NewCard("Filler", p.ID)
		c.TypeLine = "Sorcery"
		p.Graveyard.PushTop(c)
		ids = append(ids, c.InstanceID)
	}
	return ids
}

func carrionSeatIDs(ps ...*game.Player) []uuid.UUID {
	var ids []uuid.UUID
	for _, p := range ps {
		ids = append(ids, p.ID)
	}
	return ids
}

// With nobody paying the Rats deal their 2; every player is asked, the
// active player first.
func TestCarrionRatsDealTheirDamageWhenNobodyPays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rats := effectstest.CarrionPush(g, me, "Carrion Rats", effectstest.CarrionRatsOracle, 2, 1)
	carrionGraveyard(opp, 2)
	before := opp.Life
	effectstest.CarrionAttack(t, g, rats, opp)
	asked := effectstest.AppAnswerWith(t, g, effectstest.AppOnly(uuid.Nil))
	if !effectstest.SameIDs(asked, carrionSeatIDs(g.Seats...)) {
		t.Errorf("asked %v, want every seat in turn order", asked)
	}
	if g.PermanentAssignsNoCombatDamageForEffect(rats) {
		t.Error("nobody paid, and the Rats are marked")
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	if opp.Life != before-2 {
		t.Errorf("defender life %d → %d, want the Rats' 2", before, opp.Life)
	}
}

// An opponent exiling a card from their own graveyard ends the offer and
// the Rats deal no combat damage.
func TestCarrionRatsAssignNoDamageOnceAnOpponentExilesACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rats := effectstest.CarrionPush(g, me, "Carrion Rats", effectstest.CarrionRatsOracle, 2, 1)
	mine := carrionGraveyard(me, 1)
	theirs := carrionGraveyard(opp, 2)
	before := opp.Life
	effectstest.CarrionAttack(t, g, rats, opp)
	asked := effectstest.AppAnswerWith(t, g, effectstest.AppOnly(opp.ID, theirs[0]))
	if !effectstest.SameIDs(asked, carrionSeatIDs(me, opp)) {
		t.Errorf("asked %v, want the controller then the defender, and nobody after the payer", asked)
	}
	if !g.Exile.Contains(theirs[0]) || !opp.Graveyard.Contains(theirs[1]) || !me.Graveyard.Contains(mine[0]) {
		t.Error("only the payer's chosen card leaves a graveyard")
	}
	if !g.PermanentAssignsNoCombatDamageForEffect(rats) {
		t.Fatal("the Rats are not marked")
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	if opp.Life != before {
		t.Errorf("defender life %d → %d, want no damage", before, opp.Life)
	}
}

// A payer's graveyard is their own: a card in someone else's can't be
// named, and an empty graveyard can't pay (CR 118.3).
func TestCarrionRatsCantBePaidFromAnotherPlayersGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rats := effectstest.CarrionPush(g, me, "Carrion Rats", effectstest.CarrionRatsOracle, 2, 1)
	theirs := carrionGraveyard(opp, 1)
	effectstest.CarrionAttack(t, g, rats, opp)
	link := effectstest.AppOpenLink(g)
	if link.Chooser != me.ID {
		t.Fatalf("first link is for %v, want the active player", link.Chooser)
	}
	if err := g.ResolvePayUnlessWithCards(link.ID, link.Chooser, true, theirs); err == nil {
		t.Fatal("naming another player's graveyard card was accepted")
	}
	// The controller's graveyard is empty, so their "yes" is a decline and
	// the offer moves on.
	if err := g.ResolvePayUnlessWithCards(link.ID, link.Chooser, true, nil); err != nil {
		t.Fatalf("empty-graveyard yes: %v", err)
	}
	if g.PermanentAssignsNoCombatDamageForEffect(rats) {
		t.Error("an empty graveyard paid")
	}
	if next := effectstest.AppOpenLink(g); next == nil || next.Chooser != opp.ID {
		t.Fatalf("the offer did not move on to the defender: %+v", next)
	}
}

// The trigger fires on blocking too, and the Rats' controller isn't the
// active player: the attacker's controller is asked first.
func TestCarrionRatsTriggerOnBlockingToo(t *testing.T) {
	for _, paid := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		atkSeat, blkSeat := g.Seats[1], g.Seats[0]
		effectstest.AangAdvanceToMain(t, g, 1)
		raider := effectstest.CarrionPush(g, atkSeat, "Raider", "", 3, 3)
		rats := effectstest.CarrionPush(g, blkSeat, "Carrion Rats", effectstest.CarrionRatsOracle, 2, 1)
		fuel := carrionGraveyard(atkSeat, 1)
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(raider, blkSeat.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		if err := g.DeclareBlocker(rats, raider); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
		effectstest.LockInBlocks(t, g)
		effectstest.CarrionToOffer(t, g)
		if link := effectstest.AppOpenLink(g); link.Chooser != atkSeat.ID {
			t.Fatalf("first link is for %v, want the active player", link.Chooser)
		}
		who := uuid.Nil
		if paid {
			who = atkSeat.ID
		}
		effectstest.AppAnswerWith(t, g, effectstest.AppOnly(who, fuel...))
		effectstest.AdvanceTo(t, g, game.StepCombatDamage)
		c := effectstest.FindBattlefieldCardForTest(g, raider)
		if c == nil {
			t.Fatalf("paid=%v: the raider died", paid)
		}
		if want := map[bool]int{false: 2, true: 0}[paid]; c.DamageMarked != want {
			t.Errorf("paid=%v: the raider took %d from the blocking Rats, want %d", paid, c.DamageMarked, want)
		}
	}
}

// CR 400.7: Rats that left and came back before the payment are a new
// object, and the payment doesn't touch them.
func TestCarrionRatsThatCameBackAreUnaffectedByThePayment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rats := effectstest.CarrionPush(g, me, "Carrion Rats", effectstest.CarrionRatsOracle, 2, 1)
	fuel := carrionGraveyard(opp, 1)
	effectstest.CarrionAttack(t, g, rats, opp)
	g.WithWriteLock(func() {
		if err := (effects.Flicker{Target: rats}).Apply(effectstest.CtxFor(g, &game.StackItem{Controller: me.ID, SourceCardID: uuid.New()})); err != nil {
			t.Fatalf("Flicker: %v", err)
		}
	})
	back := effectstest.FindBattlefieldByName(g, "Carrion Rats")
	if back == uuid.Nil || back == rats {
		t.Fatalf("the Rats did not come back as a new object (%v → %v)", rats, back)
	}
	effectstest.AppAnswerWith(t, g, effectstest.AppOnly(opp.ID, fuel...))
	if !g.Exile.Contains(fuel[0]) {
		t.Error("the payment was still made")
	}
	if g.PermanentAssignsNoCombatDamageForEffect(back) || g.PermanentAssignsNoCombatDamageForEffect(rats) {
		t.Error("the new object is marked")
	}
}

// The Wurm costs three cards: a payer with two can't pay, a payer with
// three exiles exactly three.
func TestCarrionWurmNeedsThreeCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, short, deep := g.Seats[0], g.Seats[1], g.Seats[2]
	wurm := effectstest.CarrionPush(g, me, "Carrion Wurm", effectstest.CarrionWurmOracle, 6, 5)
	carrionGraveyard(short, 2)
	cards := carrionGraveyard(deep, 4)
	before := short.Life
	effectstest.CarrionAttack(t, g, wurm, short)
	link := effectstest.AppOpenLink(g)
	if link.PayCost != "exile three cards from your graveyard" {
		t.Errorf("offer words %q", link.PayCost)
	}
	asked := effectstest.AppAnswerWith(t, g, func(c *game.PendingChoice) (bool, []uuid.UUID) {
		switch c.Chooser {
		case short.ID:
			return true, nil // two cards: "pay" is a decline
		case deep.ID:
			return true, cards[:3]
		}
		return false, nil
	})
	if !effectstest.SameIDs(asked, carrionSeatIDs(me, short, deep)) {
		t.Errorf("asked %v, want the controller, the short graveyard, then the payer", asked)
	}
	for i, id := range cards {
		if got, want := g.Exile.Contains(id), i < 3; got != want {
			t.Errorf("card %d exiled=%v, want %v", i, got, want)
		}
	}
	if !g.PermanentAssignsNoCombatDamageForEffect(wurm) {
		t.Fatal("the Wurm is not marked")
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	if short.Life != before {
		t.Errorf("defender life %d → %d, want no damage", before, short.Life)
	}
}

// One card of three is not a payment.
func TestCarrionWurmRefusesAPaymentOfFewerThanThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wurm := effectstest.CarrionPush(g, me, "Carrion Wurm", effectstest.CarrionWurmOracle, 6, 5)
	carrionGraveyard(me, 3)
	effectstest.CarrionAttack(t, g, wurm, opp)
	link := effectstest.AppOpenLink(g)
	if err := g.ResolvePayUnlessWithCards(link.ID, link.Chooser, true, []uuid.UUID{me.Graveyard.Cards[0].InstanceID}); err == nil {
		t.Fatal("a one-card payment was accepted for a three-card price")
	}
}
