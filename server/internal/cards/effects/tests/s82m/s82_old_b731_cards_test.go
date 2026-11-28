package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b731_cards_test.go — Urza's Saga, batch 731 (#5624). Each test plays
// the real catalog entry and asserts what its text says.

const (
	b731SeedlingsOracle = "87285539-29c8-4cc7-a22c-4a4b96a61e82"
	b731PetOracle       = "afee3422-5490-411b-a36e-82e447c1bd0f"
	b731WightOracle     = "dbde53c1-c6b7-4775-87d3-e95166392e60"
	b731RangersOracle   = "e4606bc2-faaa-4dd8-9109-55ce3440cf5b"
	b731WurmOracle      = "ee45c80c-699c-4e37-97e6-1859c6601c99"
	b731RazeOracle      = "02fde433-955d-49df-bd72-270b6feac9e7"
	b731PitTrapOracle   = "671c4677-fe0a-4bda-9674-f4dcd9069bfd"
	b731CatOracle       = "ed68160e-5e3d-44f2-ac7a-3749fda10119"
	b731SanctumOracle   = "34187c71-6033-4058-aadc-2bc266f762be"
	b731QuarryOracle    = "57b4da3f-361a-4cbe-b77f-190ec33eefd8"
	b731AcademyOracle   = "dba4fd31-8931-42dd-bd86-45479c2abf74"
	b731CreatureLine    = "Creature — Test"
)

func TestOldB731EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		b731SeedlingsOracle: "Treefolk Seedlings",
		b731PetOracle:       "Imaginary Pet",
		b731WightOracle:     "Reclusive Wight",
		b731RangersOracle:   "Treetop Rangers",
		b731WurmOracle:      "Scoria Wurm",
		b731RazeOracle:      "Raze",
		b731PitTrapOracle:   "Pit Trap",
		b731CatOracle:       "Catastrophe",
		b731SanctumOracle:   "Serra's Sanctum",
		b731QuarryOracle:    "Thran Quarry",
		b731AcademyOracle:   "Tolarian Academy",
	}
	for oracle, name := range want {
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

func b731HandCount(p *game.Player, name string) int {
	n := 0
	for _, c := range p.Hand.Cards {
		if c.Name == name {
			n++
		}
	}
	return n
}

func b731OnBattlefield(g *game.Game, id uuid.UUID) bool {
	return effectstest.FindBattlefieldCardForTest(g, id) != nil
}

func b731Cast(t *testing.T, g *game.Game, name, typeLine, oracle string, p game.CastSpellParams) error {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: active.ID, Controller: active.ID})
	return g.CastSpell(active.ID, id, p)
}

func TestOldB731TreefolkSeedlingsToughnessCountsYourForests(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedlings := effectstest.B12Push(g, me.ID, "Treefolk Seedlings", "Creature — Treefolk", b731SeedlingsOracle, 2, 5)
	effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	if p, tough := effectstest.EffectivePTOf(t, g, seedlings); p != 2 || tough != 1 {
		t.Errorf("with one Forest it is %d/%d, want 2/1", p, tough)
	}
	effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	if _, tough := effectstest.EffectivePTOf(t, g, seedlings); tough != 2 {
		t.Errorf("with two of your Forests (an Island and an opposing Forest don't count) toughness is %d, want 2", tough)
	}
}

func TestOldB731TreetopRangersCanOnlyBeBlockedByFlyers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rangers := effectstest.PushCatalogPermanent(g, me.ID, "Treetop Rangers", "Creature — Elf Ranger", b731RangersOracle, false)
	wall := effectstest.ReqCreature(g, opp.ID, "Wall", "", 0, 4)
	bird := effectstest.ReqCreature(g, opp.ID, "Bird", "", 1, 1, "flying")
	reach := effectstest.ReqCreature(g, opp.ID, "Spider", "", 1, 3, "reach")
	effectstest.BrAttack(t, g, rangers)
	if err := effectstest.ReqBlock(t, g, rangers, wall); err == nil {
		t.Error("a creature without flying blocked the Rangers")
	}
	if err := effectstest.ReqBlock(t, g, rangers, reach); err == nil {
		t.Error("a creature with only reach blocked the Rangers")
	}
	if err := effectstest.ReqBlock(t, g, rangers, bird); err != nil {
		t.Errorf("a flyer could not block the Rangers: %v", err)
	}
}

func TestOldB731ImaginaryPetReturnsToHandOnlyWithACardInHand(t *testing.T) {
	for _, hasCard := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		pet := effectstest.PushCatalogPermanent(g, me.ID, "Imaginary Pet", "Creature — Illusion", b731PetOracle, false)
		me.Hand.Cards = nil
		g.Seats[1].Hand.Cards = nil
		if hasCard {
			effectstest.HandCardForTest(me, "Card", "Basic Land — Forest", "")
		}
		effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
		me.Hand.Cards = me.Hand.Cards[:len(me.Hand.Cards)]
		effectstest.PassPriorityAroundTable(t, g)
		if hasCard == b731OnBattlefield(g, pet) {
			t.Errorf("hasCard=%v: Imaginary Pet on the battlefield = %v", hasCard, b731OnBattlefield(g, pet))
		}
		if hasCard && b731HandCount(me, "Imaginary Pet") != 1 {
			t.Error("the Pet is not in its owner's hand")
		}
	}
}

func TestOldB731ReclusiveWightSacrificesItselfOnlyWithAnotherNonlandPermanent(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(g *game.Game, me *game.Player)
		dies  bool
	}{
		{"alone", func(*game.Game, *game.Player) {}, false},
		{"with a land", func(g *game.Game, me *game.Player) { effectstest.AangPushLand(g, me.ID, "Forest", false) }, false},
		{"with an opponent's permanent", func(g *game.Game, _ *game.Player) {
			effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", b731CreatureLine, 1, 1)
		}, false},
		{"with another creature", func(g *game.Game, me *game.Player) { effectstest.B12Creature(g, me.ID, "Mine", b731CreatureLine, 1, 1) }, true},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		wight := effectstest.PushCatalogPermanent(g, me.ID, "Reclusive Wight", "Creature — Zombie Minion", b731WightOracle, false)
		tc.setup(g, me)
		effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
		effectstest.PassPriorityAroundTable(t, g)
		if b731OnBattlefield(g, wight) == tc.dies {
			t.Errorf("%s: Wight on the battlefield = %v, want %v", tc.name, !tc.dies, !tc.dies)
		}
	}
}

func TestOldB731ScoriaWurmReturnsToHandOnALostFlip(t *testing.T) {
	sawKeep, sawLose := false, false
	for i := 0; i < 80 && !(sawKeep && sawLose); i++ {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		wurm := effectstest.PushCatalogPermanent(g, me.ID, "Scoria Wurm", "Creature — Wurm", b731WurmOracle, false)
		effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
		effectstest.PassPriorityAroundTable(t, g)
		flip := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
		if flip == nil {
			t.Fatal("no coin call at the upkeep")
		}
		if err := g.ResolveCoinCall(flip.ID, me.ID, "heads"); err != nil {
			t.Fatalf("coin: %v", err)
		}
		flips := effectstest.RandomEvents(g, game.EventFlipCoin)
		won := len(flips) > 0 && flips[len(flips)-1].Won
		effectstest.PassPriorityAroundTable(t, g)
		if won {
			sawKeep = true
			if !b731OnBattlefield(g, wurm) {
				t.Error("won the flip but the Wurm returned")
			}
		} else {
			sawLose = true
			if b731OnBattlefield(g, wurm) || b731HandCount(me, "Scoria Wurm") != 1 {
				t.Error("lost the flip but the Wurm did not return to its owner's hand")
			}
		}
	}
	if !sawKeep || !sawLose {
		t.Fatalf("never saw both outcomes (keep %v, lose %v)", sawKeep, sawLose)
	}
}

func TestOldB731RazeSacrificesALandAndDestroysAnother(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := effectstest.AangPushLand(g, me.ID, "Forest", false)
	theirs := effectstest.AangPushLand(g, opp.ID, "Island", false)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := b731Cast(t, g, "Raze", "Sorcery", b731RazeOracle, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}}); err == nil {
		t.Fatal("cast with no land chosen to sacrifice")
	}
	if err := b731Cast(t, g, "Raze", "Sorcery", b731RazeOracle, game.CastSpellParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if b731OnBattlefield(g, fodder) || b731OnBattlefield(g, theirs) {
		t.Error("the sacrificed land or the target survived")
	}
}

func TestOldB731PitTrapDestroysAnAttackingCreatureWithoutFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trap := effectstest.PushCatalogPermanent(g, opp.ID, "Pit Trap", "Artifact", b731PitTrapOracle, false)
	ogre := effectstest.ReqCreature(g, me.ID, "Ogre", "", 3, 3)
	bird := effectstest.ReqCreature(g, me.ID, "Bird", "", 1, 1, "flying")
	effectstest.DeclareAttack(t, g, opp.ID, ogre, bird)
	effectstest.FloatForTest(g, opp, "CC")
	if err := g.ActivateCatalogAbility(opp.ID, trap, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bird}}}); err == nil {
		t.Fatal("a flyer was accepted as the target")
	}
	effectstest.FloatForTest(g, opp, "CC")
	if err := g.ActivateCatalogAbility(opp.ID, trap, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ogre}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if b731OnBattlefield(g, trap) {
		t.Error("the Trap was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if b731OnBattlefield(g, ogre) {
		t.Error("the attacking Ogre survived")
	}
	if !b731OnBattlefield(g, bird) {
		t.Error("the flyer was destroyed")
	}
}

func TestOldB731PitTrapCantTargetAnUnattackingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trap := effectstest.PushCatalogPermanent(g, opp.ID, "Pit Trap", "Artifact", b731PitTrapOracle, false)
	idle := effectstest.ReqCreature(g, me.ID, "Idle", "", 3, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	effectstest.FloatForTest(g, opp, "CC")
	if err := g.ActivateCatalogAbility(opp.ID, trap, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: idle}}}); err == nil {
		t.Error("a creature that is not attacking was accepted")
	}
}

func TestOldB731CatastropheDestroysAllLandsOrAllCreatures(t *testing.T) {
	for _, lands := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		land := effectstest.AangPushLand(g, me.ID, "Forest", false)
		theirLand := effectstest.AangPushLand(g, opp.ID, "Island", false)
		mine := effectstest.B12Creature(g, me.ID, "Mine", b731CreatureLine, 2, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Theirs", b731CreatureLine, 2, 2)
		effectstest.CastCatalogSpell(t, g, "Catastrophe", "Sorcery", b731CatOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		pick := effectstest.LatestChoiceOfKind(g, game.PendingChoiceOptionPick)
		if pick == nil {
			t.Fatal("no choice between lands and creatures")
		}
		idx := 1
		if lands {
			idx = 0
		}
		if err := g.ResolveOptionPick(pick.ID, me.ID, idx); err != nil {
			t.Fatalf("answer: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		for _, id := range []uuid.UUID{land, theirLand} {
			if b731OnBattlefield(g, id) == lands {
				t.Errorf("lands=%v: a land's presence is %v", lands, !lands)
			}
		}
		for _, id := range []uuid.UUID{mine, theirs} {
			if b731OnBattlefield(g, id) != lands {
				t.Errorf("lands=%v: a creature's presence is %v", lands, !lands)
			}
		}
	}
}

func TestOldB731CatastropheCreaturesCantRegenerate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B12Creature(g, me.ID, "Regenerator", b731CreatureLine, 2, 2)
	effectstest.P1Regenerate(t, g, mine)
	effectstest.CastCatalogSpell(t, g, "Catastrophe", "Sorcery", b731CatOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKind(g, game.PendingChoiceOptionPick)
	if pick == nil {
		t.Fatal("no choice")
	}
	if err := g.ResolveOptionPick(pick.ID, me.ID, 1); err != nil {
		t.Fatalf("answer: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if b731OnBattlefield(g, mine) {
		t.Error("a creature with a regeneration shield survived Catastrophe")
	}
}

func b731ManaAmount(t *testing.T, g *game.Game, oracle string, controller, source uuid.UUID) int {
	t.Helper()
	spec, _ := effects.Lookup(oracle)
	for _, ab := range spec.ManaAbilities {
		if ab.ProducedAmount != nil {
			var amount game.ManaAmount
			g.WithWriteLock(func() { amount = ab.ProducedAmount(g, controller, source) })
			return amount.N
		}
	}
	t.Fatal("no scaled mana ability")
	return 0
}

func TestOldB731SerrasSanctumAddsWhiteForEachEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sanctum := effectstest.PushCatalogPermanent(g, me.ID, "Serra's Sanctum", "Legendary Land", b731SanctumOracle, false)
	if n := b731ManaAmount(t, g, b731SanctumOracle, me.ID, sanctum); n != 0 {
		t.Errorf("with no enchantment it adds %d, want 0", n)
	}
	effectstest.GaPut(g, me.ID, "Ench A", "Enchantment", "")
	effectstest.GaPut(g, me.ID, "Ench B", "Enchantment — Aura", "")
	effectstest.GaPut(g, opp.ID, "Theirs", "Enchantment", "")
	effectstest.GaPut(g, me.ID, "Rock", "Artifact", "")
	if n := b731ManaAmount(t, g, b731SanctumOracle, me.ID, sanctum); n != 2 {
		t.Errorf("with two enchantments of yours it adds %d, want 2", n)
	}
	if err := g.ActivateManaAbility(me.ID, sanctum, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !effectstest.TappedForTest(t, g, sanctum) {
		t.Error("the Sanctum did not tap")
	}
}

func TestOldB731TolarianAcademyAddsBlueForEachArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	academy := effectstest.PushCatalogPermanent(g, me.ID, "Tolarian Academy", "Legendary Land", b731AcademyOracle, false)
	if n := b731ManaAmount(t, g, b731AcademyOracle, me.ID, academy); n != 0 {
		t.Errorf("with no artifact it adds %d, want 0", n)
	}
	effectstest.GaPut(g, me.ID, "Rock", "Artifact", "")
	effectstest.GaPut(g, me.ID, "Golem", "Artifact Creature — Golem", "")
	effectstest.GaPut(g, opp.ID, "Theirs", "Artifact", "")
	effectstest.GaPut(g, me.ID, "Ench", "Enchantment", "")
	if n := b731ManaAmount(t, g, b731AcademyOracle, me.ID, academy); n != 2 {
		t.Errorf("with two artifacts of yours it adds %d, want 2", n)
	}
}

func TestOldB731ThranQuarrySacrificesAtAnyEndStepWithoutCreatures(t *testing.T) {
	for _, tc := range []struct {
		name     string
		creature bool
		theirs   bool
		survives bool
	}{
		{"no creatures", false, false, false},
		{"your creature", true, false, true},
		{"only an opponent's creature", false, true, false},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		quarry := effectstest.PushCatalogPermanent(g, me.ID, "Thran Quarry", "Land", b731QuarryOracle, false)
		if tc.creature {
			effectstest.B12Creature(g, me.ID, "Mine", b731CreatureLine, 1, 1)
		}
		if tc.theirs {
			effectstest.B12Creature(g, opp.ID, "Theirs", b731CreatureLine, 1, 1)
		}
		effectstest.AdvanceToEndStep(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if b731OnBattlefield(g, quarry) != tc.survives {
			t.Errorf("%s: Quarry on the battlefield = %v, want %v", tc.name, !tc.survives, tc.survives)
		}
	}
}

func TestOldB731ThranQuarryTriggersOnAnOpponentsEndStepToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	quarry := effectstest.PushCatalogPermanent(g, me.ID, "Thran Quarry", "Land", b731QuarryOracle, false)
	effectstest.AdvanceToEndStepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if b731OnBattlefield(g, quarry) {
		t.Error("the Quarry survived an opponent's end step with no creature of yours")
	}
}

func TestOldB731ThranQuarryTapsForAnyColor(t *testing.T) {
	spec, _ := effects.Lookup(b731QuarryOracle)
	if len(spec.ManaAbilities) != 1 {
		t.Fatalf("%d mana abilities, want 1", len(spec.ManaAbilities))
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	quarry := effectstest.PushCatalogPermanent(g, me.ID, "Thran Quarry", "Land", b731QuarryOracle, false)
	effectstest.B28TapForMana(t, g, me.ID, quarry, "R")
	if !effectstest.TappedForTest(t, g, quarry) {
		t.Error("the Quarry did not tap")
	}
}
