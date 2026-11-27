package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	o596Harbinger = "11c2e14c-cee5-480d-b1be-a6b7e245332a"
	o596Strider   = "8524a6a9-4d4c-4796-aad3-0f08f825a239"
	o596Body      = "90d33e75-17bd-4fc9-82f5-eff254135df5"
	o596Honored   = "be6d78c7-b7fb-46a9-99ea-f10d21c3da58"
	o596Garami    = "c918baa5-0e7c-4df2-8567-64a7fbef4df8"
	o596Filth     = "d76522e0-324f-4ec9-b36e-0fe7448aae69"
)

func TestO596CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		o596Harbinger: "Harbinger of Spring", o596Strider: "Moonlit Strider", o596Body: "Body of Jukai",
		o596Honored: "Kami of the Honored Dead", o596Garami: "Forked-Branch Garami", o596Filth: "Crawling Filth",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func o596Dead(g *game.Game, owner *game.Player, name string, mv string) uuid.UUID {
	id := uuid.New()
	owner.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Spirit",
		ManaCost: mv, Owner: owner.ID, Controller: owner.ID})
	return id
}

func o596Pushed(g *game.Game, owner *game.Player, name, line, oracle, cost string, p, tt int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: line, OracleID: oracle,
		ManaCost: cost, Power: p, Toughness: tt, Owner: owner.ID, Controller: owner.ID,
	})
}

func TestO596SoulshiftersReturnASpiritOfTheirValueOrLess(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, cost string
		limit              int
	}{
		{"Harbinger of Spring", o596Harbinger, "{4}{G}", 4},
		{"Moonlit Strider", o596Strider, "{3}{W}", 3},
		{"Body of Jukai", o596Body, "{7}{G}{G}", 8},
		{"Kami of the Honored Dead", o596Honored, "{5}{W}{W}", 6},
		{"Crawling Filth", o596Filth, "{5}{B}", 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			// A real cost on the creature, so its own death puts nothing legal in the graveyard
			// unless the cost fits.
			soul := o596Pushed(g, me, tc.name, "Creature — Spirit", tc.oracle, "{9}{9}", 2, 2)
			zeros := ""
			for i := 0; i < tc.limit; i++ {
				zeros += "{1}"
			}
			fits := o596Dead(g, me, "Fits", zeros)
			tooBig := o596Dead(g, me, "Too Big", zeros+"{1}")
			effectstest.KillCreature(t, g, me.ID, soul)
			p := effectstest.PassUntilPickTarget(t, g, me.ID)
			if p == nil {
				t.Fatal("no soulshift target prompt")
			}
			if !effectstest.HasID(p.PickTargetCards, fits) || effectstest.HasID(p.PickTargetCards, tooBig) {
				t.Errorf("soulshift %d offers %v, want only %v", tc.limit, p.PickTargetCards, fits)
			}
			effectstest.PickCard(t, g, me.ID, fits)
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
			effectstest.PassPriorityAroundTable(t, g)
			if !me.Hand.Contains(fits) {
				t.Error("soulshift did not return the Spirit card")
			}
		})
	}
}

func TestO596ForkedBranchGaramReturnsTwoSpirits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	garami := o596Pushed(g, me, "Forked-Branch Garami", "Creature — Spirit", o596Garami, "{9}{9}", 4, 4)
	a := o596Dead(g, me, "First", "{1}{1}")
	b := o596Dead(g, me, "Second", "{1}{1}{1}")
	effectstest.KillCreature(t, g, me.ID, garami)
	for _, pick := range []uuid.UUID{a, b} {
		p := effectstest.PassUntilPickTarget(t, g, me.ID)
		if p == nil {
			t.Fatal("a soulshift target prompt is missing")
		}
		effectstest.PickCard(t, g, me.ID, pick)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Errorf("hand has first=%v second=%v, want both returned", me.Hand.Contains(a), me.Hand.Contains(b))
	}
}

func TestO596HarbingerHasProtectionFromNonSpiritCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	h := o596Pushed(g, me, "Harbinger of Spring", "Creature — Spirit", o596Harbinger, "{4}{G}", 2, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	spirit := effectstest.B12Creature(g, opp.ID, "Ghost", "Creature — Spirit", 2, 2)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 1)
	if err := g.DeclareAttacker(bear, me.ID); err != nil {
		t.Fatalf("the Bear could not attack: %v", err)
	}
	if err := g.DeclareAttacker(spirit, me.ID); err != nil {
		t.Fatalf("the Spirit could not attack: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(h, bear); err != nil {
		t.Errorf("the Harbinger could not block a non-Spirit creature: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Pr6Marked(g, h); got != 0 {
		t.Errorf("the Harbinger was dealt %d damage by a non-Spirit creature", got)
	}
	if !g.Battlefield.Contains(h) {
		t.Error("the Harbinger (toughness 1) died to a non-Spirit creature's combat damage")
	}
}

func TestO596KamiOfTheHonoredDeadGainsLifeWhenDealtDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kami := o596Pushed(g, me, "Kami of the Honored Dead", "Creature — Spirit", o596Honored, "{5}{W}{W}", 3, 5)
	life := me.Life
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(kami, kami, 3); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life-life != 3 {
		t.Errorf("gained %d life, want 3", me.Life-life)
	}
}

func TestO596MoonlitStriderSacrificesForProtection(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	strider := o596Pushed(g, me, "Moonlit Strider", "Creature — Spirit", o596Strider, "{3}{W}", 1, 4)
	friend := effectstest.B12Creature(g, me.ID, "Friend", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, strider, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(friend)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "R")
	if g.Battlefield.Contains(strider) {
		t.Error("the Strider was not sacrificed")
	}
	c := effectstest.GaLayered(t, g, friend)
	if !game.HasKeyword(&c, game.ProtectionFromColor("R")) {
		t.Error("the target did not gain protection from red")
	}
}
