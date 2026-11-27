package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b754_cards_test.go — Tempest, batch 754 (S82 older sets, #5628):
// activated abilities. Every test plays the real catalog entry through the
// engine.

const (
	oldB754Telethopter = "8202ff77-9afc-4e5b-895b-471bbfa67cce"
	oldB754Scout       = "831e2185-0bb2-45bb-99cd-d2091bad86f0"
	oldB754Ranger      = "8f47a230-7500-4afb-97cd-dbd444c307c8"
	oldB754Orim        = "9d405e5b-0cee-4c8f-b3ef-65765853b8a4"
	oldB754Clergy      = "a5a1b4c9-474a-42b7-b412-c2d43a3353db"
	oldB754Fylamarid   = "c5829bff-9687-42b2-8bba-1f036e592cdd"
	oldB754Riders      = "d04f1f62-5363-4bf2-bef2-fb3387fc5f48"
	oldB754Toy         = "d0e33a3f-659a-4889-98f6-d0f385e259f1"
	oldB754Strings     = "eb810d5c-903e-4ab4-874c-e49a686ca367"
)

func TestOldB754EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		oldB754Telethopter, oldB754Scout, oldB754Ranger, oldB754Orim, oldB754Clergy,
		oldB754Fylamarid, oldB754Riders, oldB754Toy, oldB754Strings,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func oldB754Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	return g, me, opp
}

func oldB754Activate(g *game.Game, me *game.Player, id uuid.UUID, params game.ActivateAbilityParams) error {
	params.Strict = true
	return g.ActivateCatalogAbility(me.ID, id, 0, params)
}

func oldB754Target(ids ...uuid.UUID) game.ActivateAbilityParams {
	return game.ActivateAbilityParams{Targets: effectstest.CardRefs(ids...)}
}

func TestOldB754TelethopterTapsAnotherCreatureForFlying(t *testing.T) {
	g, me, _ := oldB754Table(t)
	copter := effectstest.B12Push(g, me.ID, "Telethopter", "Artifact Creature — Thopter", oldB754Telethopter, 3, 1)
	// Tapping itself pays: the Telethopter is an untapped creature you control.
	if err := oldB754Activate(g, me, copter, game.ActivateAbilityParams{TapIDs: []uuid.UUID{copter}}); err != nil {
		t.Fatalf("activate tapping itself: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, copter), "flying") {
		t.Error("the Telethopter paid with itself and did not gain flying")
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.UntapForTest(g, copter)
	p := game.ActivateAbilityParams{TapIDs: []uuid.UUID{bear}}
	if err := oldB754Activate(g, me, copter, p); err != nil {
		t.Fatalf("activate tapping another creature: %v", err)
	}
	if !effectstest.B13Tapped(t, g, bear) {
		t.Error("the other creature was not tapped to pay")
	}
}

func TestOldB754ScoutGivesAnyCreatureFirstStrike(t *testing.T) {
	g, me, opp := oldB754Table(t)
	scout := effectstest.B12Push(g, me.ID, "Advance Scout", "Creature — Human Soldier Scout", oldB754Scout, 1, 1)
	foe := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, scout), "first strike") {
		t.Error("the Scout lacks first strike")
	}
	effectstest.FloatForTest(g, me, "W")
	if err := oldB754Activate(g, me, scout, oldB754Target(foe)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, foe), "first strike") {
		t.Error("the target did not gain first strike")
	}
}

func TestOldB754RangerPutsALandFromHandAtSorcerySpeedOnly(t *testing.T) {
	g, me, _ := oldB754Table(t)
	ranger := effectstest.B12Push(g, me.ID, "Skyshroud Ranger", "Creature — Elf Ranger", oldB754Ranger, 1, 1)
	land := effectstest.HandCardForTest(me, "Forest", "Basic Land — Forest", "")
	bear := effectstest.HandCardForTest(me, "Bear", "Creature — Bear", "")
	effectstest.AdvanceToStepInTurn(t, g, game.StepBeginCombat)
	if err := oldB754Activate(g, me, ranger, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated outside a main phase")
	}
	effectstest.AdvanceToMain(t, g)
	if err := oldB754Activate(g, me, ranger, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate in the main phase: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, land)
	if !g.Battlefield.Contains(land) {
		t.Error("the land did not arrive")
	}
	if !me.Hand.Contains(bear) {
		t.Error("a nonland card left the hand")
	}
}

func TestOldB754HealersPreventTheirAmountOfDamage(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness, n    int
	}{
		{"Orim, Samite Healer", "Legendary Creature — Human Cleric", oldB754Orim, 1, 3, 3},
		{"Clergy en-Vec", "Creature — Human Cleric", oldB754Clergy, 1, 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := oldB754Table(t)
			healer := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 9)
			if err := oldB754Activate(g, me, healer, oldB754Target(bear)); err != nil {
				t.Fatalf("activate: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, tc.n+2) })
			if got := effectstest.DamageMarkedOn(g, bear); got != 2 {
				t.Errorf("%d damage into a %d-shield: marked %d, want 2", tc.n+2, tc.n, got)
			}
		})
	}
}

func TestOldB754SqueesToyShieldsACreatureOnly(t *testing.T) {
	g, me, opp := oldB754Table(t)
	toy := effectstest.B12Push(g, me.ID, "Squee's Toy", "Artifact", oldB754Toy, 0, 0)
	if err := oldB754Activate(g, me, toy, oldB754Target(opp.ID)); err == nil {
		t.Fatal("the Toy targeted a player")
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 9)
	if err := oldB754Activate(g, me, toy, oldB754Target(bear)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 3) })
	if got := effectstest.DamageMarkedOn(g, bear); got != 2 {
		t.Errorf("3 damage into a 1-shield: marked %d, want 2", got)
	}
}

func TestOldB754MantaRidersGainFlying(t *testing.T) {
	g, me, _ := oldB754Table(t)
	riders := effectstest.B12Push(g, me.ID, "Manta Riders", "Creature — Merfolk", oldB754Riders, 1, 1)
	effectstest.FloatForTest(g, me, "U")
	if err := oldB754Activate(g, me, riders, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, riders), "flying") {
		t.Error("the Riders did not gain flying")
	}
}

func TestOldB754PuppetStringsTapsOrUntaps(t *testing.T) {
	g, me, opp := oldB754Table(t)
	strings := effectstest.B12Push(g, me.ID, "Puppet Strings", "Artifact", oldB754Strings, 0, 0)
	foe := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, "CC")
	if err := oldB754Activate(g, me, strings, oldB754Target(foe)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B13Tapped(t, g, foe) {
		t.Fatal("the creature was not tapped")
	}
	effectstest.UntapForTest(g, strings)
	effectstest.FloatForTest(g, me, "CC")
	if err := oldB754Activate(g, me, strings, oldB754Target(foe)); err != nil {
		t.Fatalf("activate again: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B13Tapped(t, g, foe) {
		t.Error("the creature was not untapped")
	}
}

func TestOldB754FylamaridCantBeBlockedByBlueAndMakesCreaturesBlue(t *testing.T) {
	g, me, opp := oldB754Table(t)
	squid := effectstest.B12Push(g, me.ID, "Fylamarid", "Creature — Squid Beast", oldB754Fylamarid, 1, 3)
	blue := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Blue Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Colors: []string{"U"}, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	green := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Colors: []string{"G"}, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.BrAttack(t, g, squid)
	if err := g.DeclareBlocker(blue, squid); err == nil {
		t.Error("a blue creature blocked the Fylamarid")
	}
	if err := g.DeclareBlocker(green, squid); err != nil {
		t.Errorf("a green creature could not block the Fylamarid: %v", err)
	}
}

func TestOldB754FylamaridTurnsACreatureBlue(t *testing.T) {
	g, me, opp := oldB754Table(t)
	squid := effectstest.B12Push(g, me.ID, "Fylamarid", "Creature — Squid Beast", oldB754Fylamarid, 1, 3)
	green := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Colors: []string{"G"}, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.FloatForTest(g, me, "U")
	if err := oldB754Activate(g, me, squid, oldB754Target(green)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	got := effectstest.B12Card(t, g, green)
	if len(got.EffectiveColors()) != 1 || got.EffectiveColors()[0] != "U" {
		t.Errorf("the target's colors are %v, want only blue", got.EffectiveColors())
	}
}
