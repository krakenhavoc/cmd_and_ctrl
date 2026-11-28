package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	b843RimeDryad    = "0464176d-75d2-403a-aa9d-40e85f106d01"
	b843Hipparion    = "306fd9d2-5ded-4857-90f4-e62d73c6aed2"
	b843Shyft        = "40c3007a-be6c-4400-9139-12b9ca098ac4"
	b843StoneSpirit  = "63bc0a02-0d6c-4a00-8b17-1b024f830219"
	b843Legions      = "70b69803-1870-4441-954b-dba5b433fcb2"
	b843DireWolves   = "7d72a2f4-cce6-4a2b-9507-0ee1c3e1e4ae"
	b843ArcticFoxes  = "898f0824-cf4c-4e34-8407-9a32fdae6dcb"
	b843Vampire      = "8fa93c01-f9ea-4d48-a040-418d1ca6be19"
	b843Rats         = "bc6449c8-f0cd-4c07-8acf-436f97e8bbb2"
	b843Conscripts   = "d99e1d08-32ca-49c5-97cc-e064d661202f"
	b843SnowForest   = "Basic Snow Land — Forest"
	b843PlainForest  = "Basic Land — Forest"
	b843SnowSwamp    = "Basic Snow Land — Swamp"
	b843PlainSwamp   = "Basic Land — Swamp"
	b843SnowIslandTL = "Basic Snow Land — Island"
)

// Snow landwalk: only a snow land of the right type removes the blockers.
func TestB843SnowLandwalkNeedsASnowLand(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, snow, plain string
	}{
		{"Rime Dryad", b843RimeDryad, b843SnowForest, b843PlainForest},
		{"Legions of Lim-Dûl", b843Legions, b843SnowSwamp, b843PlainSwamp},
	} {
		for _, snow := range []bool{true, false} {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me, them := effectstest.SeatsForCombat(g)
			att := effectstest.B12Push(g, me.ID, tc.name, "Creature — Dryad", tc.oracle, 1, 2)
			blocker := effectstest.B12Creature(g, them.ID, "Blocker", "Creature — Bear", 2, 2)
			typeLine := tc.plain
			if snow {
				typeLine = tc.snow
			}
			effectstest.PushLandFor(g, them.ID, "Land", typeLine)
			effectstest.AttackIntoBlocks(t, g, them.ID, att)
			refused := effectstest.BlockRefusal(t, g, blocker, att) != ""
			if refused != snow {
				t.Errorf("%s vs snow=%v: block refused=%v", tc.name, snow, refused)
			}
		}
	}
}

// Hipparion: blocking a creature with power 3 or more costs {1}, paid as the block is finished; a smaller one is free.
func TestB843HipparionChargesToBlockBigCreatures(t *testing.T) {
	for _, power := range []int{2, 3} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, them := effectstest.SeatsForCombat(g)
		att := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Ogre", power, 3)
		hip := effectstest.B12Push(g, them.ID, "Hipparion", "Creature — Horse", b843Hipparion, 1, 3)
		effectstest.AttackIntoBlocks(t, g, them.ID, att)
		if err := g.DeclareBlocker(hip, att); err != nil {
			t.Fatalf("power %d: staging the block: %v", power, err)
		}
		cost := g.PriceBlockDeclaration(them.ID).Cost
		if power == 2 {
			if cost != "" {
				t.Errorf("blocking a 2-power creature costs %q, want nothing", cost)
			}
			if err := g.FinishBlocks(them.ID); err != nil {
				t.Errorf("finishing the free block: %v", err)
			}
			continue
		}
		if cost != "{1}" {
			t.Errorf("blocking a 3-power creature costs %q, want {1}", cost)
		}
		if err := g.FinishBlocks(them.ID); !errors.Is(err, game.ErrBlockTaxUnpaid) {
			t.Fatalf("finishing without {1}: %v, want ErrBlockTaxUnpaid", err)
		}
		effectstest.FloatForTest(g, them, "C")
		if err := g.FinishBlocks(them.ID); err != nil {
			t.Errorf("finishing with {1}: %v", err)
		}
	}
}

// Stone Spirit can't be blocked by creatures with flying.
func TestB843StoneSpiritIsNotBlockedByFlyers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := effectstest.SeatsForCombat(g)
	att := effectstest.B12Push(g, me.ID, "Stone Spirit", "Creature — Elemental Spirit", b843StoneSpirit, 4, 3)
	flyer := effectstest.ApaPush(g, them.ID, them.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	walker := effectstest.B12Creature(g, them.ID, "Walker", "Creature — Bear", 2, 2)
	effectstest.AttackIntoBlocks(t, g, them.ID, att)
	if effectstest.BlockRefusal(t, g, flyer, att) == "" {
		t.Error("a flying creature blocked the Stone Spirit")
	}
	if r := effectstest.BlockRefusal(t, g, walker, att); r != "" {
		t.Errorf("a ground creature was refused: %v", r)
	}
}

// Dire Wolves have banding only while their controller controls a Plains.
func TestB843DireWolvesNeedAPlainsForBanding(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wolves := effectstest.B12Push(g, me.ID, "Dire Wolves", "Creature — Wolf", b843DireWolves, 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, wolves, game.KeywordBanding) {
		t.Fatal("banding without a Plains")
	}
	plains := effectstest.B12Push(g, me.ID, "Plains", "Basic Land — Plains", "", 0, 0)
	if !effectstest.HasEffectiveKeyword(t, g, wolves, game.KeywordBanding) {
		t.Error("no banding with a Plains")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentsForEffect([]uuid.UUID{plains}) })
	if effectstest.HasEffectiveKeyword(t, g, wolves, game.KeywordBanding) {
		t.Error("banding stayed after the Plains left")
	}
}

// Arctic Foxes can't be blocked by creatures with power 2 or more while the defender controls a snow land.
func TestB843ArcticFoxesNeedASnowLandForTheRestriction(t *testing.T) {
	for _, snow := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, them := effectstest.SeatsForCombat(g)
		att := effectstest.B12Push(g, me.ID, "Arctic Foxes", "Creature — Fox", b843ArcticFoxes, 1, 1)
		big := effectstest.B12Creature(g, them.ID, "Big", "Creature — Ogre", 2, 2)
		small := effectstest.B12Creature(g, them.ID, "Small", "Creature — Rat", 1, 1)
		if snow {
			effectstest.PushLandFor(g, them.ID, "Snow-Covered Island", b843SnowIslandTL)
		}
		effectstest.AttackIntoBlocks(t, g, them.ID, att)
		if refused := effectstest.BlockRefusal(t, g, big, att) != ""; refused != snow {
			t.Errorf("snow=%v: power-2 blocker refused=%v", snow, refused)
		}
		if effectstest.BlockRefusal(t, g, small, att) != "" {
			t.Errorf("snow=%v: a power-1 blocker was refused", snow)
		}
	}
}

// Pestilence Rats: power is the number of other Rats on the battlefield, toughness 3.
func TestB843PestilenceRatsCountsOtherRats(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	a := effectstest.B12Push(g, g.Seats[0].ID, "Pestilence Rats", "Creature — Rat", b843Rats, 0, 3)
	if p, tough := effectstest.PtOf(t, g, a); p != 0 || tough != 3 {
		t.Errorf("alone: %d/%d, want 0/3", p, tough)
	}
	effectstest.B12Push(g, g.Seats[0].ID, "Pestilence Rats", "Creature — Rat", b843Rats, 0, 3)
	effectstest.B12Creature(g, g.Seats[1].ID, "Their Rat", "Creature — Rat", 1, 1)
	effectstest.B12Creature(g, g.Seats[1].ID, "Not a rat", "Creature — Bear", 1, 1)
	if p, tough := effectstest.PtOf(t, g, a); p != 2 || tough != 3 {
		t.Errorf("with two other Rats: %d/%d, want 2/3", p, tough)
	}
}

// Orcish Conscripts can't attack alone or with fewer than two others.
func TestB843OrcishConscriptsNeedTwoOtherAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	orc := effectstest.B12Push(g, me.ID, "Orcish Conscripts", "Creature — Orc", b843Conscripts, 2, 2)
	b1 := effectstest.B12Creature(g, me.ID, "B1", "Creature — Bear", 2, 2)
	b2 := effectstest.B12Creature(g, me.ID, "B2", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{orc, b1} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("staging: %v", err)
		}
	}
	if err := g.PassPriority(); err == nil {
		t.Fatal("the Conscripts attacked with only one other creature")
	}
	if err := g.DeclareAttacker(b2, opp.ID); err != nil {
		t.Fatalf("third attacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	if c := effectstest.FindBattlefieldCardForTest(g, orc); c == nil || c.AttackingTarget == uuid.Nil {
		t.Error("the Conscripts are not attacking with two others")
	}
}

func TestB843OrcishConscriptsNeedTwoOtherBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := effectstest.SeatsForCombat(g)
	a1 := effectstest.B12Creature(g, me.ID, "A1", "Creature — Bear", 1, 1)
	a2 := effectstest.B12Creature(g, me.ID, "A2", "Creature — Bear", 1, 1)
	a3 := effectstest.B12Creature(g, me.ID, "A3", "Creature — Bear", 1, 1)
	orc := effectstest.B12Push(g, them.ID, "Orcish Conscripts", "Creature — Orc", b843Conscripts, 2, 2)
	o1 := effectstest.B12Creature(g, them.ID, "O1", "Creature — Bear", 2, 2)
	o2 := effectstest.B12Creature(g, them.ID, "O2", "Creature — Bear", 2, 2)
	effectstest.AttackIntoBlocks(t, g, them.ID, a1, a2, a3)
	if err := g.DeclareBlocker(orc, a1); err == nil {
		t.Fatal("the Conscripts blocked alone")
	}
	if err := g.DeclareBlocker(o1, a2); err != nil {
		t.Fatalf("a companion block: %v", err)
	}
	if err := g.DeclareBlocker(orc, a1); err == nil {
		t.Fatal("the Conscripts blocked with only one other blocker")
	}
	if err := g.DeclareBlocker(o2, a3); err != nil {
		t.Fatalf("a second companion block: %v", err)
	}
	if err := g.DeclareBlocker(orc, a1); err != nil {
		t.Errorf("the Conscripts were refused with two other blockers: %v", err)
	}
}

// Shyft: at your upkeep you may pick new colors, which stay.
func TestB843ShyftBecomesTheChosenColorsAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	shyft := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Shyft", TypeLine: "Creature — Shapeshifter", OracleID: b843Shyft, Power: 4, Toughness: 2, Colors: []string{"U"}})
	effectstest.AdvanceTo(t, g, game.StepUpkeep)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerColor(t, g, me.ID, "RG")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB9Colors(t, g, shyft); got != "RG" {
		t.Errorf("colors %q, want RG", got)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	if got := effectstest.TmtB9Colors(t, g, shyft); got != "RG" {
		t.Errorf("the change did not last past the turn: %q", got)
	}
}

// Krovikan Vampire: a creature it damaged that died comes back under its controller's control at the end step,
// and is sacrificed when the Vampire leaves.
func TestB843KrovikanVampireReturnsWhatItKilledAndSacrificesItWhenLost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := effectstest.SeatsForCombat(g)
	vamp := effectstest.B12Push(g, me.ID, "Krovikan Vampire", "Creature — Vampire", b843Vampire, 3, 3)
	bear := effectstest.B12Creature(g, them.ID, "Their Bear", "Creature — Bear", 2, 2)
	bystander := effectstest.B12Creature(g, them.ID, "Bystander", "Creature — Bear", 1, 1)
	effectstest.Pr7aBlockedCombat(t, g, them.ID, vamp, bear)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, bear)
	if !ok || c.Controller != me.ID {
		t.Fatalf("the dead Bear is not on the battlefield under my control: ok=%v", ok)
	}
	if !g.Battlefield.Contains(bystander) {
		t.Error("an undamaged creature was disturbed")
	}
	// Lose the Vampire at instant speed: a Lightning Bolt in the end step.
	bolt := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: effectstest.LightningBoltOracle, Owner: me.ID, Controller: me.ID, Colors: []string{"R"}})
	effectstest.FloatForTest(g, me, "R")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: effectstest.CardRefs(vamp)}); err != nil {
		t.Fatalf("Bolt: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the returned Bear was not sacrificed when the Vampire left")
	}
}

func TestB843KrovikanVampireDoesNothingWithoutAVictim(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := effectstest.SeatsForCombat(g)
	effectstest.B12Push(g, me.ID, "Krovikan Vampire", "Creature — Vampire", b843Vampire, 3, 3)
	effectstest.B12Creature(g, them.ID, "Their Bear", "Creature — Bear", 2, 2)
	before := len(g.Battlefield.Cards)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.Battlefield.Cards) != before {
		t.Errorf("battlefield %d -> %d with nothing to return", before, len(g.Battlefield.Cards))
	}
}
