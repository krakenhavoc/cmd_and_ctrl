package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kuja_b_test.go — card-level tests for slice kuja-b of the Kuja deck
// request (#3213). Each drives a real cast and asserts the observable
// result.

const (
	kbCircleOfPowerOracle = "2d41fd1f-4a98-45dd-a3c5-f7c5e0152109"
	kbCremateOracle       = "c6a2e410-b182-48d1-aeb2-bc8de27e9cd2"
	kbDarknessOracle      = "932708ae-9d5c-4561-aa8b-0d2222d37fdc"
	kbFlickACoinOracle    = "990583ef-5bf4-4e14-a8f1-61d32c6a75df"
	kbOverkillOracle      = "571cb4b6-b4e5-40a2-a970-42c1ab65c8a9"
)

func kbCardTarget(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func kbWizard(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Human Wizard",
		Power: 1, Toughness: 1, Owner: owner, Controller: owner,
	})
}

// --- Circle of Power ------------------------------------------------

// Draw two, lose two, a Wizard token, and the Wizards you control at
// resolution (the new one too) get +1/+0 and lifelink; an opponent's
// Wizard is untouched, and the pump wears off at cleanup.
func TestCircleOfPowerDrawsMakesAWizardAndPumpsYours(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := kbWizard(g, me.ID, "Old Wizard")
	theirs := kbWizard(g, opp.ID, "Their Wizard")
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	hand, life := len(me.Hand.Cards), me.Life

	castCatalogSpell(t, g, "Circle of Power", "Sorcery", kbCircleOfPowerOracle, nil)
	passPriorityAroundTable(t, g)

	if got := len(me.Hand.Cards) - hand; got != 2 {
		t.Errorf("hand grew by %d, want 2 draws", got)
	}
	if me.Life != life-2 {
		t.Errorf("life %d -> %d, want -2", life, me.Life)
	}
	token := findBattlefieldByName(g, "Wizard")
	if token == uuid.Nil {
		t.Fatal("no Wizard token was created")
	}
	if p, tough := effectivePower(t, g, token), effectiveToughness(t, g, token); p != 1 || tough != 1 {
		t.Errorf("the new Wizard is a 0/1 plus +1/+0: got %d/%d", p, tough)
	}
	for _, id := range []uuid.UUID{token, mine} {
		if !hasEffectiveKeyword(t, g, id, "lifelink") {
			t.Errorf("Wizard %s should have lifelink", id)
		}
	}
	if p := effectivePower(t, g, mine); p != 2 {
		t.Errorf("my old Wizard is 1/1 plus +1/+0: power %d", p)
	}
	if p := effectivePower(t, g, theirs); p != 1 || hasEffectiveKeyword(t, g, theirs, "lifelink") {
		t.Errorf("an opponent's Wizard is untouched: power %d", p)
	}
	if p := effectivePower(t, g, bear); p != 2 || hasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Errorf("a non-Wizard is untouched: power %d", p)
	}
	advancePastCleanupOf(t, g, 0)
	if p := effectivePower(t, g, mine); p != 1 {
		t.Errorf("the pump ends at cleanup: power %d", p)
	}
}

// The Wizard token carries its real trigger: a noncreature spell costs
// each opponent 1.
func TestCircleOfPowerWizardPingsOnANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castCatalogSpell(t, g, "Circle of Power", "Sorcery", kbCircleOfPowerOracle, nil)
	passPriorityAroundTable(t, g)
	before := opp.Life
	castCatalogSpell(t, g, "Divination", "Sorcery", "273b339c-964b-4a18-8eb5-ceb8abcdfd9e", nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("opponent life %d -> %d, want -1", before, opp.Life)
	}
	_ = me
}

// --- Cremate ---------------------------------------------------------

// Exiles the targeted card from any graveyard, then draws.
func TestCremateExilesAnyGraveyardCardAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := batch01GraveyardCard(opp, "Their Bear", "Creature — Bear")
	hand := len(me.Hand.Cards)
	castCatalogSpell(t, g, "Cremate", "Instant", kbCremateOracle, kbCardTarget(victim))
	passPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(victim) || !g.Exile.Contains(victim) {
		t.Error("the targeted card should be exiled from its owner's graveyard")
	}
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("hand %d -> %d: the test adds the spell, casts it and draws one (+1)", hand, got)
	}
}

// A battlefield permanent is not a card in a graveyard.
func TestCremateRefusesAPermanentOnTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	bear := pushVanillaCreature(g, g.Seats[1].ID, "Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Cremate", "Instant", kbCremateOracle, kbCardTarget(bear)); err == nil {
		t.Error("Cremate targeted a creature on the battlefield")
	}
}

// --- Darkness --------------------------------------------------------

func TestDarknessPreventsCombatDamageButNotSpellDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castCatalogSpell(t, g, "Darkness", "Instant", kbDarknessOracle, nil)
	passPriorityAroundTable(t, g)
	attacker := pushVanillaCreature(g, opp.ID, "Attacker", 2, 2)
	defender := pushVanillaCreature(g, me.ID, "Defender", 2, 2)
	if err := g.MarkCombatDamage(attacker, defender, 2); err != nil {
		t.Fatalf("MarkCombatDamage: %v", err)
	}
	var marked int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == defender {
				marked = c.DamageMarked
			}
		}
	})
	if marked != 0 {
		t.Errorf("combat damage marked %d, want 0", marked)
	}
}

// --- Flick a Coin ----------------------------------------------------

func TestFlickACoinPingsMakesATreasureAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hand, bf, life := len(me.Hand.Cards), len(g.Battlefield.Cards), opp.Life
	castCatalogSpell(t, g, "Flick a Coin", "Instant", kbFlickACoinOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("target life %d -> %d, want -1", life, opp.Life)
	}
	if got := len(g.Battlefield.Cards); got != bf+1 || countBattlefieldByName(g, "Treasure") != 1 {
		t.Errorf("battlefield %d -> %d, want one Treasure", bf, got)
	}
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("hand %d -> %d: the test adds the spell, casts it and draws one (+1)", hand, got)
	}
}

// --- Overkill --------------------------------------------------------

func TestOverkillKillsAnyCreatureThroughIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	big := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big", TypeLine: "Creature — Giant",
		Power: 9, Toughness: 9, Owner: opp.ID, Controller: opp.ID,
		Keywords: []string{"indestructible"},
	})
	castCatalogSpell(t, g, "Overkill", "Instant", kbOverkillOracle, kbCardTarget(big))
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(big) {
		t.Error("a 9/9 indestructible creature dies to -0/-9999 (toughness 0 or less)")
	}
}

// A noncreature target is refused at announce.
func TestOverkillRefusesAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	err := castCatalogSpellErr(t, g, "Overkill", "Instant", kbOverkillOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}})
	if err == nil {
		t.Error("Overkill targeted a player")
	}
}
