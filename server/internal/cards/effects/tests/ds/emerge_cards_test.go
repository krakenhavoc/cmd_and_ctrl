package testsds

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

func TestEmergeCardsAreFullAndDeclareEmerge(t *testing.T) {
	for _, c := range []struct {
		oracle, label string
	}{
		{effectstest.AbundantMawOracle, "Emerge {6}{B}"},
		{effectstest.AdiposeOffspringOracle, "Emerge {5}{W}"},
		{effectstest.CrabominationOracle, "Emerge from artifact {5}{B}{B}"},
		{effectstest.CrestingMosasaurusOracle, "Emerge {6}{U}"},
		{effectstest.DecimatorOracle, "Emerge {6}{G}{G}{G}"},
		{effectstest.DrownyardBehemothOracle, "Emerge {7}{U}"},
		{effectstest.ElderDeepFiendOracle, "Emerge {5}{U}{U}"},
		{effectstest.ItOfTheHorridSwarmOracle, "Emerge {6}{G}"},
		{effectstest.LashweedLurkerOracle, "Emerge {5}{G}{U}"},
		{effectstest.MockeryOfNatureOracle, "Emerge {7}{G}"},
		{effectstest.TwistedRiddlekeeperOrcl, "Emerge {5}{C}{U}"},
		{effectstest.VexingScuttlerOracle, "Emerge {6}{U}"},
		{effectstest.WretchedGryffOracle, "Emerge {5}{U}"},
	} {
		spec, ok := effects.Lookup(c.oracle)
		if !ok {
			t.Errorf("%s is not registered", c.oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: completeness %v, want Full", spec.Name, spec.Completeness)
		}
		offers := game.AlternativeCostsFor(c.oracle)
		if len(offers) != 1 || offers[0].Key != effects.AltCostKeyEmerge || offers[0].Label != c.label ||
			!offers[0].ReducedBySacrificedManaValue || offers[0].Sacrifice == nil || offers[0].CardPaymentCount() != 1 {
			t.Errorf("%s offers %+v, want one %q", spec.Name, offers, c.label)
		}
	}
}

// CR 702.119a: reduced by the sacrificed creature's mana value; CR 118.7a:
// the generic part only, never below {0}; CR 202.3: a token is 0.
func TestEmergeIsPricedByTheSacrificedPermanentsManaValue(t *testing.T) {
	g, me, _ := effectstest.EmergeTable(t)
	fiend := effectstest.HandCardOf(g, me, "Elder Deep-Fiend", "Creature — Eldrazi Octopus", "{8}", effectstest.ElderDeepFiendOracle)
	riddle := effectstest.HandCardOf(g, me, "Twisted Riddlekeeper", "Creature — Eldrazi Sphinx", "{8}", effectstest.TwistedRiddlekeeperOrcl)
	four := effectstest.EmergeFodder(g, me.ID, "Four", "{2}{G}{G}", 4, 4)
	seven := effectstest.EmergeFodder(g, me.ID, "Seven", "{7}", 7, 7)
	spawn := effectstest.ApaPush(g, me.ID, me.ID, effects.EldraziSpawnToken())
	for _, c := range []struct {
		name      string
		card, sac uuid.UUID
		want      string
	}{
		{"a four-drop", fiend, four, "{1}{U}{U}"},
		{"a seven-drop", fiend, seven, "{U}{U}"},
		{"a token", fiend, spawn, "{5}{U}{U}"},
		{"Riddlekeeper's {C} stays", riddle, seven, "{C}{U}"},
	} {
		if got := effectstest.EmergePrice(t, g, me, c.card, c.sac); got != c.want {
			t.Errorf("%s: price %s, want %s", c.name, got, c.want)
		}
	}
}

// CR 601.2f: the increase is added before the reduction is taken off.
func TestEmergeReductionComesAfterAnIncrease(t *testing.T) {
	g, me, opp := effectstest.EmergeTable(t)
	effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Sphere of Resistance", TypeLine: "Artifact", ManaCost: "{2}", OracleID: effectstest.SphereOfResistanceOracle})
	fiend := effectstest.HandCardOf(g, me, "Elder Deep-Fiend", "Creature — Eldrazi Octopus", "{8}", effectstest.ElderDeepFiendOracle)
	four := effectstest.EmergeFodder(g, me.ID, "Four", "{4}", 4, 4)
	seven := effectstest.EmergeFodder(g, me.ID, "Seven", "{7}", 7, 7)
	if got := effectstest.EmergePrice(t, g, me, fiend, four); got != "{2}{U}{U}" {
		t.Errorf("over a four-drop under Sphere: %s, want {2}{U}{U}", got)
	}
	if got := effectstest.EmergePrice(t, g, me, fiend, seven); got != "{U}{U}" {
		t.Errorf("over a seven-drop under Sphere: %s, want {U}{U}", got)
	}
}

// End to end under the strict gate: three Islands pay {1}{U}{U}, the
// creature dies with the spell on the stack, and the cast trigger taps up
// to four permanents. Two Islands are refused with nothing paid.
func TestElderDeepFiendEmergesOverAFourDrop(t *testing.T) {
	g, me, opp := effectstest.EmergeTable(t)
	fiend := effectstest.HandCardOf(g, me, "Elder Deep-Fiend", "Creature — Eldrazi Octopus", "{8}", effectstest.ElderDeepFiendOracle)
	four := effectstest.EmergeFodder(g, me.ID, "Four", "{4}", 4, 4)
	theirs := effectstest.Pr7Creature(g, opp.ID, "Theirs", 2, "G")
	islands := effectstest.EmergeLands(g, me.ID, 2, "Island")
	if err := effectstest.CastEmerge(g, me, fiend, four, true); err == nil {
		t.Fatal("Elder Deep-Fiend was emerged with two Islands for {1}{U}{U}")
	}
	if effectstest.FindBattlefieldCardForTest(g, four) == nil || !me.Hand.Contains(fiend) || effectstest.TappedForTest(t, g, islands[0]) {
		t.Fatal("a refused cast paid something")
	}
	islands = append(islands, effectstest.EmergeLands(g, me.ID, 1, "Island")...)
	if err := effectstest.CastEmerge(g, me, fiend, four, true); err != nil {
		t.Fatalf("Elder Deep-Fiend over a four-drop with three Islands: %v", err)
	}
	if effectstest.FindBattlefieldCardForTest(g, four) != nil || !me.Graveyard.Contains(four) {
		t.Fatal("the four-drop was not sacrificed with the spell on the stack")
	}
	for _, id := range islands {
		if !effectstest.TappedForTest(t, g, id) {
			t.Error("an Island was left untapped paying {1}{U}{U}")
		}
	}
	var paid []game.ObjectRef
	g.ReadSnapshot(func() {
		if item := g.StackMeta[fiend]; item != nil {
			paid = append(paid, item.Paid.AltCostObjects...)
		}
	})
	if len(paid) != 1 || paid[0].ID != four {
		t.Errorf("PaidCost.AltCostObjects = %+v, want the four-drop", paid)
	}
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}); err != nil {
			t.Fatalf("pick targets: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedForTest(t, g, theirs) {
		t.Error("the cast trigger did not tap the target")
	}
	if effectstest.FindBattlefieldCardForTest(g, fiend) == nil {
		t.Error("Elder Deep-Fiend did not resolve")
	}
}

// Owner decision 3: a creature named to the emerge sacrifice may tap for
// mana first (CR 601.2g before 601.2h). An Eldrazi Spawn named to it may
// not be cracked for its own mana: that would leave the sacrifice nothing
// to pay with.
func TestEmergeSacrificedCreatureMayTapForManaFirst(t *testing.T) {
	g, me, _ := effectstest.EmergeTable(t)
	gryff := effectstest.HandCardOf(g, me, "Wretched Gryff", "Creature — Eldrazi Hippogriff", "{7}", effectstest.WretchedGryffOracle)
	elf := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Llanowar Elves", OracleID: effectstest.LlanowarElvesOracle,
		TypeLine: "Creature — Elf Druid", ManaCost: "{G}", Power: 1, Toughness: 1})
	islands := effectstest.EmergeLands(g, me.ID, 4, "Island")
	if err := effectstest.CastEmerge(g, me, gryff, elf, true); err != nil {
		t.Fatalf("Wretched Gryff for {4}{U} with four Islands and the Elves it sacrifices: %v", err)
	}
	if !me.Graveyard.Contains(elf) {
		t.Error("the Elves were not sacrificed")
	}
	for _, id := range islands {
		if !effectstest.TappedForTest(t, g, id) {
			t.Error("an Island was left untapped")
		}
	}

	g2, me2, _ := effectstest.EmergeTable(t)
	gryff2 := effectstest.HandCardOf(g2, me2, "Wretched Gryff", "Creature — Eldrazi Hippogriff", "{7}", effectstest.WretchedGryffOracle)
	spawn := effectstest.ApaPush(g2, me2.ID, me2.ID, effects.EldraziSpawnToken())
	effectstest.EmergeLands(g2, me2.ID, 5, "Island")
	if err := effectstest.CastEmerge(g2, me2, gryff2, spawn, true); err == nil {
		t.Fatal("the Spawn named to emerge was cracked for the {5}{U}'s sixth mana")
	}
	if effectstest.FindBattlefieldCardForTest(g2, spawn) == nil {
		t.Fatal("the refused cast spent the Spawn")
	}
}

// The offer's card component is a creature (or, for Crabomination, an
// artifact) you control, exactly one.
func TestEmergeRefusesTheWrongPayment(t *testing.T) {
	g, me, opp := effectstest.EmergeTable(t)
	gryff := effectstest.HandCardOf(g, me, "Wretched Gryff", "Creature — Eldrazi Hippogriff", "{7}", effectstest.WretchedGryffOracle)
	crab := effectstest.HandCardOf(g, me, "Crabomination", "Creature — Crab Demon", "{4}{B}{B}", effectstest.CrabominationOracle)
	theirs := effectstest.EmergeFodder(g, opp.ID, "Theirs", "{3}", 3, 3)
	mine := effectstest.EmergeFodder(g, me.ID, "Mine", "{3}", 3, 3)
	rock := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{5}"})
	land := effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	for name, c := range map[string]struct{ card, sac uuid.UUID }{
		"an opponent's creature":       {gryff, theirs},
		"a land":                       {gryff, land},
		"a noncreature artifact":       {gryff, rock},
		"a creature for Crabomination": {crab, mine},
	} {
		if err := effectstest.CastEmerge(g, me, c.card, c.sac, false); err == nil {
			t.Errorf("%s paid emerge", name)
		}
	}
	if err := g.CastSpell(me.ID, gryff, game.CastSpellParams{AlternativeCost: effects.AltCostKeyEmerge}); err == nil {
		t.Error("emerge was paid sacrificing nothing")
	}
	if err := g.CastSpell(me.ID, gryff, game.CastSpellParams{AlternativeCost: effects.AltCostKeyEmerge, AltCostIDs: []uuid.UUID{mine, rock}}); err == nil {
		t.Error("emerge was paid sacrificing two permanents")
	}
	if got := effectstest.EmergePrice(t, g, me, crab, rock); got != "{B}{B}" {
		t.Errorf("Crabomination over a five-mana artifact: %s, want {B}{B}", got)
	}
}

// emergeResolve casts `card` emerged over a fresh vanilla creature in
// permissive mode, answers its "you may" and target prompts with `pick`,
// and lets everything resolve.
func emergeResolve(t *testing.T, g *game.Game, me *game.Player, card uuid.UUID, targets []game.TargetRef, pick []game.TargetRef) {
	t.Helper()
	fodder := effectstest.EmergeFodder(g, me.ID, "Fodder", "{3}", 1, 1)
	if err := effectstest.CastEmerge(g, me, card, fodder, false, targets...); err != nil {
		t.Fatalf("emerge: %v", err)
	}
	// A trigger's targets are picked as it goes on the stack; its "you
	// may" is asked as it resolves (CR 603.5).
	for i := 0; i < 16; i++ {
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil && len(pick) > 0 {
			if err := g.ResolvePickTargets(p.ID, me.ID, pick); err != nil {
				t.Fatalf("pick targets: %v", err)
			}
			continue
		}
		if c := effectstest.LatestTriggerPrompt(g, me.ID); c != nil {
			if err := g.ResolveTriggerPrompt(c.ID, me.ID, true); err != nil {
				t.Fatalf("ResolveTriggerPrompt: %v", err)
			}
			continue
		}
		if effectstest.StackFullyEmpty(g) {
			break
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// Every cast trigger does its printed thing once the emerge cost is paid.
func TestEmergeCastTriggersDoTheirThing(t *testing.T) {
	ref := func(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }
	t.Run("Abundant Maw", func(t *testing.T) {
		g, me, opp := effectstest.EmergeTable(t)
		maw := effectstest.HandCardOf(g, me, "Abundant Maw", "Creature — Eldrazi Leech", "{8}", effectstest.AbundantMawOracle)
		mine, theirs := me.Life, opp.Life
		emergeResolve(t, g, me, maw, nil, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
		if opp.Life != theirs-3 || me.Life != mine+3 {
			t.Errorf("life me %d→%d, opponent %d→%d; want +3 and -3", mine, me.Life, theirs, opp.Life)
		}
	})
	t.Run("It of the Horrid Swarm", func(t *testing.T) {
		g, me, _ := effectstest.EmergeTable(t)
		it := effectstest.HandCardOf(g, me, "It of the Horrid Swarm", "Creature — Eldrazi Insect", "{8}", effectstest.ItOfTheHorridSwarmOracle)
		emergeResolve(t, g, me, it, nil, nil)
		if n, _ := effectstest.CountTokensNamed(g, me.ID, "Insect"); n != 2 {
			t.Errorf("%d Insect tokens, want 2", n)
		}
	})
	t.Run("Wretched Gryff", func(t *testing.T) {
		g, me, _ := effectstest.EmergeTable(t)
		gryff := effectstest.HandCardOf(g, me, "Wretched Gryff", "Creature — Eldrazi Hippogriff", "{7}", effectstest.WretchedGryffOracle)
		emergeResolve(t, g, me, gryff, nil, nil)
		if me.Hand.Size() != 1 {
			t.Errorf("hand %d after the Gryff, want the one card drawn", me.Hand.Size())
		}
	})
	t.Run("Decimator of the Provinces", func(t *testing.T) {
		g, me, _ := effectstest.EmergeTable(t)
		bear := effectstest.EmergeFodder(g, me.ID, "Bear", "{2}", 2, 2)
		dec := effectstest.HandCardOf(g, me, "Decimator of the Provinces", "Creature — Eldrazi Boar", "{10}", effectstest.DecimatorOracle)
		g.WithWriteLock(func() {
			for i := range me.Hand.Cards {
				if me.Hand.Cards[i].InstanceID == dec {
					me.Hand.Cards[i].Power, me.Hand.Cards[i].Toughness = 7, 7
				}
			}
		})
		emergeResolve(t, g, me, dec, nil, nil)
		g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
		c := effectstest.FindBattlefieldCardForTest(g, bear)
		if c == nil || c.CurrentPower() != 4 || c.CurrentToughness() != 4 || !game.HasKeyword(c, "trample") {
			t.Errorf("the Bear is not a 4/4 trampler: %+v", c)
		}
		if d := effectstest.FindBattlefieldCardForTest(g, dec); d == nil || d.CurrentPower() != 7 {
			t.Error("the Decimator pumped itself from the stack")
		}
	})
	t.Run("Twisted Riddlekeeper", func(t *testing.T) {
		g, me, opp := effectstest.EmergeTable(t)
		a := effectstest.Pr7Creature(g, opp.ID, "A", 2, "G")
		b := effectstest.Pr7Creature(g, opp.ID, "B", 2, "G")
		riddle := effectstest.HandCardOf(g, me, "Twisted Riddlekeeper", "Creature — Eldrazi Sphinx", "{8}", effectstest.TwistedRiddlekeeperOrcl)
		emergeResolve(t, g, me, riddle, nil, []game.TargetRef{ref(a), ref(b)})
		for _, id := range []uuid.UUID{a, b} {
			c := effectstest.FindBattlefieldCardForTest(g, id)
			if c == nil || !c.Tapped || c.Counters[game.CounterStun] != 1 {
				t.Errorf("a target is not tapped with one stun counter: %+v", c)
			}
		}
	})
	t.Run("Mockery of Nature", func(t *testing.T) {
		g, me, opp := effectstest.EmergeTable(t)
		rock := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{2}"})
		mock := effectstest.HandCardOf(g, me, "Mockery of Nature", "Creature — Eldrazi Beast", "{9}", effectstest.MockeryOfNatureOracle)
		emergeResolve(t, g, me, mock, nil, []game.TargetRef{ref(rock)})
		if effectstest.FindBattlefieldCardForTest(g, rock) != nil {
			t.Error("the artifact was not destroyed")
		}
	})
	t.Run("Lashweed Lurker", func(t *testing.T) {
		g, me, opp := effectstest.EmergeTable(t)
		theirs := effectstest.Pr7Creature(g, opp.ID, "Theirs", 2, "G")
		lash := effectstest.HandCardOf(g, me, "Lashweed Lurker", "Creature — Eldrazi Horror", "{8}", effectstest.LashweedLurkerOracle)
		emergeResolve(t, g, me, lash, nil, []game.TargetRef{ref(theirs)})
		top, err := opp.Library.Top()
		if err != nil || top.InstanceID != theirs {
			t.Error("the permanent is not on top of its owner's library")
		}
	})
	t.Run("Vexing Scuttler", func(t *testing.T) {
		g, me, _ := effectstest.EmergeTable(t)
		bolt := effectstest.PushGraveyardPermanent(me, "Bolt", "Instant", "{R}")
		vex := effectstest.HandCardOf(g, me, "Vexing Scuttler", "Creature — Eldrazi Crab", "{8}", effectstest.VexingScuttlerOracle)
		emergeResolve(t, g, me, vex, nil, []game.TargetRef{ref(bolt)})
		if !me.Hand.Contains(bolt) {
			t.Error("the instant did not return to hand")
		}
	})
}

// Adipose Offspring: X Aliens where X is the sacrificed creature's
// toughness as it last existed (its +1/+1 counter counts), even when the
// Offspring has left before its enters trigger resolves; one Alien when
// cast for its mana cost.
func TestAdiposeOffspringMakesAliensByTheSacrificedToughness(t *testing.T) {
	g, me, _ := effectstest.EmergeTable(t)
	fodder := effectstest.EmergeFodder(g, me.ID, "Wall", "{3}", 0, 4)
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, fodder)
		c.Counters = map[string]int{"+1/+1": 1}
		g.RecomputeLayersIfStaleLocked()
	})
	off := effectstest.HandCardOf(g, me, "Adipose Offspring", "Creature — Alien", "{3}{W}", effectstest.AdiposeOffspringOracle)
	if err := effectstest.CastEmerge(g, me, off, fodder, false); err != nil {
		t.Fatalf("emerge: %v", err)
	}
	// Resolve the spell, then bounce the Offspring with its trigger waiting.
	for i := 0; i < 8 && effectstest.FindBattlefieldCardForTest(g, off) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if effectstest.FindBattlefieldCardForTest(g, off) == nil {
		t.Fatal("the Offspring did not resolve")
	}
	g.WithWriteLock(func() { g.BounceCardsToHandForEffect([]uuid.UUID{off}) })
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Alien"); n != 5 {
		t.Errorf("%d Aliens, want 5 (the sacrificed creature's toughness, counter included)", n)
	}

	g2, me2, _ := effectstest.EmergeTable(t)
	off2 := effectstest.HandCardOf(g2, me2, "Adipose Offspring", "Creature — Alien", "{3}{W}", effectstest.AdiposeOffspringOracle)
	if err := g2.CastSpell(me2.ID, off2, game.CastSpellParams{}); err != nil {
		t.Fatalf("hard cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if n, _ := effectstest.CountTokensNamed(g2, me2.ID, "Alien"); n != 1 {
		t.Errorf("%d Aliens for a hard-cast Offspring, want 1", n)
	}
}

// Cresting Mosasaurus, cast: every creature but Dinosaurs goes home.
func TestCrestingMosasaurusBouncesEveryNonDinosaur(t *testing.T) {
	g, me, opp := effectstest.EmergeTable(t)
	theirs := effectstest.Pr7Creature(g, opp.ID, "Theirs", 2, "G")
	dino := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Raptor", TypeLine: "Creature — Dinosaur", ManaCost: "{2}", Power: 2, Toughness: 2})
	mosa := effectstest.HandCardOf(g, me, "Cresting Mosasaurus", "Creature — Dinosaur", "{6}{U}{U}", effectstest.CrestingMosasaurusOracle)
	emergeResolve(t, g, me, mosa, nil, nil)
	if !opp.Hand.Contains(theirs) {
		t.Error("the opponent's creature was not returned")
	}
	if effectstest.FindBattlefieldCardForTest(g, dino) == nil || effectstest.FindBattlefieldCardForTest(g, mosa) == nil {
		t.Error("a Dinosaur was returned")
	}
}

// Drownyard Behemoth has hexproof the turn it enters and not after.
func TestDrownyardBehemothHasHexproofOnlyTheTurnItEntered(t *testing.T) {
	g, me, _ := effectstest.EmergeTable(t)
	beh := effectstest.HandCardOf(g, me, "Drownyard Behemoth", "Creature — Eldrazi Crab", "{9}", effectstest.DrownyardBehemothOracle)
	emergeResolve(t, g, me, beh, nil, nil)
	hexproof := func() bool {
		var has bool
		g.WithWriteLock(func() {
			g.RecomputeLayersIfStaleLocked()
			if c := effectstest.FindBattlefieldCardForTest(g, beh); c != nil {
				has = game.HasKeyword(c, "hexproof")
			}
		})
		return has
	}
	if !hexproof() {
		t.Fatal("no hexproof the turn it entered")
	}
	effectstest.B39NextTurnOf(t, g, 0)
	if effectstest.FindBattlefieldCardForTest(g, beh) == nil {
		t.Fatal("the Behemoth left")
	}
	if hexproof() {
		t.Error("hexproof outlasted the turn it entered")
	}
}

// Crabomination: emerged from an artifact, its enters trigger exiles the
// opponent's top card, a random graveyard card and a random hand card, and
// one spell among them may be cast for free.
func TestCrabominationExilesThreeAndCastsOneFree(t *testing.T) {
	g, me, opp := effectstest.EmergeTable(t)
	bear := func() game.Card {
		return game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
			Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID}
	}
	var lib, gy, hand game.Card
	g.WithWriteLock(func() {
		lib, gy, hand = bear(), bear(), bear()
		opp.Library.PushTop(lib)
		opp.Graveyard.PushTop(gy)
		opp.Hand.Cards = []game.Card{hand}
	})
	rock := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{5}"})
	crab := effectstest.HandCardOf(g, me, "Crabomination", "Creature — Crab Demon", "{4}{B}{B}", effectstest.CrabominationOracle)
	if err := effectstest.CastEmerge(g, me, crab, rock, false); err != nil {
		t.Fatalf("emerge from artifact: %v", err)
	}
	for i := 0; i < 8 && effectstest.FindBattlefieldCardForTest(g, crab) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
			t.Fatalf("pick the opponent: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range []game.Card{lib, gy, hand} {
		if !g.Exile.Contains(c.InstanceID) {
			t.Fatalf("a card was not exiled (library %v, graveyard %v, hand %v)",
				g.Exile.Contains(lib.InstanceID), g.Exile.Contains(gy.InstanceID), g.Exile.Contains(hand.InstanceID))
		}
	}
	if err := g.CastSpell(me.ID, hand.InstanceID, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("casting an exiled card for free: %v", err)
	}
	if err := g.CastSpell(me.ID, lib.InstanceID, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("a second spell was cast from among the exiled cards")
	}
}

// The wire: an emerge offer is flagged and priced per candidate by the
// engine's pricer.
func TestEmergeOfferOnTheWireIsPricedPerCandidate(t *testing.T) {
	g, me, _ := effectstest.EmergeTable(t)
	fiend := effectstest.HandCardOf(g, me, "Elder Deep-Fiend", "Creature — Eldrazi Octopus", "{8}", effectstest.ElderDeepFiendOracle)
	four := effectstest.EmergeFodder(g, me.ID, "Four", "{4}", 4, 4)
	spawn := effectstest.ApaPush(g, me.ID, me.ID, effects.EldraziSpawnToken())
	var offer *protocol.AlternativeCostView
	v := protocol.ViewOfGameFor(g, me.ID.String())
	for _, s := range v.Seats {
		for i := range s.Hand.Cards {
			if s.Hand.Cards[i].InstanceID != fiend.String() {
				continue
			}
			for j := range s.Hand.Cards[i].AlternativeCosts {
				if s.Hand.Cards[i].AlternativeCosts[j].Key == effects.AltCostKeyEmerge {
					offer = &s.Hand.Cards[i].AlternativeCosts[j]
				}
			}
		}
	}
	if offer == nil {
		t.Fatal("the emerge offer is not on the wire")
	}
	if !offer.ReducesByManaValue || offer.SacrificeOptions == nil || offer.PayLabel != "a creature" {
		t.Fatalf("offer %+v, want a flagged sacrifice offer", offer)
	}
	want := map[uuid.UUID]protocol.AltCostPriceView{
		four:  {ManaValue: 4, Price: "{1}{U}{U}"},
		spawn: {ManaValue: 0, Price: "{5}{U}{U}"},
	}
	for id, w := range want {
		if got := offer.SacrificePrices[id.String()]; got != w {
			t.Errorf("price for %s: %+v, want %+v", id, got, w)
		}
	}
}

// Owner decision 5: the enumerator prices each payment on its own and
// offers only the ones the seat can afford — here the creatures worth five
// or more, with one Island for the Gryff's {U}.
func TestEmergeEnumeratorOffersOnlyAffordablePayments(t *testing.T) {
	g, me, _ := effectstest.EmergeTable(t)
	gryff := effectstest.HandCardOf(g, me, "Wretched Gryff", "Creature — Eldrazi Hippogriff", "{7}", effectstest.WretchedGryffOracle)
	effectstest.EmergeLands(g, me.ID, 1, "Island")
	two := effectstest.EmergeFodder(g, me.ID, "Two", "{2}", 2, 2)
	five := effectstest.EmergeFodder(g, me.ID, "Five", "{5}", 5, 5)
	six := effectstest.EmergeFodder(g, me.ID, "Six", "{6}", 6, 6)
	spawn := effectstest.ApaPush(g, me.ID, me.ID, effects.EldraziSpawnToken())
	offered := map[uuid.UUID]bool{}
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type != legal.TypeCastSpell || m.Source != gryff {
			continue
		}
		var p struct {
			AlternativeCost string      `json:"alternative_cost"`
			AltCostIDs      []uuid.UUID `json:"alt_cost_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil || p.AlternativeCost != effects.AltCostKeyEmerge || len(p.AltCostIDs) != 1 {
			continue
		}
		offered[p.AltCostIDs[0]] = true
	}
	if !offered[five] || !offered[six] {
		t.Errorf("offered %v, want the five- and six-drops", offered)
	}
	if offered[two] || offered[spawn] {
		t.Errorf("offered %v, which includes a payment the seat can't afford", offered)
	}
}

// effects.Register holds the flag to a one-permanent sacrifice.
func TestRegisterRefusesEmergeWithoutOneSacrifice(t *testing.T) {
	two := effects.Emerge("{4}")
	two.Sacrifice = two.Sacrifice.WithCount(2, 2)
	none := effects.Overload("{4}")
	none.ReducedBySacrificedManaValue = true
	for name, ac := range map[string]game.AlternativeCost{"two": two, "none": none} {
		msg := effectstest.RegisterPanics(effects.Spec{OracleID: "test-0135-emerge-" + name, Name: "Test " + name,
			AlternativeCosts: []game.AlternativeCost{ac}})
		if !strings.Contains(msg, "CR 702.119a") {
			t.Errorf("%s: panic %q, want the emerge guard", name, msg)
		}
	}
}
