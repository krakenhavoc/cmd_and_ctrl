package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b30_d_cards_test.go — Avatar: The Last Airbender Eternal, batch
// 33's cards in the combined batch 30 (#5075). Each test plays the real
// catalog entry and asserts what its text says.

const (
	avB30dDiresightOracle    = "a7b12df9-fddd-4cd4-8c09-d54a47867ab6"
	avB30dFistsOracle        = "e05abb6b-e9f4-4d9e-ad1e-7805ebc09914"
	avB30dPathOracle         = "ba13cbdb-8043-4c1b-8a85-084c86886ac8"
	avB30dJetOracle          = "b9ca8ead-c267-4170-a6c8-04498330c82b"
	avB30dSukiOracle         = "558513bb-a76d-434a-98fc-3707a5d8886a"
	avB30dBaboonOracle       = "eb036e02-f43e-46a2-90f1-931579467c9b"
	avB30dBlueSpiritOracle   = "497e9c7c-a16d-43e5-a7ec-04153aa2566a"
	avB30dTuiAndLaOracle     = "781c9391-0322-40ad-8050-07151e2e6fc7"
	avB30dAzulaOracle        = "7139ead5-86f0-46ae-8a6d-ad66d0408f73"
	avB30dAzulaTypeLine      = "Legendary Creature — Human Noble"
	avB30dBlueSpiritTypeLine = "Legendary Creature — Human Rogue Ally"
)

func TestAvatarB30dCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		avB30dDiresightOracle:  "Diresight",
		avB30dFistsOracle:      "Fists of Flame",
		avB30dPathOracle:       "Descendants' Path",
		avB30dJetOracle:        "Jet, Rebel Leader",
		avB30dSukiOracle:       "Suki, Kyoshi Captain",
		avB30dBaboonOracle:     "Baboon Spirit",
		avB30dBlueSpiritOracle: "The Blue Spirit",
		avB30dTuiAndLaOracle:   "Tui and La, Moon and Ocean",
		avB30dAzulaOracle:      "Azula, Ruthless Firebender",
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

// --- Diresight ------------------------------------------------------------------

func TestAvatarB30DiresightSurveilsThenDrawsTwoAndLosesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedLibrary(me, "Bin Me", "Keep Me", "Third", "Fourth")
	hand, life := me.Hand.Size(), me.Life
	effectstest.CastCatalogSpell(t, g, "Diresight", "Sorcery", avB30dDiresightOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("surveil prompt %+v, want one over two cards", c)
	}
	if me.Hand.Size() != hand {
		t.Fatal("drew before the surveil was answered")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{ids[0]}, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(ids[0]) {
		t.Error("the surveilled card is not in the graveyard")
	}
	if me.Hand.Size() != hand+2 || !me.Hand.Contains(ids[1]) || !me.Hand.Contains(ids[2]) {
		t.Errorf("hand %d (want %d) should hold the kept card and the next one", me.Hand.Size(), hand+2)
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
}

// --- Fists of Flame ---------------------------------------------------------------

func TestAvatarB30FistsOfFlameCountsEveryCardDrawnThisTurnLockedIn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	before := g.TurnTallyFor(me.ID).CardsDrawn
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Fists of Flame", "Instant", avB30dFistsOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 { // the spell left, one card drawn
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
	want := 2 + before + 1
	if got := effectstest.EffectivePower(t, g, bear); got != want {
		t.Errorf("power %d, want %d (+1/+0 for each of %d cards drawn)", got, want, before+1)
	}
	if effectstest.EffectiveToughness(t, g, bear) != 2 {
		t.Error("the toughness changed")
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("no trample")
	}
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	if got := effectstest.EffectivePower(t, g, bear); got != want {
		t.Errorf("power %d after a later draw, want %d (locked in)", got, want)
	}
}

func TestAvatarB30FistsOfFlameDrawsNothingWhenItsTargetIsGone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "One", "Two")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Fists of Flame", "Instant", avB30dFistsOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d: a spell with an illegal target does not resolve", me.Hand.Size(), hand)
	}
}

// --- Descendants' Path ----------------------------------------------------------------

func avB30dPathGame(t *testing.T, top game.Card) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Descendants' Path", "Enchantment", avB30dPathOracle, false)
	effectstest.B12Creature(g, me.ID, "Elf Scout", "Creature — Elf Scout", 1, 1)
	ids := effectstest.TrekB31Library(g, me, top, game.Card{Name: "Second", TypeLine: "Land"})
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, ids[0]
}

func TestAvatarB30DescendantsPathCastsASharedTypeCreatureFree(t *testing.T) {
	g, me, top := avB30dPathGame(t, game.Card{Name: "Elf Warrior", TypeLine: "Creature — Elf Warrior", ManaCost: "{3}{G}", Power: 3, Toughness: 3})
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, top)
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, top, game.CastSpellParams{}); err != nil {
		t.Fatalf("casting the revealed card: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, top) != game.ZoneBattlefield {
		t.Fatalf("the cast creature is in %v", effectstest.B12ZoneOf(g, top))
	}
}

func TestAvatarB30DescendantsPathBottomsADeclinedCard(t *testing.T) {
	g, me, top := avB30dPathGame(t, game.Card{Name: "Elf Warrior", TypeLine: "Creature — Elf Warrior", ManaCost: "{3}{G}", Power: 3, Toughness: 3})
	effectstest.DeclineResolutionCast(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, top) != game.ZoneLibrary || me.Library.Cards[0].InstanceID != top {
		t.Error("the declined card should be on the bottom of the library")
	}
}

func TestAvatarB30DescendantsPathBottomsACardThatSharesNoType(t *testing.T) {
	for _, top := range []game.Card{
		{Name: "Goblin Raider", TypeLine: "Creature — Goblin", ManaCost: "{1}{R}", Power: 2, Toughness: 1},
		{Name: "Elf Sorcery", TypeLine: "Kindred Sorcery — Elf", ManaCost: "{G}"},
	} {
		g, me, id := avB30dPathGame(t, top)
		if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
			t.Errorf("%s: a cast was offered", top.Name)
		}
		if effectstest.B12ZoneOf(g, id) != game.ZoneLibrary || me.Library.Cards[0].InstanceID != id {
			t.Errorf("%s: should be on the bottom of the library", top.Name)
		}
	}
}

// --- Jet, Rebel Leader ------------------------------------------------------------------

func TestAvatarB30JetPutsASmallCreatureAttackingAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jet := effectstest.B27Push(g, me.ID, "Jet, Rebel Leader", "Legendary Creature — Human Rebel Ally", avB30dJetOracle, "{3}{W}", 3, 4, "W")
	ids := effectstest.TrekB31Library(g, me,
		game.Card{Name: "Small Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Big Giant", TypeLine: "Creature — Giant", ManaCost: "{3}{R}", Power: 4, Toughness: 4},
		game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"},
		game.Card{Name: "Three Drop", TypeLine: "Creature — Soldier", ManaCost: "{2}{W}", Power: 3, Toughness: 3},
		game.Card{Name: "Shock", TypeLine: "Instant", ManaCost: "{R}"},
		game.Card{Name: "Sixth", TypeLine: "Creature — Bear", ManaCost: "{G}", Power: 1, Toughness: 1},
	)
	effectstest.DeclareAttack(t, g, opp.ID, jet)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick prompt")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	if !offered[ids[0]] || !offered[ids[3]] || offered[ids[1]] || offered[ids[2]] || offered[ids[4]] || offered[ids[5]] {
		t.Errorf("offered %v: want only the two creature cards of mana value 3 or less among the top five", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, ids[0])
	effectstest.TmtB10AnswerAttackTarget(t, g, me.ID)
	c, ok := effectstest.BattlefieldCard(g, ids[0])
	if !ok {
		t.Fatalf("the chosen card is in %v", effectstest.B12ZoneOf(g, ids[0]))
	}
	if !c.Tapped || c.AttackingTarget == uuid.Nil {
		t.Errorf("tapped %v, attacking %v; want tapped and attacking", c.Tapped, c.AttackingTarget)
	}
	if top, err := me.Library.Top(); err != nil || top.InstanceID != ids[5] {
		t.Error("the sixth card should now be on top")
	}
	bottom := map[uuid.UUID]bool{}
	for _, card := range me.Library.Cards[:4] {
		bottom[card.InstanceID] = true
	}
	for _, id := range ids[1:5] {
		if !bottom[id] {
			t.Errorf("%v is not among the bottom four", id)
		}
	}
}

func TestAvatarB30JetMayPutNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jet := effectstest.B27Push(g, me.ID, "Jet, Rebel Leader", "Legendary Creature — Human Rebel Ally", avB30dJetOracle, "{3}{W}", 3, 4, "W")
	ids := effectstest.TrekB31Library(g, me,
		game.Card{Name: "Small Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"},
	)
	effectstest.DeclareAttack(t, g, opp.ID, jet)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	if effectstest.B12ZoneOf(g, ids[0]) != game.ZoneLibrary || me.Library.Size() != 2 {
		t.Error("declining should leave both cards in the library")
	}
}

// --- Suki, Kyoshi Captain -----------------------------------------------------------------

func TestAvatarB30SukiPumpsOtherWarriorsAndGivesAttackersDoubleStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	suki := effectstest.B27Push(g, me.ID, "Suki, Kyoshi Captain", "Legendary Creature — Human Warrior Ally", avB30dSukiOracle, "{2}{W}", 3, 3, "W")
	attacker := effectstest.B12Creature(g, me.ID, "Kyoshi Warrior", "Creature — Human Warrior", 2, 2)
	home := effectstest.B12Creature(g, me.ID, "Home Warrior", "Creature — Human Warrior", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Warrior", "Creature — Warrior", 2, 2)
	if effectstest.EffectivePower(t, g, attacker) != 3 || effectstest.EffectiveToughness(t, g, attacker) != 3 {
		t.Error("another Warrior you control should be 3/3")
	}
	if effectstest.EffectivePower(t, g, suki) != 3 || effectstest.EffectivePower(t, g, bear) != 2 || effectstest.EffectivePower(t, g, theirs) != 2 {
		t.Error("only other Warriors you control are pumped")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{suki, attacker, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.FloatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, suki, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without the {W}")
	}
	effectstest.FloatForTest(g, me, "W")
	effectstest.B16Activate(t, g, me.ID, suki, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]bool{suki: true, attacker: true, home: false, bear: false, theirs: false} {
		if got := effectstest.HasEffectiveKeyword(t, g, id, "double strike"); got != want {
			t.Errorf("%v double strike %v, want %v", id, got, want)
		}
	}
}

// --- Baboon Spirit ---------------------------------------------------------------------------

func TestAvatarB30BaboonSpiritMakesASpiritForAnotherNontokenSpirit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Baboon Spirit", "Creature — Monkey Spirit", avB30dBaboonOracle, false)
	effectstest.TrekB31Enter(t, g, me, game.Card{Name: "Wandering Spirit", TypeLine: "Creature — Spirit", Power: 1, Toughness: 1})
	effectstest.PassPriorityAroundTable(t, g)
	if got := avB24Count(g, me.ID, "Spirit"); got != 1 {
		t.Fatalf("%d Spirit tokens, want 1", got)
	}
	var token game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spirit" && c.IsToken() {
			token = c
		}
	}
	if !token.IsCreature() || token.Power != 1 || token.Toughness != 1 || len(token.Colors) != 0 {
		t.Errorf("token %+v, want a 1/1 colorless Spirit creature", token)
	}
	// A non-Spirit and a token Spirit make nothing: one more Spirit, the
	// token made here, and no new one.
	effectstest.TrekB31Enter(t, g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.KurukSpiritToken(), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := avB24Count(g, me.ID, "Spirit"); got != 2 {
		t.Errorf("%d Spirit tokens after the Bear and a token Spirit, want 2", got)
	}
}

func TestAvatarB30BaboonSpiritBlinksAnotherCreatureUntilTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	baboon := effectstest.PushCatalogPermanent(g, me.ID, "Baboon Spirit", "Creature — Monkey Spirit", avB30dBaboonOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CCCU")
	for _, bad := range []uuid.UUID{baboon, theirs} {
		if err := g.ActivateCatalogAbility(me.ID, baboon, 0, game.ActivateAbilityParams{Strict: true,
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bad}}}); err == nil {
			t.Fatalf("targeted %v, which is not another creature you control", bad)
		}
	}
	effectstest.B16Activate(t, g, me.ID, baboon, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || effectstest.B12ZoneOf(g, bear) != game.ZoneExile {
		t.Fatal("the Bear was not exiled")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByName(g, "Bear") == false {
		t.Error("the Bear did not return at the end step")
	}
}

// --- The Blue Spirit ---------------------------------------------------------------------------

func TestAvatarB30BlueSpiritFlashesTheFirstCreatureAndDrawsInCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	effectstest.PushCatalogPermanent(g, me.ID, "The Blue Spirit", avB30dBlueSpiritTypeLine, avB30dBlueSpiritOracle, false)
	first := effectstest.MayhemHandCard(g, me, "Elf", "Creature — Elf", "{G}", "")
	second := effectstest.MayhemHandCard(g, me, "Bear", "Creature — Bear", "{G}", "")
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FloatForTest(g, me, "GG")
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the first creature spell as though it had flash: %v", err)
	}
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(first) {
		t.Fatal("the Elf did not enter")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d: a nontoken creature entered during combat", me.Hand.Size(), hand+1)
	}
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("the second creature spell of the turn was cast at instant speed")
	}
}

func TestAvatarB30BlueSpiritIgnoresTokensAndMainPhaseEntries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	effectstest.PushCatalogPermanent(g, me.ID, "The Blue Spirit", avB30dBlueSpiritTypeLine, avB30dBlueSpiritOracle, false)
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.TrekB31Enter(t, g, me, game.Card{Name: "Main Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatalf("hand %d, want %d: the creature entered in a main phase", me.Hand.Size(), hand)
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	hand = me.Hand.Size()
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 white Ally"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d: a token entered", me.Hand.Size(), hand)
	}
}

// --- Tui and La, Moon and Ocean -------------------------------------------------------------------

func TestAvatarB30TuiAndLaDrawOnTapAndGrowOnUntap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	tl := effectstest.PushCatalogPermanent(g, me.ID, "Tui and La, Moon and Ocean", "Legendary Creature — Fish Spirit", avB30dTuiAndLaOracle, false)
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(tl) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d after becoming tapped", me.Hand.Size(), hand+1)
	}
	if effectstest.CounterCount(g, tl, "+1/+1") != 0 {
		t.Error("a counter before untapping")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(tl) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, tl, "+1/+1"); got != 1 {
		t.Errorf("%d +1/+1 counters after untapping, want 1", got)
	}
	// Untapping an untapped permanent is not becoming untapped.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(tl) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, tl, "+1/+1"); got != 1 {
		t.Errorf("%d counters, want 1", got)
	}
}

func TestAvatarB30TuiAndLaDrawWhenTheyAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	tl := effectstest.PushCatalogPermanent(g, me.ID, "Tui and La, Moon and Ocean", "Legendary Creature — Fish Spirit", avB30dTuiAndLaOracle, false)
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, tl)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d: attacking taps them", me.Hand.Size(), hand+1)
	}
}

// --- Azula, Ruthless Firebender -----------------------------------------------------------------------

func avB30dExperience(g *game.Game, p *game.Player) int {
	var n int
	g.ReadSnapshot(func() { n = p.Counters[game.CounterExperience] })
	return n
}

// avB30dToAzulaDiscard resolves the attack triggers up to Azula's
// optional discard.
func avB30dToAzulaDiscard(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	// Firebending and the discard trigger are two triggers of Azula's:
	// any order reaches the discard.
	for i := 0; i < 4 && effectstest.DiscardChoiceFor(g, me.ID) == nil; i++ {
		effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		for _, c := range g.PendingChoices {
			if c != nil {
				t.Logf("pending: %s for %v", c.Kind, c.Chooser)
			}
		}
		t.Fatal("no discard prompt")
	}
}

func TestAvatarB30AzulaFirebendsAndCountsEachPlayerWhoDiscarded(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	azula := effectstest.B27Push(g, me.ID, "Azula, Ruthless Firebender", avB30dAzulaTypeLine, avB30dAzulaOracle, "{2}{B}", 3, 3, "B")
	effectstest.AdvanceToMain(t, g)
	pitch := effectstest.MayhemHandCard(g, opp, "Their Card", "Instant", "{U}", "")
	effectstest.MayhemDiscard(t, g, opp, pitch)
	mine := effectstest.MayhemHandCard(g, me, "My Card", "Instant", "{U}", "")
	effectstest.DeclareAttack(t, g, opp.ID, azula)
	avB30dToAzulaDiscard(t, g, me)
	effectstest.AnswerDiscard(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(mine) {
		t.Error("the card was not discarded")
	}
	if got := avB30dExperience(g, me); got != 2 {
		t.Errorf("%d experience counters, want 2 (you and the opponent discarded)", got)
	}
	if got := effectstest.PoolCount(me, "R"); got != 1 {
		t.Errorf("%d {R} floating after firebending 1, want 1", got)
	}
}

func TestAvatarB30AzulaDecliningStillCountsOthers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	azula := effectstest.B27Push(g, me.ID, "Azula, Ruthless Firebender", avB30dAzulaTypeLine, avB30dAzulaOracle, "{2}{B}", 3, 3, "B")
	effectstest.AdvanceToMain(t, g)
	pitch := effectstest.MayhemHandCard(g, opp, "Their Card", "Instant", "{U}", "")
	effectstest.MayhemDiscard(t, g, opp, pitch)
	keep := effectstest.MayhemHandCard(g, me, "My Card", "Instant", "{U}", "")
	effectstest.DeclareAttack(t, g, opp.ID, azula)
	avB30dToAzulaDiscard(t, g, me)
	effectstest.AnswerDiscard(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(keep) {
		t.Error("the card left the hand")
	}
	if got := avB30dExperience(g, me); got != 1 {
		t.Errorf("%d experience counters, want 1 (only the opponent discarded)", got)
	}
}

func TestAvatarB30AzulaPumpsPerExperienceCounterAndGainsMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	azula := effectstest.B27Push(g, me.ID, "Azula, Ruthless Firebender", avB30dAzulaTypeLine, avB30dAzulaOracle, "{2}{B}", 3, 3, "B")
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.AddPlayerCounterByForEffect(me.ID, me.ID, game.CounterExperience, 3) })
	effectstest.FloatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, azula, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without {B}")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, "CCB")
	effectstest.B16Activate(t, g, me.ID, azula, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EffectivePower(t, g, azula) != 6 || effectstest.EffectiveToughness(t, g, azula) != 6 {
		t.Errorf("Azula %d/%d, want 6/6", effectstest.EffectivePower(t, g, azula), effectstest.EffectiveToughness(t, g, azula))
	}
	if !effectstest.HasEffectiveKeyword(t, g, azula, "menace") {
		t.Error("no menace")
	}
	g.WithWriteLock(func() { _ = g.AddPlayerCounterByForEffect(me.ID, me.ID, game.CounterExperience, 1) })
	if effectstest.EffectivePower(t, g, azula) != 6 {
		t.Error("a later experience counter changed the locked-in pump")
	}
}
