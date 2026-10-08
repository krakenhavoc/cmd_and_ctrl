package effects

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// emerge_cards_test.go — ADR 0135 §4 (#2416): emerge (CR 702.119) against
// the real catalog. The offer is a one-permanent sacrifice whose price is
// reduced by that permanent's mana value (CR 702.119a), taken off the
// generic part only (CR 118.7a) after any increase (CR 601.2f); the
// sacrificed creature may tap for mana before it goes (owner decision 3);
// the enumerator prices each payment on its own (owner decision 5).

const (
	abundantMawOracle        = "3676b745-001d-46e6-880f-2fe26476a38d"
	adiposeOffspringOracle   = "1e21e57e-3bc9-41a8-9746-57b583a5ad63"
	crabominationOracle      = "5c26c30b-bc4a-4d20-8406-ef9022256e0d"
	crestingMosasaurusOracle = "bf245433-8e3b-47d7-b628-3cdae93c384d"
	decimatorOracle          = "7f522ded-09fd-457e-9efe-0a6324925e4c"
	drownyardBehemothOracle  = "bfe1a52a-f2f8-4c62-8502-037f2d1395b2"
	elderDeepFiendOracle     = "4eafe717-4ba4-4901-8c67-11757230eb54"
	itOfTheHorridSwarmOracle = "f4ad190a-6c82-4923-af54-b4cce8d6465f"
	lashweedLurkerOracle     = "a7dd75bf-ecda-406c-baa2-8c051f138809"
	mockeryOfNatureOracle    = "5da5240e-5d87-494f-aec9-64b7a3f0d935"
	twistedRiddlekeeperOrcl  = "c7ad20a3-51bf-4a22-a09c-fff8cae22765"
	vexingScuttlerOracle     = "2a5605a8-606e-4866-9473-7a2240ac6b8c"
	wretchedGryffOracle      = "83d0c4bc-aab9-4c14-a7dc-a344fcce4fea"
)

func TestEmergeCardsAreFullAndDeclareEmerge(t *testing.T) {
	for _, c := range []struct {
		oracle, label string
	}{
		{abundantMawOracle, "Emerge {6}{B}"},
		{adiposeOffspringOracle, "Emerge {5}{W}"},
		{crabominationOracle, "Emerge from artifact {5}{B}{B}"},
		{crestingMosasaurusOracle, "Emerge {6}{U}"},
		{decimatorOracle, "Emerge {6}{G}{G}{G}"},
		{drownyardBehemothOracle, "Emerge {7}{U}"},
		{elderDeepFiendOracle, "Emerge {5}{U}{U}"},
		{itOfTheHorridSwarmOracle, "Emerge {6}{G}"},
		{lashweedLurkerOracle, "Emerge {5}{G}{U}"},
		{mockeryOfNatureOracle, "Emerge {7}{G}"},
		{twistedRiddlekeeperOrcl, "Emerge {5}{C}{U}"},
		{vexingScuttlerOracle, "Emerge {6}{U}"},
		{wretchedGryffOracle, "Emerge {5}{U}"},
	} {
		spec, ok := Lookup(c.oracle)
		if !ok {
			t.Errorf("%s is not registered", c.oracle)
			continue
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s: completeness %v, want Full", spec.Name, spec.Completeness)
		}
		offers := game.AlternativeCostsFor(c.oracle)
		if len(offers) != 1 || offers[0].Key != AltCostKeyEmerge || offers[0].Label != c.label ||
			!offers[0].ReducedBySacrificedManaValue || offers[0].Sacrifice == nil || offers[0].CardPaymentCount() != 1 {
			t.Errorf("%s offers %+v, want one %q", spec.Name, offers, c.label)
		}
	}
}

// emergeTable is a main phase on seat 0's turn with an empty hand.
func emergeTable(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	return g, me, g.Seats[1]
}

// emergeFodder puts a creature with mana cost `cost` on `owner`'s
// battlefield.
func emergeFodder(g *game.Game, owner uuid.UUID, name, cost string, power, toughness int) uuid.UUID {
	return apaPush(g, owner, owner, game.Card{Name: name, TypeLine: "Creature — Test", ManaCost: cost, Power: power, Toughness: toughness})
}

func emergeLands(g *game.Game, owner uuid.UUID, n int, name string) []uuid.UUID {
	var out []uuid.UUID
	for i := 0; i < n; i++ {
		out = append(out, pushLandFor(g, owner, name, "Basic Land — "+name))
	}
	return out
}

// castEmerge casts `card` for its emerge cost sacrificing `sac`. Strict
// casts pay their mana through the auto-tapper; the others are permissive.
func castEmerge(g *game.Game, p *game.Player, card, sac uuid.UUID, strict bool, targets ...game.TargetRef) error {
	return g.CastSpell(p.ID, card, game.CastSpellParams{
		AlternativeCost: AltCostKeyEmerge, AltCostIDs: []uuid.UUID{sac},
		Strict: strict, AutoTap: strict, Targets: targets,
	})
}

// emergePrice is what the engine's one pricer charges for `card` emerged
// over `sac`.
func emergePrice(t *testing.T, g *game.Game, p *game.Player, card, sac uuid.UUID) string {
	t.Helper()
	var out string
	g.WithWriteLock(func() {
		c, ok := g.LookupCardForEffect(card)
		if !ok {
			t.Fatal("the emerge card is gone")
		}
		price, err := g.PriceCastForEffect(p.ID, c, game.CastSpellParams{AlternativeCost: AltCostKeyEmerge, AltCostIDs: []uuid.UUID{sac}})
		if err != nil {
			t.Fatalf("PriceCastForEffect: %v", err)
		}
		out = price.Total.String()
	})
	return out
}

// CR 702.119a: reduced by the sacrificed creature's mana value; CR 118.7a:
// the generic part only, never below {0}; CR 202.3: a token is 0.
func TestEmergeIsPricedByTheSacrificedPermanentsManaValue(t *testing.T) {
	g, me, _ := emergeTable(t)
	fiend := handCardOf(g, me, "Elder Deep-Fiend", "Creature — Eldrazi Octopus", "{8}", elderDeepFiendOracle)
	riddle := handCardOf(g, me, "Twisted Riddlekeeper", "Creature — Eldrazi Sphinx", "{8}", twistedRiddlekeeperOrcl)
	four := emergeFodder(g, me.ID, "Four", "{2}{G}{G}", 4, 4)
	seven := emergeFodder(g, me.ID, "Seven", "{7}", 7, 7)
	spawn := apaPush(g, me.ID, me.ID, EldraziSpawnToken())
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
		if got := emergePrice(t, g, me, c.card, c.sac); got != c.want {
			t.Errorf("%s: price %s, want %s", c.name, got, c.want)
		}
	}
}

// CR 601.2f: the increase is added before the reduction is taken off.
func TestEmergeReductionComesAfterAnIncrease(t *testing.T) {
	g, me, opp := emergeTable(t)
	apaPush(g, opp.ID, opp.ID, game.Card{Name: "Sphere of Resistance", TypeLine: "Artifact", ManaCost: "{2}", OracleID: sphereOfResistanceOracle})
	fiend := handCardOf(g, me, "Elder Deep-Fiend", "Creature — Eldrazi Octopus", "{8}", elderDeepFiendOracle)
	four := emergeFodder(g, me.ID, "Four", "{4}", 4, 4)
	seven := emergeFodder(g, me.ID, "Seven", "{7}", 7, 7)
	if got := emergePrice(t, g, me, fiend, four); got != "{2}{U}{U}" {
		t.Errorf("over a four-drop under Sphere: %s, want {2}{U}{U}", got)
	}
	if got := emergePrice(t, g, me, fiend, seven); got != "{U}{U}" {
		t.Errorf("over a seven-drop under Sphere: %s, want {U}{U}", got)
	}
}

// End to end under the strict gate: three Islands pay {1}{U}{U}, the
// creature dies with the spell on the stack, and the cast trigger taps up
// to four permanents. Two Islands are refused with nothing paid.
func TestElderDeepFiendEmergesOverAFourDrop(t *testing.T) {
	g, me, opp := emergeTable(t)
	fiend := handCardOf(g, me, "Elder Deep-Fiend", "Creature — Eldrazi Octopus", "{8}", elderDeepFiendOracle)
	four := emergeFodder(g, me.ID, "Four", "{4}", 4, 4)
	theirs := pr7Creature(g, opp.ID, "Theirs", 2, "G")
	islands := emergeLands(g, me.ID, 2, "Island")
	if err := castEmerge(g, me, fiend, four, true); err == nil {
		t.Fatal("Elder Deep-Fiend was emerged with two Islands for {1}{U}{U}")
	}
	if findBattlefieldCardForTest(g, four) == nil || !me.Hand.Contains(fiend) || tappedForTest(t, g, islands[0]) {
		t.Fatal("a refused cast paid something")
	}
	islands = append(islands, emergeLands(g, me.ID, 1, "Island")...)
	if err := castEmerge(g, me, fiend, four, true); err != nil {
		t.Fatalf("Elder Deep-Fiend over a four-drop with three Islands: %v", err)
	}
	if findBattlefieldCardForTest(g, four) != nil || !me.Graveyard.Contains(four) {
		t.Fatal("the four-drop was not sacrificed with the spell on the stack")
	}
	for _, id := range islands {
		if !tappedForTest(t, g, id) {
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
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}); err != nil {
			t.Fatalf("pick targets: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if !tappedForTest(t, g, theirs) {
		t.Error("the cast trigger did not tap the target")
	}
	if findBattlefieldCardForTest(g, fiend) == nil {
		t.Error("Elder Deep-Fiend did not resolve")
	}
}

// Owner decision 3: a creature named to the emerge sacrifice may tap for
// mana first (CR 601.2g before 601.2h). An Eldrazi Spawn named to it may
// not be cracked for its own mana: that would leave the sacrifice nothing
// to pay with.
func TestEmergeSacrificedCreatureMayTapForManaFirst(t *testing.T) {
	g, me, _ := emergeTable(t)
	gryff := handCardOf(g, me, "Wretched Gryff", "Creature — Eldrazi Hippogriff", "{7}", wretchedGryffOracle)
	elf := apaPush(g, me.ID, me.ID, game.Card{Name: "Llanowar Elves", OracleID: llanowarElvesOracle,
		TypeLine: "Creature — Elf Druid", ManaCost: "{G}", Power: 1, Toughness: 1})
	islands := emergeLands(g, me.ID, 4, "Island")
	if err := castEmerge(g, me, gryff, elf, true); err != nil {
		t.Fatalf("Wretched Gryff for {4}{U} with four Islands and the Elves it sacrifices: %v", err)
	}
	if !me.Graveyard.Contains(elf) {
		t.Error("the Elves were not sacrificed")
	}
	for _, id := range islands {
		if !tappedForTest(t, g, id) {
			t.Error("an Island was left untapped")
		}
	}

	g2, me2, _ := emergeTable(t)
	gryff2 := handCardOf(g2, me2, "Wretched Gryff", "Creature — Eldrazi Hippogriff", "{7}", wretchedGryffOracle)
	spawn := apaPush(g2, me2.ID, me2.ID, EldraziSpawnToken())
	emergeLands(g2, me2.ID, 5, "Island")
	if err := castEmerge(g2, me2, gryff2, spawn, true); err == nil {
		t.Fatal("the Spawn named to emerge was cracked for the {5}{U}'s sixth mana")
	}
	if findBattlefieldCardForTest(g2, spawn) == nil {
		t.Fatal("the refused cast spent the Spawn")
	}
}

// The offer's card component is a creature (or, for Crabomination, an
// artifact) you control, exactly one.
func TestEmergeRefusesTheWrongPayment(t *testing.T) {
	g, me, opp := emergeTable(t)
	gryff := handCardOf(g, me, "Wretched Gryff", "Creature — Eldrazi Hippogriff", "{7}", wretchedGryffOracle)
	crab := handCardOf(g, me, "Crabomination", "Creature — Crab Demon", "{4}{B}{B}", crabominationOracle)
	theirs := emergeFodder(g, opp.ID, "Theirs", "{3}", 3, 3)
	mine := emergeFodder(g, me.ID, "Mine", "{3}", 3, 3)
	rock := apaPush(g, me.ID, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{5}"})
	land := pushLandFor(g, me.ID, "Island", "Basic Land — Island")
	for name, c := range map[string]struct{ card, sac uuid.UUID }{
		"an opponent's creature":       {gryff, theirs},
		"a land":                       {gryff, land},
		"a noncreature artifact":       {gryff, rock},
		"a creature for Crabomination": {crab, mine},
	} {
		if err := castEmerge(g, me, c.card, c.sac, false); err == nil {
			t.Errorf("%s paid emerge", name)
		}
	}
	if err := g.CastSpell(me.ID, gryff, game.CastSpellParams{AlternativeCost: AltCostKeyEmerge}); err == nil {
		t.Error("emerge was paid sacrificing nothing")
	}
	if err := g.CastSpell(me.ID, gryff, game.CastSpellParams{AlternativeCost: AltCostKeyEmerge, AltCostIDs: []uuid.UUID{mine, rock}}); err == nil {
		t.Error("emerge was paid sacrificing two permanents")
	}
	if got := emergePrice(t, g, me, crab, rock); got != "{B}{B}" {
		t.Errorf("Crabomination over a five-mana artifact: %s, want {B}{B}", got)
	}
}

// emergeResolve casts `card` emerged over a fresh vanilla creature in
// permissive mode, answers its "you may" and target prompts with `pick`,
// and lets everything resolve.
func emergeResolve(t *testing.T, g *game.Game, me *game.Player, card uuid.UUID, targets []game.TargetRef, pick []game.TargetRef) {
	t.Helper()
	fodder := emergeFodder(g, me.ID, "Fodder", "{3}", 1, 1)
	if err := castEmerge(g, me, card, fodder, false, targets...); err != nil {
		t.Fatalf("emerge: %v", err)
	}
	for i := 0; i < 4; i++ {
		answered := false
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == me.ID {
				answerLatestTriggerPrompt(t, g, me.ID, true)
				answered = true
				break
			}
		}
		if p := latestPickTarget(g, me.ID); p != nil && len(pick) > 0 {
			if err := g.ResolvePickTargets(p.ID, me.ID, pick); err != nil {
				t.Fatalf("pick targets: %v", err)
			}
			answered = true
		}
		if !answered {
			break
		}
	}
	passPriorityAroundTable(t, g)
}

// Every cast trigger does its printed thing once the emerge cost is paid.
func TestEmergeCastTriggersDoTheirThing(t *testing.T) {
	ref := func(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }
	t.Run("Abundant Maw", func(t *testing.T) {
		g, me, opp := emergeTable(t)
		maw := handCardOf(g, me, "Abundant Maw", "Creature — Eldrazi Leech", "{8}", abundantMawOracle)
		mine, theirs := me.Life, opp.Life
		emergeResolve(t, g, me, maw, nil, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
		if opp.Life != theirs-3 || me.Life != mine+3 {
			t.Errorf("life me %d→%d, opponent %d→%d; want +3 and -3", mine, me.Life, theirs, opp.Life)
		}
	})
	t.Run("It of the Horrid Swarm", func(t *testing.T) {
		g, me, _ := emergeTable(t)
		it := handCardOf(g, me, "It of the Horrid Swarm", "Creature — Eldrazi Insect", "{8}", itOfTheHorridSwarmOracle)
		emergeResolve(t, g, me, it, nil, nil)
		if n, _ := countTokensNamed(g, me.ID, "Insect"); n != 2 {
			t.Errorf("%d Insect tokens, want 2", n)
		}
	})
	t.Run("Wretched Gryff", func(t *testing.T) {
		g, me, _ := emergeTable(t)
		gryff := handCardOf(g, me, "Wretched Gryff", "Creature — Eldrazi Hippogriff", "{7}", wretchedGryffOracle)
		emergeResolve(t, g, me, gryff, nil, nil)
		if me.Hand.Size() != 1 {
			t.Errorf("hand %d after the Gryff, want the one card drawn", me.Hand.Size())
		}
	})
	t.Run("Decimator of the Provinces", func(t *testing.T) {
		g, me, _ := emergeTable(t)
		bear := emergeFodder(g, me.ID, "Bear", "{2}", 2, 2)
		dec := handCardOf(g, me, "Decimator of the Provinces", "Creature — Eldrazi Boar", "{10}", decimatorOracle)
		g.WithWriteLock(func() {
			for i := range me.Hand.Cards {
				if me.Hand.Cards[i].InstanceID == dec {
					me.Hand.Cards[i].Power, me.Hand.Cards[i].Toughness = 7, 7
				}
			}
		})
		emergeResolve(t, g, me, dec, nil, nil)
		g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
		c := findBattlefieldCardForTest(g, bear)
		if c == nil || c.CurrentPower() != 4 || c.CurrentToughness() != 4 || !game.HasKeyword(c, "trample") {
			t.Errorf("the Bear is not a 4/4 trampler: %+v", c)
		}
		if d := findBattlefieldCardForTest(g, dec); d == nil || d.CurrentPower() != 7 {
			t.Error("the Decimator pumped itself from the stack")
		}
	})
	t.Run("Twisted Riddlekeeper", func(t *testing.T) {
		g, me, opp := emergeTable(t)
		a := pr7Creature(g, opp.ID, "A", 2, "G")
		b := pr7Creature(g, opp.ID, "B", 2, "G")
		riddle := handCardOf(g, me, "Twisted Riddlekeeper", "Creature — Eldrazi Sphinx", "{8}", twistedRiddlekeeperOrcl)
		emergeResolve(t, g, me, riddle, nil, []game.TargetRef{ref(a), ref(b)})
		for _, id := range []uuid.UUID{a, b} {
			c := findBattlefieldCardForTest(g, id)
			if c == nil || !c.Tapped || c.Counters[game.CounterStun] != 1 {
				t.Errorf("a target is not tapped with one stun counter: %+v", c)
			}
		}
	})
	t.Run("Mockery of Nature", func(t *testing.T) {
		g, me, opp := emergeTable(t)
		rock := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{2}"})
		mock := handCardOf(g, me, "Mockery of Nature", "Creature — Eldrazi Beast", "{9}", mockeryOfNatureOracle)
		emergeResolve(t, g, me, mock, nil, []game.TargetRef{ref(rock)})
		if findBattlefieldCardForTest(g, rock) != nil {
			t.Error("the artifact was not destroyed")
		}
	})
	t.Run("Lashweed Lurker", func(t *testing.T) {
		g, me, opp := emergeTable(t)
		theirs := pr7Creature(g, opp.ID, "Theirs", 2, "G")
		lash := handCardOf(g, me, "Lashweed Lurker", "Creature — Eldrazi Horror", "{8}", lashweedLurkerOracle)
		emergeResolve(t, g, me, lash, nil, []game.TargetRef{ref(theirs)})
		top, err := opp.Library.Top()
		if err != nil || top.InstanceID != theirs {
			t.Error("the permanent is not on top of its owner's library")
		}
	})
	t.Run("Vexing Scuttler", func(t *testing.T) {
		g, me, _ := emergeTable(t)
		bolt := pushGraveyardPermanent(me, "Bolt", "Instant", "{R}")
		vex := handCardOf(g, me, "Vexing Scuttler", "Creature — Eldrazi Crab", "{8}", vexingScuttlerOracle)
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
	g, me, _ := emergeTable(t)
	fodder := emergeFodder(g, me.ID, "Wall", "{3}", 0, 4)
	g.WithWriteLock(func() {
		c := findBattlefieldCardForTest(g, fodder)
		c.Counters = map[string]int{"+1/+1": 1}
		g.RecomputeLayersIfStaleLocked()
	})
	off := handCardOf(g, me, "Adipose Offspring", "Creature — Alien", "{3}{W}", adiposeOffspringOracle)
	if err := castEmerge(g, me, off, fodder, false); err != nil {
		t.Fatalf("emerge: %v", err)
	}
	// Resolve the spell, then bounce the Offspring with its trigger waiting.
	for i := 0; i < 8 && findBattlefieldCardForTest(g, off) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if findBattlefieldCardForTest(g, off) == nil {
		t.Fatal("the Offspring did not resolve")
	}
	g.WithWriteLock(func() { g.BounceCardsToHandForEffect([]uuid.UUID{off}) })
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Alien"); n != 5 {
		t.Errorf("%d Aliens, want 5 (the sacrificed creature's toughness, counter included)", n)
	}

	g2, me2, _ := emergeTable(t)
	off2 := handCardOf(g2, me2, "Adipose Offspring", "Creature — Alien", "{3}{W}", adiposeOffspringOracle)
	if err := g2.CastSpell(me2.ID, off2, game.CastSpellParams{}); err != nil {
		t.Fatalf("hard cast: %v", err)
	}
	passPriorityAroundTable(t, g2)
	if n, _ := countTokensNamed(g2, me2.ID, "Alien"); n != 1 {
		t.Errorf("%d Aliens for a hard-cast Offspring, want 1", n)
	}
}

// Cresting Mosasaurus, cast: every creature but Dinosaurs goes home.
func TestCrestingMosasaurusBouncesEveryNonDinosaur(t *testing.T) {
	g, me, opp := emergeTable(t)
	theirs := pr7Creature(g, opp.ID, "Theirs", 2, "G")
	dino := apaPush(g, me.ID, me.ID, game.Card{Name: "Raptor", TypeLine: "Creature — Dinosaur", ManaCost: "{2}", Power: 2, Toughness: 2})
	mosa := handCardOf(g, me, "Cresting Mosasaurus", "Creature — Dinosaur", "{6}{U}{U}", crestingMosasaurusOracle)
	emergeResolve(t, g, me, mosa, nil, nil)
	if !opp.Hand.Contains(theirs) {
		t.Error("the opponent's creature was not returned")
	}
	if findBattlefieldCardForTest(g, dino) == nil || findBattlefieldCardForTest(g, mosa) == nil {
		t.Error("a Dinosaur was returned")
	}
}

// Drownyard Behemoth has hexproof the turn it enters and not after.
func TestDrownyardBehemothHasHexproofOnlyTheTurnItEntered(t *testing.T) {
	g, me, _ := emergeTable(t)
	beh := handCardOf(g, me, "Drownyard Behemoth", "Creature — Eldrazi Crab", "{9}", drownyardBehemothOracle)
	emergeResolve(t, g, me, beh, nil, nil)
	hexproof := func() bool {
		var has bool
		g.WithWriteLock(func() {
			g.RecomputeLayersIfStaleLocked()
			if c := findBattlefieldCardForTest(g, beh); c != nil {
				has = game.HasKeyword(c, "hexproof")
			}
		})
		return has
	}
	if !hexproof() {
		t.Fatal("no hexproof the turn it entered")
	}
	b39NextTurnOf(t, g, 0)
	if findBattlefieldCardForTest(g, beh) == nil {
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
	g, me, opp := emergeTable(t)
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
	rock := apaPush(g, me.ID, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{5}"})
	crab := handCardOf(g, me, "Crabomination", "Creature — Crab Demon", "{4}{B}{B}", crabominationOracle)
	if err := castEmerge(g, me, crab, rock, false); err != nil {
		t.Fatalf("emerge from artifact: %v", err)
	}
	for i := 0; i < 8 && findBattlefieldCardForTest(g, crab) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
			t.Fatalf("pick the opponent: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
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
	g, me, _ := emergeTable(t)
	fiend := handCardOf(g, me, "Elder Deep-Fiend", "Creature — Eldrazi Octopus", "{8}", elderDeepFiendOracle)
	four := emergeFodder(g, me.ID, "Four", "{4}", 4, 4)
	spawn := apaPush(g, me.ID, me.ID, EldraziSpawnToken())
	var offer *protocol.AlternativeCostView
	v := protocol.ViewOfGameFor(g, me.ID.String())
	for _, s := range v.Seats {
		for i := range s.Hand.Cards {
			if s.Hand.Cards[i].InstanceID != fiend.String() {
				continue
			}
			for j := range s.Hand.Cards[i].AlternativeCosts {
				if s.Hand.Cards[i].AlternativeCosts[j].Key == AltCostKeyEmerge {
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
	g, me, _ := emergeTable(t)
	gryff := handCardOf(g, me, "Wretched Gryff", "Creature — Eldrazi Hippogriff", "{7}", wretchedGryffOracle)
	emergeLands(g, me.ID, 1, "Island")
	two := emergeFodder(g, me.ID, "Two", "{2}", 2, 2)
	five := emergeFodder(g, me.ID, "Five", "{5}", 5, 5)
	six := emergeFodder(g, me.ID, "Six", "{6}", 6, 6)
	spawn := apaPush(g, me.ID, me.ID, EldraziSpawnToken())
	offered := map[uuid.UUID]bool{}
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type != legal.TypeCastSpell || m.Source != gryff {
			continue
		}
		var p struct {
			AlternativeCost string      `json:"alternative_cost"`
			AltCostIDs      []uuid.UUID `json:"alt_cost_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil || p.AlternativeCost != AltCostKeyEmerge || len(p.AltCostIDs) != 1 {
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
	two := Emerge("{4}")
	two.Sacrifice = two.Sacrifice.WithCount(2, 2)
	none := Overload("{4}")
	none.ReducedBySacrificedManaValue = true
	for name, ac := range map[string]game.AlternativeCost{"two": two, "none": none} {
		msg := registerPanics(Spec{OracleID: "test-0135-emerge-" + name, Name: "Test " + name,
			AlternativeCosts: []game.AlternativeCost{ac}})
		if !strings.Contains(msg, "CR 702.119a") {
			t.Errorf("%s: panic %q, want the emerge guard", name, msg)
		}
	}
}
