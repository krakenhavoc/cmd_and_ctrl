package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b260_cards_test.go — Kaladesh, batch 260 (#5539). Every test plays
// the real catalog entry through the engine.

const (
	kldB260Guardian     = "1342bd59-91d3-4a5b-b853-542a52d7da6d"
	kldB260Courage      = "1aaa85e1-4770-457f-a9af-3eeb56c71447"
	kldB260Might        = "38462d26-f252-4055-9f1a-59bc3ab2c75e"
	kldB260Trinketeer   = "aa57f607-c5eb-45a7-bfb1-adce0aef49c7"
	kldB260StartEngines = "bbe3726b-e14d-48d1-8bbf-40d48f1ef8c7"
	kldB260BuiltToSmash = "c363f338-0777-45e6-b13c-f15ba026b63a"
	kldB260Exemplar     = "c51b0667-a24e-44c0-bc95-bf61b393f3cf"
	kldB260Bear         = "Creature — Bear"
)

func TestKldB260EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		kldB260Guardian:     "Guardian of the Great Conduit",
		kldB260Courage:      "Ornamental Courage",
		kldB260Might:        "Engineered Might",
		kldB260Trinketeer:   "Master Trinketeer",
		kldB260StartEngines: "Start Your Engines",
		kldB260BuiltToSmash: "Built to Smash",
		kldB260Exemplar:     "Toolcraft Exemplar",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Guardian of the Great Conduit -------------------------------------------

func TestKldB260GuardianIsBiggerAndVigilantWhileYouControlANissa(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	guardian := effectstest.B12Push(g, me.ID, "Guardian of the Great Conduit", "Creature — Elemental", kldB260Guardian, 2, 4)
	effectstest.AssertKeywords(t, g, guardian, "reach")
	if p := effectstest.EffectivePower(t, g, guardian); p != 2 {
		t.Fatalf("power %d with no Nissa, want 2", p)
	}
	if effectstest.HasEffectiveKeyword(t, g, guardian, "vigilance") {
		t.Fatal("vigilance without a Nissa planeswalker")
	}
	// A planeswalker that is not a Nissa, and an opponent's Nissa, change nothing.
	effectstest.B12Push(g, me.ID, "Jace", "Legendary Planeswalker — Jace", "", 0, 0)
	effectstest.B12Push(g, opp.ID, "Nissa, Their Walker", "Legendary Planeswalker — Nissa", "", 0, 0)
	if p := effectstest.EffectivePower(t, g, guardian); p != 2 {
		t.Errorf("power %d, want 2: only your own Nissa counts", p)
	}
	nissa := effectstest.B12Push(g, me.ID, "Nissa, Steward of Elements", "Legendary Planeswalker — Nissa", "", 0, 0)
	if p, tough := effectstest.EffectivePower(t, g, guardian), effectstest.EffectiveToughness(t, g, guardian); p != 4 || tough != 4 {
		t.Errorf("with a Nissa the Guardian is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, guardian, "vigilance") {
		t.Error("no vigilance with a Nissa planeswalker")
	}
	effectstest.B27Kill(g, nissa)
	if p := effectstest.EffectivePower(t, g, guardian); p != 2 {
		t.Errorf("power %d after the Nissa left, want 2", p)
	}
	if effectstest.HasEffectiveKeyword(t, g, guardian, "vigilance") {
		t.Error("still vigilant after the Nissa left")
	}
}

// --- Ornamental Courage ----------------------------------------------------------

func TestKldB260OrnamentalCourageUntapsAndPumps(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	bear := effectstest.B12Creature(g, me.ID, "Bear", kldB260Bear, 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.CastCatalogSpell(t, g, "Ornamental Courage", "Instant", kldB260Courage, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.FindBattlefieldCardByID(g, bear)
	if c == nil || c.Tapped {
		t.Fatal("the bear was not untapped")
	}
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || tough != 5 {
		t.Errorf("bear is %d/%d, want 3/5", p, tough)
	}
}

func TestKldB260OrnamentalCourageDoesNothingWithoutItsTarget(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	bear := effectstest.B12Creature(g, me.ID, "Bear", kldB260Bear, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Ornamental Courage", "Instant", kldB260Courage, effectstest.B16TargetCard(bear))
	effectstest.B27Kill(g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardByID(g, bear) != nil {
		t.Error("the bear came back")
	}
}

// --- Built to Smash ------------------------------------------------------------------

func TestKldB260BuiltToSmashPumpsAnAttackerAndGivesAnArtifactTrample(t *testing.T) {
	for _, artifact := range []bool{false, true} {
		name := "plain creature"
		typeLine := kldB260Bear
		if artifact {
			name = "artifact creature"
			typeLine = "Artifact Creature — Construct"
		}
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			attacker := effectstest.B12Creature(g, me.ID, "Attacker", typeLine, 2, 2)
			idle := effectstest.B12Creature(g, me.ID, "Idle", kldB260Bear, 2, 2)
			effectstest.MscM1Attacking(g, attacker, opp.ID)
			if err := effectstest.CastCatalogSpellErr(t, g, "Built to Smash", "Instant", kldB260BuiltToSmash, effectstest.B16TargetCard(idle)); err == nil {
				t.Error("a creature that is not attacking was accepted")
			}
			effectstest.CastCatalogSpell(t, g, "Built to Smash", "Instant", kldB260BuiltToSmash, effectstest.B16TargetCard(attacker))
			effectstest.PassPriorityAroundTable(t, g)
			if p, tough := effectstest.EffectivePower(t, g, attacker), effectstest.EffectiveToughness(t, g, attacker); p != 5 || tough != 5 {
				t.Errorf("attacker is %d/%d, want 5/5", p, tough)
			}
			if got := effectstest.HasEffectiveKeyword(t, g, attacker, "trample"); got != artifact {
				t.Errorf("trample = %v, want %v", got, artifact)
			}
		})
	}
}

// --- Engineered Might -----------------------------------------------------------------

func TestKldB260EngineeredMightBothModes(t *testing.T) {
	t.Run("target creature", func(t *testing.T) {
		g, me, _ := effectstest.SpendTable(t)
		a := effectstest.B12Creature(g, me.ID, "A", kldB260Bear, 2, 2)
		b := effectstest.B12Creature(g, me.ID, "B", kldB260Bear, 2, 2)
		effectstest.CastCatalogSpellWithModesTargets(t, g, kldB260Might, []int{0}, effectstest.B16TargetCard(a))
		effectstest.PassPriorityAroundTable(t, g)
		if p, tough := effectstest.EffectivePower(t, g, a), effectstest.EffectiveToughness(t, g, a); p != 7 || tough != 7 {
			t.Errorf("the target is %d/%d, want 7/7", p, tough)
		}
		effectstest.AssertKeywords(t, g, a, "trample")
		if p := effectstest.EffectivePower(t, g, b); p != 2 {
			t.Errorf("the other bear is power %d, want 2", p)
		}
	})
	t.Run("creatures you control", func(t *testing.T) {
		g, me, opp := effectstest.SpendTable(t)
		a := effectstest.B12Creature(g, me.ID, "A", kldB260Bear, 2, 2)
		b := effectstest.B12Creature(g, me.ID, "B", kldB260Bear, 3, 3)
		theirs := effectstest.B12Creature(g, opp.ID, "Theirs", kldB260Bear, 2, 2)
		effectstest.CastCatalogSpellWithModesTargets(t, g, kldB260Might, []int{1}, nil)
		effectstest.PassPriorityAroundTable(t, g)
		for id, want := range map[uuid.UUID]int{a: 4, b: 5} {
			if p := effectstest.EffectivePower(t, g, id); p != want {
				t.Errorf("power %d, want %d", p, want)
			}
			if !effectstest.HasEffectiveKeyword(t, g, id, "vigilance") {
				t.Error("no vigilance")
			}
		}
		if p := effectstest.EffectivePower(t, g, theirs); p != 2 {
			t.Errorf("an opponent's creature was pumped to %d", p)
		}
	})
}

// --- Start Your Engines ------------------------------------------------------------------

func TestKldB260StartYourEnginesAnimatesVehiclesAndPumpsEveryCreature(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	bear := effectstest.B12Creature(g, me.ID, "Bear", kldB260Bear, 2, 2)
	vehicle := effectstest.B12Push(g, me.ID, "Test Vehicle", "Artifact — Vehicle", "", 3, 3)
	theirVehicle := effectstest.B12Push(g, opp.ID, "Their Vehicle", "Artifact — Vehicle", "", 3, 3)
	if effectstest.EffectiveCardForTest(t, g, vehicle).IsCreature() {
		t.Fatal("a Vehicle is a creature before it is crewed")
	}
	effectstest.CastCatalogSpell(t, g, "Start Your Engines", "Sorcery", kldB260StartEngines, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.EffectiveCardForTest(t, g, vehicle); !c.IsCreature() || !c.IsArtifact() {
		t.Error("the Vehicle did not become an artifact creature")
	}
	if effectstest.EffectiveCardForTest(t, g, theirVehicle).IsCreature() {
		t.Error("an opponent's Vehicle was animated")
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Errorf("bear power %d, want 4", p)
	}
	if p := effectstest.EffectivePower(t, g, vehicle); p != 5 {
		t.Errorf("the animated Vehicle has power %d, want 5 (3 + 2)", p)
	}
}

// --- Master Trinketeer --------------------------------------------------------------------

func TestKldB260MasterTrinketeerPumpsServosAndThopters(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	master := effectstest.B12Push(g, me.ID, "Master Trinketeer", "Creature — Dwarf Artificer", kldB260Trinketeer, 3, 2)
	servo := effectstest.B12Push(g, me.ID, "Servo", "Token Artifact Creature — Servo", "", 1, 1)
	thopter := effectstest.B12Push(g, me.ID, "Thopter", "Artifact Creature — Thopter", "", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", kldB260Bear, 2, 2)
	theirServo := effectstest.B12Push(g, opp.ID, "Servo", "Token Artifact Creature — Servo", "", 1, 1)
	for id, want := range map[uuid.UUID]int{master: 3, servo: 2, thopter: 2, bear: 2, theirServo: 1} {
		if p := effectstest.EffectivePower(t, g, id); p != want {
			t.Errorf("power %d, want %d", p, want)
		}
	}
}

func TestKldB260MasterTrinketeerMakesAServo(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	master := effectstest.B12Push(g, me.ID, "Master Trinketeer", "Creature — Dwarf Artificer", kldB260Trinketeer, 3, 2)
	if err := g.ActivateCatalogAbility(me.ID, master, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("{3}{W} paid from an empty pool")
	}
	effectstest.ApaMana(me, "W", "W", "W", "W")
	effectstest.B16Activate(t, g, me.ID, master, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Servo", me.ID); n != 1 {
		t.Fatalf("%d Servos, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Servo" {
			if p := effectstest.EffectivePower(t, g, c.InstanceID); p != 2 {
				t.Errorf("the new Servo has power %d, want 2 (1/1 plus the anthem)", p)
			}
		}
	}
}

// --- Toolcraft Exemplar --------------------------------------------------------------------

func TestKldB260ToolcraftExemplarScalesWithArtifacts(t *testing.T) {
	for _, c := range []struct {
		artifacts    int
		power, tough int
		firstStrike  bool
		description  string
	}{
		{0, 1, 1, false, "no artifact: nothing happens"},
		{1, 3, 2, false, "one artifact: +2/+1"},
		{2, 3, 2, false, "two artifacts: +2/+1 only"},
		{3, 3, 2, true, "three artifacts: first strike too"},
	} {
		t.Run(c.description, func(t *testing.T) {
			g, me, _ := effectstest.SpendTable(t)
			ex := effectstest.B12Push(g, me.ID, "Toolcraft Exemplar", "Creature — Dwarf Artificer", kldB260Exemplar, 1, 1)
			for i := 0; i < c.artifacts; i++ {
				effectstest.PushCatalogPermanent(g, me.ID, "Trinket", "Artifact", "", false)
			}
			effectstest.AdvanceTo(t, g, game.StepBeginCombat)
			effectstest.PassPriorityAroundTable(t, g)
			if p, tough := effectstest.EffectivePower(t, g, ex), effectstest.EffectiveToughness(t, g, ex); p != c.power || tough != c.tough {
				t.Errorf("Exemplar is %d/%d, want %d/%d", p, tough, c.power, c.tough)
			}
			if got := effectstest.HasEffectiveKeyword(t, g, ex, "first strike"); got != c.firstStrike {
				t.Errorf("first strike = %v, want %v", got, c.firstStrike)
			}
		})
	}
}

func TestKldB260ToolcraftExemplarIgnoresAnOpponentsBeginningOfCombat(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	ex := effectstest.B12Push(g, me.ID, "Toolcraft Exemplar", "Creature — Dwarf Artificer", kldB260Exemplar, 1, 1)
	effectstest.PushCatalogPermanent(g, me.ID, "Trinket", "Artifact", "", false)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 1)
	if g.Seats[g.Turn.ActiveSeat].ID != opp.ID {
		t.Skip("could not reach an opponent's combat")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, ex); p != 1 {
		t.Errorf("power %d on an opponent's turn, want 1", p)
	}
}
