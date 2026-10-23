package effects

import (
	"testing"

	"github.com/google/uuid"

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
	cane := castCatalogSpell(t, g, "Blue Mage's Cane", "Artifact — Equipment", blueMagesCaneOracle, nil)
	passPriorityAroundTable(t, g)
	hero := ysBHero(t, g, cane)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	})
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(hero, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	return hero
}

func TestBlueMagesCaneIsRegisteredAndEquipsAWizardWithPlusTwoToughness(t *testing.T) {
	spec, ok := Lookup(blueMagesCaneOracle)
	if !ok || spec.Name != "Blue Mage's Cane" || spec.Completeness != CompletenessFull {
		t.Fatalf("registration: ok=%v spec=%+v", ok, spec)
	}
	g := newCatalogGame(t)
	cane := castCatalogSpell(t, g, "Blue Mage's Cane", "Artifact — Equipment", blueMagesCaneOracle, nil)
	passPriorityAroundTable(t, g)
	hero := ysBHero(t, g, cane)
	v := s58p6View(t, g, hero)
	if v.CurrentPower() != 1 || v.CurrentToughness() != 3 {
		t.Errorf("Hero is %d/%d, want 1/3", v.CurrentPower(), v.CurrentToughness())
	}
	if !v.HasSubtype("Wizard") || !v.HasSubtype("Hero") {
		t.Errorf("Hero subtypes %v, want Hero and Wizard", v.Effective().Subtypes)
	}
}

func TestBlueMagesCaneExilesTheDefendersSpellAndOffersACopyForThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := gyCard(opp, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	bear := gyCard(opp, "Bear", "Creature — Bear", "{1}{G}", "")
	mine := gyCard(me, "Divination", "Sorcery", "{2}{U}", divinationOracle)

	caneAttack(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the attack trigger")
	}
	if !hasID(pick.PickTargetCards, bolt) || hasID(pick.PickTargetCards, bear) || hasID(pick.PickTargetCards, mine) {
		t.Errorf("candidates = %v, want only the defender's instant %v", pick.PickTargetCards, bolt)
	}
	pickTriggerTarget(t, g, me.ID, bolt)
	passPriorityAroundTable(t, g)

	if !inExile(g, bolt) {
		t.Fatal("the targeted card is exiled")
	}
	cp := copyOf(t, g, "Lightning Bolt")
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
	targets := []game.TargetRef{playerTarget(opp.ID)}
	if err := castCopy(g, me.ID, cp.InstanceID, true, targets...); err == nil {
		t.Fatal("cast the copy with an empty pool under strict mana: {3} must be charged")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	before := opp.Life
	if err := castCopy(g, me.ID, cp.InstanceID, true, targets...); err != nil {
		t.Fatalf("cast the copy for {3}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("life = %d, want %d: the copy resolves as Lightning Bolt", opp.Life, before-3)
	}
	if n := len(cardCopies(g)); n != 0 {
		t.Errorf("%d copies left: a resolved copy ceases to exist", n)
	}
	if !inExile(g, bolt) {
		t.Error("the original stays exiled")
	}
}

func TestBlueMagesCaneUncastCopyCeases(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	divination := gyCard(opp, "Divination", "Sorcery", "{2}{U}", divinationOracle)

	caneAttack(t, g)
	pickTriggerTarget(t, g, me.ID, divination)
	passPriorityAroundTable(t, g)
	if len(cardCopies(g)) != 1 {
		t.Fatalf("copies = %d, want 1", len(cardCopies(g)))
	}
	// Declining ends the offer, and the table moves on.
	declineResolutionCasts(t, g, me.ID)
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	if n := len(cardCopies(g)); n != 0 {
		t.Errorf("%d copies left: an uncast copy ceases to exist", n)
	}
	if !inExile(g, divination) {
		t.Error("the original stays exiled")
	}
}

func TestBlueMagesCaneWithNothingToExileMakesNoCopy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gyCard(g.Seats[1], "Bear", "Creature — Bear", "{1}{G}", "")

	caneAttack(t, g)
	// Up to one: with no instant or sorcery in the defender's graveyard
	// the trigger has nothing to choose and resolves doing nothing.
	if pick := latestPickTarget(g, me.ID); pick != nil {
		t.Fatalf("a target prompt with no legal card: %v", pick.PickTargetCards)
	}
	passPriorityAroundTable(t, g)
	if n := len(cardCopies(g)); n != 0 {
		t.Errorf("%d copies with nothing exiled", n)
	}
}
