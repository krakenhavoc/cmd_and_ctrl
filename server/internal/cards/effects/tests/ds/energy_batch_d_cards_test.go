package testsds

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// Every row declares the energy its text prints, and every card is Full.
func TestEnergyBatchDCardsDeclareTheirEnergy(t *testing.T) {
	for _, tc := range []struct {
		oracle string
		row    int
		energy int
		x      bool
	}{
		{effectstest.EdHeliosOne, 1, 0, true},
		{effectstest.EdMultiformWonder, 0, 1, false},
		{effectstest.EdMultiformWonder, 1, 1, false},
		{effectstest.EdPeemaAetherSeer, 0, 3, false},
		{effectstest.EdPeemaTrailblazer, 0, 6, false},
		{effectstest.EdIronworks, 0, 3, false},
		{effectstest.EdRoilCartographer, 0, 6, false},
		{effectstest.EdScurryOfGremlins, 0, 4, false},
		{effectstest.EdSphinxRevelation, 0, 0, true},
		{effectstest.EdStoneIdol, 0, 6, false},
		{effectstest.EdSynthEradicator, 0, 3, false},
	} {
		c := effectstest.SpecActivatedEnergy(t, tc.oracle, tc.row)
		if c.Energy != tc.energy || c.EnergyX != tc.x {
			t.Errorf("%s row %d: energy %d x %v, want %d %v", tc.oracle, tc.row, c.Energy, c.EnergyX, tc.energy, tc.x)
		}
		if spec, _ := effects.Lookup(tc.oracle); spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: completeness %v, want Full", spec.Name, spec.Completeness)
		}
	}
}

// edShort asserts an activation is refused for want of energy and pays
// nothing.
func edShort(t *testing.T, g *game.Game, p *game.Player, source uuid.UUID, row int, params game.ActivateAbilityParams) {
	t.Helper()
	before := effectstest.EnergyOf(p)
	wasTapped := false
	if c := effectstest.FindBattlefieldCardForTest(g, source); c != nil {
		wasTapped = c.Tapped
	}
	if err := g.ActivateCatalogAbility(p.ID, source, row, params); !errors.Is(err, game.ErrInsufficientEnergy) {
		t.Fatalf("row %d with %d energy: err = %v, want ErrInsufficientEnergy", row, before, err)
	}
	if effectstest.EnergyOf(p) != before {
		t.Errorf("a refused activation changed energy %d → %d", before, effectstest.EnergyOf(p))
	}
	if c := effectstest.FindBattlefieldCardForTest(g, source); c != nil && c.Tapped != wasTapped {
		t.Error("a refused activation tapped its source")
	}
}

// HELIOS One: X is the energy paid and the mana value the target must
// have; the land is sacrificed; the enumerator offers the X that fits.
func TestHeliosOneDestroysAtTheManaValueItPaysFor(t *testing.T) {
	g, me, opp := effectstest.P7Table(t)
	helios := effectstest.EdCard(t, g, me, "HELIOS One", effectstest.EdHeliosOne, "Land", 0, 0)
	two := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Two Drop", TypeLine: "Artifact", ManaCost: "{2}"})
	three := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Three Drop", TypeLine: "Artifact", ManaCost: "{3}"})

	effectstest.P7Activate(t, g, me, helios, 0, game.ActivateAbilityParams{})
	if effectstest.EnergyOf(me) != 1 {
		t.Fatalf("{1}, {T}: energy %d, want 1", effectstest.EnergyOf(me))
	}
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, helios).Tapped = false })
	effectstest.SetEnergy(t, g, me, 2)
	edShort(t, g, me, helios, 1, game.ActivateAbilityParams{XValue: 3, Targets: effectstest.EdCardTarget(three)})
	if err := g.ActivateCatalogAbility(me.ID, helios, 1, game.ActivateAbilityParams{XValue: 2, Targets: effectstest.EdCardTarget(three)}); err == nil {
		t.Fatal("X=2 accepted a mana-value-3 target")
	}

	// The enumerator offers X=2 at the two-drop, and nothing else.
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
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

	effectstest.P7Activate(t, g, me, helios, 1, game.ActivateAbilityParams{XValue: 2, Targets: effectstest.EdCardTarget(two)})
	if effectstest.EnergyOf(me) != 0 {
		t.Errorf("energy %d, want 0", effectstest.EnergyOf(me))
	}
	if effectstest.FindBattlefieldCardForTest(g, two) != nil {
		t.Error("the two-drop survived")
	}
	if effectstest.FindBattlefieldCardForTest(g, helios) != nil {
		t.Error("HELIOS One was not sacrificed")
	}
	if effectstest.FindBattlefieldCardForTest(g, three) == nil {
		t.Error("the three-drop was destroyed")
	}
}

// Multiform Wonder: each row pays one energy, and the choice is made as
// it resolves.
func TestMultiformWonderPaysOneEnergyAndAsks(t *testing.T) {
	g, me, _ := effectstest.P7Table(t)
	w := effectstest.EdCard(t, g, me, "Multiform Wonder", effectstest.EdMultiformWonder, "Artifact Creature — Construct", 3, 3)
	edShort(t, g, me, w, 0, game.ActivateAbilityParams{})
	effectstest.SetEnergy(t, g, me, 2)

	if err := g.ActivateCatalogAbility(me.ID, w, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 2)
	if c := effectstest.ApaLive(g, w); !game.HasKeyword(c, "lifelink") || game.HasKeyword(c, "flying") {
		t.Errorf("keywords %v, want lifelink only", c.Keywords)
	}

	if err := g.ActivateCatalogAbility(me.ID, w, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ConfirmChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no +2/-2 or -2/+2 question")
	}
	if err := g.ResolveConfirm(c.ID, me.ID, false); err != nil {
		t.Fatal(err)
	}
	if live := effectstest.ApaLive(g, w); live.CurrentPower() != 1 || live.CurrentToughness() != 5 {
		t.Errorf("P/T %d/%d, want 1/5", live.CurrentPower(), live.CurrentToughness())
	}
	if effectstest.EnergyOf(me) != 0 {
		t.Errorf("energy %d, want 0", effectstest.EnergyOf(me))
	}
}

// Peema Aether-Seer: three energy makes the target block this turn.
func TestPeemaAetherSeerMakesItsTargetBlock(t *testing.T) {
	g, me, opp := effectstest.P7Table(t)
	seer := effectstest.EdCard(t, g, me, "Peema Aether-Seer", effectstest.EdPeemaAetherSeer, "Creature — Elf Druid", 3, 2)
	bear := effectstest.EdCard(t, g, opp, "Opp Bear", "", "Creature — Bear", 2, 2)
	effectstest.SetEnergy(t, g, me, 2)
	edShort(t, g, me, seer, 0, game.ActivateAbilityParams{Targets: effectstest.EdCardTarget(bear)})
	effectstest.SetEnergy(t, g, me, 3)
	effectstest.P7Activate(t, g, me, seer, 0, game.ActivateAbilityParams{Targets: effectstest.EdCardTarget(bear)})
	if effectstest.EnergyOf(me) != 0 {
		t.Errorf("energy %d, want 0", effectstest.EnergyOf(me))
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
	g, me, opp := effectstest.P7Table(t)
	tb := effectstest.EdCard(t, g, me, "Peema Trailblazer", effectstest.EdPeemaTrailblazer, "Creature — Elephant Warrior", 3, 3)
	effectstest.AttackWith(t, g, opp.ID, tb)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EnergyOf(me) != 3 {
		t.Fatalf("energy after three combat damage %d, want 3", effectstest.EnergyOf(me))
	}
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)

	edShort(t, g, me, tb, 0, game.ActivateAbilityParams{})
	effectstest.SetEnergy(t, g, me, 6)
	hand := me.Hand.Size()
	effectstest.P7Activate(t, g, me, tb, 0, game.ActivateAbilityParams{})
	if got := effectstest.ApaLive(g, tb).Counters[game.CounterPlusOne]; got != 2 {
		t.Errorf("+1/+1 counters %d, want 2", got)
	}
	if got := me.Hand.Size() - hand; got != 5 {
		t.Errorf("drew %d, want 5 (its power after the counters)", got)
	}
	effectstest.SetEnergy(t, g, me, 6)
	if err := g.ActivateCatalogAbility(me.ID, tb, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrAbilityExhausted) {
		t.Errorf("second exhaust: err = %v, want ErrAbilityExhausted", err)
	}
}

// Phyrexian Ironworks: one energy per attack, however many attack; three
// energy makes a Phyrexian Golem.
func TestPhyrexianIronworksGetsEnergyAndMakesAGolem(t *testing.T) {
	g, me, opp := effectstest.P7Table(t)
	works := effectstest.PushPermanentForTest(g, me.ID, "Phyrexian Ironworks", effectstest.EdIronworks, "Artifact")
	effectstest.SetEnergy(t, g, me, 2)
	edShort(t, g, me, works, 0, game.ActivateAbilityParams{})
	effectstest.SetEnergy(t, g, me, 3)
	effectstest.P7Activate(t, g, me, works, 0, game.ActivateAbilityParams{})
	if effectstest.EnergyOf(me) != 0 {
		t.Errorf("energy %d, want 0", effectstest.EnergyOf(me))
	}
	golem := effectstest.FindBattlefieldByName(g, "Phyrexian Golem")
	if golem == uuid.Nil {
		t.Fatal("no Phyrexian Golem")
	}
	if c := effectstest.ApaLive(g, golem); c.CurrentPower() != 3 || c.CurrentToughness() != 3 || !c.IsArtifact() {
		t.Errorf("Golem %d/%d artifact=%v, want a 3/3 artifact", c.CurrentPower(), c.CurrentToughness(), c.IsArtifact())
	}

	a := effectstest.EdCard(t, g, me, "Bear A", "", "Creature — Bear", 2, 2)
	b := effectstest.EdCard(t, g, me, "Bear B", "", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EnergyOf(me) != 1 {
		t.Errorf("energy after one attack with two creatures %d, want 1", effectstest.EnergyOf(me))
	}
}

// Roil Cartographer: six energy and {T} draw three.
func TestRoilCartographerDrawsThreeForSix(t *testing.T) {
	g, me, _ := effectstest.P7Table(t)
	rc := effectstest.EdCard(t, g, me, "Roil Cartographer", effectstest.EdRoilCartographer, "Creature — Merfolk Rogue", 1, 3)
	effectstest.SetEnergy(t, g, me, 5)
	edShort(t, g, me, rc, 0, game.ActivateAbilityParams{})
	effectstest.SetEnergy(t, g, me, 7)
	hand := me.Hand.Size()
	effectstest.P7Activate(t, g, me, rc, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d, want 3", got)
	}
	if effectstest.EnergyOf(me) != 1 {
		t.Errorf("energy %d, want 1", effectstest.EnergyOf(me))
	}
}

// Scurry of Gremlins: two Gremlins, then energy for every creature you
// control (them included); four energy pumps the team and gives haste.
func TestScurryOfGremlinsCountsTheGremlinsAndPumps(t *testing.T) {
	g, me, _ := effectstest.P7Table(t)
	bear := effectstest.EdCard(t, g, me, "Bear", "", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Scurry of Gremlins", "Enchantment", effectstest.EdScurryOfGremlins, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EnergyOf(me) != 3 {
		t.Fatalf("energy %d, want 3 (two Gremlins and the Bear)", effectstest.EnergyOf(me))
	}
	scurry := effectstest.FindBattlefieldByName(g, "Scurry of Gremlins")
	edShort(t, g, me, scurry, 0, game.ActivateAbilityParams{})
	effectstest.SetEnergy(t, g, me, 4)
	effectstest.P7Activate(t, g, me, scurry, 0, game.ActivateAbilityParams{})
	if c := effectstest.ApaLive(g, bear); c.CurrentPower() != 3 || !game.HasKeyword(c, "haste") {
		t.Errorf("Bear %d power haste=%v, want 3 and haste", c.CurrentPower(), game.HasKeyword(c, "haste"))
	}
	if effectstest.EnergyOf(me) != 0 {
		t.Errorf("energy %d, want 0", effectstest.EnergyOf(me))
	}
}

// Sphinx of the Revelation: life gained becomes energy; X energy draws X.
func TestSphinxOfTheRevelationTurnsLifeIntoCards(t *testing.T) {
	g, me, _ := effectstest.P7Table(t)
	sphinx := effectstest.EdCard(t, g, me, "Sphinx of the Revelation", effectstest.EdSphinxRevelation, "Artifact Creature — Sphinx", 4, 5)
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EnergyOf(me) != 3 {
		t.Fatalf("energy after gaining 3 life %d, want 3", effectstest.EnergyOf(me))
	}
	edShort(t, g, me, sphinx, 0, game.ActivateAbilityParams{XValue: 4})
	hand := me.Hand.Size()
	effectstest.P7Activate(t, g, me, sphinx, 0, game.ActivateAbilityParams{XValue: 2})
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
	if effectstest.EnergyOf(me) != 1 {
		t.Errorf("energy %d, want 1", effectstest.EnergyOf(me))
	}
}

// Stone Idol Generator: one energy per attacking creature; six makes the
// 6/12 Construct.
func TestStoneIdolGeneratorCountsAttackersAndMakesAConstruct(t *testing.T) {
	g, me, opp := effectstest.P7Table(t)
	gen := effectstest.PushPermanentForTest(g, me.ID, "Stone Idol Generator", effectstest.EdStoneIdol, "Artifact")
	effectstest.SetEnergy(t, g, me, 5)
	edShort(t, g, me, gen, 0, game.ActivateAbilityParams{})
	effectstest.SetEnergy(t, g, me, 6)
	effectstest.P7Activate(t, g, me, gen, 0, game.ActivateAbilityParams{})
	construct := effectstest.FindBattlefieldByName(g, "Construct")
	if construct == uuid.Nil {
		t.Fatal("no Construct")
	}
	if c := effectstest.ApaLive(g, construct); c.CurrentPower() != 6 || c.CurrentToughness() != 12 || !game.HasKeyword(c, "trample") {
		t.Errorf("Construct %d/%d, want a 6/12 with trample", c.CurrentPower(), c.CurrentToughness())
	}

	a := effectstest.EdCard(t, g, me, "Bear A", "", "Creature — Bear", 2, 2)
	b := effectstest.EdCard(t, g, me, "Bear B", "", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EnergyOf(me) != 2 {
		t.Errorf("energy after two attackers %d, want 2", effectstest.EnergyOf(me))
	}
}

// Synth Eradicator: three energy pings for 3; the attack trigger exiles
// the top card and either gives two energy or the card for the turn.
func TestSynthEradicatorPingsAndImpulses(t *testing.T) {
	g, me, opp := effectstest.P7Table(t)
	synth := effectstest.EdCard(t, g, me, "Synth Eradicator", effectstest.EdSynthEradicator, "Artifact Creature — Synth Soldier", 3, 3)
	effectstest.SetEnergy(t, g, me, 2)
	edShort(t, g, me, synth, 0, game.ActivateAbilityParams{Targets: effectstest.P8Player(opp.ID)})
	effectstest.SetEnergy(t, g, me, 3)
	life := opp.Life
	effectstest.P7Activate(t, g, me, synth, 0, game.ActivateAbilityParams{Targets: effectstest.P8Player(opp.ID)})
	if opp.Life != life-3 || effectstest.EnergyOf(me) != 0 {
		t.Errorf("life %d energy %d, want %d and 0", opp.Life, effectstest.EnergyOf(me), life-3)
	}

	for _, take := range []bool{true, false} {
		g, me, opp := effectstest.P7Table(t)
		synth := effectstest.EdCard(t, g, me, "Synth Eradicator", effectstest.EdSynthEradicator, "Artifact Creature — Synth Soldier", 3, 3)
		top := effectstest.P8LibraryTop(g, me, game.Card{Name: "Top Card", TypeLine: "Sorcery", ManaCost: "{R}"})
		effectstest.DeclareAttack(t, g, opp.ID, synth)
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.ConfirmChoiceFor(g, me.ID)
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
			if effectstest.EnergyOf(me) != 2 || perm != nil {
				t.Errorf("took the energy: energy %d permission %v, want 2 and none", effectstest.EnergyOf(me), perm)
			}
		} else if effectstest.EnergyOf(me) != 0 || perm == nil {
			t.Errorf("declined: energy %d permission %v, want 0 and a permission", effectstest.EnergyOf(me), perm)
		}
	}
}
