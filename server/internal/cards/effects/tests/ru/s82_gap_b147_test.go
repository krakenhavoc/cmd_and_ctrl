package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b147_test.go — GAP batch 147 (#5948): Jumpstart 2022 creatures
// Blight Keeper, Conductor of Cacophony, Cruel Sadist, Demon of
// Catastrophes, Disciple of Perdition and Dread Presence.

const (
	gap147Keeper    = "1dc73412-c91c-48c9-bde6-b169add79b37"
	gap147Conductor = "94daa0c2-4790-4193-aaa8-ba492e8c8fe7"
	gap147Sadist    = "9fceca3c-fc10-4e46-a727-122aafcef349"
	gap147Demon     = "2e0b838d-c858-4aa5-999e-c4ef9d29571b"
	gap147Disciple  = "9ac62ff0-6d13-4091-ba84-fbfbe742a4f4"
	gap147Presence  = "1690a782-9e54-4231-9a72-abae5ed7050c"
)

func TestGap147EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gap147Keeper: "Blight Keeper", gap147Conductor: "Conductor of Cacophony", gap147Sadist: "Cruel Sadist",
		gap147Demon: "Demon of Catastrophes", gap147Disciple: "Disciple of Perdition", gap147Presence: "Dread Presence",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func gap147Counters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return effectstest.CounterCount(g, id, game.CounterPlusOne)
}

func TestGap147BlightKeeperDrainsFourForEightManaAndASacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	keeper := effectstest.PushCatalogPermanent(g, me.ID, "Blight Keeper", "Creature — Bat Imp", gap147Keeper, false)
	target := game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}
	effectstest.FloatForTest(g, me, "BCCCCCC")
	if err := g.ActivateCatalogAbility(me.ID, keeper, 0, target); err == nil {
		t.Fatal("activated with seven mana, the cost is eight")
	}
	effectstest.FloatForTest(g, me, "BCCCCCCC")
	meLife, oppLife := me.Life, opp.Life
	effectstest.B16Activate(t, g, me.ID, keeper, 0, target)
	if g.Battlefield.Contains(keeper) {
		t.Error("the sacrifice is a cost: the Keeper must be gone as the ability goes on the stack")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-4 || me.Life != meLife+4 {
		t.Errorf("life me %d opp %d, want %d and %d", me.Life, opp.Life, meLife+4, oppLife-4)
	}
}

func TestGap147DemonOfCatastrophesWantsASacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "BBCC")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Demon of Catastrophes", TypeLine: "Creature — Demon", ManaCost: "{2}{B}{B}",
			OracleID: gap147Demon, Owner: me.ID, Controller: me.ID, Power: 6, Toughness: 6})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast with no creature to sacrifice")
	}
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.FloatForTest(g, me, "BBCC")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatal("the Demon did not enter")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "flying") || !effectstest.HasKeywordOnBattlefield(t, g, id, "trample") || c.Power != 6 {
		t.Errorf("want a 6/6 flying trampler, got %d/%d", c.Power, c.Toughness)
	}
}

func TestGap147ConductorEntersWithTwoCountersAndPingsEveryoneElse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "BCCC")
	cond := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: cond, Name: "Conductor of Cacophony", TypeLine: "Creature — Demon", ManaCost: "{3}{B}",
			OracleID: gap147Conductor, Owner: me.ID, Controller: me.ID, Power: 2, Toughness: 1})
	})
	if err := g.CastSpell(me.ID, cond, effectstest.Strict); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := gap147Counters(t, g, cond); got != 2 {
		t.Fatalf("+1/+1 counters on entry = %d, want 2", got)
	}
	ally := effectstest.B12Creature(g, me.ID, "Ally Bear", "Creature — Bear", 2, 2)
	runt := effectstest.B12Creature(g, opp.ID, "Runt", "Creature — Rat", 1, 1)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)

	effectstest.FloatForTest(g, me, "BB")
	meLife, oppLife := me.Life, opp.Life
	effectstest.B16Activate(t, g, me.ID, cond, 0, game.ActivateAbilityParams{Strict: true})
	if got := gap147Counters(t, g, cond); got != 1 {
		t.Errorf("the counter is a cost: %d left, want 1", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(runt) {
		t.Error("the 1/1 survived 1 damage")
	}
	for _, id := range []uuid.UUID{ally, ogre, cond} {
		c, ok := effectstest.BattlefieldCard(g, id)
		if !ok {
			t.Fatalf("%s died", id)
		}
		want := 1
		if id == cond {
			want = 0
		}
		if c.DamageMarked != want {
			t.Errorf("%s has %d damage, want %d", c.Name, c.DamageMarked, want)
		}
	}
	if me.Life != meLife-1 || opp.Life != oppLife-1 {
		t.Errorf("life me %d opp %d, want %d and %d: each player, the controller too", me.Life, opp.Life, meLife-1, oppLife-1)
	}

	// The last counter can be spent: the Conductor is then a 2/1.
	effectstest.FloatForTest(g, me, "B")
	effectstest.B16Activate(t, g, me.ID, cond, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if got := gap147Counters(t, g, cond); got != 0 {
		t.Errorf("counters left %d, want 0", got)
	}
	effectstest.FloatForTest(g, me, "B")
	if err := g.ActivateCatalogAbility(me.ID, cond, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with no +1/+1 counter to remove")
	}
}

func gap147SadistBoard(t *testing.T, counters int) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	sadist := effectstest.PushCatalogPermanent(g, me.ID, "Cruel Sadist", "Creature — Human Assassin", gap147Sadist, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(sadist, game.CounterPlusOne, counters) })
	return g, me, opp, sadist
}

func TestGap147CruelSadistGrowsForALifeAndRefusesAtZeroLife(t *testing.T) {
	g, me, _, sadist := gap147SadistBoard(t, 0)
	effectstest.FloatForTest(g, me, "B")
	life := me.Life
	effectstest.B16Activate(t, g, me.ID, sadist, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if got := gap147Counters(t, g, sadist); got != 1 {
		t.Errorf("counters %d, want 1", got)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d: the life is paid", me.Life, life-1)
	}
	if !effectstest.B16Tapped(t, g, sadist) {
		t.Error("the Sadist did not tap")
	}

	g, me, _, sadist = gap147SadistBoard(t, 0)
	effectstest.FloatForTest(g, me, "B")
	me.Life = 0
	if err := g.ActivateCatalogAbility(me.ID, sadist, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("paid 1 life at 0 life")
	}
}

func TestGap147CruelSadistDealsXDamageForXCountersRemoved(t *testing.T) {
	g, me, opp, sadist := gap147SadistBoard(t, 3)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.FloatForTest(g, me, "BCC")
	effectstest.B16Activate(t, g, me.ID, sadist, 1, game.ActivateAbilityParams{
		Strict:           true,
		Targets:          []game.TargetRef{{Kind: game.TargetCard, ID: ogre}},
		CounterSourceIDs: []uuid.UUID{sadist},
		CounterCounts:    []int{2},
	})
	if got := gap147Counters(t, g, sadist); got != 1 {
		t.Errorf("counters %d, want 1 after removing 2 as a cost", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, ogre)
	if !ok || c.DamageMarked != 2 {
		t.Fatalf("the Ogre should have 2 damage and live: ok=%v %+v", ok, c.DamageMarked)
	}

	g, me, opp, sadist = gap147SadistBoard(t, 3)
	ogre = effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.FloatForTest(g, me, "BCC")
	effectstest.B16Activate(t, g, me.ID, sadist, 1, game.ActivateAbilityParams{
		Strict:           true,
		Targets:          []game.TargetRef{{Kind: game.TargetCard, ID: ogre}},
		CounterSourceIDs: []uuid.UUID{sadist},
		CounterCounts:    []int{3},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("3 damage did not kill a 3/3")
	}
}

func TestGap147CruelSadistRefusesAPlayerTarget(t *testing.T) {
	g, me, opp, sadist := gap147SadistBoard(t, 1)
	effectstest.FloatForTest(g, me, "BCC")
	err := g.ActivateCatalogAbility(me.ID, sadist, 1, game.ActivateAbilityParams{
		Strict:           true,
		Targets:          []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		CounterSourceIDs: []uuid.UUID{sadist},
		CounterCounts:    []int{1},
	})
	if err == nil {
		t.Error("a player is not a creature")
	}
}

func gap147Library(me *game.Player) {
	effectstest.DftMPLibraryTop(me, effectstest.C2b15Bear(), effectstest.C2b15Bear(), effectstest.C2b15Bear())
}

func TestGap147DiscipleOfPerditionModes(t *testing.T) {
	setup := func(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
		t.Helper()
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		gap147Library(me)
		d := effectstest.PushCatalogPermanent(g, me.ID, "Disciple of Perdition", "Creature — Human Warlock", gap147Disciple, false)
		return g, me, opp, d
	}
	t.Run("draw and lose life", func(t *testing.T) {
		g, me, _, d := setup(t)
		hand, life := me.Hand.Size(), me.Life
		effectstest.KillCreature(t, g, me.ID, d)
		c := effectstest.ModePickChoiceFor(g, me.ID)
		if c == nil {
			t.Fatal("no mode prompt")
		}
		if err := g.ResolveModePick(c.ID, me.ID, []int{0, 1}); err == nil {
			t.Fatal("both modes at a life total other than 13")
		}
		effectstest.AnswerMode(t, g, c, me.ID, 0)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Hand.Size() != hand+1 || me.Life != life-1 {
			t.Errorf("hand %d life %d, want %d and %d", me.Hand.Size(), me.Life, hand+1, life-1)
		}
	})
	t.Run("exile an opponent's graveyard", func(t *testing.T) {
		g, me, opp, d := setup(t)
		bear := effectstest.C2b15Bear()
		bear.InstanceID, bear.Owner, bear.Controller = uuid.New(), opp.ID, opp.ID
		g.WithWriteLock(func() { opp.Graveyard.PushTop(bear) })
		life, oppLife := me.Life, opp.Life
		effectstest.KillCreature(t, g, me.ID, d)
		effectstest.AnswerMode(t, g, effectstest.ModePickChoiceFor(g, me.ID), me.ID, 1)
		effectstest.B04WaitForPick(t, g, me.ID)
		if err := g.ResolvePickTarget(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err == nil {
			t.Fatal("the controller is not an opponent")
		}
		if err := g.ResolvePickTarget(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if opp.Graveyard.Contains(bear.InstanceID) {
			t.Error("the graveyard card was not exiled")
		}
		if opp.Life != oppLife-1 || me.Life != life {
			t.Errorf("life me %d opp %d, want %d and %d", me.Life, opp.Life, life, oppLife-1)
		}
	})
	t.Run("exactly 13 life may choose both", func(t *testing.T) {
		g, me, opp, d := setup(t)
		me.Life = 13
		hand, oppLife := me.Hand.Size(), opp.Life
		effectstest.KillCreature(t, g, me.ID, d)
		c := effectstest.ModePickChoiceFor(g, me.ID)
		if err := g.ResolveModePick(c.ID, me.ID, []int{0, 1}); err != nil {
			t.Fatalf("both modes at exactly 13 life: %v", err)
		}
		effectstest.B04WaitForPick(t, g, me.ID)
		effectstest.PickPlayer(t, g, me.ID, opp.ID)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Hand.Size() != hand+1 || me.Life != 12 || opp.Life != oppLife-1 {
			t.Errorf("hand %d life %d opp %d, want %d, 12 and %d", me.Hand.Size(), me.Life, opp.Life, hand+1, oppLife-1)
		}
	})
	t.Run("one mode is still enough at 13", func(t *testing.T) {
		g, me, _, d := setup(t)
		me.Life = 13
		effectstest.KillCreature(t, g, me.ID, d)
		c := effectstest.ModePickChoiceFor(g, me.ID)
		if err := g.ResolveModePick(c.ID, me.ID, []int{0}); err != nil {
			t.Fatalf("one mode at 13 life: %v", err)
		}
	})
}

func TestGap147DreadPresenceModes(t *testing.T) {
	setup := func(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
		t.Helper()
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		gap147Library(me)
		p := effectstest.PushCatalogPermanent(g, me.ID, "Dread Presence", "Creature — Nightmare", gap147Presence, false)
		return g, me, opp, p
	}
	t.Run("a Forest does not trigger it", func(t *testing.T) {
		g, me, _, _ := setup(t)
		effectstest.B12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
		if effectstest.ModePickChoiceFor(g, me.ID) != nil || g.Stack.Size() != 0 {
			t.Error("a non-Swamp land triggered Dread Presence")
		}
	})
	t.Run("draw and lose life", func(t *testing.T) {
		g, me, _, _ := setup(t)
		hand, life := me.Hand.Size(), me.Life
		effectstest.B12PlayFromHand(t, g, "Swamp", "Basic Land — Swamp", "", game.CastSpellParams{})
		c := effectstest.ModePickChoiceFor(g, me.ID)
		if c == nil {
			t.Fatal("no mode prompt for a Swamp entering")
		}
		effectstest.AnswerMode(t, g, c, me.ID, 0)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Hand.Size() != hand+1 || me.Life != life-1 {
			t.Errorf("hand %d life %d, want %d and %d", me.Hand.Size(), me.Life, hand+1, life-1)
		}
	})
	t.Run("damage and life gain", func(t *testing.T) {
		g, me, opp, _ := setup(t)
		ogre := effectstest.B12Creature(g, opp.ID, "Runt", "Creature — Rat", 2, 2)
		life := me.Life
		effectstest.B12PlayFromHand(t, g, "Swamp", "Basic Land — Swamp", "", game.CastSpellParams{})
		effectstest.AnswerMode(t, g, effectstest.ModePickChoiceFor(g, me.ID), me.ID, 1)
		effectstest.B04WaitForPick(t, g, me.ID)
		effectstest.PickCard(t, g, me.ID, ogre)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(ogre) {
			t.Error("2 damage did not kill the 2/2")
		}
		if me.Life != life+2 {
			t.Errorf("life %d, want %d", me.Life, life+2)
		}
	})
	t.Run("damage a player", func(t *testing.T) {
		g, me, opp, _ := setup(t)
		oppLife := opp.Life
		effectstest.B12PlayFromHand(t, g, "Swamp", "Basic Land — Swamp", "", game.CastSpellParams{})
		effectstest.AnswerMode(t, g, effectstest.ModePickChoiceFor(g, me.ID), me.ID, 1)
		effectstest.B04WaitForPick(t, g, me.ID)
		effectstest.PickPlayer(t, g, me.ID, opp.ID)
		effectstest.PassPriorityAroundTable(t, g)
		if opp.Life != oppLife-2 {
			t.Errorf("opponent life %d, want %d", opp.Life, oppLife-2)
		}
	})
}
