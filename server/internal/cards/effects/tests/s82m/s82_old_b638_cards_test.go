package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b638_cards_test.go — Planeshift, batch 638 (#5615). Each test plays
// the real catalog entry and asserts what its text says.

const (
	b638StoneKavuOracle      = "01565f61-a3c8-42ca-99f3-07cfc64034a2"
	b638MireKavuOracle       = "0bc58096-7925-4870-bb95-58cce51ea8e7"
	b638DaringLeapOracle     = "10bc00f7-d501-46e3-9f23-af1ddc3f6c4c"
	b638CalderaKavuOracle    = "7cdcfbdf-5ec7-451b-8c96-f19d75a4d76c"
	b638AmphibiousKavuOracle = "9c9b1871-e776-4373-ba89-e009131cb6f7"
	b638GerrardsOracle       = "de8ebe4b-d86d-478b-946e-f83e1409b63f"
	b638AlphaKavuOracle      = "f1a76cca-829a-4576-a560-f1af34b80b87"
	b638KavuType             = "Creature — Kavu"
	b638BearType             = "Creature — Bear"
)

func TestOldB638EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b638StoneKavuOracle:      "Stone Kavu",
		b638MireKavuOracle:       "Mire Kavu",
		b638DaringLeapOracle:     "Daring Leap",
		b638CalderaKavuOracle:    "Caldera Kavu",
		b638AmphibiousKavuOracle: "Amphibious Kavu",
		b638GerrardsOracle:       "Gerrard's Command",
		b638AlphaKavuOracle:      "Alpha Kavu",
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

func b638Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	return g, g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func b638Activate(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, index int, mana string, targets ...game.TargetRef) {
	t.Helper()
	effectstest.DcMana(t, g, me.ID, mana)
	if err := effectstest.InvB1Activate(g, me.ID, id, index, targets...); err != nil {
		t.Fatalf("activate %d: %v", index, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func b638WantPT(t *testing.T, g *game.Game, id uuid.UUID, p, tough int) {
	t.Helper()
	if gp, gt := effectstest.PtOf(t, g, id); gp != p || gt != tough {
		t.Errorf("P/T is %d/%d, want %d/%d", gp, gt, p, tough)
	}
}

func TestOldB638StoneKavuPumpsWithRedAndWhiteAsOftenAsYouPay(t *testing.T) {
	g, me, _ := b638Table(t)
	kavu := effectstest.B12Push(g, me.ID, "Stone Kavu", b638KavuType, b638StoneKavuOracle, 3, 3)
	b638Activate(t, g, me, kavu, 0, "{R}")
	b638Activate(t, g, me, kavu, 0, "{R}")
	b638WantPT(t, g, kavu, 5, 3)
	b638Activate(t, g, me, kavu, 1, "{W}")
	b638WantPT(t, g, kavu, 5, 4)
}

func TestOldB638MireKavuGrowsOnlyWhileYouControlASwamp(t *testing.T) {
	g, me, opp := b638Table(t)
	kavu := effectstest.B12Push(g, me.ID, "Mire Kavu", b638KavuType, b638MireKavuOracle, 3, 2)
	effectstest.B12Permanent(g, opp.ID, "Their Swamp", "Basic Land — Swamp")
	b638WantPT(t, g, kavu, 3, 2)
	swamp := effectstest.B12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	b638WantPT(t, g, kavu, 4, 3)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(swamp, game.DestroyOptions{}) })
	b638WantPT(t, g, kavu, 3, 2)
}

func TestOldB638DaringLeapGivesPlusOneFlyingAndFirstStrike(t *testing.T) {
	g, _, opp := b638Table(t)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", b638BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Daring Leap", "Instant", b638DaringLeapOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	b638WantPT(t, g, bear, 3, 3)
	for _, kw := range []string{"flying", "first strike"} {
		if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the target did not gain %s", kw)
		}
	}
}

func TestOldB638GerrardsCommandUntapsAndGivesPlusThree(t *testing.T) {
	g, me, _ := b638Table(t)
	bear := effectstest.B12Creature(g, me.ID, "Bear", b638BearType, 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.CastCatalogSpell(t, g, "Gerrard's Command", "Instant", b638GerrardsOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	b638WantPT(t, g, bear, 5, 5)
	if effectstest.FindCardAnywhere(t, g, bear).Tapped {
		t.Error("the creature is still tapped")
	}
}

func TestOldB638CalderaKavuPumpsAndBecomesTheChosenColor(t *testing.T) {
	g, me, _ := b638Table(t)
	kavu := effectstest.B27Push(g, me.ID, "Caldera Kavu", b638KavuType, b638CalderaKavuOracle, "{2}{R}", 2, 2, "R")
	b638Activate(t, g, me, kavu, 0, "{B}{B}")
	b638WantPT(t, g, kavu, 3, 3)
	effectstest.DcMana(t, g, me.ID, "{G}")
	if err := effectstest.InvB1Activate(g, me.ID, kavu, 1); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "U")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB9Colors(t, g, kavu); got != "U" {
		t.Errorf("colors = %q, want only blue", got)
	}
}

func TestOldB638AlphaKavuShrinksPowerAndGrowsToughnessOfAKavuOnly(t *testing.T) {
	g, me, opp := b638Table(t)
	alpha := effectstest.B12Push(g, me.ID, "Alpha Kavu", b638KavuType, b638AlphaKavuOracle, 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Other Kavu", b638KavuType, 3, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", b638BearType, 2, 2)
	effectstest.DcMana(t, g, me.ID, "{G}{G}")
	if err := effectstest.InvB1Activate(g, me.ID, alpha, 0, effectstest.CardTarget(bear)...); err == nil {
		t.Fatal("a non-Kavu creature was accepted as the target")
	}
	if err := effectstest.InvB1Activate(g, me.ID, alpha, 0, effectstest.CardTarget(other)...); err != nil {
		t.Fatalf("activate on another Kavu: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	b638WantPT(t, g, other, 2, 4)
	effectstest.DcMana(t, g, me.ID, "{G}{G}")
	if err := effectstest.InvB1Activate(g, me.ID, alpha, 0, effectstest.CardTarget(alpha)...); err != nil {
		t.Fatalf("activate on itself: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	b638WantPT(t, g, alpha, 1, 3)
}

// b638BlockFight has me attack opp with attacker and opp block it with each of
// blockers, then lets the block triggers settle.
func b638Block(t *testing.T, g *game.Game, me, opp *game.Player, attacker uuid.UUID, blockers ...uuid.UUID) {
	t.Helper()
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	for _, b := range blockers {
		if err := g.DeclareBlocker(b, attacker); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
	}
	effectstest.FinishBlockDeclarations(t, g)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB638AmphibiousKavuGetsPlusThreeOnceWhenBlockedByBlueOrBlackCreatures(t *testing.T) {
	g, me, opp := b638Table(t)
	kavu := effectstest.B27Push(g, me.ID, "Amphibious Kavu", b638KavuType, b638AmphibiousKavuOracle, "{2}{G}", 2, 2, "G")
	blue := effectstest.B27Push(g, opp.ID, "Blue Wall", "Creature — Wall", "", "{1}{U}", 0, 9, "U")
	black := effectstest.B27Push(g, opp.ID, "Black Wall", "Creature — Wall", "", "{1}{B}", 0, 9, "B")
	b638Block(t, g, me, opp, kavu, blue, black)
	b638WantPT(t, g, kavu, 5, 5) // one trigger for "one or more", not two
}

func TestOldB638AmphibiousKavuIgnoresOtherColors(t *testing.T) {
	g, me, opp := b638Table(t)
	kavu := effectstest.B27Push(g, me.ID, "Amphibious Kavu", b638KavuType, b638AmphibiousKavuOracle, "{2}{G}", 2, 2, "G")
	red := effectstest.B27Push(g, opp.ID, "Red Wall", "Creature — Wall", "", "{1}{R}", 0, 9, "R")
	b638Block(t, g, me, opp, kavu, red)
	b638WantPT(t, g, kavu, 2, 2)
}

func TestOldB638AmphibiousKavuGetsPlusThreeWhenItBlocksABlackCreature(t *testing.T) {
	g, me, opp := b638Table(t)
	attacker := effectstest.B27Push(g, me.ID, "Black Bear", b638BearType, "", "{1}{B}", 2, 2, "B")
	kavu := effectstest.B27Push(g, opp.ID, "Amphibious Kavu", b638KavuType, b638AmphibiousKavuOracle, "{2}{G}", 2, 2, "G")
	b638Block(t, g, me, opp, attacker, kavu)
	b638WantPT(t, g, kavu, 5, 5)
}

func TestOldB638AmphibiousKavuBlockingAGreenCreatureStaysSmall(t *testing.T) {
	g, me, opp := b638Table(t)
	attacker := effectstest.B27Push(g, me.ID, "Green Bear", b638BearType, "", "{1}{G}", 2, 2, "G")
	kavu := effectstest.B27Push(g, opp.ID, "Amphibious Kavu", b638KavuType, b638AmphibiousKavuOracle, "{2}{G}", 2, 2, "G")
	b638Block(t, g, me, opp, attacker, kavu)
	b638WantPT(t, g, kavu, 2, 2)
}
