package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b636_cards_test.go — Planeshift Auras, batch 636 (#5615). Each test
// plays the real catalog entry through the engine and asserts what its text
// says.

const (
	oldB636KeldonMantle     = "170cc0c4-0fba-4751-8613-e934b0438364"
	oldB636Insolence        = "33a6d78a-a3c7-44f1-8ce9-d6632123fec1"
	oldB636SisaysIngenuity  = "3a487f20-6b1d-4be3-b51f-0c8f0a611601"
	oldB636MultanisHarmony  = "81c223a5-5ff2-43e5-bd0b-25c4f72b2150"
	oldB636Hobble           = "c2398875-0912-4210-90bd-b2d568a580cf"
	oldB636SleepingPotion   = "c8fbcef0-2cbf-4a59-8322-374dd9153ebf"
	oldB636SinisterStrength = "deb499bd-71d4-4430-8a58-a31f5ab1b239"
	oldB636HeroicDefiance   = "f6209f18-9f40-44b4-9a8c-3445eda89928"
)

var oldB636Cards = map[string]string{
	oldB636KeldonMantle: "Keldon Mantle", oldB636Insolence: "Insolence", oldB636SisaysIngenuity: "Sisay's Ingenuity",
	oldB636MultanisHarmony: "Multani's Harmony", oldB636Hobble: "Hobble", oldB636SleepingPotion: "Sleeping Potion",
	oldB636SinisterStrength: "Sinister Strength", oldB636HeroicDefiance: "Heroic Defiance",
}

func TestOldB636EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range oldB636Cards {
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

func oldB636Setup(t *testing.T, colours ...string) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.TmpB1Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2, colours...)
	return g, me, opp, bear
}

func TestOldB636KeldonMantleRegeneratesPumpsAndGrantsTrample(t *testing.T) {
	g, me, opp, bear := oldB636Setup(t)
	aura := effectstest.Enchant(t, g, "Keldon Mantle", oldB636KeldonMantle, bear)
	activate := func(idx int, mana string) {
		t.Helper()
		effectstest.FloatMana(t, g, me, mana)
		effectstest.B16Activate(t, g, me.ID, aura, idx, game.ActivateAbilityParams{})
		effectstest.PassPriorityAroundTable(t, g)
	}
	activate(1, "{R}")
	if p, th := effectstest.EffectivePTOf(t, g, bear); p != 3 || th != 2 {
		t.Errorf("after {R} the creature is %d/%d, want 3/2", p, th)
	}
	activate(2, "{G}")
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("{G} did not give the creature trample")
	}
	activate(0, "{B}")
	src := effectstest.B12Creature(g, opp.ID, "Source", "Creature — Giant", 9, 9)
	effectstest.DamageAndSettle(t, g, src, bear, 5)
	if !g.Battlefield.Contains(bear) {
		t.Error("the regeneration shield did not save the creature")
	}
}

func TestOldB636InsolenceDamagesTheCreaturesControllerWhenItBecomesTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.Enchant(t, g, "Insolence", oldB636Insolence, theirs)
	life, mine := opp.Life, me.Life
	effectstest.TapForTest(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("the creature's controller is at %d, want %d", opp.Life, life-2)
	}
	if me.Life != mine {
		t.Errorf("the Aura's controller lost life (%d -> %d)", mine, me.Life)
	}
}

func TestOldB636SisaysIngenuityDrawsAndGrantsAColourChange(t *testing.T) {
	g, me, _, bear := oldB636Setup(t, "G")
	other := effectstest.TmpB1Push(g, g.Seats[1].ID, "Other", "Creature — Bear", "", 2, 2, "G")
	hand := me.Hand.Size()
	effectstest.Enchant(t, g, "Sisay's Ingenuity", oldB636SisaysIngenuity, bear)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want a card drawn as the Aura enters", hand, me.Hand.Size())
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{U}")
	effectstest.TspB1ActivateGranted(t, g, me.ID, bear, effectstest.CardTarget(other))
	if c := effectstest.AnswerColor(t, g, me.ID, "R"); c == nil {
		t.Fatal("no colour prompt")
	}
	effectstest.PassPriorityAroundTable(t, g)
	got := effectstest.LayeredCard(t, g, other).EffectiveColors()
	if len(got) != 1 || got[0] != "R" {
		t.Errorf("the target's colours are %v, want [R]", got)
	}
}

func TestOldB636MultanisHarmonyGrantsAnyColorMana(t *testing.T) {
	g, me, _, bear := oldB636Setup(t)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	effectstest.Enchant(t, g, "Multani's Harmony", oldB636MultanisHarmony, bear)
	if n := effectstest.GrantedManaRowsOf(g, bear); n != 1 {
		t.Errorf("the enchanted creature has %d granted mana rows, want 1", n)
	}
	if n := effectstest.GrantedManaRowsOf(g, other); n != 0 {
		t.Errorf("another creature has %d granted mana rows, want 0", n)
	}
}

func TestOldB636HobbleDrawsStopsAttackingAndStopsBlockingBlackCreatures(t *testing.T) {
	for _, tc := range []struct {
		name    string
		colours []string
		want    game.Restriction
	}{
		{"nonblack", []string{"G"}, game.CantAttack},
		{"black", []string{"B"}, game.CantAttack | game.CantBlock},
		{"black and green", []string{"B", "G"}, game.CantAttack | game.CantBlock},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _, bear := oldB636Setup(t, tc.colours...)
			hand := me.Hand.Size()
			effectstest.Enchant(t, g, "Hobble", oldB636Hobble, bear)
			if me.Hand.Size() != hand+1 {
				t.Errorf("hand %d -> %d, want a card drawn", hand, me.Hand.Size())
			}
			effectstest.AssertRestrictions(t, g, bear, tc.want)
		})
	}
}

func TestOldB636SleepingPotionTapsLocksAndIsSacrificedWhenTheCreatureIsTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	aura := effectstest.Enchant(t, g, "Sleeping Potion", oldB636SleepingPotion, theirs)
	if !effectstest.TappedNow(g, theirs) {
		t.Fatal("the enchanted creature was not tapped as the Aura entered")
	}
	spec, _ := effects.Lookup(oldB636SleepingPotion)
	if len(spec.UntapStepRestrictions) != 1 {
		t.Errorf("Sleeping Potion declares %d untap-step restrictions, want 1", len(spec.UntapStepRestrictions))
	}
	if !g.Battlefield.Contains(aura) {
		t.Fatal("the Aura left the battlefield before the creature was targeted")
	}
	if _, err := b281Cast(t, g, me, "Slaying Fire", "Instant", "{2}{R}", effectstest.SlayingFireOracle, "CCR",
		game.CastSpellParams{Targets: b281Target(theirs)}); err != nil {
		t.Fatalf("targeting the creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("Sleeping Potion was not sacrificed when its creature became a target")
	}
}

func TestOldB636SinisterStrengthPumpsAndMakesTheCreatureBlack(t *testing.T) {
	g, _, _, bear := oldB636Setup(t, "G", "W")
	effectstest.Enchant(t, g, "Sinister Strength", oldB636SinisterStrength, bear)
	if p, th := effectstest.EffectivePTOf(t, g, bear); p != 5 || th != 3 {
		t.Errorf("the creature is %d/%d, want 5/3", p, th)
	}
	if cols := effectstest.LayeredCard(t, g, bear).EffectiveColors(); len(cols) != 1 || cols[0] != "B" {
		t.Errorf("the creature's colours are %v, want exactly [B]", cols)
	}
}

func TestOldB636HeroicDefiancePumpsUnlessItSharesAMostCommonColor(t *testing.T) {
	for _, tc := range []struct {
		name    string
		creat   []string
		others  [][]string // other permanents' colours
		pumped  bool
		comment string
	}{
		{"colourless creature", nil, nil, true, "shares no colour"},
		{"tied with the white Aura", []string{"R"}, nil, false, "red, white tied at one"},
		{"white creature", []string{"W"}, nil, false, "white is the most common"},
		{"outnumbered colour", []string{"G"}, [][]string{{"R"}, {"R"}}, true, "red is the most common, green is not"},
		{"shares the most common", []string{"R"}, [][]string{{"R"}, {"R"}}, false, "red is the most common"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, _, _, bear := oldB636Setup(t, tc.creat...)
			for _, cols := range tc.others {
				effectstest.TmpB1Push(g, g.Seats[1].ID, "Other", "Creature — Bear", "", 1, 1, cols...)
			}
			effectstest.Enchant(t, g, "Heroic Defiance", oldB636HeroicDefiance, bear)
			p, th := effectstest.EffectivePTOf(t, g, bear)
			want := 2
			if tc.pumped {
				want = 5
			}
			if p != want || th != want {
				t.Errorf("%s: the creature is %d/%d, want %d/%d", tc.comment, p, th, want, want)
			}
		})
	}
}
