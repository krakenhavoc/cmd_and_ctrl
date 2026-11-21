package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blue_mages_cane_test.go — Blue Mage's Cane (#3657, seam #3783): the
// granted attack trigger exiles a card from the defending player's
// graveyard and offers a copy of it for {3} in place of its mana cost.

const blueMagesCaneOracle = "6290233b-b015-4909-8aba-bf8d65cc37b5"

// caneAttack casts the Cane, attacks the next seat with its Hero and
// returns the Hero, leaving the game in the declare-attackers step with the attack
// trigger's target prompt open (or the trigger on the stack).
func caneAttack(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	opp := g.Seats[1]
	cane := effectstest.CastCatalogSpell(t, g, "Blue Mage's Cane", "Artifact — Equipment", blueMagesCaneOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	hero := effectstest.YsBHero(t, g, cane)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(hero, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	return hero
}

func TestBlueMagesCaneIsRegisteredAndEquipsAWizardWithPlusTwoToughness(t *testing.T) {
	spec, ok := effects.Lookup(blueMagesCaneOracle)
	if !ok || spec.Name != "Blue Mage's Cane" || spec.Completeness != effects.CompletenessFull {
		t.Fatalf("registration: ok=%v spec=%+v", ok, spec)
	}
	g := effectstest.NewCatalogGame(t)
	cane := effectstest.CastCatalogSpell(t, g, "Blue Mage's Cane", "Artifact — Equipment", blueMagesCaneOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	hero := effectstest.YsBHero(t, g, cane)
	v := effectstest.S58p6View(t, g, hero)
	if v.CurrentPower() != 1 || v.CurrentToughness() != 3 {
		t.Errorf("Hero is %d/%d, want 1/3", v.CurrentPower(), v.CurrentToughness())
	}
	if !v.HasSubtype("Wizard") || !v.HasSubtype("Hero") {
		t.Errorf("Hero subtypes %v, want Hero and Wizard", v.Effective().Subtypes)
	}
}

func TestBlueMagesCaneExilesTheDefendersSpellAndOffersACopyForThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.GyCard(opp, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)
	bear := effectstest.GyCard(opp, "Bear", "Creature — Bear", "{1}{G}", "")
	mine := effectstest.GyCard(me, "Divination", "Sorcery", "{2}{U}", effectstest.DivinationOracle)

	caneAttack(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the attack trigger")
	}
	if !effectstest.HasID(pick.PickTargetCards, bolt) || effectstest.HasID(pick.PickTargetCards, bear) || effectstest.HasID(pick.PickTargetCards, mine) {
		t.Errorf("candidates = %v, want only the defender's instant %v", pick.PickTargetCards, bolt)
	}
	effectstest.PickTriggerTarget(t, g, me.ID, bolt)
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.InExile(g, bolt) {
		t.Fatal("the targeted card is exiled")
	}
	cp := effectstest.CopyOf(t, g, "Lightning Bolt")
	if cp.InstanceID == bolt || cp.Owner != me.ID {
		t.Errorf("copy = %+v: a new object owned by the Cane's controller", cp)
	}
	g.ReadSnapshot(func() {
		perm := g.CastPermissionOnCardForEffect(cp, game.ZoneExile)
		if perm == nil || perm.Player != me.ID {
			t.Fatalf("the controller has no permission over the copy: %+v", perm)
		}
		if perm.Cost != "{3}" {
			t.Errorf("permission cost = %q, want {3} in place of the printed {R}", perm.Cost)
		}
	})

	// The price is really {3}: an empty pool under strict mana cannot pay it.
	targets := []game.TargetRef{effectstest.PlayerTarget(opp.ID)}
	if err := effectstest.CastCopy(g, me.ID, cp.InstanceID, true, targets...); err == nil {
		t.Fatal("cast the copy with an empty pool under strict mana: {3} must be charged")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	before := opp.Life
	if err := effectstest.CastCopy(g, me.ID, cp.InstanceID, true, targets...); err != nil {
		t.Fatalf("cast the copy for {3}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("life = %d, want %d: the copy resolves as Lightning Bolt", opp.Life, before-3)
	}
	if n := len(effectstest.CardCopies(g)); n != 0 {
		t.Errorf("%d copies left: a resolved copy ceases to exist", n)
	}
	if !effectstest.InExile(g, bolt) {
		t.Error("the original stays exiled")
	}
}

func TestBlueMagesCaneUncastCopyCeases(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	divination := effectstest.GyCard(opp, "Divination", "Sorcery", "{2}{U}", effectstest.DivinationOracle)

	caneAttack(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, divination)
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.CardCopies(g)) != 1 {
		t.Fatalf("copies = %d, want 1", len(effectstest.CardCopies(g)))
	}
	// Declining ends the offer, and the table moves on.
	effectstest.DeclineResolutionCasts(t, g, me.ID)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.CardCopies(g)); n != 0 {
		t.Errorf("%d copies left: an uncast copy ceases to exist", n)
	}
	if !effectstest.InExile(g, divination) {
		t.Error("the original stays exiled")
	}
}

func TestBlueMagesCaneWithNothingToExileMakesNoCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.GyCard(g.Seats[1], "Bear", "Creature — Bear", "{1}{G}", "")

	caneAttack(t, g)
	// Up to one: with no instant or sorcery in the defender's graveyard
	// the trigger has nothing to choose and resolves doing nothing.
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		t.Fatalf("a target prompt with no legal card: %v", pick.PickTargetCards)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.CardCopies(g)); n != 0 {
		t.Errorf("%d copies with nothing exiled", n)
	}
}
