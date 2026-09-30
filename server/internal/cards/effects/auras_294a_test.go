package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// auras_294a_test.go — slice 294-a, the Auras. Each test asserts
// behaviour: what the enchanted permanent is, what it can do, and what
// happens when the Aura or its host leaves.

const (
	spiritMantleOracle         = "81a51328-b995-4f3b-90bc-a20ae21dd254"
	etherealArmorOracle        = "dbba75f5-2404-4bd5-982b-f6c4effa5316"
	aqueousFormOracle          = "3378fef3-4d8c-4d9e-a338-4abb6f70d410"
	sagesReverieOracle         = "bd92cacd-c7d8-43a2-a261-ea4754fefdc5"
	freedFromTheRealOracle     = "03d32ace-dd99-48c6-bcae-9b915d041aef"
	shinyImpetusOracle         = "aae76e5c-f5e0-4d18-b465-e6a829be908a"
	stickyFingersOracle        = "fa8ddc13-5108-4fe9-b3aa-dc28df476f7b"
	imprisonedInTheMoonOracle  = "339a7418-9a29-4eae-a3c2-ea590d175936"
	amphibianDownpourOracle    = "a4280cea-386b-450d-a7e9-29c7aa58ce5a"
	kayasGhostformOracle       = "b15c6203-6311-44eb-916d-a4c271960c74"
	giftOfImmortalityOracle    = "2c11c72f-5b04-47d8-bdf8-8d9d197ebc1b"
	shelteredByGhostsOracle    = "d13fc657-c6fc-4394-bc70-691050550226"
	mantleOfTheAncientsOracle  = "7424560f-557f-4bc9-a3e7-eb890c73aaa3"
	mechanizedProductionOracle = "39d8406a-90c3-460c-b4e8-0f590573db51"
)

// auraCast casts the named Aura from the active seat's hand onto
// `host` and settles the spell and whatever it triggers.
func auraCast(t *testing.T, g *game.Game, name, oracle string, host uuid.UUID) uuid.UUID {
	t.Helper()
	aura := castCatalogSpell(t, g, name, auraTypeLine, oracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: host}})
	passPriorityAroundTable(t, g)
	return aura
}

// auraBear puts a vanilla 2/2 that has been in play a while under owner.
func auraBear(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
}

// auraRestrictions is the effective restriction set of a permanent.
func auraRestrictions(t *testing.T, g *game.Game, id uuid.UUID) game.Restriction {
	t.Helper()
	var r game.Restriction
	var found bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				r, found = c.Effective().Restrictions, true
				return
			}
		}
	})
	if !found {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return r
}

// auraSeedLibrary gives a seat enough library to draw from.
func auraSeedLibrary(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		pushLibraryCardForTest(p, game.Card{Name: "Filler", TypeLine: "Sorcery"})
	}
}

// --- Spirit Mantle ------------------------------------------------

func TestSpiritMantlePumpsAndGivesProtectionFromCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := auraBear(g, me.ID)
	auraCast(t, g, "Spirit Mantle", spiritMantleOracle, bear)

	if p, tt := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("enchanted bear is %d/%d, want 3/3", p, tt)
	}
	if !containsString(effectiveAbilities(t, g, bear), "protection from creatures") {
		t.Errorf("abilities %v lack protection from creatures", effectiveAbilities(t, g, bear))
	}
	var prot bool
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(bear)
		prot = game.ProtectedFrom(&c, &game.Characteristic{Types: []string{"Creature"}})
	})
	if !prot {
		t.Error("the enchanted creature is not protected from a creature source")
	}
}

// --- Ethereal Armor -----------------------------------------------

func TestEtherealArmorCountsYourEnchantmentsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := auraBear(g, me.ID)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Enchantment", TypeLine: "Enchantment",
		Owner: me.ID, Controller: me.ID,
	})
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Enchantment", TypeLine: "Enchantment",
		Owner: opp.ID, Controller: opp.ID,
	})
	auraCast(t, g, "Ethereal Armor", etherealArmorOracle, bear)

	// Two enchantments are mine: the Armor itself and one other.
	if p, tt := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tt != 4 {
		t.Errorf("enchanted bear is %d/%d, want 4/4", p, tt)
	}
	if !containsString(effectiveAbilities(t, g, bear), "first strike") {
		t.Errorf("abilities %v lack first strike", effectiveAbilities(t, g, bear))
	}
}

// --- Aqueous Form -------------------------------------------------

func TestAqueousFormMakesTheCreatureUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := auraBear(g, me.ID)
	if auraRestrictions(t, g, bear)&game.CantBeBlocked != 0 {
		t.Fatal("setup: the bear is already unblockable")
	}
	aura := auraCast(t, g, "Aqueous Form", aqueousFormOracle, bear)

	if auraRestrictions(t, g, bear)&game.CantBeBlocked == 0 {
		t.Error("the enchanted creature can be blocked")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	if auraRestrictions(t, g, bear)&game.CantBeBlocked != 0 {
		t.Error("the creature is still unblockable after the Aura left")
	}
}

func TestAqueousFormScriesWhenTheEnchantedCreatureAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	auraSeedLibrary(me, 4)
	bear := auraBear(g, me.ID)
	aura := auraCast(t, g, "Aqueous Form", aqueousFormOracle, bear)

	declareAttack(t, g, opp.ID, bear)
	if triggerOnStack(g, aura) == nil {
		t.Fatal("the attack put no Aqueous Form trigger on the stack")
	}
	passPriorityAroundTable(t, g)
	answerScryKeepAll(t, g, me.ID)
}

// --- Sage's Reverie -----------------------------------------------

func TestSagesReverieDrawsAndPumpsPerAuraOnCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	auraSeedLibrary(me, 6)
	bear := auraBear(g, me.ID)
	auraCast(t, g, "Spirit Mantle", spiritMantleOracle, bear)
	handBefore := me.Hand.Size()

	auraCast(t, g, "Sage's Reverie", sagesReverieOracle, bear)

	// Two Auras of mine are on a creature (Spirit Mantle and the
	// Reverie): two cards drawn, +2/+2 from the Reverie, +1/+1 from
	// the Mantle.
	if got := me.Hand.Size() - handBefore; got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
	if p, tt := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 5 || tt != 5 {
		t.Errorf("enchanted bear is %d/%d, want 5/5", p, tt)
	}
}

func TestSagesReverieIgnoresAurasOnLandsAndOpponentsAuras(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	auraSeedLibrary(me, 6)
	bear := auraBear(g, me.ID)
	land := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
		Owner: me.ID, Controller: me.ID,
	})
	// An Aura of mine on a land, and an opponent's Aura on my creature.
	onLand := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Wild Growth", TypeLine: auraTypeLine,
		Owner: me.ID, Controller: me.ID,
	})
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Aura", TypeLine: auraTypeLine,
		Owner: opp.ID, Controller: opp.ID,
	})
	g.WithWriteLock(func() {
		_ = g.AttachForEffect(onLand, game.TargetRef{Kind: game.TargetCard, ID: land})
		_ = g.AttachForEffect(theirs, game.TargetRef{Kind: game.TargetCard, ID: bear})
	})
	handBefore := me.Hand.Size()

	auraCast(t, g, "Sage's Reverie", sagesReverieOracle, bear)

	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("drew %d cards, want 1 (only the Reverie counts)", got)
	}
	if p := effectivePower(t, g, bear); p != 3 {
		t.Errorf("power %d, want 3", p)
	}
}

// --- Freed from the Real ------------------------------------------

func TestFreedFromTheRealTapsAndUntapsTheEnchantedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := auraBear(g, opp.ID)
	aura := auraCast(t, g, "Freed from the Real", freedFromTheRealOracle, theirs)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}"); err != nil {
		t.Fatalf("fund: %v", err)
	}
	tapped := func() bool {
		var v bool
		g.ReadSnapshot(func() { c, _ := g.LookupCardForEffect(theirs); v = c.Tapped })
		return v
	}

	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the tap ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !tapped() {
		t.Fatal("{U}: Tap enchanted creature did not tap it — the Aura's controller may activate it on an opponent's creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, aura, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the untap ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if tapped() {
		t.Error("{U}: Untap enchanted creature left it tapped")
	}
}
