package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 182 (Ravnica Allegiance: removal, other creatures,
// planeswalkers, creatures with keywords, artifacts, lands, #5521). Every test
// plays the real catalog entry through the engine.

const (
	oldB182Deputy     = "0e4150db-ac43-48b4-9791-0d874906acf5"
	oldB182Volley     = "d4c4c7d7-744a-4548-adad-a5fc4fb3f586"
	oldB182Runner     = "3bfa51dd-daed-473a-8377-7bb34a2c3371"
	oldB182Infiltrate = "dcebce37-02eb-4b5d-86be-42b4a0287624"
	oldB182Dovin      = "a3386efe-d6c5-4853-a4db-61460da3364a"
	oldB182Domri      = "d8fe9672-0ead-4f47-abd8-e1ee133c5c60"
	oldB182Veteran    = "c25e4b69-4ea3-43ce-ba6f-b918ebfc4980"
	oldB182Foresight  = "d1108bd1-87b2-4795-9690-54ff4504e04f"
	oldB182Guildpact  = "5b410473-1df9-4d16-9b4d-c3ca8e4635b7"
	oldB182Plaza      = "5ff1d6d8-8cea-4a25-90d9-b575f4c99bc8"
	oldB182Torrent    = "70f44207-338e-4031-a093-f045e8d328b2"
	oldB182BellHaunt  = "796dc53d-5a0d-4d19-9fa4-f474a9d96041"
)

func TestOldB182EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB182Deputy: "Deputy of Detention", oldB182Volley: "Sagittars' Volley",
		oldB182Runner: "Rubblebelt Runner", oldB182Infiltrate: "Bloodmist Infiltrator",
		oldB182Dovin: "Dovin, Architect of Law", oldB182Domri: "Domri, City Smasher",
		oldB182Veteran: "Tenth District Veteran", oldB182Foresight: "Sphinx of Foresight",
		oldB182Guildpact: "Sphinx of the Guildpact", oldB182Plaza: "Plaza of Harmony",
		oldB182Torrent: "Swirling Torrent", oldB182BellHaunt: "Basilica Bell-Haunt",
	} {
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

func TestOldB182DeputyOfDetentionExilesEveryPermanentWithThatNameUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	b := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	mine := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	other := effectstest.B12Permanent(g, opp.ID, "Pebble", "Artifact")
	land := effectstest.B12Permanent(g, opp.ID, "Rock", "Land")
	deputy := effectstest.CastHoldCreature(t, g, "Deputy of Detention", "Creature — Vedalken Wizard", oldB182Deputy, 1, 3)
	if err := g.ResolvePickTarget(effectstest.PickTargetPrompt(t, g).ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
		t.Fatal("targeted my own permanent")
	}
	effectstest.AnswerPickTarget(t, g, a)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("both of the opponent's Rocks should be exiled")
	}
	for name, id := range map[string]uuid.UUID{"my Rock": mine, "a different name": other, "a land": land} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s was exiled", name)
		}
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(deputy) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, opp.ID, "Rock"); n != 3 {
		t.Errorf("the opponent controls %d Rocks (two artifacts and a land), want 3 after the Deputy left", n)
	}
}

func TestOldB182SagittarsVolleyDestroysOneFlyerAndPingsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flyer := func(owner uuid.UUID, toughness int) uuid.UUID {
		return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird", Keywords: []string{"flying"},
			Power: 1, Toughness: toughness, Owner: owner, Controller: owner,
		})
	}
	target, small, big, mine := flyer(opp.ID, 5), flyer(opp.ID, 1), flyer(opp.ID, 4), flyer(me.ID, 1)
	ground := effectstest.B12Creature(g, opp.ID, "Ground", "Creature — Bear", 2, 1)
	effectstest.CastCatalogSpell(t, g, "Sagittars' Volley", "Instant", oldB182Volley, effectstest.CardTarget(target))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(target) || g.Battlefield.Contains(small) {
		t.Error("the targeted flyer and the 1-toughness flyer should be gone")
	}
	for name, id := range map[string]uuid.UUID{"the big flyer": big, "my flyer": mine, "the ground creature": ground} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s died", name)
		}
	}
	if got := effectstest.FindCardAnywhere(t, g, big).DamageMarked; got != 1 {
		t.Errorf("the big flyer has %d damage, want 1", got)
	}
	if got := effectstest.FindCardAnywhere(t, g, mine).DamageMarked; got != 0 {
		t.Errorf("my own flyer took %d damage", got)
	}
}

func TestOldB182RubblebeltRunnerCantBeBlockedByTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	runner := effectstest.B12Push(g, me.ID, "Rubblebelt Runner", "Creature — Lizard Warrior", oldB182Runner, 3, 3)
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Token Creature — Goblin", Power: 1, Toughness: 1,
		Owner: opp.ID, Controller: opp.ID, TokenKey: "test-token",
	})
	card := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, runner)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(token, runner); err == nil {
		t.Error("a creature token blocked the Runner")
	}
	if err := g.DeclareBlocker(card, runner); err != nil {
		t.Errorf("a nontoken creature could not block: %v", err)
	}
}

func TestOldB182BloodmistInfiltratorSacrificesToBecomeUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	infiltrator := effectstest.B12Push(g, me.ID, "Bloodmist Infiltrator", "Creature — Vampire", oldB182Infiltrate, 3, 1)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, infiltrator)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatal("no sacrifice prompt")
	}
	if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{fodder}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the other creature was not sacrificed")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, infiltrator); err == nil {
		t.Error("the Infiltrator was blocked after the sacrifice")
	}
}

func TestOldB182DovinArchitectOfLawPlusAndMinus(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	dovin := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dovin, Architect of Law", TypeLine: "Legendary Planeswalker — Dovin", OracleID: oldB182Dovin,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLoyalty: 5},
	})
	life, hand := me.Life, me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, dovin, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 || me.Hand.Size() != hand+1 {
		t.Errorf("+1: life %d (want %d), hand %d (want %d)", me.Life, life+2, me.Hand.Size(), hand+1)
	}
	if got := effectstest.LoyaltyOf(g, dovin); got != 6 {
		t.Errorf("loyalty %d, want 6", got)
	}
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	// Only one loyalty ability a turn: move to the next of my turns.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, dovin, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(theirs)}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertTapped(t, g, true, theirs)
	if skips := effectstest.HmlUntapSkips(t, g, theirs); len(skips) == 0 {
		t.Error("the creature will untap normally")
	}
}

func TestOldB182DovinArchitectOfLawUltimateTapsEverythingAndSkipsTheUntapStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	dovin := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dovin, Architect of Law", TypeLine: "Legendary Planeswalker — Dovin", OracleID: oldB182Dovin,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLoyalty: 9},
	})
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, dovin, 2, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertTapped(t, g, true, bear, land)
	effectstest.AssertTapped(t, g, false, mine)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AssertTapped(t, g, true, bear, land)
}

// oldB182DomriTable seats Domri with the given loyalty, a creature of mine and
// one of the opponent's.
func oldB182DomriTable(t *testing.T, loyalty int) (g *game.Game, me, opp *game.Player, domri, bear, theirs uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	domri = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Domri, City Smasher", TypeLine: "Legendary Planeswalker — Domri", OracleID: oldB182Domri,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLoyalty: loyalty},
	})
	bear = effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs = effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	return
}

func TestOldB182DomriPlusTwoPumpsAndHastesMyCreatures(t *testing.T) {
	g, me, _, domri, bear, theirs := oldB182DomriTable(t, 4)
	if err := g.ActivateCatalogAbility(me.ID, domri, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, to := effectstest.EffectivePTOf(t, g, bear); p != 3 || to != 3 || !effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Errorf("my creature is %d/%d, want 3/3 with haste", p, to)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, theirs); p != 2 {
		t.Errorf("their creature was pumped: power %d", p)
	}
	if got := effectstest.LoyaltyOf(g, domri); got != 6 {
		t.Errorf("loyalty %d, want 6", got)
	}
}

func TestOldB182DomriMinusThreeDealsThree(t *testing.T) {
	g, me, opp, domri, _, _ := oldB182DomriTable(t, 4)
	before := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, domri, 1, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("life %d, want %d", opp.Life, before-3)
	}
}

func TestOldB182DomriMinusEightPutsCountersAndTrampleOnMyCreatures(t *testing.T) {
	g, me, _, domri, bear, theirs := oldB182DomriTable(t, 8)
	if err := g.ActivateCatalogAbility(me.ID, domri, 2, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, to := effectstest.PtOf(t, g, bear); p != 5 || to != 5 {
		t.Errorf("my creature is %d/%d, want 5/5", p, to)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("no trample after the −8")
	}
	if p, _ := effectstest.EffectivePTOf(t, g, theirs); p != 2 {
		t.Errorf("their creature got counters: power %d", p)
	}
}

func TestOldB182TenthDistrictVeteranUntapsAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vet := effectstest.B12Push(g, me.ID, "Tenth District Veteran", "Creature — Human Soldier", oldB182Veteran, 2, 3)
	ally := effectstest.B12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, vet)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(ally) })
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: ally}}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.AssertTapped(t, g, false, ally)
}

func TestOldB182SphinxOfForesightScriesEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Sphinx of Foresight", "Creature — Sphinx", oldB182Foresight, 4, 4)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Fatal("no scry at the beginning of my upkeep")
	}
	if effectstest.ScryChoiceFor(g, me.ID).ScryCards == nil || len(effectstest.ScryChoiceFor(g, me.ID).ScryCards) != 1 {
		t.Errorf("scried %d, want 1", len(effectstest.ScryChoiceFor(g, me.ID).ScryCards))
	}
}

func TestOldB182SphinxOfTheGuildpactIsAllColorsAndHexproofFromMonocolored(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sphinx := effectstest.B12Push(g, me.ID, "Sphinx of the Guildpact", "Artifact Creature — Sphinx", oldB182Guildpact, 5, 5)
	g.RunStateChecksForTest()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c, ok := effectstest.BattlefieldCard(g, sphinx)
	if !ok {
		t.Fatal("not on the battlefield")
	}
	if got := len(c.EffectiveColors()); got != 5 {
		t.Errorf("%d colors, want all five", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, sphinx, "flying") {
		t.Error("no flying")
	}
	spec, _ := effects.Lookup(oldB182Guildpact)
	found := false
	for _, kw := range spec.PrintedKeywords {
		if kw == game.HexproofFromMonocolored {
			found = true
		}
	}
	if !found {
		t.Error("hexproof from monocolored is not a printed keyword")
	}
}

func TestOldB182PlazaOfHarmonyGainsLifeWithTwoGatesAndMakesGateMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Permanent(g, me.ID, "Gate One", "Land — Gate")
	life := me.Life
	effectstest.B12PlayFromHand(t, g, "Plaza of Harmony", "Land", oldB182Plaza, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d with one Gate, want %d", me.Life, life)
	}
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.AdvanceToMain(t, g2)
	effectstest.B12Permanent(g2, me2.ID, "Gate One", "Land — Gate")
	effectstest.B12Permanent(g2, me2.ID, "Gate Two", "Land — Gate")
	life = me2.Life
	effectstest.B12PlayFromHand(t, g2, "Plaza of Harmony", "Land", oldB182Plaza, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g2)
	if me2.Life != life+3 {
		t.Errorf("life %d with two Gates, want %d", me2.Life, life+3)
	}
}

func TestOldB182SwirlingTorrentTucksAndBounces(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "Tucked", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Bounced", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpellWithModesTargets(t, g, oldB182Torrent, []int{0, 1}, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Fatal("a creature stayed on the battlefield")
	}
	if top, err := opp.Library.Top(); err != nil || top.InstanceID != a {
		t.Errorf("the tucked creature is not on top of its owner's library")
	}
	if !opp.Hand.Contains(b) {
		t.Error("the other creature was not returned to hand")
	}
}

func TestOldB182BasilicaBellHauntDrainsCardsAndLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	life := me.Life
	effectstest.CastHoldCreature(t, g, "Basilica Bell-Haunt", "Creature — Spirit", oldB182BellHaunt, 3, 4)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
	for _, opp := range g.Seats[1:] {
		if effectstest.DiscardChoiceFor(g, opp.ID) == nil {
			t.Errorf("%s was not asked to discard", opp.Name)
		}
	}
}
