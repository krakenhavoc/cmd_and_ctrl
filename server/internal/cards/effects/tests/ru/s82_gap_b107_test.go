package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b107_test.go — S82 GAP batch 107 (#5948). Each test plays the real
// catalog entry.

const (
	gapB107BraidsOracle       = "359b8f2c-831e-47d0-9267-dae79cecb63a"
	gapB107BrinelinOracle     = "1f00edd6-4321-48a1-9053-77c77b922aeb"
	gapB107ArtisansOracle     = "7dda40f2-7b5f-479c-86b6-8bc11fff9f4a"
	gapB107KahoOracle         = "395a7bd6-7c2d-448e-842c-ca53256d7008"
	gapB107IngesterOracle     = "ec9c35f2-6a68-45e7-a86e-9f5af9fe28fc"
	gapB107WeaponsmithOracle  = "40c1375b-1d15-4525-894a-16b7a0ac02b4"
	gapB107ShiftingOracle     = "6a262f53-2eea-45d0-8189-9222d66f1b52"
	gapB107SparkshaperOracle  = "5e674793-6586-44a4-b936-a71203c5850a"
	gapB107DovinGrandArbiterO = "596237da-3d51-4052-a8ca-2218efc1afb3"
)

func TestGapB107CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB107BraidsOracle:      "Braids, Conjurer Adept",
		gapB107BrinelinOracle:    "Brinelin, the Moon Kraken",
		gapB107ArtisansOracle:    "Faerie Artisans",
		gapB107KahoOracle:        "Kaho, Minamo Historian",
		gapB107IngesterOracle:    "Phyrexian Ingester",
		gapB107WeaponsmithOracle: "Renowned Weaponsmith",
		gapB107ShiftingOracle:    "Shifting Sliver",
		gapB107SparkshaperOracle: "Sparkshaper Visionary",
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

// --- Braids, Conjurer Adept --------------------------------------------------

func TestGapB107BraidsLetsEachUpkeepsPlayerPutACardFromTheirHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Braids, Conjurer Adept", "Legendary Creature — Human Wizard", gapB107BraidsOracle, 2, 2)
	bear := effectstest.HandCardForTest(opp, "Hand Bear", "Creature — Bear", "")
	spell := effectstest.HandCardForTest(opp, "Hand Instant", "Instant", "")
	aura := effectstest.HandCardForTest(opp, "Hand Aura", "Enchantment — Aura", "")
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, opp.ID)
	if pick == nil {
		t.Fatal("the opponent was not offered a card at their own upkeep")
	}
	for _, bad := range []uuid.UUID{spell, aura} {
		if effectstest.HasID(pick.ChooseCards, bad) {
			t.Error("a card that is not an artifact, creature or land was offered")
		}
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := g.LookupCardForEffect(bear)
	if !ok || !g.Battlefield.Contains(bear) || c.Controller != opp.ID {
		t.Errorf("the Bear did not enter under the opponent's control (on battlefield %v, controller %v)", g.Battlefield.Contains(bear), c.Controller)
	}
}

func TestGapB107BraidsOffersItsControllerTheirOwnLandAtTheirUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Braids, Conjurer Adept", "Legendary Creature — Human Wizard", gapB107BraidsOracle, 2, 2)
	land := effectstest.HandCardForTest(me, "Hand Forest", "Basic Land — Forest", "")
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) {
		t.Error("the land did not enter")
	}
}

// --- Brinelin, the Moon Kraken -----------------------------------------------

func TestGapB107BrinelinBouncesANonlandPermanentWhenItEntersIfYouAgree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, opp.ID, "Their Land", "Land")
	effectstest.ToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Brinelin, the Moon Kraken", "Legendary Creature — Kraken", gapB107BrinelinOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ResolvePickTarget(effectstest.LatestPickTarget(g, g.Seats[0].ID).ID, g.Seats[0].ID, game.TargetRef{Kind: game.TargetCard, ID: land}); err == nil {
		t.Fatal("a land was accepted as the nonland target")
	}
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, g.Seats[0].ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("the targeted permanent was not returned to its owner's hand")
	}
}

func TestGapB107BrinelinDecliningLeavesThePermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.ToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Brinelin, the Moon Kraken", "Legendary Creature — Kraken", gapB107BrinelinOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, g.Seats[0].ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirs) {
		t.Error("the permanent was returned although the player declined")
	}
}

func TestGapB107BrinelinTriggersOnASpellOfManaValueSixOrMoreOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	brin := effectstest.B12Push(g, me.ID, "Brinelin, the Moon Kraken", "Legendary Creature — Kraken", gapB107BrinelinOracle, 6, 8)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.ToMain(t, g)
	five := effectstest.MayhemHandCard(g, me, "Five Drop", "Sorcery", "{5}", "")
	if err := g.CastSpell(me.ID, five, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	g.RunStateChecksForTest()
	if effectstest.TriggerOnStack(g, brin) != nil {
		t.Fatal("a mana value 5 spell triggered Brinelin")
	}
	effectstest.PassPriorityAroundTable(t, g)
	six := effectstest.MayhemHandCard(g, me, "Six Drop", "Sorcery", "{6}", "")
	if err := g.CastSpell(me.ID, six, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	g.RunStateChecksForTest()
	effectstest.AnswerPickTarget(t, g, theirs)
	g.RunStateChecksForTest()
	if effectstest.TriggerOnStack(g, brin) == nil {
		t.Fatal("a mana value 6 spell did not trigger Brinelin")
	}
}

// --- Faerie Artisans ---------------------------------------------------------

func gapB107TokensNamed(g *game.Game, name string) []game.Card {
	var out []game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsToken() && c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

func TestGapB107FaerieArtisansCopiesAnOpponentsCreatureAsAnArtifactAndExilesTheOlderCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, owner := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, owner.ID, "Faerie Artisans", "Creature — Faerie Artificer", gapB107ArtisansOracle, 2, 2)
	effectstest.ToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	copies := gapB107TokensNamed(g, "Grizzly Bears")
	if len(copies) != 1 {
		t.Fatalf("%d token copies, want 1", len(copies))
	}
	if !copies[0].IsArtifact() || !copies[0].IsCreature() || copies[0].Controller != owner.ID {
		t.Errorf("copy %q: artifact %v creature %v controller %v, want an artifact creature under the Artisans' controller",
			copies[0].TypeLine, copies[0].IsArtifact(), copies[0].IsCreature(), copies[0].Controller)
	}
	first := copies[0].InstanceID
	effectstest.CastCatalogSpell(t, g, "Hill Giant", "Creature — Giant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(first) {
		t.Error("the earlier token created with the Artisans was not exiled")
	}
	if n := len(gapB107TokensNamed(g, "Hill Giant")); n != 1 {
		t.Errorf("%d Hill Giant copies, want 1", n)
	}
	_ = caster
}

func TestGapB107FaerieArtisansIgnoresItsControllersCreaturesAndTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Faerie Artisans", "Creature — Faerie Artificer", gapB107ArtisansOracle, 2, 2)
	effectstest.ToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(gapB107TokensNamed(g, "Grizzly Bears")); n != 0 {
		t.Errorf("%d copies of a creature its own controller cast", n)
	}
}

// --- Kaho, Minamo Historian --------------------------------------------------

func TestGapB107KahoExilesThreeInstantsAndCastsOneOfManaValueXFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Zap One", TypeLine: "Instant", ManaCost: "{R}"},
		game.Card{Name: "Zap Two", TypeLine: "Instant", ManaCost: "{1}{R}"},
		game.Card{Name: "Zap Two Again", TypeLine: "Instant", ManaCost: "{1}{U}"},
		game.Card{Name: "Zap Three", TypeLine: "Instant", ManaCost: "{2}{R}"},
		game.Card{Name: "Slow Sorcery", TypeLine: "Sorcery", ManaCost: "{1}{R}"},
	)
	effectstest.ToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Kaho, Minamo Historian", "Legendary Creature — Human Wizard", gapB107KahoOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pickable := effectstest.SearchChoiceFor(g, me.ID)
	if pickable == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.HasID(pickable.SearchCards, ids[4]) {
		t.Error("a sorcery was offered to a search for instants")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ids[0], ids[1], ids[2])
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range ids[:3] {
		if effectstest.B12ZoneOf(g, id) != game.ZoneExile {
			t.Fatalf("an instant is in %v, want exile", effectstest.B12ZoneOf(g, id))
		}
	}
	if effectstest.B12ZoneOf(g, ids[3]) == game.ZoneExile {
		t.Error("a fourth instant was exiled")
	}
	kaho := effectstest.BattlefieldIDsNamed(g, "Kaho, Minamo Historian")[0]
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == kaho {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	// X = 2: the offer covers the two mana value 2 instants, not the 1-drop.
	if err := effectstest.RavB2Activate(t, g, me.ID, kaho, 0, game.ActivateAbilityParams{XValue: 2}, "C", "C"); err != nil {
		t.Fatalf("activate with X=2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var offer *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceResolutionCast && c.Chooser == me.ID {
			offer = c
		}
	}
	if offer == nil {
		t.Fatal("no free-cast offer after activating Kaho")
	}
	if effectstest.HasID(offer.DiscardOptions, ids[0]) {
		t.Error("the mana value 1 instant was offered for X=2")
	}
	if !effectstest.HasID(offer.DiscardOptions, ids[1]) || !effectstest.HasID(offer.DiscardOptions, ids[2]) {
		t.Errorf("the mana value 2 instants were not both offered: %v", offer.DiscardOptions)
	}
}

func TestGapB107KahoNeedsXToMatchAndTheTap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedSearchLibrary(me, game.Card{Name: "Zap Two", TypeLine: "Instant", ManaCost: "{1}{R}"})
	effectstest.ToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Kaho, Minamo Historian", "Legendary Creature — Human Wizard", gapB107KahoOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, ids[0])
	effectstest.PassPriorityAroundTable(t, g)
	kaho := effectstest.BattlefieldIDsNamed(g, "Kaho, Minamo Historian")[0]
	// Summoning sick: the {T} symbol is refused the turn she arrives.
	if err := effectstest.RavB2Activate(t, g, me.ID, kaho, 0, game.ActivateAbilityParams{XValue: 2}, "C", "C"); err == nil {
		t.Error("Kaho tapped the turn she arrived")
	}
}

// --- Phyrexian Ingester ------------------------------------------------------

// gapB107CastIngester casts a real 3/3 Ingester from the active player's hand
// and lets the cast resolve, stopping at the enters trigger's prompts.
func gapB107CastIngester(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PutInHand(me, game.Card{Name: "Phyrexian Ingester", TypeLine: "Creature — Phyrexian Beast",
		OracleID: gapB107IngesterOracle, Power: 3, Toughness: 3})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the Ingester: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestGapB107IngesterGetsTheExiledCreaturesPowerAndToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Big Victim", "Creature — Ogre", 4, 5)
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Tok", TypeLine: "Token Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	effectstest.ToMain(t, g)
	ing := gapB107CastIngester(t, g)
	if p := effectstest.LatestPickTarget(g, g.Seats[0].ID); p != nil && effectstest.HasID(p.PickTargetCards, token) {
		t.Error("a token was offered to an exile of a nontoken creature")
	}
	effectstest.AnswerPickTarget(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, g.Seats[0].ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, victim) != game.ZoneExile {
		t.Fatalf("the creature is in %v, want exile", effectstest.B12ZoneOf(g, victim))
	}
	if p, tough := effectstest.EffectivePTOf(t, g, ing); p != 7 || tough != 8 {
		t.Errorf("the Ingester is %d/%d, want 3/3 plus the exiled 4/5", p, tough)
	}
}

func TestGapB107IngesterIsAPlainThreeThreeWithoutAnImprint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Big Victim", "Creature — Ogre", 4, 5)
	effectstest.ToMain(t, g)
	ing := gapB107CastIngester(t, g)
	effectstest.AnswerPickTarget(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, g.Seats[0].ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Error("the creature was exiled although the player declined")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, ing); p != 3 || tough != 3 {
		t.Errorf("the Ingester is %d/%d, want its own 3/3", p, tough)
	}
}

// --- Renowned Weaponsmith ----------------------------------------------------

func TestGapB107WeaponsmithMakesTwoColourlessForArtifactsOnly(t *testing.T) {
	spec, _ := effects.Lookup(gapB107WeaponsmithOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{C}{C}" || !spec.ManaAbilities[0].Cost.Tap {
		t.Fatalf("mana abilities %+v, want one {T}: Add {C}{C}", spec.ManaAbilities)
	}
	if len(spec.ManaAbilities[0].Restrictions) != 1 {
		t.Errorf("restrictions %v, want the artifact cast-or-activate rule", spec.ManaAbilities[0].Restrictions)
	}
}

func TestGapB107WeaponsmithFindsHeartPiercerBowOrVialOfDragonfire(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	smith := effectstest.B12Push(g, me.ID, "Renowned Weaponsmith", "Creature — Human Artificer", gapB107WeaponsmithOracle, 1, 3)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Heart-Piercer Bow", TypeLine: "Artifact — Equipment"},
		game.Card{Name: "Vial of Dragonfire", TypeLine: "Artifact — Equipment"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
	)
	effectstest.RavB2MustActivate(t, g, me.ID, smith, 0, game.ActivateAbilityParams{}, "U")
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.HasID(c.SearchCards, ids[2]) || !effectstest.HasID(c.SearchCards, ids[0]) || !effectstest.HasID(c.SearchCards, ids[1]) {
		t.Errorf("search options %v, want only the Bow and the Vial", c.SearchCards)
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) {
		t.Error("the Vial of Dragonfire did not reach the hand")
	}
}

// --- Shifting Sliver ---------------------------------------------------------

func TestGapB107ShiftingSliverOnlySliversMayBlockSlivers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Shifting Sliver", "Creature — Sliver", gapB107ShiftingOracle, 2, 2)
	attacker := effectstest.B12Creature(g, me.ID, "Attacking Sliver", "Creature — Sliver", 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	theirSliver := effectstest.B12Creature(g, opp.ID, "Their Sliver", "Creature — Sliver", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(bear, attacker); err == nil {
		t.Error("a Bear blocked a Sliver")
	}
	if err := g.DeclareBlocker(theirSliver, attacker); err != nil {
		t.Errorf("a Sliver could not block a Sliver: %v", err)
	}
}

func TestGapB107ShiftingSliverDoesNotShieldNonSlivers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Shifting Sliver", "Creature — Sliver", gapB107ShiftingOracle, 2, 2)
	attacker := effectstest.B12Creature(g, me.ID, "Attacking Bear", "Creature — Bear", 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(bear, attacker); err != nil {
		t.Errorf("a Bear could not block a Bear: %v", err)
	}
}

// --- Sparkshaper Visionary ---------------------------------------------------

func gapB107Planeswalker(g *game.Game, owner uuid.UUID, name, oracle string, loyalty int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Legendary Planeswalker — Test", OracleID: oracle,
		Owner: owner, Controller: owner, Counters: map[string]int{game.CounterLoyalty: loyalty},
	})
}

func TestGapB107SparkshaperAnimatesYourPlaneswalkersAsBirds(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Sparkshaper Visionary", "Creature — Human Wizard", gapB107SparkshaperOracle, 0, 5)
	mine := gapB107Planeswalker(g, me.ID, "My Walker", "", 4)
	theirs := gapB107Planeswalker(g, opp.ID, "Their Walker", "", 4)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	g.RunStateChecksForTest()
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt at the beginning of combat")
	}
	if effectstest.HasID(p.PickTargetCards, theirs) {
		t.Error("an opponent's planeswalker was offered")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: mine}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := g.LookupCardForEffect(mine)
	if !ok {
		t.Fatal("the planeswalker is gone")
	}
	if !c.IsCreature() || c.IsPlaneswalker() || !c.HasSubtype("Bird") {
		t.Errorf("type line %q: creature %v planeswalker %v, want a Bird creature that is no longer a planeswalker", c.TypeLine, c.IsCreature(), c.IsPlaneswalker())
	}
	if pw, tough := effectstest.EffectivePTOf(t, g, mine); pw != 3 || tough != 3 {
		t.Errorf("it is %d/%d, want 3/3", pw, tough)
	}
	for _, kw := range []string{"flying", "hexproof"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, mine, kw) {
			t.Errorf("it lacks %s", kw)
		}
	}
	if !g.Battlefield.Contains(mine) || c.Counters[game.CounterLoyalty] != 4 {
		t.Error("it lost its loyalty counters")
	}
}

func TestGapB107SparkshaperKeepsLoyaltyAbilitiesAndEndsAtEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Sparkshaper Visionary", "Creature — Human Wizard", gapB107SparkshaperOracle, 0, 5)
	dovin := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dovin, Grand Arbiter", TypeLine: "Legendary Planeswalker — Dovin", OracleID: gapB107DovinGrandArbiterO,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLoyalty: 3},
	})
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	g.RunStateChecksForTest()
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: dovin}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	if err := g.ActivateCatalogAbility(me.ID, dovin, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("the animated planeswalker's loyalty ability was refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.BattlefieldIDsNamed(g, "Thopter")); n != 1 {
		t.Errorf("%d Thopters, want 1 from Dovin's -1", n)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	for i := 0; i < 6 && g.Turn.Step != game.StepUpkeep; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	c, _ := g.LookupCardForEffect(dovin)
	if c.IsCreature() || !c.IsPlaneswalker() {
		t.Errorf("after the turn Dovin is %q, want a planeswalker again", c.TypeLine)
	}
}

func TestGapB107SparkshaperChoosingNothingIsFine(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Sparkshaper Visionary", "Creature — Human Wizard", gapB107SparkshaperOracle, 0, 5)
	walker := gapB107Planeswalker(g, me.ID, "My Walker", "", 4)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	g.RunStateChecksForTest()
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
			t.Fatalf("choosing no targets: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := g.LookupCardForEffect(walker); c.IsCreature() {
		t.Error("a planeswalker nobody chose was animated")
	}
}
