package testsms

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msh_b1_cards_test.go — Marvel Super Heroes, half B, first batch (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	mshB1SynthoidsOracle   = "10916f33-aaa6-43a1-9835-b60fd061161f"
	mshB1AbsorbingOracle   = "ec8e310f-e871-498f-8d80-c3617c5e1433"
	mshB1Agent13Oracle     = "ddf7201a-6672-416d-be6b-202dafbeccab"
	mshB1CoulsonOracle     = "dae20bcd-dc24-4f9d-9400-a6f013b49f14"
	mshB1HydraOracle       = "31816924-1430-4581-86e8-325e2ca16217"
	mshB1InvasionOracle    = "7c854a07-e031-458e-a335-9c31865020b8"
	mshB1AntManOracle      = "b0ed9fbe-324d-467a-a326-c6f85a4c688a"
	mshB1ZolaOracle        = "95161e23-e4c4-411f-bce5-58a02abc0958"
	mshB1AttumaOracle      = "e3b70a85-4296-411f-a673-7992b0c8a9aa"
	mshB1DisassembledOracl = "34493cf5-7586-48ef-806d-57273b2ab034"
	mshB1UnderSiegeOracle  = "d911cf65-7273-4162-a766-7d4107699dc2"
)

func TestMshB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mshB1SynthoidsOracle:   "A.I.M. Synthoids",
		mshB1AbsorbingOracle:   "Absorbing Man",
		mshB1Agent13Oracle:     "Agent 13, Sharon Carter",
		mshB1CoulsonOracle:     "Agent Phil Coulson",
		mshB1HydraOracle:       "Agents of HYDRA",
		mshB1InvasionOracle:    "Alien Invasion",
		mshB1AntManOracle:      "Ant-Man, Colony Commander",
		mshB1ZolaOracle:        "Arnim Zola, Bio-Fanatic",
		mshB1AttumaOracle:      "Attuma, Atlantean Warlord",
		mshB1DisassembledOracl: "Avengers Disassembled",
		mshB1UnderSiegeOracle:  "Avengers: Under Siege",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: name %q, completeness %s", name, spec.Name, spec.Completeness)
		}
	}
}

// --- A.I.M. Synthoids ---------------------------------------------------------

func TestMshB1SynthoidsSurveilsTwoWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "Top A", "Top B", "Third")

	effectstest.CastCatalogSpell(t, g, "A.I.M. Synthoids", "Artifact Creature — Robot Villain", mshB1SynthoidsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("entering did not queue a surveil")
	}
	if len(c.ScryCards) != 2 {
		t.Fatalf("looked at %d cards, want 2 (surveil 2)", len(c.ScryCards))
	}
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{c.ScryCards[0]}, []uuid.UUID{c.ScryCards[1]}); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if got := effectstest.GraveyardNames(me); len(got) != 1 || got[0] != "Top A" {
		t.Errorf("graveyard = %v, want [Top A]", got)
	}
	if got := effectstest.LibraryTopNames(me, 1); got[0] != "Top B" {
		t.Errorf("library top = %q, want Top B", got[0])
	}
}

// --- Agents of HYDRA ----------------------------------------------------------

func TestMshB1HydraLeavesAVillainWithMenaceWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	agent := effectstest.PushCatalogPermanent(g, me.ID, "Agents of HYDRA", "Creature — Human Spy Villain", mshB1HydraOracle, false)

	effectstest.HobDestroy(t, g, agent)

	if g.Battlefield.Contains(agent) {
		t.Fatal("the Agent survived")
	}
	if got := effectstest.MshB1Tokens(g, me.ID, "Villain"); got != 1 {
		t.Fatalf("%d Villain tokens, want 1", got)
	}
	v := effectstest.MshB1Token(t, g, me.ID, "Villain")
	if v.Power != 2 || v.Toughness != 1 || len(v.Colors) != 1 || v.Colors[0] != "B" {
		t.Errorf("token = %d/%d %v, want a 2/1 black Villain", v.Power, v.Toughness, v.Colors)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, v.InstanceID, "menace") {
		t.Error("the Villain token lacks menace")
	}
}

func TestMshB1HydraDoesNothingWhenItIsExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	agent := effectstest.PushCatalogPermanent(g, me.ID, "Agents of HYDRA", "Creature — Human Spy Villain", mshB1HydraOracle, false)

	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(agent); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.MshB1Tokens(g, me.ID, "Villain"); got != 0 {
		t.Errorf("%d Villain tokens from an exile, want 0 (it did not die)", got)
	}
}

// --- Agent 13, Sharon Carter --------------------------------------------------

func TestMshB1Agent13InvestigatesWhenACreatureAttacksAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Agent 13, Sharon Carter", "Legendary Creature — Human Spy Hero", mshB1Agent13Oracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.MshB1Tokens(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues after a lone attack by another creature, want 1", got)
	}
}

func TestMshB1Agent13TriggersForHerselfAttackingAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	agent := effectstest.PushCatalogPermanent(g, me.ID, "Agent 13, Sharon Carter", "Legendary Creature — Human Spy Hero", mshB1Agent13Oracle, false)

	effectstest.DeclareAttack(t, g, opp.ID, agent)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.MshB1Tokens(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues after Agent 13 attacked alone, want 1", got)
	}
}

func TestMshB1Agent13DoesNotTriggerWhenTwoCreaturesAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	agent := effectstest.PushCatalogPermanent(g, me.ID, "Agent 13, Sharon Carter", "Legendary Creature — Human Spy Hero", mshB1Agent13Oracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	effectstest.DeclareAttack(t, g, opp.ID, agent, bear)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.MshB1Tokens(g, me.ID, "Clue"); got != 0 {
		t.Errorf("%d Clues after two creatures attacked, want 0 (they did not attack alone)", got)
	}
}

func TestMshB1Agent13IgnoresAnOpponentsLoneAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, opp.ID, "Agent 13, Sharon Carter", "Legendary Creature — Human Spy Hero", mshB1Agent13Oracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.MshB1Tokens(g, opp.ID, "Clue"); got != 0 {
		t.Errorf("the opponent's Agent 13 investigated for my attacker (%d Clues)", got)
	}
	if got := effectstest.MshB1Tokens(g, me.ID, "Clue"); got != 0 {
		t.Errorf("I got %d Clues from their Agent 13", got)
	}
}

// --- Agent Phil Coulson -------------------------------------------------------

func TestMshB1CoulsonHasVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	coulson := effectstest.PushCatalogPermanent(g, me.ID, "Agent Phil Coulson", "Legendary Creature — Human Spy Hero", mshB1CoulsonOracle, false)
	if !effectstest.HasKeywordOnBattlefield(t, g, coulson, "vigilance") {
		t.Error("Agent Phil Coulson lacks vigilance")
	}
}

func TestMshB1CoulsonPutsACounterOnEachOtherHeroYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	coulson := effectstest.PushCatalogPermanent(g, me.ID, "Agent Phil Coulson", "Legendary Creature — Human Spy Hero", mshB1CoulsonOracle, false)
	heroA := effectstest.B12Creature(g, me.ID, "Hero A", "Creature — Human Hero", 2, 2)
	heroB := effectstest.B12Creature(g, me.ID, "Hero B", "Creature — Mutant Hero", 1, 1)
	villain := effectstest.B12Creature(g, me.ID, "Villain", "Creature — Human Villain", 2, 2)
	theirHero := effectstest.B12Creature(g, opp.ID, "Their Hero", "Creature — Human Hero", 2, 2)

	effectstest.B16Activate(t, g, me.ID, coulson, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)

	for name, id := range map[string]uuid.UUID{"Hero A": heroA, "Hero B": heroB} {
		if n := effectstest.HobA8Counters(t, g, id); n != 1 {
			t.Errorf("%s has %d counters, want 1", name, n)
		}
	}
	for name, id := range map[string]uuid.UUID{"Coulson": coulson, "the Villain": villain, "their Hero": theirHero} {
		if n := effectstest.HobA8Counters(t, g, id); n != 0 {
			t.Errorf("%s has %d counters, want 0", name, n)
		}
	}
	if !effectstest.B16Tapped(t, g, coulson) {
		t.Error("the ability did not tap Coulson")
	}
}

func TestMshB1CoulsonNeedsToBeUntappedAndIsNotAnAttackCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	coulson := effectstest.PushCatalogPermanent(g, me.ID, "Agent Phil Coulson", "Legendary Creature — Human Spy Hero", mshB1CoulsonOracle, false)
	effectstest.B16Activate(t, g, me.ID, coulson, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, coulson, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("a tapped Coulson activated again")
	}
}

// --- Attuma, Atlantean Warlord ------------------------------------------------

func TestMshB1AttumaPumpsOnlyYourOtherMerfolk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	attuma := effectstest.PushCatalogPermanent(g, me.ID, "Attuma, Atlantean Warlord", "Legendary Creature — Merfolk Warrior Villain", mshB1AttumaOracle, false)
	mine := effectstest.B12Creature(g, me.ID, "My Merfolk", "Creature — Merfolk", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Merfolk", "Creature — Merfolk", 2, 2)

	if p, tough := effectstest.EffectivePTOf(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("my Merfolk = %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("a Bear = %d/%d, want 2/2", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("their Merfolk = %d/%d, want 2/2 (you control)", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, attuma); p != 1 || tough != 1 {
		t.Errorf("Attuma = %d/%d, want his own 1/1 test body (not pumped by himself)", p, tough)
	}
}

func TestMshB1AttumaDrawsOnceForTwoMerfolkAttackingOnePlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attuma := effectstest.PushCatalogPermanent(g, me.ID, "Attuma, Atlantean Warlord", "Legendary Creature — Merfolk Warrior Villain", mshB1AttumaOracle, false)
	other := effectstest.B12Creature(g, me.ID, "My Merfolk", "Creature — Merfolk", 2, 2)
	before := len(me.Hand.Cards)

	effectstest.DeclareAttack(t, g, opp.ID, attuma, other)
	effectstest.PassPriorityAroundTable(t, g)

	if got := len(me.Hand.Cards); got != before+1 {
		t.Errorf("hand %d -> %d, want exactly one draw for two Merfolk attacking one player", before, got)
	}
}

func TestMshB1AttumaDrawsOncePerPlayerAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	attuma := effectstest.PushCatalogPermanent(g, me.ID, "Attuma, Atlantean Warlord", "Legendary Creature — Merfolk Warrior Villain", mshB1AttumaOracle, false)
	other := effectstest.B12Creature(g, me.ID, "My Merfolk", "Creature — Merfolk", 2, 2)
	before := len(me.Hand.Cards)

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attuma, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(other, third.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	if got := len(me.Hand.Cards); got != before+2 {
		t.Errorf("hand %d -> %d, want a draw for each of the two players attacked", before, got)
	}
}

func TestMshB1AttumaIgnoresANonMerfolkAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Attuma, Atlantean Warlord", "Legendary Creature — Merfolk Warrior Villain", mshB1AttumaOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	before := len(me.Hand.Cards)

	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)

	if got := len(me.Hand.Cards); got != before {
		t.Errorf("hand %d -> %d, want no draw when only a Bear attacks", before, got)
	}
}

func TestMshB1AttumaCountsAnotherMerfolkAttackingAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Attuma, Atlantean Warlord", "Legendary Creature — Merfolk Warrior Villain", mshB1AttumaOracle, false)
	other := effectstest.B12Creature(g, me.ID, "My Merfolk", "Creature — Merfolk", 2, 2)
	before := len(me.Hand.Cards)

	effectstest.DeclareAttack(t, g, opp.ID, other)
	effectstest.PassPriorityAroundTable(t, g)

	if got := len(me.Hand.Cards); got != before+1 {
		t.Errorf("hand %d -> %d, want one draw for a lone Merfolk attacker", before, got)
	}
}

// --- Arnim Zola, Bio-Fanatic --------------------------------------------------

func mshB1Zola(t *testing.T, g *game.Game, creaturesInYard int) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	zola := effectstest.PushCatalogPermanent(g, me.ID, "Arnim Zola, Bio-Fanatic", "Legendary Artifact Creature — Scientist Villain", mshB1ZolaOracle, false)
	for i := 0; i < creaturesInYard; i++ {
		effectstest.PushCatalogGraveyardCard(me, "Dead Bear", "Creature — Bear", "", 2, 2)
	}
	return zola
}

func TestMshB1ZolaMakesATappedVillainWithTwoCreatureCardsInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	zola := mshB1Zola(t, g, 2)

	if err := effectstest.HobB4Activate(t, g, me.ID, zola, 0, "{C}{C}{C}", nil); err != nil {
		t.Fatalf("activation refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.MshB1Tokens(g, me.ID, "Villain"); got != 1 {
		t.Fatalf("%d Villain tokens, want 1", got)
	}
	v := effectstest.MshB1Token(t, g, me.ID, "Villain")
	if !v.Tapped {
		t.Error("the Villain entered untapped")
	}
	if v.Power != 2 || v.Toughness != 1 || len(v.Colors) != 1 || v.Colors[0] != "B" {
		t.Errorf("token = %d/%d %v, want a 2/1 black Villain", v.Power, v.Toughness, v.Colors)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, v.InstanceID, "menace") {
		t.Error("the Villain lacks menace")
	}
	if !effectstest.B16Tapped(t, g, zola) {
		t.Error("Zola did not tap for the cost")
	}
}

func TestMshB1ZolaNeedsTwoCreatureCardsInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	zola := mshB1Zola(t, g, 1)
	// A noncreature card does not count.
	effectstest.PushCatalogGraveyardCard(me, "Dead Rock", "Artifact", "", 0, 0)

	if err := effectstest.HobB4Activate(t, g, me.ID, zola, 0, "{C}{C}{C}", nil); err == nil {
		t.Fatal("activated with one creature card in the graveyard")
	}
	if got := effectstest.MshB1Tokens(g, me.ID, "Villain"); got != 0 {
		t.Errorf("%d Villains from a refused activation", got)
	}
}

func TestMshB1ZolaCountsOnlyYourOwnGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zola := mshB1Zola(t, g, 1)
	effectstest.PushCatalogGraveyardCard(opp, "Their Dead Bear", "Creature — Bear", "", 2, 2)
	effectstest.PushCatalogGraveyardCard(opp, "Their Other Bear", "Creature — Bear", "", 2, 2)

	if err := effectstest.HobB4Activate(t, g, me.ID, zola, 0, "{C}{C}{C}", nil); err == nil {
		t.Fatal("activated off an opponent's graveyard")
	}
}

func TestMshB1ZolaCostsThreeMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	zola := mshB1Zola(t, g, 2)

	if err := effectstest.HobB4Activate(t, g, me.ID, zola, 0, "{C}{C}", nil); err == nil {
		t.Fatal("activated with only two mana")
	}
}

// --- Alien Invasion -----------------------------------------------------------

func mshB1InvasionAtCombat(t *testing.T, g *game.Game, counters int) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	invasion := effectstest.PushCatalogPermanent(g, me.ID, "Alien Invasion", "Enchantment", mshB1InvasionOracle, false)
	if counters > 0 {
		if err := g.AddCounterBy(me.ID, invasion, "invasion", counters); err != nil {
			t.Fatalf("AddCounterBy: %v", err)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	return invasion
}

func mshB1InvasionCounters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s left the battlefield", id)
	}
	return c.Counters["invasion"]
}

func TestMshB1InvasionFirstAlienIsAOneOneAndAddsTheFirstCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	invasion := mshB1InvasionAtCombat(t, g, 0)

	if got := effectstest.MshB1Tokens(g, me.ID, "Alien"); got != 1 {
		t.Fatalf("%d Aliens, want 1", got)
	}
	alien := effectstest.MshB1Token(t, g, me.ID, "Alien")
	if alien.CurrentPower() != 1 || alien.CurrentToughness() != 1 {
		t.Errorf("Alien = %d/%d, want 1/1", alien.CurrentPower(), alien.CurrentToughness())
	}
	if len(alien.Colors) != 1 || alien.Colors[0] != "R" {
		t.Errorf("Alien colours = %v, want red", alien.Colors)
	}
	if n := mshB1InvasionCounters(t, g, invasion); n != 1 {
		t.Errorf("%d invasion counters, want 1", n)
	}
}

func TestMshB1InvasionAlienHasHasteAndMustAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mshB1InvasionAtCombat(t, g, 0)

	alien := effectstest.MshB1Token(t, g, me.ID, "Alien")
	if !effectstest.HasKeywordOnBattlefield(t, g, alien.InstanceID, "haste") {
		t.Error("the Alien lacks haste")
	}
	if len(alien.Effective().AttackRequirements) == 0 {
		t.Error("the Alien has no attack requirement (it attacks each combat if able)")
	}
}

func TestMshB1InvasionLaterAliensGrowByTheInvasionCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	invasion := mshB1InvasionAtCombat(t, g, 2)

	alien := effectstest.MshB1Token(t, g, me.ID, "Alien")
	if alien.CurrentPower() != 3 || alien.CurrentToughness() != 3 {
		t.Errorf("Alien with two invasion counters out = %d/%d, want 3/3", alien.CurrentPower(), alien.CurrentToughness())
	}
	if n := effectstest.HobA8Counters(t, g, alien.InstanceID); n != 2 {
		t.Errorf("Alien has %d +1/+1 counters, want 2", n)
	}
	if n := mshB1InvasionCounters(t, g, invasion); n != 3 {
		t.Errorf("%d invasion counters, want 3 (the new counter comes after the Alien's)", n)
	}
}

func TestMshB1InvasionAcrossTwoTurnsMakesAOneOneThenATwoTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	invasion := mshB1InvasionAtCombat(t, g, 0)
	first := effectstest.MshB1Token(t, g, me.ID, "Alien")
	// The first Alien must attack: put it away so the walk is not gated.
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(first.InstanceID); err != nil {
			t.Fatal(err)
		}
	})

	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.MshB1Tokens(g, me.ID, "Alien"); got != 1 {
		t.Fatalf("%d Aliens on the second turn, want 1", got)
	}
	second := effectstest.MshB1Token(t, g, me.ID, "Alien")
	if second.CurrentPower() != 2 || second.CurrentToughness() != 2 {
		t.Errorf("second Alien = %d/%d, want 2/2", second.CurrentPower(), second.CurrentToughness())
	}
	if n := mshB1InvasionCounters(t, g, invasion); n != 2 {
		t.Errorf("%d invasion counters, want 2", n)
	}
}

func TestMshB1InvasionOnlyTriggersOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, opp.ID, "Alien Invasion", "Enchantment", mshB1InvasionOracle, false)
	// It is my combat, not the controller's: nothing happens.
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.MshB1Tokens(g, opp.ID, "Alien"); got != 0 {
		t.Errorf("%d Aliens at someone else's combat, want 0", got)
	}
}

// --- Ant-Man, Colony Commander ------------------------------------------------

func mshB1AntManAttack(t *testing.T, g *game.Game) (antMan, bear uuid.UUID) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	antMan = effectstest.PushCatalogPermanent(g, me.ID, "Ant-Man, Colony Commander", "Legendary Creature — Human Rogue Hero", mshB1AntManOracle, false)
	bear = effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, antMan)
	// The pool empties at every step change, so the {1} is floated now.
	effectstest.B06AddMana(me, "G")
	effectstest.PassPriorityAroundTable(t, g)
	return antMan, bear
}

func TestMshB1AntManPaysOneToPutACounterOnTargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	_, bear := mshB1AntManAttack(t, g)

	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.HobA8Counters(t, g, bear); n != 1 {
		t.Errorf("Bear has %d counters, want 1", n)
	}
	if got := effectstest.MshB1Tokens(g, me.ID, "Insect"); got != 1 {
		t.Errorf("%d Insects, want 1 (a +1/+1 counter was put on a creature)", got)
	}
	if len(effectstest.PoolColors(me)) != 0 {
		t.Errorf("pool = %v after paying {1}", effectstest.PoolColors(me))
	}
}

func TestMshB1AntManCanTargetAnyCreatureIncludingAnOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mshB1AntManAttack(t, g)

	effectstest.AnswerPayUnless(t, g, me.ID, true)
	pick := effectstest.PickTargetPrompt(t, g)
	found := false
	for _, id := range pick.PickTargetCards {
		if id == theirs {
			found = true
		}
	}
	if !found {
		t.Fatalf("an opponent's creature was not offered: %v", pick.PickTargetCards)
	}
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.HobA8Counters(t, g, theirs); n != 1 {
		t.Errorf("their Bear has %d counters, want 1", n)
	}
	if got := effectstest.MshB1Tokens(g, me.ID, "Insect"); got != 1 {
		t.Errorf("%d Insects, want 1 (it says a creature, not a creature you control)", got)
	}
}

func TestMshB1AntManDecliningTheMayPayDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	_, bear := mshB1AntManAttack(t, g)

	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.HobA8Counters(t, g, bear); n != 0 {
		t.Errorf("Bear has %d counters after declining", n)
	}
	if got := effectstest.MshB1Tokens(g, me.ID, "Insect"); got != 0 {
		t.Errorf("%d Insects after declining", got)
	}
	if len(effectstest.PoolColors(me)) != 1 {
		t.Errorf("the {1} was spent anyway: pool = %v", effectstest.PoolColors(me))
	}
}

func TestMshB1AntManInsectOnlyOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Ant-Man, Colony Commander", "Legendary Creature — Human Rogue Hero", mshB1AntManOracle, false)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)

	for _, id := range []uuid.UUID{a, b, a} {
		if err := g.AddCounterBy(me.ID, id, game.CounterPlusOne, 1); err != nil {
			t.Fatalf("AddCounterBy: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}

	if got := effectstest.MshB1Tokens(g, me.ID, "Insect"); got != 1 {
		t.Errorf("%d Insects after three placements in one turn, want 1", got)
	}

	// The next turn it triggers again.
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	if err := g.AddCounterBy(me.ID, b, game.CounterPlusOne, 1); err != nil {
		t.Fatalf("AddCounterBy: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.MshB1Tokens(g, me.ID, "Insect"); got != 2 {
		t.Errorf("%d Insects after a placement on a new turn, want 2", got)
	}
}

func TestMshB1AntManIgnoresOpponentsCountersAndOtherKinds(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Ant-Man, Colony Commander", "Legendary Creature — Human Rogue Hero", mshB1AntManOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")

	if err := g.AddCounterBy(opp.ID, bear, game.CounterPlusOne, 1); err != nil {
		t.Fatalf("AddCounterBy: %v", err)
	}
	if err := g.AddCounterBy(me.ID, bear, game.CounterMinusOne, 1); err != nil {
		t.Fatalf("AddCounterBy: %v", err)
	}
	if err := g.AddCounterBy(me.ID, rock, game.CounterPlusOne, 1); err != nil {
		t.Fatalf("AddCounterBy: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.MshB1Tokens(g, me.ID, "Insect"); got != 0 {
		t.Errorf("%d Insects from counters that do not count, want 0", got)
	}
}

func TestMshB1AntManInsectIsAOneOneGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Ant-Man, Colony Commander", "Legendary Creature — Human Rogue Hero", mshB1AntManOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	if err := g.AddCounterBy(me.ID, bear, game.CounterPlusOne, 1); err != nil {
		t.Fatalf("AddCounterBy: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	insect := effectstest.MshB1Token(t, g, me.ID, "Insect")
	if insect.Power != 1 || insect.Toughness != 1 || len(insect.Colors) != 1 || insect.Colors[0] != "G" {
		t.Errorf("Insect = %d/%d %v, want a 1/1 green Insect", insect.Power, insect.Toughness, insect.Colors)
	}
}

// --- Absorbing Man ------------------------------------------------------------

func mshB1AbsorbingMan(g *game.Game) uuid.UUID {
	me := g.Seats[0]
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Absorbing Man", OracleID: mshB1AbsorbingOracle,
		TypeLine: "Legendary Creature — Human Villain",
		Power:    4, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
}

func mshB1HasType(c game.Card, word string) bool {
	for _, w := range strings.Fields(strings.ReplaceAll(c.TypeLine, "—", " ")) {
		if w == word {
			return true
		}
	}
	return false
}

func TestMshB1AbsorbingManBecomesACopyOfATargetLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	man := mshB1AbsorbingMan(g)
	forest := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")

	effectstest.AdvanceToMain(t, g)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: forest}, true)

	c := effectstest.GaLayered(t, g, man)
	if c.Name != "Absorbing Man" {
		t.Errorf("name = %q, want Absorbing Man", c.Name)
	}
	for _, word := range []string{"Legendary", "Land", "Creature", "Forest", "Human", "Villain"} {
		if !mshB1HasType(c, word) {
			t.Errorf("type line %q lacks %s", c.TypeLine, word)
		}
	}
	if p, tough := effectstest.EffectivePTOf(t, g, man); p != 4 || tough != 4 {
		t.Errorf("P/T = %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, man, "vigilance") {
		t.Error("he lacks vigilance")
	}
	if !c.IsLand() || !c.IsCreature() {
		t.Error("he should be a land creature while the copy lasts")
	}
}

func TestMshB1AbsorbingManCopiesAnArtifactsAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	man := mshB1AbsorbingMan(g)
	rock := effectstest.PushCatalogPermanent(g, me.ID, "Arnim Zola, Bio-Fanatic", "Legendary Artifact Creature — Scientist Villain", mshB1ZolaOracle, false)

	effectstest.AdvanceToMain(t, g)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: rock}, true)

	c := effectstest.GaLayered(t, g, man)
	if c.OracleID != mshB1ZolaOracle {
		t.Errorf("oracle id = %q, want the copied card's (so it has its abilities)", c.OracleID)
	}
	if c.Name != "Absorbing Man" {
		t.Errorf("name = %q, want Absorbing Man", c.Name)
	}
	if !mshB1HasType(c, "Artifact") || !mshB1HasType(c, "Creature") || !mshB1HasType(c, "Scientist") || !mshB1HasType(c, "Villain") || !mshB1HasType(c, "Human") {
		t.Errorf("type line = %q, want the artifact's types plus Human Villain creature", c.TypeLine)
	}
}

func TestMshB1AbsorbingManOffersOnlyArtifactsNonAuraEnchantmentsAndLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mshB1AbsorbingMan(g)
	land := effectstest.SeedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	ench := effectstest.B12Permanent(g, opp.ID, "Shrine", "Enchantment")
	aura := effectstest.B12Permanent(g, opp.ID, "Pacifism", "Enchantment — Aura")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)

	effectstest.AdvanceToMain(t, g)
	pick := effectstest.PickTargetPrompt(t, g)
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.PickTargetCards {
		offered[id] = true
	}
	for name, id := range map[string]uuid.UUID{"a land": land, "an artifact": rock, "a non-Aura enchantment": ench} {
		if !offered[id] {
			t.Errorf("%s was not offered", name)
		}
	}
	for name, id := range map[string]uuid.UUID{"an Aura": aura, "a creature": bear} {
		if offered[id] {
			t.Errorf("%s was offered", name)
		}
	}
	_ = me
}

func TestMshB1AbsorbingManCanChooseNoTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	man := mshB1AbsorbingMan(g)
	effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")

	effectstest.AdvanceToMain(t, g)
	pick := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTargets(pick.ID, pick.Chooser, nil); err != nil {
		t.Fatalf("ResolvePickTargets(none): %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.GaLayered(t, g, man)
	if c.IsLand() || c.Name != "Absorbing Man" {
		t.Errorf("he changed with no target: %q %q", c.Name, c.TypeLine)
	}
}

func TestMshB1AbsorbingManRevertsWhenYourNextTurnBegins(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	man := mshB1AbsorbingMan(g)
	forest := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")

	effectstest.AdvanceToMain(t, g)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: forest}, true)

	// Past the end of this turn, through every opponent's turn: still a copy.
	effectstest.AdvanceToNextSeatsTurn(t, g)
	for g.Turn.ActiveSeat != 0 {
		if c := effectstest.GaLayered(t, g, man); !c.IsLand() {
			t.Fatalf("the copy ended during seat %d's turn (type line %q)", g.Turn.ActiveSeat, c.TypeLine)
		}
		effectstest.AdvanceToNextSeatsTurn(t, g)
	}
	// Seat 0's turn has begun: the copy is over before the first main phase.
	for i := 0; i < 8 && g.Turn.Step != game.StepUpkeep; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	c := effectstest.GaLayered(t, g, man)
	if c.IsLand() || mshB1HasType(c, "Forest") {
		t.Errorf("still a copy on my next turn: %q", c.TypeLine)
	}
}

// --- Avengers Disassembled ----------------------------------------------------

func TestMshB1DisassembledDamageModeHitsEveryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 2, 3)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 4, 4)

	effectstest.B02cCastModalSpell(t, g, "Avengers Disassembled", "Sorcery", mshB1DisassembledOracl, []int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(small) {
		t.Error("the 2/3 survived 3 damage")
	}
	if !g.Battlefield.Contains(big) {
		t.Fatal("the 4/4 died to 3 damage")
	}
	if n := effectstest.DamageMarkedOn(g, big); n != 3 {
		t.Errorf("the 4/4 has %d damage, want 3", n)
	}
}

func TestMshB1DisassembledDamageModeDoesNotHitPlayers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	myLife, oppLife := me.Life, opp.Life

	effectstest.B02cCastModalSpell(t, g, "Avengers Disassembled", "Sorcery", mshB1DisassembledOracl, []int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if me.Life != myLife || opp.Life != oppLife {
		t.Errorf("life changed: me %d->%d, opp %d->%d; it hits creatures only", myLife, me.Life, oppLife, opp.Life)
	}
}

func TestMshB1DisassembledLandModeDestroysAndOffersATappedBasic(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.SeedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
	forest := effectstest.StapleLibraryCard(opp, "Forest", "Basic Land — Forest")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)

	effectstest.B02cCastModalSpell(t, g, "Avengers Disassembled", "Sorcery", mshB1DisassembledOracl, []int{1}, effectstest.S58Refs(land))
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(land) {
		t.Error("the land survived")
	}
	if effectstest.SearchChoiceFor(g, g.Seats[0].ID) != nil {
		t.Error("the caster was offered the search; it belongs to the land's controller")
	}
	effectstest.AnswerSearchByID(t, g, opp.ID, forest)
	got, ok := effectstest.BattlefieldCard(g, forest)
	if !ok || !got.Tapped {
		t.Errorf("the basic must be on the battlefield tapped, got %+v (found %v)", got, ok)
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the land-only mode killed a creature")
	}
	if n := effectstest.DamageMarkedOn(g, bear); n != 0 {
		t.Errorf("the land-only mode dealt %d damage", n)
	}
}

func TestMshB1DisassembledLandModeSearchIsOptional(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.SeedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
	forest := effectstest.StapleLibraryCard(opp, "Forest", "Basic Land — Forest")

	effectstest.B02cCastModalSpell(t, g, "Avengers Disassembled", "Sorcery", mshB1DisassembledOracl, []int{1}, effectstest.S58Refs(land))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchFailToFind(t, g, opp.ID)

	if g.Battlefield.Contains(forest) {
		t.Error("the basic was fetched after declining")
	}
}

func TestMshB1DisassembledBothModesDamageThenDestroy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.SeedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
	forest := effectstest.StapleLibraryCard(opp, "Forest", "Basic Land — Forest")
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 3)

	effectstest.B02cCastModalSpell(t, g, "Avengers Disassembled", "Sorcery", mshB1DisassembledOracl, []int{0, 1}, effectstest.S58Refs(land))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, opp.ID, forest)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(small) {
		t.Error("the 2/3 survived the damage mode")
	}
	if g.Battlefield.Contains(land) {
		t.Error("the land survived the destroy mode")
	}
	if !g.Battlefield.Contains(forest) {
		t.Error("the fetched basic did not arrive")
	}
}

func TestMshB1DisassembledNeedsAtLeastOneMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	err := effectstest.CastCatalogSpellErr(t, g, "Avengers Disassembled", "Sorcery", mshB1DisassembledOracl, nil)
	if err == nil {
		t.Error("cast with no mode chosen")
	}
}

// --- Avengers: Under Siege ----------------------------------------------------

func mshB1CastUnderSiege(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := effectstest.CastCatalogSpell(t, g, "Avengers: Under Siege", "Enchantment — Saga", mshB1UnderSiegeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestMshB1UnderSiegeChapterOneMakesTwoVillainsWithMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mshB1CastUnderSiege(t, g)

	if got := effectstest.MshB1Tokens(g, me.ID, "Villain"); got != 2 {
		t.Fatalf("%d Villain tokens after chapter I, want 2", got)
	}
	v := effectstest.MshB1Token(t, g, me.ID, "Villain")
	if v.Power != 2 || v.Toughness != 1 || len(v.Colors) != 1 || v.Colors[0] != "B" {
		t.Errorf("token = %d/%d %v, want a 2/1 black Villain", v.Power, v.Toughness, v.Colors)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, v.InstanceID, "menace") {
		t.Error("the Villain lacks menace")
	}
}

func TestMshB1UnderSiegeChapterTwoSparesVillainsAndHitsEveryOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirBear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	theirVillain := effectstest.B12Creature(g, opp.ID, "Their Villain", "Creature — Human Villain", 2, 5)
	mshB1CastUnderSiege(t, g)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(theirBear) {
		t.Error("a non-Villain 2/2 survived 2 damage")
	}
	if n := effectstest.DamageMarkedOn(g, theirVillain); n != 0 {
		t.Errorf("a Villain was dealt %d damage by chapter II", n)
	}
	if got := effectstest.MshB1Tokens(g, me.ID, "Villain"); got != 2 {
		t.Errorf("my %d Villain tokens, want both to survive", got)
	}
	for i, p := range g.Seats {
		want := lives[p.ID]
		if i != 0 {
			want -= 2
		}
		if p.Life != want {
			t.Errorf("seat %d life = %d, want %d (each opponent loses 2, I lose none)", i, p.Life, want)
		}
	}
}

func TestMshB1UnderSiegeChapterThreeMakesATreasurePerVillainYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "My Villain", "Creature — Human Villain", 2, 5)
	effectstest.B12Creature(g, opp.ID, "Their Villain", "Creature — Human Villain", 2, 5)
	saga := mshB1CastUnderSiege(t, g)

	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.MshB1Tokens(g, me.ID, "Treasure"); got != 3 {
		t.Errorf("%d Treasures, want 3 (two tokens and my own Villain, not theirs)", got)
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}
