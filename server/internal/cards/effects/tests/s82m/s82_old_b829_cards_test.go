package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b829_cards_test.go — Ice Age, batch 829 (S82 OLD, #5636): Auras.
// Every test plays the real catalog entry through the engine.

const (
	oldB829BlueScarab = "47c12d2f-b458-4848-9a01-0fc37ef3d942"
	oldB829Seizures   = "550921f5-ec65-460d-9ada-f2c5a4c4b2c4"
	oldB829Visage     = "6f67058b-ed14-4e3c-9af3-d61570870e36"
	oldB829Minion     = "80459049-bd89-4736-82f8-03c603644033"
	oldB829Stonehands = "8beccaff-d30f-4164-a264-cb7119cee4f1"
	oldB829SoulKiss   = "8ea941aa-ee0c-4830-912a-0c672546aa0c"
	oldB829Flare      = "9bdc79c9-c8b0-4db1-89d9-0ca7920b6576"
	oldB829Prismatic  = "ad5e2cb0-00d7-4dad-b20e-23bd41ada398"
	oldB829Aggression = "b6c38a2b-2fe8-4080-80ce-cfa12aaf57f0"
	oldB829Bear       = "Creature — Bear"
)

func TestOldB829EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB829BlueScarab: "Blue Scarab", oldB829Seizures: "Seizures", oldB829Visage: "Imposing Visage",
		oldB829Minion: "Errant Minion", oldB829Stonehands: "Stonehands", oldB829SoulKiss: "Soul Kiss",
		oldB829Flare: "Essence Flare", oldB829Prismatic: "Prismatic Ward", oldB829Aggression: "Aggression",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// oldB829ToUpkeepOf advances to the upkeep of seat (at least one step on) and
// lets its triggers go on the stack.
func oldB829ToUpkeepOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 300; i++ {
		if g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == seat && i > 0 {
			break
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

func oldB829AuraAbility(t *testing.T, g *game.Game, me *game.Player, aura uuid.UUID, mana string) error {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, mana)
	err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true})
	if err == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	return err
}

func TestOldB829BlueScarabPumpsOnlyWhileAnOpponentControlsABluePermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB829Bear, 2, 2)
	effectstest.Enchant(t, g, "Blue Scarab", oldB829BlueScarab, bear)
	if p, tough := effectstest.PtOf(t, g, bear); p != 2 || tough != 2 {
		t.Fatalf("with no blue permanent the Bear is %d/%d, want 2/2", p, tough)
	}
	mine := effectstest.B12Creature(g, me.ID, "Own Blue", oldB829Bear, 1, 1)
	effectstest.AlliSetColors(g, mine, "U")
	if p, _ := effectstest.PtOf(t, g, bear); p != 2 {
		t.Fatalf("the Aura's controller's own blue permanent pumped the Bear: power %d", p)
	}
	theirs := effectstest.B12Creature(g, opp.ID, "Their Blue", oldB829Bear, 1, 1)
	effectstest.AlliSetColors(g, theirs, "U")
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Fatalf("with an opponent's blue permanent the Bear is %d/%d, want 4/4", p, tough)
	}
}

func TestOldB829BlueScarabCantBeBlockedByBlueCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB829Bear, 2, 2)
	effectstest.Enchant(t, g, "Blue Scarab", oldB829BlueScarab, bear)
	blue := effectstest.B12Creature(g, opp.ID, "Blue Blocker", oldB829Bear, 2, 2)
	effectstest.AlliSetColors(g, blue, "U")
	green := effectstest.B12Creature(g, opp.ID, "Green Blocker", oldB829Bear, 2, 2)
	effectstest.AlliSetColors(g, green, "G")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blue, bear); err == nil {
		t.Error("a blue creature blocked the enchanted creature")
	}
	if err := g.DeclareBlocker(green, bear); err != nil {
		t.Errorf("a green creature could not block it: %v", err)
	}
}

func TestOldB829ImposingVisageGrantsMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB829Bear, 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, bear, "menace") {
		t.Fatal("the Bear has menace before the Aura")
	}
	effectstest.Enchant(t, g, "Imposing Visage", oldB829Visage, bear)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "menace") {
		t.Error("the enchanted creature lacks menace")
	}
}

func TestOldB829StonehandsPumpsToughnessAndFirebreathes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB829Bear, 2, 2)
	aura := effectstest.Enchant(t, g, "Stonehands", oldB829Stonehands, bear)
	if p, tough := effectstest.PtOf(t, g, bear); p != 2 || tough != 4 {
		t.Fatalf("the Bear is %d/%d, want 2/4", p, tough)
	}
	if err := oldB829AuraAbility(t, g, me, aura, "{R}"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 4 {
		t.Errorf("after {R} the Bear is %d/%d, want 3/4", p, tough)
	}
}

func TestOldB829SoulKissIsLimitedToThreeActivationsAndPaysLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB829Bear, 2, 2)
	aura := effectstest.Enchant(t, g, "Soul Kiss", oldB829SoulKiss, bear)
	life := effectstest.LifeOf(g, me.ID)
	for i := 0; i < 3; i++ {
		if err := oldB829AuraAbility(t, g, me, aura, "{B}"); err != nil {
			t.Fatalf("activation %d: %v", i+1, err)
		}
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 8 || tough != 8 {
		t.Errorf("after three activations the Bear is %d/%d, want 8/8", p, tough)
	}
	if got := effectstest.LifeOf(g, me.ID); got != life-3 {
		t.Errorf("life %d, want %d", got, life-3)
	}
	if err := oldB829AuraAbility(t, g, me, aura, "{B}"); err == nil {
		t.Error("a fourth activation in one turn was allowed")
	}
}

func TestOldB829EssenceFlarePumpsPowerAndShrinksEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB829Bear, 2, 2)
	effectstest.Enchant(t, g, "Essence Flare", oldB829Flare, bear)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 2 {
		t.Fatalf("the Bear is %d/%d, want 4/2", p, tough)
	}
	oldB829ToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 1 {
		t.Errorf("after an upkeep the Bear is %d/%d, want 4/1", p, tough)
	}
}

func TestOldB829SeizuresDamagesTheControllerUnlessTheyPay(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[1]
		bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB829Bear, 2, 2)
		effectstest.Enchant(t, g, "Seizures", oldB829Seizures, bear)
		life := effectstest.LifeOf(g, opp.ID)
		effectstest.TapForTest(t, g, bear)
		effectstest.PassPriorityAroundTable(t, g)
		if pay {
			effectstest.FillPool(opp, 3)
		}
		effectstest.AnswerPayUnless(t, g, opp.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		want := life - 3
		if pay {
			want = life
		}
		if got := effectstest.LifeOf(g, opp.ID); got != want {
			t.Errorf("pay=%v: the creature's controller has %d life, want %d", pay, got, want)
		}
	}
}

func TestOldB829PrismaticWardPreventsDamageFromTheChosenColorOnly(t *testing.T) {
	for _, tc := range []struct {
		chosen string
		want   int
	}{{"R", 0}, {"G", 3}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		bear := effectstest.B12Creature(g, me.ID, "Bear", oldB829Bear, 2, 5)
		effectstest.Enchant(t, g, "Prismatic Ward", oldB829Prismatic, bear)
		effectstest.AnswerColor(t, g, me.ID, tc.chosen)
		effectstest.PassPriorityAroundTable(t, g)
		bolt := effectstest.PutInHand(me, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", OracleID: effectstest.LightningBoltOracle, Colors: []string{"R"}, ManaCost: "{R}"})
		effectstest.FloatMana(t, g, me, "{R}")
		if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err != nil {
			t.Fatalf("cast Bolt: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.DamageOn(t, g, bear); got != tc.want {
			t.Errorf("chosen %s: the Bear has %d damage, want %d", tc.chosen, got, tc.want)
		}
	}
}

func TestOldB829ErrantMinionLetsThePlayerPayToPreventItsDamage(t *testing.T) {
	for _, tc := range []struct{ pay, lost int }{{2, 0}, {1, 1}, {0, 2}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		bear := effectstest.B12Creature(g, me.ID, "Bear", oldB829Bear, 2, 2)
		effectstest.Enchant(t, g, "Errant Minion", oldB829Minion, bear)
		life := effectstest.LifeOf(g, me.ID)
		oldB829ToUpkeepOf(t, g, 0)
		effectstest.FillPool(me, 2)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerNumber(t, g, me.ID, tc.pay)
		if tc.pay > 0 {
			effectstest.AnswerPayUnless(t, g, me.ID, true)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.LifeOf(g, me.ID); got != life-tc.lost {
			t.Errorf("paid %d: life %d, want %d", tc.pay, got, life-tc.lost)
		}
	}
}

func TestOldB829AggressionGrantsFirstStrikeAndTrampleAndDestroysAnIdleCreature(t *testing.T) {
	for _, attacks := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bear := effectstest.B12Creature(g, me.ID, "Bear", oldB829Bear, 2, 2)
		effectstest.Enchant(t, g, "Aggression", oldB829Aggression, bear)
		if !effectstest.HasEffectiveKeyword(t, g, bear, "first strike") || !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
			t.Fatal("the enchanted creature lacks first strike or trample")
		}
		if attacks {
			effectstest.AttackWith(t, g, opp.ID, bear)
		}
		effectstest.AdvanceToEndStep(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if alive := effects.OnBattlefield(g, bear); alive != attacks {
			t.Errorf("attacked=%v: the creature is on the battlefield = %v", attacks, alive)
		}
	}
}

func TestOldB829AggressionCantEnchantAWall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wall := effectstest.B12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 4)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Aggression", effectstest.AuraTypeLine, oldB829Aggression, effectstest.CardTargets(wall)); err == nil {
		t.Error("Aggression enchanted a Wall")
	}
}
