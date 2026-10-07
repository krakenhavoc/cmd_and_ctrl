package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// energy_batch_c_cards_test.go — ADR 0129 PR 1 (#1995): energy cards
// whose activations pay energy (CR 107.14) and whose triggers give it.

const (
	ecConsulateSurveillance = "06d62e22-df9c-488e-8d2d-793eaf4cd75c"
	ecGontisAetherHeart     = "69428825-3c40-486d-b051-14e97a598ce6"
	ecAssemblyLine          = "6e24a5bd-0ce3-4dc1-98c6-60d527bfdcfd"
	ecDeadlockTrap          = "726c911d-6543-4400-a8fa-b8a5c9f0c15d"
	ecDemonOfDarkSchemes    = "2b2f2abd-6c0f-49a5-a44b-f366614c944d"
	ecDrMadisonLi           = "9c4ff8fe-7d69-42b9-bad1-9e2c8a3e29f1"
	ecDynavoltTower         = "59f56dd9-cc82-4f0d-a522-0d8ed274fa40"
	ecEraOfInnovation       = "355603d1-f62d-407a-8fa9-74e7dca8ecf9"
)

// ecRefusedWhenShort sets p's energy to one less than the row's printed
// energy and asserts the activation is refused with nothing paid.
func ecRefusedWhenShort(t *testing.T, g *game.Game, p *game.Player, source uuid.UUID, index int, oracle string, params game.ActivateAbilityParams) {
	t.Helper()
	need := specActivatedEnergy(t, oracle, index).Energy
	setEnergy(t, g, p, need-1)
	err := g.ActivateCatalogAbility(p.ID, source, index, params)
	if !errors.Is(err, game.ErrInsufficientEnergy) {
		t.Fatalf("row %d with %d energy: err = %v, want ErrInsufficientEnergy", index, need-1, err)
	}
	if energyOf(p) != need-1 {
		t.Fatalf("a refused activation spent energy: %d, want %d", energyOf(p), need-1)
	}
}

func TestEnergyBatchCConsulateSurveillance(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	setEnergy(t, g, me, 0)
	id := castCatalogSpell(t, g, "Consulate Surveillance", "Enchantment", ecConsulateSurveillance, nil)
	passPriorityAroundTable(t, g)
	if energyOf(me) != 4 {
		t.Fatalf("enters: energy %d, want 4", energyOf(me))
	}
	ecRefusedWhenShort(t, g, me, id, 0, ecConsulateSurveillance, game.ActivateAbilityParams{})
	setEnergy(t, g, me, 4)
	src := pr7Creature(g, opp.ID, "Pinger", 2, "R")
	pr7Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	if energyOf(me) != 2 {
		t.Fatalf("paid: energy %d, want 2", energyOf(me))
	}
	pr7Choose(t, g, me.ID, src)
	life := me.Life
	pr7Hit(t, g, src, me.ID, 2)
	if me.Life != life {
		t.Errorf("the chosen source's damage was dealt: life %d, want %d", me.Life, life)
	}
}

func TestEnergyBatchCGontisAetherHeart(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	setEnergy(t, g, me, 0)
	heart := castCatalogSpell(t, g, "Gonti's Aether Heart", "Legendary Artifact", ecGontisAetherHeart, nil)
	passPriorityAroundTable(t, g)
	if energyOf(me) != 2 {
		t.Fatalf("the Heart enters: energy %d, want 2", energyOf(me))
	}
	castCatalogSpell(t, g, "Plain Bauble", "Artifact", "00000000-0000-0000-0000-0000000e0c01", nil)
	passPriorityAroundTable(t, g)
	if energyOf(me) != 4 {
		t.Fatalf("another artifact enters: energy %d, want 4", energyOf(me))
	}
	castCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "00000000-0000-0000-0000-0000000e0c02", nil)
	passPriorityAroundTable(t, g)
	if energyOf(me) != 4 {
		t.Fatalf("a non-artifact entered and paid: energy %d, want 4", energyOf(me))
	}
	ecRefusedWhenShort(t, g, me, heart, 0, ecGontisAetherHeart, game.ActivateAbilityParams{})
	setEnergy(t, g, me, 9)
	turns := len(g.ExtraTurns)
	pr7Activate(t, g, me.ID, heart, 0, game.ActivateAbilityParams{})
	if energyOf(me) != 1 {
		t.Errorf("paid: energy %d, want 1", energyOf(me))
	}
	if !g.Exile.Contains(heart) {
		t.Error("the Heart is not in exile")
	}
	if len(g.ExtraTurns) != turns+1 {
		t.Errorf("extra turns %d, want %d", len(g.ExtraTurns), turns+1)
	}
}

func TestEnergyBatchCAutomatedAssemblyLine(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	setEnergy(t, g, me, 0)
	line := b12Push(g, me.ID, "Automated Assembly Line", "Artifact", ecAssemblyLine, 0, 0)
	golem := b12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b784AttackEach(t, g, [2]uuid.UUID{golem, g.Seats[1].ID}, [2]uuid.UUID{bear, g.Seats[2].ID})
	if n := b784Triggers(g, line); n != 1 {
		t.Fatalf("one artifact creature hit one player: %d triggers, want 1", n)
	}
	passPriorityAroundTable(t, g)
	if energyOf(me) != 1 {
		t.Fatalf("energy %d, want 1", energyOf(me))
	}

	ecRefusedWhenShort(t, g, me, line, 0, ecAssemblyLine, game.ActivateAbilityParams{})
	setEnergy(t, g, me, 3)
	robots := countOnBattlefield(g, "Robot", me.ID)
	pr7Activate(t, g, me.ID, line, 0, game.ActivateAbilityParams{})
	if energyOf(me) != 0 || countOnBattlefield(g, "Robot", me.ID) != robots+1 {
		t.Fatalf("paid: energy %d, robots %d", energyOf(me), countOnBattlefield(g, "Robot", me.ID)-robots)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Robot" && c.Controller == me.ID && (!c.Tapped || c.Power != 3 || c.Toughness != 3 || !c.IsArtifact()) {
			t.Errorf("token %+v, want a tapped 3/3 artifact", c)
		}
	}
}

func TestEnergyBatchCDeadlockTrap(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	setEnergy(t, g, me, 0)
	trap := castCatalogSpell(t, g, "Deadlock Trap", "Artifact", ecDeadlockTrap, nil)
	passPriorityAroundTable(t, g)
	if energyOf(me) != 2 {
		t.Fatalf("enters: energy %d, want 2", energyOf(me))
	}
	if c := findBattlefieldCardForTest(g, trap); c == nil || !c.Tapped {
		t.Fatal("Deadlock Trap entered untapped")
	}
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, trap).Tapped = false })
	victim := pushCatalogPermanent(g, opp.ID, "Pyromancy", "Enchantment Creature", p8Pyromancy, false)
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: victim}}
	ecRefusedWhenShort(t, g, me, trap, 0, ecDeadlockTrap, game.ActivateAbilityParams{Targets: targets})
	setEnergy(t, g, me, 1)
	pr7Activate(t, g, me.ID, trap, 0, game.ActivateAbilityParams{Targets: targets})
	if energyOf(me) != 0 {
		t.Errorf("paid: energy %d, want 0", energyOf(me))
	}
	v := apaLive(g, victim)
	if !v.Tapped {
		t.Error("the target was not tapped")
	}
	if game.CanActivateAbilities(v) {
		t.Error("the target's activated abilities can still be activated this turn")
	}
}

func TestEnergyBatchCDemonOfDarkSchemes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	setEnergy(t, g, me, 0)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	ox := b12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 4)
	demon := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: demon, Name: "Demon of Dark Schemes", TypeLine: "Creature — Demon",
		OracleID: ecDemonOfDarkSchemes, Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, demon, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the Demon: %v", err)
	}
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, bear) != nil {
		t.Fatal("the 2/2 survived -2/-2")
	}
	if findBattlefieldCardForTest(g, ox) == nil {
		t.Fatal("the 2/4 died")
	}
	if d := apaLive(g, demon); d == nil || d.CurrentPower() != 5 {
		t.Fatal("the Demon shrank itself")
	}
	if energyOf(me) != 1 {
		t.Fatalf("another creature died: energy %d, want 1", energyOf(me))
	}
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: bear}}
	ecRefusedWhenShort(t, g, me, demon, 0, ecDemonOfDarkSchemes, game.ActivateAbilityParams{Targets: targets})
	setEnergy(t, g, me, 4)
	pr7Activate(t, g, me.ID, demon, 0, game.ActivateAbilityParams{Targets: targets})
	if energyOf(me) != 0 {
		t.Errorf("paid: energy %d, want 0", energyOf(me))
	}
	c := findBattlefieldCardForTest(g, bear)
	if c == nil || c.Controller != me.ID || !c.Tapped {
		t.Fatalf("the bear is %+v, want tapped on the battlefield under my control", c)
	}
}

func TestEnergyBatchCDrMadisonLi(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	setEnergy(t, g, me, 0)
	li := b12Push(g, me.ID, "Dr. Madison Li", "Legendary Creature — Human Scientist", ecDrMadisonLi, 2, 3)
	castCatalogSpell(t, g, "Plain Bauble", "Artifact", "00000000-0000-0000-0000-0000000e0c03", nil)
	passPriorityAroundTable(t, g)
	if energyOf(me) != 1 {
		t.Fatalf("an artifact spell: energy %d, want 1", energyOf(me))
	}

	// Row 0: +1/+0, trample and haste.
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: bear}}
	ecRefusedWhenShort(t, g, me, li, 0, ecDrMadisonLi, game.ActivateAbilityParams{Targets: targets})
	setEnergy(t, g, me, 1)
	pr7Activate(t, g, me.ID, li, 0, game.ActivateAbilityParams{Targets: targets})
	b := apaLive(g, bear)
	if energyOf(me) != 0 || b.CurrentPower() != 3 || !game.HasKeyword(b, "trample") || !game.HasKeyword(b, "haste") {
		t.Errorf("row 0: energy %d, bear %d power trample %v haste %v", energyOf(me), b.CurrentPower(), game.HasKeyword(b, "trample"), game.HasKeyword(b, "haste"))
	}

	// Row 1: draw a card.
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, li).Tapped = false })
	ecRefusedWhenShort(t, g, me, li, 1, ecDrMadisonLi, game.ActivateAbilityParams{})
	setEnergy(t, g, me, 3)
	hand := me.Hand.Size()
	pr7Activate(t, g, me.ID, li, 1, game.ActivateAbilityParams{})
	if energyOf(me) != 0 || me.Hand.Size() != hand+1 {
		t.Errorf("row 1: energy %d, drew %d", energyOf(me), me.Hand.Size()-hand)
	}

	// Row 2: an artifact card from your graveyard, tapped.
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, li).Tapped = false })
	relic := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: relic, Name: "Relic", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	})
	targets = []game.TargetRef{{Kind: game.TargetCard, ID: relic}}
	ecRefusedWhenShort(t, g, me, li, 2, ecDrMadisonLi, game.ActivateAbilityParams{Targets: targets})
	setEnergy(t, g, me, 5)
	pr7Activate(t, g, me.ID, li, 2, game.ActivateAbilityParams{Targets: targets})
	if c := findBattlefieldCardForTest(g, relic); energyOf(me) != 0 || c == nil || !c.Tapped {
		t.Errorf("row 2: energy %d, relic %+v, want it tapped on the battlefield", energyOf(me), c)
	}
}

func TestEnergyBatchCDynavoltTower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	setEnergy(t, g, me, 0)
	tower := b12Push(g, me.ID, "Dynavolt Tower", "Artifact", ecDynavoltTower, 0, 0)
	castCatalogSpell(t, g, "Plain Ponder", "Sorcery", "00000000-0000-0000-0000-0000000e0c04", nil)
	passPriorityAroundTable(t, g)
	if energyOf(me) != 2 {
		t.Fatalf("a sorcery: energy %d, want 2", energyOf(me))
	}
	castCatalogSpell(t, g, "Plain Bauble", "Artifact", "00000000-0000-0000-0000-0000000e0c05", nil)
	passPriorityAroundTable(t, g)
	if energyOf(me) != 2 {
		t.Fatalf("an artifact spell paid: energy %d, want 2", energyOf(me))
	}
	targets := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	ecRefusedWhenShort(t, g, me, tower, 0, ecDynavoltTower, game.ActivateAbilityParams{Targets: targets})
	setEnergy(t, g, me, 5)
	life := opp.Life
	pr7Activate(t, g, me.ID, tower, 0, game.ActivateAbilityParams{Targets: targets})
	if energyOf(me) != 0 || opp.Life != life-3 {
		t.Errorf("energy %d, opponent lost %d, want 0 and 3", energyOf(me), life-opp.Life)
	}
}

func TestEnergyBatchCEraOfInnovation(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	setEnergy(t, g, me, 0)
	era := b12Push(g, me.ID, "Era of Innovation", "Enchantment", ecEraOfInnovation, 0, 0)
	castCatalogSpell(t, g, "Plain Bauble", "Artifact", "00000000-0000-0000-0000-0000000e0c06", nil)
	passPriorityAroundTable(t, g)
	b06AddMana(me, "C")
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if energyOf(me) != 2 {
		t.Fatalf("paid {1}: energy %d, want 2", energyOf(me))
	}
	castCatalogSpell(t, g, "Plain Artificer", "Creature — Human Artificer", "00000000-0000-0000-0000-0000000e0c07", nil)
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if energyOf(me) != 2 {
		t.Fatalf("declined: energy %d, want 2", energyOf(me))
	}

	ecRefusedWhenShort(t, g, me, era, 0, ecEraOfInnovation, game.ActivateAbilityParams{})
	setEnergy(t, g, me, 6)
	hand := me.Hand.Size()
	pr7Activate(t, g, me.ID, era, 0, game.ActivateAbilityParams{})
	if energyOf(me) != 0 || me.Hand.Size() != hand+3 || findBattlefieldCardForTest(g, era) != nil {
		t.Errorf("energy %d, drew %d, era still out %v", energyOf(me), me.Hand.Size()-hand, findBattlefieldCardForTest(g, era) != nil)
	}
}
