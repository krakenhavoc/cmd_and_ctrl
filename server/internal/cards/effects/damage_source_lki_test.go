package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// damage_source_lki_test.go is the card-level proof for #1396: damage
// from a permanent that has LEFT the battlefield carries the lifelink
// and deathtouch it had as it last existed (CR 608.2h), and the object a
// trigger named is never confused with a new object the same card has
// become (CR 400.7). The #1379 half — the last-known POWER — is
// resolution_lki_test.go.

const manaVaultOracle = "736892cb-a34b-4bb9-b56c-e26e3db207a2"

// Warstorm Surge: a lifelinker killed in response still gains its
// controller the life its last-known power deals.
func TestWarstormSurgeDepartedLifelinkerStillGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Warstorm Surge", "Enchantment", b06WarstormSurgeOracle, false)
	lifeBefore, oppBefore := me.Life, opp.Life

	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, game.Card{
			Name: "Beast", TypeLine: "Token Creature — Beast",
			Power: 3, Toughness: 3, Keywords: []string{"lifelink"},
		}, 1)
	})
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	removeInResponse(t, g, findBattlefieldByName(g, "Beast"), false)
	passPriorityAroundTable(t, g)

	if opp.Life != oppBefore-3 {
		t.Errorf("the departed 3/3 deals 3: %d → %d", oppBefore, opp.Life)
	}
	if me.Life != lifeBefore+3 {
		t.Errorf("the departed lifelinker gains its controller 3: %d → %d", lifeBefore, me.Life)
	}
	spec, _ := Lookup(b06WarstormSurgeOracle)
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Warstorm Surge is complete now: %v %v", spec.Completeness, spec.Caveats)
	}
}

// Warstorm Surge, CR 400.7: the creature is bounced in response and put
// straight back (a raw move, so the Surge does not trigger again), and
// the new object has lifelink the departed one never had. The damage is
// the departed object's, so nobody gains life.
func TestWarstormSurgeDamageIsNotTheReturnedObjects(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushCatalogPermanent(g, me.ID, "Warstorm Surge", "Enchantment", b06WarstormSurgeOracle, false)
	elf := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: elf, Name: "Elf", TypeLine: "Creature — Elf",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, elf, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Elf: %v", err)
	}
	passUntilOnBattlefield(t, g, elf)
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	lifeBefore, oppBefore := me.Life, opp.Life

	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(elf); err != nil {
			t.Fatalf("BounceToHandForEffect: %v", err)
		}
		if _, err := game.MoveCard(me.Hand, g.Battlefield, elf); err != nil {
			t.Fatalf("MoveCard back: %v", err)
		}
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == elf {
				g.Battlefield.Cards[i].Keywords = []string{"lifelink"}
			}
		}
		g.BumpLayerVersionForTest()
		g.RecomputeLayersIfStaleLocked()
	})
	passPriorityAroundTable(t, g)

	if opp.Life != oppBefore-1 {
		t.Errorf("the departed 1/1 deals 1: %d → %d", oppBefore, opp.Life)
	}
	if me.Life != lifeBefore {
		t.Errorf("the returned Elf's lifelink must not count — it is a new object "+
			"that dealt nothing: %d → %d", lifeBefore, me.Life)
	}
}

// Murderous Redcap with deathtouch (the Basilisk Collar line — the
// keyword rides the card here, standing in for the equipment's grant) is
// sacrificed in response to its own enter trigger. Its ping still has
// deathtouch: the 5/5 dies to 2.
func TestMurderousRedcapDepartedDeathtouchStillKills(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := b04Creature(g, opp.ID, "Big Beast", 5, 5)
	redcap := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: redcap, Name: "Murderous Redcap", TypeLine: "Creature — Goblin Assassin",
		OracleID: b40MurderousRedcapOracle, Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID, Keywords: []string{"deathtouch"},
	})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, redcap, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Redcap: %v", err)
	}
	passUntilOnBattlefield(t, g, redcap)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	removeInResponse(t, g, redcap, false)
	passPriorityAroundTable(t, g)

	if _, ok := battlefieldCard(g, victim); ok {
		t.Error("the 5/5 survived 2 damage from a departed deathtouch Redcap (CR 608.2h, #1396)")
	}
	spec, _ := Lookup(b40MurderousRedcapOracle)
	if len(spec.Caveats) != 1 {
		t.Errorf("only the persist caveat stays: %v", spec.Caveats)
	}
}

// Mana Vault is destroyed with its draw-step trigger waiting. The
// intervening if is re-checked against its last-known tapped status, so
// it still deals its 1 damage.
func TestManaVaultThatLeftStillDealsItsDrawStepDamage(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[0]
	vault := pushPermanentForTest(g, owner.ID, "Mana Vault", manaVaultOracle, "Artifact")
	if err := g.TapCard(vault, true); err != nil {
		t.Fatal(err)
	}
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	choice := untapStaticPendingPay(t, g, owner.ID)
	if err := g.ResolvePayUnless(choice.ID, owner.ID, false); err != nil {
		t.Fatalf("decline Mana Vault upkeep offer: %v", err)
	}
	lifeBefore := owner.Life
	if _, err := g.AdvanceStep(); err != nil || g.Turn.Step != game.StepDraw {
		t.Fatalf("advance to draw: turn=%+v err=%v", g.Turn, err)
	}
	if triggerOnStack(g, vault) == nil {
		t.Fatal("tapped Mana Vault did not trigger at draw step")
	}
	removeInResponse(t, g, vault, false)
	passPriorityAroundTable(t, g)
	if owner.Life != lifeBefore-1 {
		t.Errorf("the departed tapped Vault deals 1: life %d → %d", lifeBefore, owner.Life)
	}
}
