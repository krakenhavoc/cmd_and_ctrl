package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b271_test.go — Conspiracy: Take the Crown, S82 older-sets batch
// 271 (#5540): card draw and library, instants and sorceries, monstrosity.
// Each test plays the real catalog entry.

const (
	oldB271LaceOracle       = "07beecfd-1720-42a1-93f8-f06604c3c228"
	oldB271CovenantOracle   = "3edac7df-2a88-4718-865e-703dff9c972e"
	oldB271DoseOracle       = "ad470afc-d1a5-4e64-8192-fb6308ed9bfe"
	oldB271LayOracle        = "cedc52eb-66a6-4b43-87f1-9bb9f4d4871e"
	oldB271RewardOracle     = "2fbb6703-3517-4117-9ca4-45e90a6216ce"
	oldB271VoidOracle       = "4d5a970b-dd36-4871-a2aa-36ba21917cdf"
	oldB271UnnerveOracle    = "8386f43f-8106-415c-a9a6-7fd3f5b060f3"
	oldB271BesmirchOracle   = "fc80380c-7d75-4556-aaf2-bf7b2c757e75"
	oldB271CrustaceanOracle = "9a6d8d37-fcf9-4974-89f5-3500e9da3518"
	oldB271SlimeOracle      = "c2153f2b-4a9f-40b6-8862-0e4143686c8c"
	oldB271LeucrocotaOracle = "dcc5ca3f-5309-4833-bee4-f2508382ea06"
	oldB271GiantOracle      = "dd2c0046-dc31-44f7-8601-7ec77515b926"
)

func TestOldB271EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB271LaceOracle: "Lace with Moonglove", oldB271CovenantOracle: "Covenant of Minds", oldB271DoseOracle: "Reviving Dose",
		oldB271LayOracle: "Lay of the Land", oldB271RewardOracle: "Faith's Reward", oldB271VoidOracle: "Into the Void",
		oldB271UnnerveOracle: "Unnerve", oldB271BesmirchOracle: "Besmirch", oldB271CrustaceanOracle: "Skittering Crustacean",
		oldB271SlimeOracle: "Splitting Slime", oldB271LeucrocotaOracle: "Ravenous Leucrocota", oldB271GiantOracle: "Stoneshock Giant",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Lace with Moonglove --------------------------------------------------

func TestOldB271LaceWithMoongloveGivesDeathtouchAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Lace with Moonglove", "Instant", oldB271LaceOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasString(effectstest.EffectiveAbilities(t, g, bear), "deathtouch") {
		t.Error("the target did not gain deathtouch")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
	effectstest.AdvanceToStepOf(t, g, 1, game.StepPrecombatMain)
	if effectstest.HasString(effectstest.EffectiveAbilities(t, g, bear), "deathtouch") {
		t.Error("deathtouch outlasted the turn")
	}
}

func TestOldB271LaceWithMoongloveRefusesAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Lace with Moonglove", "Instant", oldB271LaceOracle,
		[]game.TargetRef{effectstest.PlayerTarget(g.Seats[1].ID)}); err == nil {
		t.Error("Lace with Moonglove targeted a player")
	}
}

// --- Reviving Dose --------------------------------------------------------

func TestOldB271RevivingDoseGainsThreeAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life, hand := me.Life, me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Reviving Dose", "Instant", oldB271DoseOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life-life != 3 {
		t.Errorf("gained %d life, want 3", me.Life-life)
	}
	if me.Hand.Size()-hand != 1 {
		t.Errorf("drew %d cards, want 1", me.Hand.Size()-hand)
	}
}

// --- Lay of the Land ------------------------------------------------------

func TestOldB271LayOfTheLandFetchesABasicLandToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Wastes Tower", "Land"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	effectstest.CastCatalogSpell(t, g, "Lay of the Land", "Sorcery", oldB271LayOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1", hand, me.Hand.Size())
	}
}

// --- Faith's Reward -------------------------------------------------------

func TestOldB271FaithsRewardReturnsOnlyWhatDiedThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	died := effectstest.B38Graveyard(g, me.ID, "Stale", "Creature — Bear", "{1}", 2, 2) // never on the battlefield this turn
	diedNow := effectstest.B12Creature(g, me.ID, "Fresh", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, me.ID, diedNow)
	discarded := effectstest.HandCardFull(me, "Discarded", "Creature — Bear", "{1}", "", nil)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, discarded); err != nil {
		t.Fatal(err)
	}
	effectstest.CastCatalogSpell(t, g, "Faith's Reward", "Instant", oldB271RewardOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(diedNow) {
		t.Error("the permanent that died this turn did not return")
	}
	if g.Battlefield.Contains(died) || g.Battlefield.Contains(discarded) {
		t.Errorf("a card that did not die this turn returned (stale=%v discarded=%v)", g.Battlefield.Contains(died), g.Battlefield.Contains(discarded))
	}
}

// --- Into the Void --------------------------------------------------------

func TestOldB271IntoTheVoidBouncesUpToTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Into the Void", "Sorcery", oldB271VoidOracle, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) || !g.Battlefield.Contains(c) {
		t.Errorf("battlefield after: a=%v b=%v c=%v, want only c", g.Battlefield.Contains(a), g.Battlefield.Contains(b), g.Battlefield.Contains(c))
	}
	if !opp.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Error("a bounced creature is not in its owner's hand")
	}
}

func TestOldB271IntoTheVoidRefusesThreeTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	ids := []uuid.UUID{
		effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2),
		effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2),
		effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 2, 2),
	}
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Into the Void", "Sorcery", oldB271VoidOracle, effectstest.CardRefs(ids...)); err == nil {
		t.Error("Into the Void accepted three targets")
	}
}

// --- Unnerve --------------------------------------------------------------

func TestOldB271UnnerveMakesEachOpponentDiscardTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats[1:] {
		before[p.ID] = p.Hand.Size()
	}
	effectstest.CastCatalogSpell(t, g, "Unnerve", "Sorcery", oldB271UnnerveOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats[1:] {
		effectstest.DiscardFromHand(t, g, p.ID)
		if got := before[p.ID] - p.Hand.Size(); got != 2 {
			t.Errorf("seat %v discarded %d cards, want 2", p.ID, got)
		}
	}
	_ = me
}

// --- Besmirch -------------------------------------------------------------

func TestOldB271BesmirchStealsGivesHasteUntapsAndGoads(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == ogre {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.CastCatalogSpell(t, g, "Besmirch", "Sorcery", oldB271BesmirchOracle, effectstest.CardRefs(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.E2Card(t, g, ogre)
	if c.Controller != me.ID {
		t.Fatalf("controller %v, want me", c.Controller)
	}
	if c.Tapped {
		t.Error("the creature was not untapped")
	}
	if !effectstest.HasString(effectstest.EffectiveAbilities(t, g, ogre), "haste") {
		t.Error("the creature did not gain haste")
	}
	if !c.IsGoadedBy(me.ID) {
		t.Error("the creature was not goaded")
	}
	// Goaded, it must attack, and not the player who goaded it.
	third := g.Seats[2]
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ogre, me.ID); err == nil {
		t.Error("the goaded creature was allowed to attack the player who goaded it")
	}
	if err := g.DeclareAttacker(ogre, third.ID); err != nil {
		t.Fatalf("goaded creature could not attack another player: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceToStepOf(t, g, 1, game.StepPrecombatMain)
	if got := effectstest.E2Card(t, g, ogre).Controller; got != opp.ID {
		t.Error("control did not return at end of turn")
	}
}

// --- Covenant of Minds ----------------------------------------------------

func oldB271Covenant(t *testing.T, g *game.Game, accept bool) (revealed []uuid.UUID) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 3; i++ {
		revealed = append(revealed, effectstest.P8LibraryTop(g, me, game.Card{Name: "Top", TypeLine: "Sorcery"}))
	}
	effectstest.CastCatalogSpell(t, g, "Covenant of Minds", "Sorcery", oldB271CovenantOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, opp.ID, accept)
	effectstest.PassPriorityAroundTable(t, g)
	return revealed
}

func TestOldB271CovenantOfMindsGivesYouTheCardsWhenTheyAgree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	revealed := oldB271Covenant(t, g, true)
	for _, id := range revealed {
		if !me.Hand.Contains(id) {
			t.Errorf("revealed card %v is not in hand", id)
		}
	}
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("hand grew by %d, want 3", got)
	}
}

func TestOldB271CovenantOfMindsBinsThemAndDrawsFiveWhenTheyRefuse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	revealed := oldB271Covenant(t, g, false)
	for _, id := range revealed {
		if !me.Graveyard.Contains(id) {
			t.Errorf("revealed card %v is not in the graveyard", id)
		}
	}
	if got := me.Hand.Size() - hand; got != 5 {
		t.Errorf("hand grew by %d, want 5", got)
	}
}

// --- Monstrosity creatures ------------------------------------------------

func TestOldB271SkitteringCrustaceanHasHexproofOnlyOnceMonstrous(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	crab := effectstest.PushMonster(g, me.ID, "Skittering Crustacean", oldB271CrustaceanOracle, 2, 3)
	if effectstest.HasString(effectstest.EffectiveAbilities(t, g, crab), "hexproof") {
		t.Fatal("hexproof before monstrosity")
	}
	effectstest.ActivateMonstrosity(t, g, me.ID, crab, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasString(effectstest.EffectiveAbilities(t, g, crab), "hexproof") {
		t.Error("no hexproof once monstrous")
	}
	if p := effectstest.E2Card(t, g, crab).CurrentPower(); p != 6 {
		t.Errorf("power %d, want 6 (2 + four counters)", p)
	}
}

func TestOldB271RavenousLeucrocotaGrowsByThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	beast := effectstest.PushMonster(g, me.ID, "Ravenous Leucrocota", oldB271LeucrocotaOracle, 2, 4)
	effectstest.ActivateMonstrosity(t, g, me.ID, beast, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.E2Card(t, g, beast).CurrentPower(); p != 5 {
		t.Errorf("power %d, want 5", p)
	}
	if !effectstest.MonstrousNow(g, beast) {
		t.Error("not monstrous")
	}
}

func TestOldB271SplittingSlimeMakesAnUnmonstrousCopyWithNoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	slime := effectstest.PushMonster(g, me.ID, "Splitting Slime", oldB271SlimeOracle, 3, 3)
	effectstest.ActivateMonstrosity(t, g, me.ID, slime, 0)
	effectstest.PassPriorityAroundTable(t, g)
	var token uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID != slime && c.Controller == me.ID && c.Name == "Splitting Slime" {
				token = c.InstanceID
			}
		}
	})
	if token == uuid.Nil {
		t.Fatal("no token copy of the Slime")
	}
	if effectstest.MonstrousNow(g, token) {
		t.Error("the token copy is monstrous")
	}
	if p := effectstest.E2Card(t, g, token).CurrentPower(); p != 3 {
		t.Errorf("the token's power is %d, want 3 (no counters)", p)
	}
	if p := effectstest.E2Card(t, g, slime).CurrentPower(); p != 6 {
		t.Errorf("the Slime's power is %d, want 6", p)
	}
}

func TestOldB271StoneshockGiantStopsOnlyGroundBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := effectstest.PushMonster(g, me.ID, "Stoneshock Giant", oldB271GiantOracle, 5, 4)
	ground := effectstest.PushVanillaCreature(g, opp.ID, "Ground", 1, 1)
	air := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Air", TypeLine: "Creature — Bird",
		Keywords: []string{"flying"}, Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 1, 1)
	effectstest.ActivateMonstrosity(t, g, me.ID, giant, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Fatal("the trigger asked for a target")
	}
	effectstest.Pr7bAttack(t, g, opp.ID, giant, mine)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(ground, giant); err == nil {
		t.Error("a creature without flying blocked after the Giant became monstrous")
	}
	if err := g.DeclareBlocker(air, giant); err != nil {
		t.Errorf("a creature with flying could not block: %v", err)
	}
}
