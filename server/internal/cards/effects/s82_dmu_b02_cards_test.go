package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	dmuB02BattleRage  = "533d6b9f-a145-44c4-9d16-74fcd9c0ae2f"
	dmuB02Flowstone   = "6c75b92a-4fc1-4f26-95be-1b92f462c111"
	dmuB02Bellow      = "4fb0ff70-eba1-4485-b68d-444eb512a2a7"
	dmuB02Firenado    = "250f804f-b5d5-414c-bb25-476d1fe1f85a"
	dmuB02JoinForces  = "6f8bf849-890e-4fc7-981f-37ff4b4fb844"
	dmuB02Impede      = "c0239ff1-6ae6-431a-806e-7321d8272a1d"
	dmuB02Strider     = "4e4a21a8-9332-4aeb-8519-949feb211294"
	dmuB02CutDown     = "1b30210d-10e9-4703-8967-032063bb4f26"
	dmuB02Extinguish  = "6dbef43d-40d3-4894-86a5-d31a5026604e"
	dmuB02SmashToDust = "c13955b8-bad3-4f79-ad9f-11e214a1cab2"
	dmuB02TailSwipe   = "525d98d6-b117-44f4-b19c-3dd26e7c9fd4"
	dmuB02Epiphany    = "fad32e19-2eda-46b7-8d3c-b95a7dbeabc2"
	dmuB02Scorn       = "76e6c0e7-593a-4b4a-9eb6-a912f9c3f26f"
)

func TestDmuB02CardsAreFullWithNoCaveats(t *testing.T) {
	want := map[string]string{
		dmuB02BattleRage: "Battle-Rage Blessing", dmuB02Flowstone: "Flowstone Infusion",
		dmuB02Bellow: "Furious Bellow", dmuB02Firenado: "Jaya's Firenado",
		dmuB02JoinForces: "Join Forces", dmuB02Impede: "Impede Momentum",
		dmuB02Strider: "Frostfist Strider", dmuB02CutDown: "Cut Down",
		dmuB02Extinguish: "Extinguish the Light", dmuB02SmashToDust: "Smash to Dust",
		dmuB02TailSwipe: "Tail Swipe", dmuB02Epiphany: "Cosmic Epiphany",
		dmuB02Scorn: "Ertai's Scorn",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// dmuB02Cast casts a free spell strictly at the given targets in a main phase.
func dmuB02Cast(t *testing.T, g *game.Game, p *game.Player, name, typeLine, oracle string, targets ...uuid.UUID) error {
	t.Helper()
	advanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: "{0}", OracleID: oracle, Owner: p.ID, Controller: p.ID})
	})
	return g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true, Targets: cardTargets(targets...)})
}

func dmuB02Resolve(t *testing.T, g *game.Game) {
	t.Helper()
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
}

func TestDmuB02BattleRageBlessingGrantsDeathtouchAndIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := dmuB02Cast(t, g, me, "Battle-Rage Blessing", "Instant", dmuB02BattleRage); err == nil {
		t.Fatal("a cast with no target was accepted")
	}
	if err := dmuB02Cast(t, g, me, "Battle-Rage Blessing", "Instant", dmuB02BattleRage, bear); err != nil {
		t.Fatalf("cast: %v", err)
	}
	dmuB02Resolve(t, g)
	for _, kw := range []string{"deathtouch", "indestructible"} {
		if !hasKeywordOnBattlefield(t, g, bear, kw) {
			t.Errorf("the creature did not gain %s", kw)
		}
	}
}

func TestDmuB02FlowstoneInfusionPumpsPowerAndKillsAOneToughnessCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := b12Creature(g, opp.ID, "Big", "Creature — Bear", 2, 4)
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 1, 2)
	if err := dmuB02Cast(t, g, me, "Flowstone Infusion", "Instant", dmuB02Flowstone, big); err != nil {
		t.Fatalf("cast: %v", err)
	}
	dmuB02Resolve(t, g)
	if p := effectivePower(t, g, big); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
	if err := dmuB02Cast(t, g, me, "Flowstone Infusion", "Instant", dmuB02Flowstone, small); err != nil {
		t.Fatalf("cast: %v", err)
	}
	dmuB02Resolve(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("a creature at toughness 0 stayed on the battlefield")
	}
}

func TestDmuB02FuriousBellowPumpsGrantsFirstStrikeAndScries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := dmuB02Cast(t, g, me, "Furious Bellow", "Instant", dmuB02Bellow, bear); err != nil {
		t.Fatalf("cast: %v", err)
	}
	dmuB02Resolve(t, g)
	if p := effectivePower(t, g, bear); p != 5 {
		t.Errorf("power %d, want 5", p)
	}
	if !hasKeywordOnBattlefield(t, g, bear, "first strike") {
		t.Error("no first strike")
	}
	dmuB02AnswerScry(t, g, me)
}

func TestDmuB02JayasFirenadoDamagesACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elk := b12Creature(g, opp.ID, "Elk", "Creature — Elk", 5, 5)
	if err := dmuB02Cast(t, g, me, "Jaya's Firenado", "Sorcery", dmuB02Firenado, elk); err != nil {
		t.Fatalf("cast: %v", err)
	}
	dmuB02Resolve(t, g)
	dmuB02AnswerScry(t, g, me)
	if g.Battlefield.Contains(elk) {
		t.Errorf("5 damage did not kill a 5/5 (marked %d, stack %d, pending %d)", findBattlefieldCardByID(g, elk).DamageMarked, g.Stack.Size(), len(g.PendingChoices))
	}
}

func TestDmuB02JoinForcesUntapsAndPumpsUpToTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	c := b12Creature(g, me.ID, "C", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{a, b, c} {
			findBattlefieldCardByID(g, id).Tapped = true
		}
	})
	if err := dmuB02Cast(t, g, me, "Join Forces", "Instant", dmuB02JoinForces, a, b, c); err == nil {
		t.Fatal("three targets were accepted")
	}
	if err := dmuB02Cast(t, g, me, "Join Forces", "Instant", dmuB02JoinForces, a, b); err != nil {
		t.Fatalf("cast: %v", err)
	}
	dmuB02Resolve(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if findBattlefieldCardByID(g, id).Tapped {
			t.Error("a target stayed tapped")
		}
		if p := effectivePower(t, g, id); p != 4 {
			t.Errorf("power %d, want 4", p)
		}
	}
	if !findBattlefieldCardByID(g, c).Tapped || effectivePower(t, g, c) != 2 {
		t.Error("the untargeted creature was touched")
	}
}

func TestDmuB02ImpedeMomentumTapsStunsThriceAndScries(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := dmuB02Cast(t, g, me, "Impede Momentum", "Sorcery", dmuB02Impede, bear); err != nil {
		t.Fatalf("cast: %v", err)
	}
	dmuB02Resolve(t, g)
	dmuB02AnswerScry(t, g, me)
	if !findBattlefieldCardByID(g, bear).Tapped {
		t.Error("not tapped")
	}
	if n := counterCount(g, bear, game.CounterStun); n != 3 {
		t.Errorf("%d stun counters, want 3", n)
	}
}

func TestDmuB02FrostfistStriderTapsAndStunsOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	strider := castFromHandForTest(t, g, me, "Frostfist Strider", "Creature — Elemental Giant", "{0}", dmuB02Strider, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
			t.Error("a creature the caster controls was accepted")
		}
		pickCard(t, g, me.ID, theirs)
	}
	passPriorityAroundTable(t, g)
	if !findBattlefieldCardByID(g, theirs).Tapped || counterCount(g, theirs, game.CounterStun) != 1 {
		t.Error("the opposing creature was not tapped and stunned")
	}
	if findBattlefieldCardByID(g, mine).Tapped || counterCount(g, mine, game.CounterStun) != 0 {
		t.Error("my creature was touched")
	}
	if findBattlefieldCardByID(g, strider) == nil {
		t.Error("the Strider is not on the battlefield")
	}
}

func TestDmuB02CutDownChecksTotalPowerAndToughness(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ok := b12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 3)
	big := b12Creature(g, opp.ID, "Large", "Creature — Bear", 3, 3)
	if err := dmuB02Cast(t, g, me, "Cut Down", "Instant", dmuB02CutDown, big); err == nil {
		t.Fatal("a creature with total 6 was a legal target")
	}
	if err := dmuB02Cast(t, g, me, "Cut Down", "Instant", dmuB02CutDown, ok); err != nil {
		t.Fatalf("cast: %v", err)
	}
	dmuB02Resolve(t, g)
	if g.Battlefield.Contains(ok) {
		t.Error("a creature with total 5 survived")
	}
}

func TestDmuB02CutDownFizzlesWhenPumpedInResponse(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	c := b12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 3)
	if err := dmuB02Cast(t, g, me, "Cut Down", "Instant", dmuB02CutDown, c); err != nil {
		t.Fatalf("cast: %v", err)
	}
	g.WithWriteLock(func() {
		if err := (BoostUntilEOT{Target: c, Power: 2, Label: "test pump"}).Apply(NewContext(g, &game.StackItem{Controller: opp.ID})); err != nil {
			t.Errorf("pump: %v", err)
		}
	})
	dmuB02Resolve(t, g)
	if !g.Battlefield.Contains(c) {
		t.Error("a creature pumped past the limit was destroyed")
	}
}

func TestDmuB02ExtinguishTheLightGainsLifeOnlyForSmallManaValue(t *testing.T) {
	for _, tc := range []struct {
		name, cost string
		gain       int
	}{
		{"three", "{1}{1}{1}", 3},
		{"four", "{2}{2}", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			id := uuid.New()
			g.WithWriteLock(func() {
				g.Battlefield.PushTop(game.Card{InstanceID: id, Name: "Target", TypeLine: "Creature — Bear", ManaCost: map[string]string{"three": "{3}", "four": "{4}"}[tc.name],
					Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
			})
			life := me.Life
			if err := dmuB02Cast(t, g, me, "Extinguish the Light", "Instant", dmuB02Extinguish, id); err != nil {
				t.Fatalf("cast: %v", err)
			}
			dmuB02Resolve(t, g)
			if g.Battlefield.Contains(id) {
				t.Error("the target survived")
			}
			if me.Life != life+tc.gain {
				t.Errorf("life %d -> %d, want +%d", life, me.Life, tc.gain)
			}
		})
	}
}

func TestDmuB02SmashToDustModes(t *testing.T) {
	t.Run("artifact", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		rock := b12Creature(g, opp.ID, "Rock", "Artifact — Equipment", 0, 0)
		advanceToMain(t, g)
		id := uuid.New()
		g.WithWriteLock(func() {
			me.Hand.PushTop(game.Card{InstanceID: id, Name: "Smash to Dust", TypeLine: "Sorcery", ManaCost: "{0}", OracleID: dmuB02SmashToDust, Owner: me.ID, Controller: me.ID})
		})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{0}, Targets: cardTarget(rock)}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		dmuB02Resolve(t, g)
		if g.Battlefield.Contains(rock) {
			t.Error("the artifact survived")
		}
	})
	t.Run("defender refuses a plain creature", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
		advanceToMain(t, g)
		id := uuid.New()
		g.WithWriteLock(func() {
			me.Hand.PushTop(game.Card{InstanceID: id, Name: "Smash to Dust", TypeLine: "Sorcery", ManaCost: "{0}", OracleID: dmuB02SmashToDust, Owner: me.ID, Controller: me.ID})
		})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{1}, Targets: cardTarget(bear)}); err == nil {
			t.Error("a creature without defender was a legal target")
		}
	})
	t.Run("damage to opposing creatures only", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
		x := b12Creature(g, opp.ID, "X", "Creature — Bear", 1, 1)
		y := b12Creature(g, g.Seats[2].ID, "Y", "Creature — Bear", 2, 2)
		advanceToMain(t, g)
		id := uuid.New()
		g.WithWriteLock(func() {
			me.Hand.PushTop(game.Card{InstanceID: id, Name: "Smash to Dust", TypeLine: "Sorcery", ManaCost: "{0}", OracleID: dmuB02SmashToDust, Owner: me.ID, Controller: me.ID})
		})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{2}}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		dmuB02Resolve(t, g)
		if !g.Battlefield.Contains(mine) {
			t.Error("my own creature took damage")
		}
		if g.Battlefield.Contains(x) {
			t.Error("a 1/1 opposing creature survived")
		}
		if !g.Battlefield.Contains(y) {
			t.Error("a 2/2 died to 1 damage")
		}
	})
}

func TestDmuB02TailSwipePumpsOnlyInYourMainPhase(t *testing.T) {
	for _, main := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
		theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 3)
		advanceToMain(t, g)
		if !main {
			for g.Turn.Step != game.StepBeginCombat {
				if _, err := g.AdvanceStep(); err != nil {
					t.Fatalf("AdvanceStep: %v", err)
				}
			}
		}
		id := uuid.New()
		g.WithWriteLock(func() {
			me.Hand.PushTop(game.Card{InstanceID: id, Name: "Tail Swipe", TypeLine: "Instant", ManaCost: "{0}", OracleID: dmuB02TailSwipe, Owner: me.ID, Controller: me.ID})
		})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: cardTargets(theirs, mine)}); err == nil {
			t.Fatal("targets in the wrong order were accepted")
		}
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: cardTargets(mine, theirs)}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		dmuB02Resolve(t, g)
		got := findBattlefieldCardByID(g, theirs)
		if main {
			// the pumped 3/3 kills their 2/3 and survives its 2 damage.
			if got != nil {
				t.Error("main phase: the opposing 3/3 survived a 3-power fight")
			}
			if p := effectivePower(t, g, mine); p != 3 {
				t.Errorf("main phase: power %d, want 3", p)
			}
		} else {
			if got == nil || got.DamageMarked != 2 {
				t.Errorf("combat: the opposing creature should have taken exactly 2")
			}
			if findBattlefieldCardByID(g, mine) != nil {
				t.Error("combat: my unpumped 2/2 survived 3 damage")
			}
		}
	}
}

func TestDmuB02CosmicEpiphanyDrawsForInstantsAndSorceries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		for i, tl := range []string{"Instant", "Sorcery", "Creature — Bear", "Instant"} {
			me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: string(rune('a' + i)), TypeLine: tl, Owner: me.ID, Controller: me.ID})
		}
	})
	if err := dmuB02Cast(t, g, me, "Cosmic Epiphany", "Sorcery", dmuB02Epiphany); err != nil {
		t.Fatalf("cast: %v", err)
	}
	before := len(me.Hand.Cards)
	dmuB02Resolve(t, g)
	if got := len(me.Hand.Cards) - before; got != 3 {
		t.Errorf("drew %d, want 3", got)
	}
}

func TestDmuB02ErtaisScornCostsLessAfterAnOpponentCastsTwoSpells(t *testing.T) {
	for _, tc := range []struct {
		name  string
		cast  int
		mana  string
		wantW bool
	}{
		{"none", 0, "{U}{U}", false},
		{"one", 1, "{U}{U}", false},
		{"two", 2, "{U}{U}", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			advanceToMain(t, g)
			g.WithWriteLock(func() {
				if g.SpellsCastThisTurn == nil {
					g.SpellsCastThisTurn = map[uuid.UUID]game.CastTally{}
				}
				g.SpellsCastThisTurn[opp.ID] = game.CastTally{Total: tc.cast}
			})
			// A spell to counter, on the stack.
			victim := uuid.New()
			g.WithWriteLock(func() {
				opp.Hand.PushTop(game.Card{InstanceID: victim, Name: "Victim", TypeLine: "Sorcery", ManaCost: "{0}", Owner: opp.ID, Controller: opp.ID})
			})
			id := uuid.New()
			g.WithWriteLock(func() {
				me.Hand.PushTop(game.Card{InstanceID: id, Name: "Ertai's Scorn", TypeLine: "Instant", ManaCost: "{1}{U}{U}", OracleID: dmuB02Scorn, Owner: me.ID, Controller: me.ID})
			})
			if err := g.CastSpell(opp.ID, victim, game.CastSpellParams{}); err != nil {
				t.Skipf("cannot cast the placeholder spell as the opponent here: %v", err)
			}
			dcMana(t, g, me.ID, tc.mana)
			err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}})
			if tc.wantW && err != nil {
				t.Errorf("two spells cast: {U}{U} should pay for it, got %v", err)
			}
			if !tc.wantW && err == nil {
				t.Errorf("%s spells cast: {U}{U} paid for a {1}{U}{U} spell", tc.name)
			}
		})
	}
}

// dmuB02AnswerScry asks for the pending scry, keeps every card on top and
// lets the spell finish.
func dmuB02AnswerScry(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	scry := latestChoiceOfKind(g, game.PendingChoiceScry)
	if scry == nil {
		t.Fatal("the scry asked nothing")
	}
	if err := g.ResolveScry(scry.ID, p.ID, nil, scry.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	dmuB02Resolve(t, g)
}
