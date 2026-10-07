package effects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// energy_batch_d_cards_test.go — ADR 0129 PR 1's energy cards, batch D:
// each pays energy as printed (CR 107.14), is refused when short
// (CR 118.3) with nothing paid, and does its effect.

const (
	edHeliosOne        = "cfb1a656-0bf1-484d-b099-33087914250b"
	edMultiformWonder  = "9849c743-9492-4a1b-83bf-1be54614b85f"
	edPeemaAetherSeer  = "4a00f56e-e558-4239-bc75-fb0af7639042"
	edPeemaTrailblazer = "88600bd4-4dcc-4788-bba4-78b6cf5ad8f8"
	edIronworks        = "b3c87cb9-a37a-428e-95d1-e6dd90b31097"
	edRoilCartographer = "3cd7ef87-4cb4-42ff-842f-e9a465dab165"
	edScurryOfGremlins = "ff24d18d-9c5d-4ffe-846f-a8d568b32724"
	edSphinxRevelation = "9b8f6c6a-49a1-4061-a2c4-5be612e138fa"
	edStoneIdol        = "411c3636-a5b3-42fa-8020-ceacf581a4f3"
	edSynthEradicator  = "c68f09d8-a1eb-449b-8bf5-4251d7a14337"
)

// Every row declares the energy its text prints, and every card is Full.
func TestEnergyBatchDCardsDeclareTheirEnergy(t *testing.T) {
	for _, tc := range []struct {
		oracle string
		row    int
		energy int
		x      bool
	}{
		{edHeliosOne, 1, 0, true},
		{edMultiformWonder, 0, 1, false},
		{edMultiformWonder, 1, 1, false},
		{edPeemaAetherSeer, 0, 3, false},
		{edPeemaTrailblazer, 0, 6, false},
		{edIronworks, 0, 3, false},
		{edRoilCartographer, 0, 6, false},
		{edScurryOfGremlins, 0, 4, false},
		{edSphinxRevelation, 0, 0, true},
		{edStoneIdol, 0, 6, false},
		{edSynthEradicator, 0, 3, false},
	} {
		c := specActivatedEnergy(t, tc.oracle, tc.row)
		if c.Energy != tc.energy || c.EnergyX != tc.x {
			t.Errorf("%s row %d: energy %d x %v, want %d %v", tc.oracle, tc.row, c.Energy, c.EnergyX, tc.energy, tc.x)
		}
		if spec, _ := Lookup(tc.oracle); spec.Completeness != CompletenessFull {
			t.Errorf("%s: completeness %v, want Full", spec.Name, spec.Completeness)
		}
	}
}

// edShort asserts an activation is refused for want of energy and pays
// nothing.
func edShort(t *testing.T, g *game.Game, p *game.Player, source uuid.UUID, row int, params game.ActivateAbilityParams) {
	t.Helper()
	before := energyOf(p)
	wasTapped := false
	if c := findBattlefieldCardForTest(g, source); c != nil {
		wasTapped = c.Tapped
	}
	if err := g.ActivateCatalogAbility(p.ID, source, row, params); !errors.Is(err, game.ErrInsufficientEnergy) {
		t.Fatalf("row %d with %d energy: err = %v, want ErrInsufficientEnergy", row, before, err)
	}
	if energyOf(p) != before {
		t.Errorf("a refused activation changed energy %d → %d", before, energyOf(p))
	}
	if c := findBattlefieldCardForTest(g, source); c != nil && c.Tapped != wasTapped {
		t.Error("a refused activation tapped its source")
	}
}

func edCard(t *testing.T, g *game.Game, owner *game.Player, name, oracle, typeLine string, power, toughness int) uuid.UUID {
	t.Helper()
	return apaPush(g, owner.ID, owner.ID, game.Card{Name: name, OracleID: oracle, TypeLine: typeLine, Power: power, Toughness: toughness})
}

func edCardTarget(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// HELIOS One: X is the energy paid and the mana value the target must
// have; the land is sacrificed; the enumerator offers the X that fits.
func TestHeliosOneDestroysAtTheManaValueItPaysFor(t *testing.T) {
	g, me, opp := p7Table(t)
	helios := edCard(t, g, me, "HELIOS One", edHeliosOne, "Land", 0, 0)
	two := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Two Drop", TypeLine: "Artifact", ManaCost: "{2}"})
	three := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Three Drop", TypeLine: "Artifact", ManaCost: "{3}"})

	p7Activate(t, g, me, helios, 0, game.ActivateAbilityParams{})
	if energyOf(me) != 1 {
		t.Fatalf("{1}, {T}: energy %d, want 1", energyOf(me))
	}
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, helios).Tapped = false })
	setEnergy(t, g, me, 2)
	edShort(t, g, me, helios, 1, game.ActivateAbilityParams{XValue: 3, Targets: edCardTarget(three)})
	if err := g.ActivateCatalogAbility(me.ID, helios, 1, game.ActivateAbilityParams{XValue: 2, Targets: edCardTarget(three)}); err == nil {
		t.Fatal("X=2 accepted a mana-value-3 target")
	}

	// The enumerator offers X=2 at the two-drop, and nothing else.
	floatMana(t, g, me, "{C}{C}{C}")
	moves := legal.EnumerateFor(g, me.ID)
	found := false
	for _, m := range moves {
		if m.Source != helios || m.Kind != legal.KindActivate {
			continue
		}
		var p struct {
			AbilityIndex int `json:"ability_index"`
			XValue       int `json:"x_value"`
			Targets      []struct {
				ID string `json:"id"`
			} `json:"targets"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		if p.AbilityIndex != 1 {
			continue
		}
		if p.XValue == 2 && len(p.Targets) == 1 && p.Targets[0].ID == two.String() {
			found = true
			if m.Cost == nil || m.Cost.Energy != 2 {
				t.Errorf("cost = %+v, want energy 2", m.Cost)
			}
			continue
		}
		t.Errorf("unexpected HELIOS move %q %s", m.Label, m.Params)
	}
	if !found {
		t.Fatal("the enumerator did not offer X=2 at the two-drop")
	}

	p7Activate(t, g, me, helios, 1, game.ActivateAbilityParams{XValue: 2, Targets: edCardTarget(two)})
	if energyOf(me) != 0 {
		t.Errorf("energy %d, want 0", energyOf(me))
	}
	if findBattlefieldCardForTest(g, two) != nil {
		t.Error("the two-drop survived")
	}
	if findBattlefieldCardForTest(g, helios) != nil {
		t.Error("HELIOS One was not sacrificed")
	}
	if findBattlefieldCardForTest(g, three) == nil {
		t.Error("the three-drop was destroyed")
	}
}

// Multiform Wonder: each row pays one energy, and the choice is made as
// it resolves.
func TestMultiformWonderPaysOneEnergyAndAsks(t *testing.T) {
	g, me, _ := p7Table(t)
	w := edCard(t, g, me, "Multiform Wonder", edMultiformWonder, "Artifact Creature — Construct", 3, 3)
	edShort(t, g, me, w, 0, game.ActivateAbilityParams{})
	setEnergy(t, g, me, 2)

	if err := g.ActivateCatalogAbility(me.ID, w, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 2)
	if c := apaLive(g, w); !game.HasKeyword(c, "lifelink") || game.HasKeyword(c, "flying") {
		t.Errorf("keywords %v, want lifelink only", c.Keywords)
	}

	if err := g.ActivateCatalogAbility(me.ID, w, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	c := confirmChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no +2/-2 or -2/+2 question")
	}
	if err := g.ResolveConfirm(c.ID, me.ID, false); err != nil {
		t.Fatal(err)
	}
	if live := apaLive(g, w); live.CurrentPower() != 1 || live.CurrentToughness() != 5 {
		t.Errorf("P/T %d/%d, want 1/5", live.CurrentPower(), live.CurrentToughness())
	}
	if energyOf(me) != 0 {
		t.Errorf("energy %d, want 0", energyOf(me))
	}
}

// Peema Aether-Seer: three energy makes the target block this turn.
func TestPeemaAetherSeerMakesItsTargetBlock(t *testing.T) {
	g, me, opp := p7Table(t)
	seer := edCard(t, g, me, "Peema Aether-Seer", edPeemaAetherSeer, "Creature — Elf Druid", 3, 2)
	bear := edCard(t, g, opp, "Opp Bear", "", "Creature — Bear", 2, 2)
	setEnergy(t, g, me, 2)
	edShort(t, g, me, seer, 0, game.ActivateAbilityParams{Targets: edCardTarget(bear)})
	setEnergy(t, g, me, 3)
	p7Activate(t, g, me, seer, 0, game.ActivateAbilityParams{Targets: edCardTarget(bear)})
	if energyOf(me) != 0 {
		t.Errorf("energy %d, want 0", energyOf(me))
	}
	found := false
	for _, e := range g.ScopedEffects {
		for _, m := range e.Mods {
			if m.Kind == game.ModAddBlockRequirement && m.Text == string(game.BlockRequirementBlocks) {
				found = true
			}
		}
	}
	if !found {
		t.Error("no blocks-if-able requirement was put on the target")
	}
}

// Peema Trailblazer: combat damage to a player gives that much energy;
// the exhaust pays six, adds two counters, then draws the greatest
// power, once.
func TestPeemaTrailblazerGetsDamageAsEnergyAndExhausts(t *testing.T) {
	g, me, opp := p7Table(t)
	tb := edCard(t, g, me, "Peema Trailblazer", edPeemaTrailblazer, "Creature — Elephant Warrior", 3, 3)
	attackWith(t, g, opp.ID, tb)
	passPriorityAroundTable(t, g)
	if energyOf(me) != 3 {
		t.Fatalf("energy after three combat damage %d, want 3", energyOf(me))
	}
	advanceTo(t, g, game.StepPostcombatMain)

	edShort(t, g, me, tb, 0, game.ActivateAbilityParams{})
	setEnergy(t, g, me, 6)
	hand := me.Hand.Size()
	p7Activate(t, g, me, tb, 0, game.ActivateAbilityParams{})
	if got := apaLive(g, tb).Counters[game.CounterPlusOne]; got != 2 {
		t.Errorf("+1/+1 counters %d, want 2", got)
	}
	if got := me.Hand.Size() - hand; got != 5 {
		t.Errorf("drew %d, want 5 (its power after the counters)", got)
	}
	setEnergy(t, g, me, 6)
	if err := g.ActivateCatalogAbility(me.ID, tb, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrAbilityExhausted) {
		t.Errorf("second exhaust: err = %v, want ErrAbilityExhausted", err)
	}
}

// Phyrexian Ironworks: one energy per attack, however many attack; three
// energy makes a Phyrexian Golem.
func TestPhyrexianIronworksGetsEnergyAndMakesAGolem(t *testing.T) {
	g, me, opp := p7Table(t)
	works := pushPermanentForTest(g, me.ID, "Phyrexian Ironworks", edIronworks, "Artifact")
	setEnergy(t, g, me, 2)
	edShort(t, g, me, works, 0, game.ActivateAbilityParams{})
	setEnergy(t, g, me, 3)
	p7Activate(t, g, me, works, 0, game.ActivateAbilityParams{})
	if energyOf(me) != 0 {
		t.Errorf("energy %d, want 0", energyOf(me))
	}
	golem := findBattlefieldByName(g, "Phyrexian Golem")
	if golem == uuid.Nil {
		t.Fatal("no Phyrexian Golem")
	}
	if c := apaLive(g, golem); c.CurrentPower() != 3 || c.CurrentToughness() != 3 || !c.IsArtifact() {
		t.Errorf("Golem %d/%d artifact=%v, want a 3/3 artifact", c.CurrentPower(), c.CurrentToughness(), c.IsArtifact())
	}

	a := edCard(t, g, me, "Bear A", "", "Creature — Bear", 2, 2)
	b := edCard(t, g, me, "Bear B", "", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, a, b)
	passPriorityAroundTable(t, g)
	if energyOf(me) != 1 {
		t.Errorf("energy after one attack with two creatures %d, want 1", energyOf(me))
	}
}

// Roil Cartographer: six energy and {T} draw three.
func TestRoilCartographerDrawsThreeForSix(t *testing.T) {
	g, me, _ := p7Table(t)
	rc := edCard(t, g, me, "Roil Cartographer", edRoilCartographer, "Creature — Merfolk Rogue", 1, 3)
	setEnergy(t, g, me, 5)
	edShort(t, g, me, rc, 0, game.ActivateAbilityParams{})
	setEnergy(t, g, me, 7)
	hand := me.Hand.Size()
	p7Activate(t, g, me, rc, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d, want 3", got)
	}
	if energyOf(me) != 1 {
		t.Errorf("energy %d, want 1", energyOf(me))
	}
}

// Scurry of Gremlins: two Gremlins, then energy for every creature you
// control (them included); four energy pumps the team and gives haste.
func TestScurryOfGremlinsCountsTheGremlinsAndPumps(t *testing.T) {
	g, me, _ := p7Table(t)
	bear := edCard(t, g, me, "Bear", "", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Scurry of Gremlins", "Enchantment", edScurryOfGremlins, nil)
	passPriorityAroundTable(t, g)
	if energyOf(me) != 3 {
		t.Fatalf("energy %d, want 3 (two Gremlins and the Bear)", energyOf(me))
	}
	scurry := findBattlefieldByName(g, "Scurry of Gremlins")
	edShort(t, g, me, scurry, 0, game.ActivateAbilityParams{})
	setEnergy(t, g, me, 4)
	p7Activate(t, g, me, scurry, 0, game.ActivateAbilityParams{})
	if c := apaLive(g, bear); c.CurrentPower() != 3 || !game.HasKeyword(c, "haste") {
		t.Errorf("Bear %d power haste=%v, want 3 and haste", c.CurrentPower(), game.HasKeyword(c, "haste"))
	}
	if energyOf(me) != 0 {
		t.Errorf("energy %d, want 0", energyOf(me))
	}
}

// Sphinx of the Revelation: life gained becomes energy; X energy draws X.
func TestSphinxOfTheRevelationTurnsLifeIntoCards(t *testing.T) {
	g, me, _ := p7Table(t)
	sphinx := edCard(t, g, me, "Sphinx of the Revelation", edSphinxRevelation, "Artifact Creature — Sphinx", 4, 5)
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if energyOf(me) != 3 {
		t.Fatalf("energy after gaining 3 life %d, want 3", energyOf(me))
	}
	edShort(t, g, me, sphinx, 0, game.ActivateAbilityParams{XValue: 4})
	hand := me.Hand.Size()
	p7Activate(t, g, me, sphinx, 0, game.ActivateAbilityParams{XValue: 2})
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
	if energyOf(me) != 1 {
		t.Errorf("energy %d, want 1", energyOf(me))
	}
}

// Stone Idol Generator: one energy per attacking creature; six makes the
// 6/12 Construct.
func TestStoneIdolGeneratorCountsAttackersAndMakesAConstruct(t *testing.T) {
	g, me, opp := p7Table(t)
	gen := pushPermanentForTest(g, me.ID, "Stone Idol Generator", edStoneIdol, "Artifact")
	setEnergy(t, g, me, 5)
	edShort(t, g, me, gen, 0, game.ActivateAbilityParams{})
	setEnergy(t, g, me, 6)
	p7Activate(t, g, me, gen, 0, game.ActivateAbilityParams{})
	construct := findBattlefieldByName(g, "Construct")
	if construct == uuid.Nil {
		t.Fatal("no Construct")
	}
	if c := apaLive(g, construct); c.CurrentPower() != 6 || c.CurrentToughness() != 12 || !game.HasKeyword(c, "trample") {
		t.Errorf("Construct %d/%d, want a 6/12 with trample", c.CurrentPower(), c.CurrentToughness())
	}

	a := edCard(t, g, me, "Bear A", "", "Creature — Bear", 2, 2)
	b := edCard(t, g, me, "Bear B", "", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, a, b)
	passPriorityAroundTable(t, g)
	if energyOf(me) != 2 {
		t.Errorf("energy after two attackers %d, want 2", energyOf(me))
	}
}

// Synth Eradicator: three energy pings for 3; the attack trigger exiles
// the top card and either gives two energy or the card for the turn.
func TestSynthEradicatorPingsAndImpulses(t *testing.T) {
	g, me, opp := p7Table(t)
	synth := edCard(t, g, me, "Synth Eradicator", edSynthEradicator, "Artifact Creature — Synth Soldier", 3, 3)
	setEnergy(t, g, me, 2)
	edShort(t, g, me, synth, 0, game.ActivateAbilityParams{Targets: p8Player(opp.ID)})
	setEnergy(t, g, me, 3)
	life := opp.Life
	p7Activate(t, g, me, synth, 0, game.ActivateAbilityParams{Targets: p8Player(opp.ID)})
	if opp.Life != life-3 || energyOf(me) != 0 {
		t.Errorf("life %d energy %d, want %d and 0", opp.Life, energyOf(me), life-3)
	}

	for _, take := range []bool{true, false} {
		g, me, opp := p7Table(t)
		synth := edCard(t, g, me, "Synth Eradicator", edSynthEradicator, "Artifact Creature — Synth Soldier", 3, 3)
		top := p8LibraryTop(g, me, game.Card{Name: "Top Card", TypeLine: "Sorcery", ManaCost: "{R}"})
		declareAttack(t, g, opp.ID, synth)
		passPriorityAroundTable(t, g)
		c := confirmChoiceFor(g, me.ID)
		if c == nil {
			t.Fatal("no energy-or-card question")
		}
		if !g.Exile.Contains(top) {
			t.Fatal("the top card was not exiled before the question")
		}
		if err := g.ResolveConfirm(c.ID, me.ID, take); err != nil {
			t.Fatal(err)
		}
		var perm *game.CastPermission
		g.WithWriteLock(func() { perm = g.CastPermissionOnCardByIDForEffect(top) })
		if take {
			if energyOf(me) != 2 || perm != nil {
				t.Errorf("took the energy: energy %d permission %v, want 2 and none", energyOf(me), perm)
			}
		} else if energyOf(me) != 0 || perm == nil {
			t.Errorf("declined: energy %d permission %v, want 0 and a permission", energyOf(me), perm)
		}
	}
}
