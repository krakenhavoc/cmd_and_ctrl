package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// multi_block_leftovers_test.go — the remaining #1706 proof cards left
// over from #1713's batch: Kemba's Legion, Cenn's Tactician, Foriysian
// Totem, Iona's Blessing and Vanguard's Shield. Same helpers and same
// shape as multi_block_cards_test.go: each blocks as many attackers as
// its own line grants, and refuses one past capacity. The card under
// test is always controlled by `opp`, attacked by `me` (the active
// seat), the same convention Palace Guard and High Ground use.
//
// Two cards (Iona's Blessing, Vanguard's Shield) grant their bonus
// through a sorcery-speed action (casting an Aura, equipping) that
// only the active player may take, so their setup runs during OPP's
// own turn (advanceToMainOf) before crossing into `me`'s next turn to
// attack — the bonus itself is not until-end-of-turn, so it survives
// the turn change. Foriysian Totem's animation IS until end of turn,
// so it and its attack must stay on the SAME turn; its activated
// ability carries no "activate only as a sorcery" clause (it is
// instant speed, as printed), so `opp` activates it without becoming
// the active player.

const (
	kembasLegionOracle    = "338ad8c4-c89f-4e78-be28-7224fac0fd0f"
	cennsTacticianOracle  = "9c4a3ca6-dcf9-4986-81da-bcfd46414bee"
	foriysianTotemOracle  = "d053ea00-e727-4030-8c64-0d32ae57f169"
	ionasBlessingOracle   = "fc5bde87-231e-424a-a8fa-cd9910a09d43"
	vanguardsShieldOracle = "bc628018-f5d0-4bd2-91e7-ee4e4a5af9aa"
)

// TestKembasLegionBlocksAnAdditionalCreaturePerEquipment — the count is
// live off the number of Equipment attached, and adds up.
func TestKembasLegionBlocksAnAdditionalCreaturePerEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	legion := pushCatalogPermanent(g, opp.ID, "Kemba's Legion", "Creature — Cat Soldier", kembasLegionOracle, false)
	swordA := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sword A", TypeLine: "Artifact — Equipment", Owner: opp.ID, Controller: opp.ID,
	})
	swordB := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sword B", TypeLine: "Artifact — Equipment", Owner: opp.ID, Controller: opp.ID,
	})
	if got := mbCapacity(t, g, legion); got != 1 {
		t.Fatalf("unequipped capacity = %d, want 1", got)
	}
	if !hasString(effectiveAbilities(t, g, legion), "vigilance") {
		t.Error("no vigilance")
	}

	g.WithWriteLock(func() {
		if err := g.AttachForEffect(swordA, game.TargetRef{Kind: game.TargetCard, ID: legion}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
	})
	if got := mbCapacity(t, g, legion); got != 2 {
		t.Fatalf("capacity with one Equipment attached = %d, want 2", got)
	}
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(swordB, game.TargetRef{Kind: game.TargetCard, ID: legion}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
	})
	if got := mbCapacity(t, g, legion); got != 3 {
		t.Fatalf("capacity with two Equipment attached = %d, want 3", got)
	}

	atks := mbAttack(t, g, me, opp, 4)
	if err := mbBlock(g, legion, atks[:3]...); err != nil {
		t.Fatalf("blocking three: %v", err)
	}
	if err := mbBlock(g, legion, atks[3]); err == nil {
		t.Fatal("a fourth block was accepted")
	}
}

// TestCennsTacticiansCounterGrantsTheExtraBlock — only a creature that
// actually carries a +1/+1 counter gets the extra block, read live off
// the counter rather than off the ability that placed it.
func TestCennsTacticiansCounterGrantsTheExtraBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tactician := pushCatalogPermanent(g, opp.ID, "Cenn's Tactician", "Creature — Kithkin Soldier", cennsTacticianOracle, false)
	soldier := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Foot Soldier", TypeLine: "Creature — Human Soldier",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	if got := mbCapacity(t, g, soldier); got != 1 {
		t.Fatalf("capacity with no counter = %d, want 1", got)
	}

	dcMana(t, g, opp.ID, "{W}")
	if err := g.ActivateCatalogAbility(opp.ID, tactician, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: soldier}},
	}); err != nil {
		t.Fatalf("activate Cenn's Tactician: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := dcCard(t, g, soldier).Counters[game.CounterPlusOne]; n != 1 {
		t.Fatalf("+1/+1 counters on the Soldier = %d, want 1", n)
	}
	if got := mbCapacity(t, g, soldier); got != 2 {
		t.Fatalf("capacity with a +1/+1 counter = %d, want 2", got)
	}

	atks := mbAttack(t, g, me, opp, 3)
	if err := mbBlock(g, soldier, atks[:2]...); err != nil {
		t.Fatalf("blocking two: %v", err)
	}
	if err := mbBlock(g, soldier, atks[2]); err == nil {
		t.Fatal("a third block was accepted")
	}
}

// TestForiysianTotemBlocksAnAdditionalCreatureOnlyWhileAnimated — the
// block line is gated on the middle ability's animation, live, and
// gone again once the animation ends.
func TestForiysianTotemBlocksAnAdditionalCreatureOnlyWhileAnimated(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	totem := pushCatalogPermanent(g, opp.ID, "Foriysian Totem", "Artifact", foriysianTotemOracle, false)
	if dcCard(t, g, totem).IsCreature() {
		t.Fatal("setup: Foriysian Totem starts as a plain artifact")
	}

	dcMana(t, g, opp.ID, "{C}{C}{C}{C}{R}")
	if err := g.ActivateCatalogAbility(opp.ID, totem, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the animation ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := dcCard(t, g, totem)
	if !c.IsCreature() || c.Effective().Power != 4 || c.Effective().Toughness != 4 {
		t.Fatalf("Foriysian Totem is not a 4/4 creature: %+v", c.Effective())
	}
	if !containsString(c.Effective().Colors, "R") {
		t.Errorf("colours = %v, want red", c.Effective().Colors)
	}
	if !hasString(effectiveAbilities(t, g, totem), "trample") {
		t.Error("no trample")
	}
	if got := mbCapacity(t, g, totem); got != 2 {
		t.Fatalf("capacity while animated = %d, want 2", got)
	}

	atks := mbAttack(t, g, me, opp, 3)
	if err := mbBlock(g, totem, atks[:2]...); err != nil {
		t.Fatalf("blocking two: %v", err)
	}
	if err := mbBlock(g, totem, atks[2]); err == nil {
		t.Fatal("a third block was accepted")
	}

	dcCleanup(g)
	if dcCard(t, g, totem).IsCreature() {
		t.Fatal("after cleanup Foriysian Totem is still a creature")
	}
}

// TestIonasBlessingPumpsGrantsVigilanceAndTheExtraBlock — the Aura's
// three clauses land on the enchanted creature at once. Cast during
// opp's own turn (Auras are sorcery speed), then crosses into a later
// turn for the attack — the bonus is not until-end-of-turn, so it
// survives the turn change.
func TestIonasBlessingPumpsGrantsVigilanceAndTheExtraBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	if got := mbCapacity(t, g, bear); got != 1 {
		t.Fatalf("unenchanted capacity = %d, want 1", got)
	}

	advanceToMainOf(t, g, 1)
	blessing := castCatalogSpell(t, g, "Iona's Blessing", auraTypeLine, ionasBlessingOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, blessing); host.Kind != game.TargetCard || host.ID != bear {
		t.Fatalf("Iona's Blessing AttachedTo = %+v, want the Bear", host)
	}
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tg != 4 {
		t.Fatalf("enchanted Bear is %d/%d, want 4/4", p, tg)
	}
	if !hasString(effectiveAbilities(t, g, bear), "vigilance") {
		t.Error("no vigilance")
	}
	if got := mbCapacity(t, g, bear); got != 2 {
		t.Fatalf("enchanted capacity = %d, want 2", got)
	}

	advanceToMainOf(t, g, 0)
	atks := mbAttack(t, g, me, opp, 3)
	if err := mbBlock(g, bear, atks[:2]...); err != nil {
		t.Fatalf("blocking two: %v", err)
	}
	if err := mbBlock(g, bear, atks[2]); err == nil {
		t.Fatal("a third block was accepted")
	}
}

// TestVanguardsShieldPumpsAndGrantsTheExtraBlock — Bonesplitter's
// shape plus the block line. Equipped during opp's own turn (equip is
// sorcery speed), then crosses into a later turn for the attack.
func TestVanguardsShieldPumpsAndGrantsTheExtraBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shield := pushCatalogPermanent(g, opp.ID, "Vanguard's Shield", "Artifact — Equipment", vanguardsShieldOracle, false)
	host := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	if got := mbCapacity(t, g, host); got != 1 {
		t.Fatalf("unequipped capacity = %d, want 1", got)
	}

	advanceToMainOf(t, g, 1)
	floatMana(t, g, opp, "{C}{C}{C}")
	equipTo(t, g, opp.ID, shield, host)
	if p, tg := effectivePower(t, g, host), effectiveToughness(t, g, host); p != 2 || tg != 5 {
		t.Fatalf("equipped Bear is %d/%d, want 2/5", p, tg)
	}
	if got := mbCapacity(t, g, host); got != 2 {
		t.Fatalf("equipped capacity = %d, want 2", got)
	}

	advanceToMainOf(t, g, 0)
	atks := mbAttack(t, g, me, opp, 3)
	if err := mbBlock(g, host, atks[:2]...); err != nil {
		t.Fatalf("blocking two: %v", err)
	}
	if err := mbBlock(g, host, atks[2]); err == nil {
		t.Fatal("a third block was accepted")
	}
}
