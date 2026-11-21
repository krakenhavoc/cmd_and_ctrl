package testsms

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func myriadAttacker(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	effectstest.B21Push(g, me.ID, "Cybermen Squadron", "Artifact Creature — Cyberman", effectstest.S58CybermenSquadronOracle, 5, 5)
	return effectstest.PushTribalCreature(g, me.ID, "Test Bot", "Artifact Creature — Construct", 3, 3)
}

func TestMyriadMayBeDeclinedForEachOpponentSeparately(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bot := myriadAttacker(t, g)
	me := g.Seats[0].ID
	effectstest.DeclareAttack(t, g, g.Seats[1].ID, bot)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me, false)
	if len(effectstest.S58Named(g, me, "Test Bot")) != 1 {
		t.Fatal("no tokens may enter before all optional decisions are made")
	}
	effectstest.AnswerMayChoice(t, g, me, true)
	copies := effectstest.S58Named(g, me, "Test Bot")
	if len(copies) != 2 {
		t.Fatalf("Test Bots = %d, want original and one copy", len(copies))
	}
	for _, c := range copies {
		if c.InstanceID != bot && c.AttackingTarget != g.Seats[3].ID {
			t.Errorf("copy attacks %v, want only the accepted opponent", c.AttackingTarget)
		}
	}
}

func TestMyriadDoubledCopiesChoosePlayerAndPlaneswalkerIndependently(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bot := myriadAttacker(t, g)
	me, other := g.Seats[0].ID, g.Seats[2].ID
	effectstest.PushTokenReplacementCard(g, effectstest.ParallelLivesOracle, "Parallel Lives", "Enchantment", me)
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Walker", TypeLine: "Planeswalker", Owner: other, Controller: other, StartingLoyalty: 6, Counters: map[string]int{game.CounterLoyalty: 6}})
	battle := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Battle", TypeLine: "Battle — Siege", Owner: other, Controller: other, ProtectorPlayerID: other, Counters: map[string]int{game.CounterDefense: 6}})
	effectstest.DeclareAttack(t, g, g.Seats[1].ID, bot)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me, true)
	effectstest.AnswerMayChoice(t, g, me, false)
	pc := effectstest.PayAmountPrompt(g, me)
	if pc == nil || pc.PayAmount.Max != 2 {
		t.Fatalf("want assignment of two doubled copies, got %+v", pc)
	}
	before := g.Clone()
	for run := 0; run < 2; run++ {
		if run > 0 {
			g.WithWriteLock(func() { g.RestoreFrom(before) })
			pc = effectstest.PayAmountPrompt(g, me)
		}
		if err := g.ResolvePayAmount(pc.ID, me, 1); err != nil {
			t.Fatal(err)
		}
		if effectstest.PayAmountPrompt(g, me) != nil {
			t.Fatal("battle must not be an attack destination; only player and walker need allocation")
		}
		attacks := map[uuid.UUID]int{}
		for _, c := range effectstest.S58Named(g, me, "Test Bot") {
			if c.InstanceID != bot {
				attacks[c.AttackingTarget]++
			}
		}
		if attacks[other] != 1 || attacks[walker] != 1 || attacks[battle] != 0 || len(attacks) != 2 {
			t.Fatalf("run %d: copies attack %v, want one at player and one at walker", run, attacks)
		}
		if len(g.DelayedTriggers) != 1 {
			t.Fatal("the whole creation must schedule exactly one delayed exile")
		}
	}
}

func TestMyriadCopiesEnterSimultaneouslyAndSeeEachOther(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B21Push(g, me.ID, "Legion Loyalty", "Enchantment", effectstest.B27LegionLoyaltyOracle, 0, 0)
	warden := effectstest.B21Push(g, me.ID, "Soul Warden", "Creature — Human Cleric", "f3fad295-1af2-4ecc-8546-b121ad6be27b", 1, 1)
	effectstest.DeclareAttack(t, g, g.Seats[1].ID, warden)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	life := me.Life
	effectstest.B21SettleStack(t, g)
	if me.Life != life+4 {
		t.Fatalf("life = %d, want %d: original sees both copies and each copy sees the other", me.Life, life+4)
	}
}

func TestMyriadIsRemovedWithTheRecipientsAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bot := myriadAttacker(t, g)
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(bot), []game.Mod{game.LoseAllAbilitiesMod()}, g.UntilEndOfTurnDuration(), "Remove myriad")
	})
	effectstest.DeclareAttack(t, g, g.Seats[1].ID, bot)
	if effectstest.TriggersOnStackFrom(g, bot) != 0 {
		t.Fatal("creature that lost all abilities still triggered granted myriad")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 || len(effectstest.S58Named(g, g.Seats[0].ID, "Test Bot")) != 1 {
		t.Fatal("removed myriad still asked a question or made a copy")
	}
}

func TestTwoMyriadGrantsTriggerIndependently(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bot := myriadAttacker(t, g)
	me := g.Seats[0].ID
	effectstest.B21Push(g, me, "Cybermen Squadron", "Artifact Creature — Cyberman", effectstest.S58CybermenSquadronOracle, 5, 5)
	effectstest.DeclareAttack(t, g, g.Seats[1].ID, bot)
	if effectstest.TriggerOrderPrompt(g) != nil {
		effectstest.AnswerTriggerOrderInOfferedOrder(t, g)
	}
	if got := effectstest.TriggersOnStackFrom(g, bot); got != 2 {
		t.Fatalf("myriad triggers = %d, want two instances on the attacker", got)
	}
	for i := 0; i < 2; i++ {
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerMayChoice(t, g, me, true)
		effectstest.AnswerMayChoice(t, g, me, true)
	}
	if got := len(effectstest.S58Named(g, me, "Test Bot")); got != 5 || len(g.DelayedTriggers) != 2 {
		t.Fatalf("bots = %d, delayed exiles = %d; want 5 and 2", got, len(g.DelayedTriggers))
	}
}

func TestMyriadCopiesAnAttackerThatLeftBeforeResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bot := myriadAttacker(t, g)
	me := g.Seats[0].ID
	effectstest.DeclareAttack(t, g, g.Seats[1].ID, bot)
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(bot); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me, true)
	effectstest.AnswerMayChoice(t, g, me, true)
	if got := len(effectstest.S58Named(g, me, "Test Bot")); got != 2 {
		t.Fatalf("departed attacker made %d copies, want 2 from its battlefield information", got)
	}
}
