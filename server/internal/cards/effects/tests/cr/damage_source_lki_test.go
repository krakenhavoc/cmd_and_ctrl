package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Warstorm Surge: a lifelinker killed in response still gains its
// controller the life its last-known power deals.
func TestWarstormSurgeDepartedLifelinkerStillGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Warstorm Surge", "Enchantment", effectstest.B06WarstormSurgeOracle, false)
	lifeBefore, oppBefore := me.Life, opp.Life

	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, game.Card{
			Name: "Beast", TypeLine: "Token Creature — Beast",
			Power: 3, Toughness: 3, Keywords: []string{"lifelink"},
		}, 1)
	})
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.RemoveInResponse(t, g, effectstest.FindBattlefieldByName(g, "Beast"), false)
	effectstest.PassPriorityAroundTable(t, g)

	if opp.Life != oppBefore-3 {
		t.Errorf("the departed 3/3 deals 3: %d → %d", oppBefore, opp.Life)
	}
	if me.Life != lifeBefore+3 {
		t.Errorf("the departed lifelinker gains its controller 3: %d → %d", lifeBefore, me.Life)
	}
	spec, _ := effects.Lookup(effectstest.B06WarstormSurgeOracle)
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Warstorm Surge is complete now: %v %v", spec.Completeness, spec.Caveats)
	}
}

// Warstorm Surge, CR 400.7: the creature is bounced in response and put
// straight back (a raw move, so the Surge does not trigger again), and
// the new object has lifelink the departed one never had. The damage is
// the departed object's, so nobody gains life.
func TestWarstormSurgeDamageIsNotTheReturnedObjects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, me.ID, "Warstorm Surge", "Enchantment", effectstest.B06WarstormSurgeOracle, false)
	elf := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: elf, Name: "Elf", TypeLine: "Creature — Elf",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, elf, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Elf: %v", err)
	}
	effectstest.PassUntilOnBattlefield(t, g, elf)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
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
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := effectstest.B04Creature(g, opp.ID, "Big Beast", 5, 5)
	redcap := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: redcap, Name: "Murderous Redcap", TypeLine: "Creature — Goblin Assassin",
		OracleID: effectstest.B40MurderousRedcapOracle, Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID, Keywords: []string{"deathtouch"},
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, redcap, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Redcap: %v", err)
	}
	effectstest.PassUntilOnBattlefield(t, g, redcap)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.RemoveInResponse(t, g, redcap, false)
	effectstest.PassPriorityAroundTable(t, g)

	if _, ok := effectstest.BattlefieldCard(g, victim); ok {
		t.Error("the 5/5 survived 2 damage from a departed deathtouch Redcap (CR 608.2h, #1396)")
	}
	spec, _ := effects.Lookup(effectstest.B40MurderousRedcapOracle)
	if len(spec.Caveats) != 0 {
		t.Errorf("persist landed with #2075, so no caveat stays: %v", spec.Caveats)
	}
}

// Mana Vault is destroyed with its draw-step trigger waiting. The
// intervening if is re-checked against its last-known tapped status, so
// it still deals its 1 damage.
func TestManaVaultThatLeftStillDealsItsDrawStepDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[0]
	vault := effectstest.PushPermanentForTest(g, owner.ID, "Mana Vault", effectstest.ManaVaultOracle, "Artifact")
	if err := g.TapCard(vault, true); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	choice := effectstest.UntapStaticPendingPay(t, g, owner.ID)
	if err := g.ResolvePayUnless(choice.ID, owner.ID, false); err != nil {
		t.Fatalf("decline Mana Vault upkeep offer: %v", err)
	}
	lifeBefore := owner.Life
	if _, err := g.AdvanceStep(); err != nil || g.Turn.Step != game.StepDraw {
		t.Fatalf("advance to draw: turn=%+v err=%v", g.Turn, err)
	}
	if effectstest.TriggerOnStack(g, vault) == nil {
		t.Fatal("tapped Mana Vault did not trigger at draw step")
	}
	effectstest.RemoveInResponse(t, g, vault, false)
	effectstest.PassPriorityAroundTable(t, g)
	if owner.Life != lifeBefore-1 {
		t.Errorf("the departed tapped Vault deals 1: life %d → %d", lifeBefore, owner.Life)
	}
}
