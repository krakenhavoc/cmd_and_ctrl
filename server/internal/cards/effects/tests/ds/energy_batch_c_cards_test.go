package testsds

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
	need := effectstest.SpecActivatedEnergy(t, oracle, index).Energy
	effectstest.SetEnergy(t, g, p, need-1)
	err := g.ActivateCatalogAbility(p.ID, source, index, params)
	if !errors.Is(err, game.ErrInsufficientEnergy) {
		t.Fatalf("row %d with %d energy: err = %v, want ErrInsufficientEnergy", index, need-1, err)
	}
	if effectstest.EnergyOf(p) != need-1 {
		t.Fatalf("a refused activation spent energy: %d, want %d", effectstest.EnergyOf(p), need-1)
	}
}

func TestEnergyBatchCConsulateSurveillance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SetEnergy(t, g, me, 0)
	id := effectstest.CastCatalogSpell(t, g, "Consulate Surveillance", "Enchantment", ecConsulateSurveillance, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EnergyOf(me) != 4 {
		t.Fatalf("enters: energy %d, want 4", effectstest.EnergyOf(me))
	}
	ecRefusedWhenShort(t, g, me, id, 0, ecConsulateSurveillance, game.ActivateAbilityParams{})
	effectstest.SetEnergy(t, g, me, 4)
	src := effectstest.Pr7Creature(g, opp.ID, "Pinger", 2, "R")
	effectstest.Pr7Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	if effectstest.EnergyOf(me) != 2 {
		t.Fatalf("paid: energy %d, want 2", effectstest.EnergyOf(me))
	}
	effectstest.Pr7Choose(t, g, me.ID, src)
	life := me.Life
	effectstest.Pr7Hit(t, g, src, me.ID, 2)
	if me.Life != life {
		t.Errorf("the chosen source's damage was dealt: life %d, want %d", me.Life, life)
	}
}

func TestEnergyBatchCGontisAetherHeart(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SetEnergy(t, g, me, 0)
	heart := effectstest.CastCatalogSpell(t, g, "Gonti's Aether Heart", "Legendary Artifact", ecGontisAetherHeart, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EnergyOf(me) != 2 {
		t.Fatalf("the Heart enters: energy %d, want 2", effectstest.EnergyOf(me))
	}
	effectstest.CastCatalogSpell(t, g, "Plain Bauble", "Artifact", "00000000-0000-0000-0000-0000000e0c01", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EnergyOf(me) != 4 {
		t.Fatalf("another artifact enters: energy %d, want 4", effectstest.EnergyOf(me))
	}
	effectstest.CastCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "00000000-0000-0000-0000-0000000e0c02", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EnergyOf(me) != 4 {
		t.Fatalf("a non-artifact entered and paid: energy %d, want 4", effectstest.EnergyOf(me))
	}
	ecRefusedWhenShort(t, g, me, heart, 0, ecGontisAetherHeart, game.ActivateAbilityParams{})
	effectstest.SetEnergy(t, g, me, 9)
	turns := len(g.ExtraTurns)
	effectstest.Pr7Activate(t, g, me.ID, heart, 0, game.ActivateAbilityParams{})
	if effectstest.EnergyOf(me) != 1 {
		t.Errorf("paid: energy %d, want 1", effectstest.EnergyOf(me))
	}
	if !g.Exile.Contains(heart) {
		t.Error("the Heart is not in exile")
	}
	if len(g.ExtraTurns) != turns+1 {
		t.Errorf("extra turns %d, want %d", len(g.ExtraTurns), turns+1)
	}
}

func TestEnergyBatchCAutomatedAssemblyLine(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SetEnergy(t, g, me, 0)
	line := effectstest.B12Push(g, me.ID, "Automated Assembly Line", "Artifact", ecAssemblyLine, 0, 0)
	golem := effectstest.B12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B784AttackEach(t, g, [2]uuid.UUID{golem, g.Seats[1].ID}, [2]uuid.UUID{bear, g.Seats[2].ID})
	if n := effectstest.B784Triggers(g, line); n != 1 {
		t.Fatalf("one artifact creature hit one player: %d triggers, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EnergyOf(me) != 1 {
		t.Fatalf("energy %d, want 1", effectstest.EnergyOf(me))
	}

	ecRefusedWhenShort(t, g, me, line, 0, ecAssemblyLine, game.ActivateAbilityParams{})
	effectstest.SetEnergy(t, g, me, 3)
	robots := effectstest.CountOnBattlefield(g, "Robot", me.ID)
	effectstest.Pr7Activate(t, g, me.ID, line, 0, game.ActivateAbilityParams{})
	if effectstest.EnergyOf(me) != 0 || effectstest.CountOnBattlefield(g, "Robot", me.ID) != robots+1 {
		t.Fatalf("paid: energy %d, robots %d", effectstest.EnergyOf(me), effectstest.CountOnBattlefield(g, "Robot", me.ID)-robots)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Robot" && c.Controller == me.ID && (!c.Tapped || c.Power != 3 || c.Toughness != 3 || !c.IsArtifact()) {
			t.Errorf("token %+v, want a tapped 3/3 artifact", c)
		}
	}
}

func TestEnergyBatchCDeadlockTrap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SetEnergy(t, g, me, 0)
	trap := effectstest.CastCatalogSpell(t, g, "Deadlock Trap", "Artifact", ecDeadlockTrap, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EnergyOf(me) != 2 {
		t.Fatalf("enters: energy %d, want 2", effectstest.EnergyOf(me))
	}
	if c := effectstest.FindBattlefieldCardForTest(g, trap); c == nil || !c.Tapped {
		t.Fatal("Deadlock Trap entered untapped")
	}
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, trap).Tapped = false })
	victim := effectstest.PushCatalogPermanent(g, opp.ID, "Pyromancy", "Enchantment Creature", effectstest.P8Pyromancy, false)
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: victim}}
	ecRefusedWhenShort(t, g, me, trap, 0, ecDeadlockTrap, game.ActivateAbilityParams{Targets: targets})
	effectstest.SetEnergy(t, g, me, 1)
	effectstest.Pr7Activate(t, g, me.ID, trap, 0, game.ActivateAbilityParams{Targets: targets})
	if effectstest.EnergyOf(me) != 0 {
		t.Errorf("paid: energy %d, want 0", effectstest.EnergyOf(me))
	}
	v := effectstest.ApaLive(g, victim)
	if !v.Tapped {
		t.Error("the target was not tapped")
	}
	if game.CanActivateAbilities(v) {
		t.Error("the target's activated abilities can still be activated this turn")
	}
}

func TestEnergyBatchCDemonOfDarkSchemes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SetEnergy(t, g, me, 0)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	ox := effectstest.B12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 4)
	demon := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: demon, Name: "Demon of Dark Schemes", TypeLine: "Creature — Demon",
		OracleID: ecDemonOfDarkSchemes, Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, demon, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the Demon: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, bear) != nil {
		t.Fatal("the 2/2 survived -2/-2")
	}
	if effectstest.FindBattlefieldCardForTest(g, ox) == nil {
		t.Fatal("the 2/4 died")
	}
	if d := effectstest.ApaLive(g, demon); d == nil || d.CurrentPower() != 5 {
		t.Fatal("the Demon shrank itself")
	}
	if effectstest.EnergyOf(me) != 1 {
		t.Fatalf("another creature died: energy %d, want 1", effectstest.EnergyOf(me))
	}
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: bear}}
	ecRefusedWhenShort(t, g, me, demon, 0, ecDemonOfDarkSchemes, game.ActivateAbilityParams{Targets: targets})
	effectstest.SetEnergy(t, g, me, 4)
	effectstest.Pr7Activate(t, g, me.ID, demon, 0, game.ActivateAbilityParams{Targets: targets})
	if effectstest.EnergyOf(me) != 0 {
		t.Errorf("paid: energy %d, want 0", effectstest.EnergyOf(me))
	}
	c := effectstest.FindBattlefieldCardForTest(g, bear)
	if c == nil || c.Controller != me.ID || !c.Tapped {
		t.Fatalf("the bear is %+v, want tapped on the battlefield under my control", c)
	}
}

func TestEnergyBatchCDrMadisonLi(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SetEnergy(t, g, me, 0)
	li := effectstest.B12Push(g, me.ID, "Dr. Madison Li", "Legendary Creature — Human Scientist", ecDrMadisonLi, 2, 3)
	effectstest.CastCatalogSpell(t, g, "Plain Bauble", "Artifact", "00000000-0000-0000-0000-0000000e0c03", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EnergyOf(me) != 1 {
		t.Fatalf("an artifact spell: energy %d, want 1", effectstest.EnergyOf(me))
	}

	// Row 0: +1/+0, trample and haste.
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: bear}}
	ecRefusedWhenShort(t, g, me, li, 0, ecDrMadisonLi, game.ActivateAbilityParams{Targets: targets})
	effectstest.SetEnergy(t, g, me, 1)
	effectstest.Pr7Activate(t, g, me.ID, li, 0, game.ActivateAbilityParams{Targets: targets})
	b := effectstest.ApaLive(g, bear)
	if effectstest.EnergyOf(me) != 0 || b.CurrentPower() != 3 || !game.HasKeyword(b, "trample") || !game.HasKeyword(b, "haste") {
		t.Errorf("row 0: energy %d, bear %d power trample %v haste %v", effectstest.EnergyOf(me), b.CurrentPower(), game.HasKeyword(b, "trample"), game.HasKeyword(b, "haste"))
	}

	// Row 1: draw a card.
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, li).Tapped = false })
	ecRefusedWhenShort(t, g, me, li, 1, ecDrMadisonLi, game.ActivateAbilityParams{})
	effectstest.SetEnergy(t, g, me, 3)
	hand := me.Hand.Size()
	effectstest.Pr7Activate(t, g, me.ID, li, 1, game.ActivateAbilityParams{})
	if effectstest.EnergyOf(me) != 0 || me.Hand.Size() != hand+1 {
		t.Errorf("row 1: energy %d, drew %d", effectstest.EnergyOf(me), me.Hand.Size()-hand)
	}

	// Row 2: an artifact card from your graveyard, tapped.
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, li).Tapped = false })
	relic := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: relic, Name: "Relic", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	})
	targets = []game.TargetRef{{Kind: game.TargetCard, ID: relic}}
	ecRefusedWhenShort(t, g, me, li, 2, ecDrMadisonLi, game.ActivateAbilityParams{Targets: targets})
	effectstest.SetEnergy(t, g, me, 5)
	effectstest.Pr7Activate(t, g, me.ID, li, 2, game.ActivateAbilityParams{Targets: targets})
	if c := effectstest.FindBattlefieldCardForTest(g, relic); effectstest.EnergyOf(me) != 0 || c == nil || !c.Tapped {
		t.Errorf("row 2: energy %d, relic %+v, want it tapped on the battlefield", effectstest.EnergyOf(me), c)
	}
}

func TestEnergyBatchCDynavoltTower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SetEnergy(t, g, me, 0)
	tower := effectstest.B12Push(g, me.ID, "Dynavolt Tower", "Artifact", ecDynavoltTower, 0, 0)
	effectstest.CastCatalogSpell(t, g, "Plain Ponder", "Sorcery", "00000000-0000-0000-0000-0000000e0c04", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EnergyOf(me) != 2 {
		t.Fatalf("a sorcery: energy %d, want 2", effectstest.EnergyOf(me))
	}
	effectstest.CastCatalogSpell(t, g, "Plain Bauble", "Artifact", "00000000-0000-0000-0000-0000000e0c05", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EnergyOf(me) != 2 {
		t.Fatalf("an artifact spell paid: energy %d, want 2", effectstest.EnergyOf(me))
	}
	targets := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	ecRefusedWhenShort(t, g, me, tower, 0, ecDynavoltTower, game.ActivateAbilityParams{Targets: targets})
	effectstest.SetEnergy(t, g, me, 5)
	life := opp.Life
	effectstest.Pr7Activate(t, g, me.ID, tower, 0, game.ActivateAbilityParams{Targets: targets})
	if effectstest.EnergyOf(me) != 0 || opp.Life != life-3 {
		t.Errorf("energy %d, opponent lost %d, want 0 and 3", effectstest.EnergyOf(me), life-opp.Life)
	}
}

func TestEnergyBatchCEraOfInnovation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SetEnergy(t, g, me, 0)
	era := effectstest.B12Push(g, me.ID, "Era of Innovation", "Enchantment", ecEraOfInnovation, 0, 0)
	effectstest.CastCatalogSpell(t, g, "Plain Bauble", "Artifact", "00000000-0000-0000-0000-0000000e0c06", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B06AddMana(me, "C")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EnergyOf(me) != 2 {
		t.Fatalf("paid {1}: energy %d, want 2", effectstest.EnergyOf(me))
	}
	effectstest.CastCatalogSpell(t, g, "Plain Artificer", "Creature — Human Artificer", "00000000-0000-0000-0000-0000000e0c07", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EnergyOf(me) != 2 {
		t.Fatalf("declined: energy %d, want 2", effectstest.EnergyOf(me))
	}

	ecRefusedWhenShort(t, g, me, era, 0, ecEraOfInnovation, game.ActivateAbilityParams{})
	effectstest.SetEnergy(t, g, me, 6)
	hand := me.Hand.Size()
	effectstest.Pr7Activate(t, g, me.ID, era, 0, game.ActivateAbilityParams{})
	if effectstest.EnergyOf(me) != 0 || me.Hand.Size() != hand+3 || effectstest.FindBattlefieldCardForTest(g, era) != nil {
		t.Errorf("energy %d, drew %d, era still out %v", effectstest.EnergyOf(me), me.Hand.Size()-hand, effectstest.FindBattlefieldCardForTest(g, era) != nil)
	}
}
