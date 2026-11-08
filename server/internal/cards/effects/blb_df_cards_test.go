package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blb_df_cards_test.go — Bloomburrow cards starting with D to F (#4467).
// Each test plays the real catalog entry and asserts what its text says.

const (
	blbDFDuoOracle          = "43ccc568-2509-49b6-99d7-c56900d1565a"
	blbDFWaveriderOracle    = "e4d52559-7624-48b2-96ba-e52ada6c507a"
	blbDFDewdropOracle      = "07634620-a83a-4ca0-8a60-2a21368a2661"
	blbDFDowndraftOracle    = "a45065e9-951a-4c6d-80b6-bd86a49f7439"
	blbDFAmbusherOracle     = "1e0a2709-2cd2-4954-81ca-60654ab03d7c"
	blbDFDruidOracle        = "9740ce6d-ef6b-4d70-839a-8bd5e0e1e0a1"
	blbDFEarlyWinterOracle  = "fbf2653e-9cc4-442f-bc73-6e2ba6e067cb"
	blbDFCrabOracle         = "3c345e24-1726-45ac-91ca-e348d1e7acfe"
	blbDFFinchOracle        = "76c2e0b4-d4fa-4e98-b363-0a76eb481304"
	blbDFGeckoOracle        = "d767463b-0fc9-43fc-a85f-e60f030f9b96"
	blbDFFlameLashOracle    = "c65e804b-9222-4fa8-88d1-df06a3279f22"
	blbDFBellOracle         = "ea4c5b8f-a5ff-47dd-8e47-5ae7b21926b4"
	blbDFSparkshooterOracle = "69038520-7156-48d9-aae8-ba3cc76a5993"
	blbDFCrabTypeLine       = "Creature — Elemental Crab"
	blbDFGeckoTypeLine      = "Creature — Lizard Warlock"
)

func TestBlbDFEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		blbDFDuoOracle:          "Daggerfang Duo",
		blbDFWaveriderOracle:    "Daring Waverider",
		blbDFDewdropOracle:      "Dewdrop Cure",
		blbDFDowndraftOracle:    "Dire Downdraft",
		blbDFAmbusherOracle:     "Downwind Ambusher",
		blbDFDruidOracle:        "Druid of the Spade",
		blbDFEarlyWinterOracle:  "Early Winter",
		blbDFCrabOracle:         "Eddymurk Crab",
		blbDFFinchOracle:        "Finch Formation",
		blbDFGeckoOracle:        "Flamecache Gecko",
		blbDFFlameLashOracle:    "Flame Lash",
		blbDFBellOracle:         "Fountainport Bell",
		blbDFSparkshooterOracle: "Frilled Sparkshooter",
	} {
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

// blbDFOpponentLosesLife makes `p` lose one life, as if dealt damage.
func blbDFOpponentLosesLife(t *testing.T, g *game.Game, p uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, p, -1); err != nil {
			t.Fatalf("lose life: %v", err)
		}
	})
}

// --- Daggerfang Duo ----------------------------------------------------------

func TestBlbDFDaggerfangDuoMayMillTwo(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := newCatalogGame(t)
		me, _, _ := blbB1Seats(g)
		library := me.Library.Size()
		duo := castHoldCreature(t, g, "Daggerfang Duo", "Creature — Rat Squirrel", blbDFDuoOracle, 3, 2)
		passPriorityAroundTable(t, g)
		answerLatestTriggerPrompt(t, g, me.ID, accept)
		passPriorityAroundTable(t, g)
		want := 0
		if accept {
			want = 2
		}
		if got := library - me.Library.Size(); got != want {
			t.Errorf("accept=%v: milled %d cards, want %d", accept, got, want)
		}
		if !hasKeywordOnBattlefield(t, g, duo, "deathtouch") {
			t.Error("the Duo lacks deathtouch")
		}
	}
}

// --- Daring Waverider --------------------------------------------------------

func TestBlbDFDaringWaveriderOffersOnlyCheapInstantsAndSorceries(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	cheap := gyCard(me, "Old Sorcery", "Sorcery", "{2}{U}", "")
	instant := gyCard(me, "Old Instant", "Instant", "{3}{U}", "")
	dear := gyCard(me, "Big Sorcery", "Sorcery", "{4}{U}", "")
	creature := gyCard(me, "Old Bear", "Creature — Bear", "{1}{G}", "")
	castHoldCreature(t, g, "Daring Waverider", "Creature — Otter Wizard", blbDFWaveriderOracle, 4, 4)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	for _, id := range []uuid.UUID{cheap, instant} {
		if !hasID(p.PickTargetCards, id) {
			t.Errorf("a legal card was not offered: %v", id)
		}
	}
	for _, id := range []uuid.UUID{dear, creature} {
		if hasID(p.PickTargetCards, id) {
			t.Errorf("an illegal card was offered: %v", id)
		}
	}
}

func TestBlbDFDaringWaveriderCastsItFreeAndExilesIt(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	sorcery := gyCard(me, "Old Sorcery", "Sorcery", "{2}{U}", "")
	castHoldCreature(t, g, "Daring Waverider", "Creature — Otter Wizard", blbDFWaveriderOracle, 4, 4)
	pickCard(t, g, me.ID, sorcery)
	passPriorityAroundTable(t, g)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, sorcery)
	if err := castThroughResolutionCast(t, g, me.ID, sorcery, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("casting for free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(sorcery) || !inExile(g, sorcery) {
		t.Error("the sorcery should be exiled, not back in the graveyard")
	}
}

// --- Dewdrop Cure ------------------------------------------------------------

func TestBlbDFDewdropCureReturnsUpToTwoUnpromised(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	a := gyCard(me, "Two Drop", "Creature — Bear", "{1}{W}", "")
	b := gyCard(me, "One Drop", "Creature — Mouse", "{W}", "")
	c := gyCard(me, "Other Two Drop", "Creature — Bear", "{1}{G}", "")
	big := gyCard(me, "Three Drop", "Creature — Beast", "{2}{G}", "")
	if _, err := castWithGift(t, g, "Dewdrop Cure", "Sorcery", blbDFDewdropOracle, cardTarget(big), uuid.Nil); err == nil {
		t.Fatal("a mana value 3 creature card was accepted")
	}
	if _, err := castWithGift(t, g, "Dewdrop Cure", "Sorcery", blbDFDewdropOracle, cardRefs(a, b, c), uuid.Nil); err == nil {
		t.Fatal("three targets were accepted without the gift")
	}
	mustCastWithGift(t, g, "Dewdrop Cure", "Sorcery", blbDFDewdropOracle, cardRefs(a, b), uuid.Nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) || g.Battlefield.Contains(c) {
		t.Error("the two chosen creature cards should return, and only them")
	}
}

func TestBlbDFDewdropCurePromisedReturnsThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	a := gyCard(me, "Two Drop", "Creature — Bear", "{1}{W}", "")
	b := gyCard(me, "One Drop", "Creature — Mouse", "{W}", "")
	c := gyCard(me, "Other Two Drop", "Creature — Bear", "{1}{G}", "")
	oppHand := opp.Hand.Size()
	mustCastWithGift(t, g, "Dewdrop Cure", "Sorcery", blbDFDewdropOracle, cardRefs(a, b, c), opp.ID)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%v did not return with the gift promised", id)
		}
	}
	if got := opp.Hand.Size(); got != oppHand+1 {
		t.Errorf("the gift's opponent has %d cards, want %d", got, oppHand+1)
	}
}

// --- Dire Downdraft ----------------------------------------------------------

func TestBlbDFDireDowndraftCostsOneLessAtATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	idle := b12Creature(g, opp.ID, "Idle", "Creature — Bear", 2, 2)
	tapped := b12Creature(g, opp.ID, "Tapped", "Creature — Bear", 2, 2)
	b16Tap(g, tapped)
	advanceToMain(t, g)
	card := game.Card{Name: "Dire Downdraft", TypeLine: "Instant", ManaCost: "{3}{U}", OracleID: blbDFDowndraftOracle}
	if err := socB19CastWithPool(g, me, card, "U", 3, game.CastSpellParams{Targets: cardRefs(idle)}); err == nil {
		t.Fatal("cast for {2}{U} at an untapped creature that is not attacking")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	card.InstanceID = uuid.New()
	if err := socB19CastWithPool(g, me, card, "U", 3, game.CastSpellParams{Targets: cardRefs(tapped)}); err != nil {
		t.Fatalf("cast for {2}{U} at a tapped creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID) != nil {
		t.Fatal("the caster was asked; the owner chooses")
	}
	c := latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, opp.ID)
	if c == nil {
		t.Fatalf("the owner was not asked: %+v", g.PendingChoices)
	}
	if err := g.ResolvePutInLibrary(c.ID, opp.ID, []uuid.UUID{tapped}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if g.Battlefield.Contains(tapped) || opp.Library.Cards[0].InstanceID != tapped {
		t.Error("the creature should be on the bottom of its owner's library")
	}
}

// --- Downwind Ambusher -------------------------------------------------------

func blbDFAmbusherMode(t *testing.T, g *game.Game, me uuid.UUID, mode int) {
	t.Helper()
	castRingCreature(t, g, "Downwind Ambusher", "Creature — Skunk Assassin", blbDFAmbusherOracle, 4, 2)
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me)
	if pick == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(pick.ID, me, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
}

func TestBlbDFDownwindAmbusherShrinksAnOpposingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	mine := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	blbDFAmbusherMode(t, g, me.ID, 0)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
		t.Fatal("my own creature was accepted; the target is a creature an opponent controls")
	}
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if pw, tough := ptOf(t, g, theirs); pw != 1 || tough != 1 {
		t.Errorf("the opposing creature is %d/%d, want 1/1", pw, tough)
	}
}

func TestBlbDFDownwindAmbusherDestroysADamagedOpposingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	hurt := blbB1Push(g, opp.ID, "Hurt Bear", "Creature — Bear", "", 3, 3)
	fresh := blbB1Push(g, opp.ID, "Fresh Bear", "Creature — Bear", "", 3, 3)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, hurt, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	blbDFAmbusherMode(t, g, me.ID, 1)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if hasID(p.PickTargetCards, fresh) {
		t.Errorf("an undamaged creature was offered: %v", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, hurt)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(hurt) || !g.Battlefield.Contains(fresh) {
		t.Error("only the damaged creature should be destroyed")
	}
}

// --- Druid of the Spade ------------------------------------------------------

func TestBlbDFDruidOfTheSpadeGrowsWhileYouControlAToken(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	druid := blbB1Push(g, me.ID, "Druid of the Spade", "Creature — Rabbit Druid", blbDFDruidOracle, 2, 3)
	blbB1Push(g, opp.ID, "Their Squirrel", "Token Creature — Squirrel", "", 1, 1)
	if pw, tough := ptOf(t, g, druid); pw != 2 || tough != 3 {
		t.Fatalf("with only an opponent's token, the Druid is %d/%d, want 2/3", pw, tough)
	}
	if hasKeywordOnBattlefield(t, g, druid, "trample") {
		t.Fatal("with only an opponent's token, the Druid has trample")
	}
	blbB1Push(g, me.ID, "Rabbit", "Token Creature — Rabbit", "", 1, 1)
	if pw, tough := ptOf(t, g, druid); pw != 4 || tough != 3 {
		t.Errorf("with a token of mine, the Druid is %d/%d, want 4/3", pw, tough)
	}
	if !hasKeywordOnBattlefield(t, g, druid, "trample") {
		t.Error("with a token of mine, the Druid lacks trample")
	}
}

// --- Early Winter ------------------------------------------------------------

func blbDFCastEarlyWinter(t *testing.T, g *game.Game, mode int, target game.TargetRef) error {
	t.Helper()
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := pushCatalogHandCard(me, "Early Winter", "Instant", blbDFEarlyWinterOracle)
	return g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{mode}, Targets: []game.TargetRef{target}})
}

func TestBlbDFEarlyWinterExilesACreature(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	bear := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	if err := blbDFCastEarlyWinter(t, g, 0, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(bear) {
		t.Error("the creature was not exiled")
	}
}

func TestBlbDFEarlyWinterLetsTheOpponentChooseAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	keep := b12Permanent(g, opp.ID, "Their Kept Enchantment", "Enchantment")
	lose := b12Permanent(g, opp.ID, "Their Lost Enchantment", "Enchantment")
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if err := blbDFCastEarlyWinter(t, g, 1, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err == nil {
		t.Fatal("Early Winter targeted its own caster")
	}
	if err := blbDFCastEarlyWinter(t, g, 1, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("the caster was asked to choose: the opponent chooses")
	}
	pick := chooseCardsChoiceFor(g, opp.ID)
	if pick == nil || !hasID(pick.ChooseCards, keep) || !hasID(pick.ChooseCards, lose) || hasID(pick.ChooseCards, rock) {
		t.Fatalf("the opponent's pick %+v should offer exactly their enchantments", pick)
	}
	answerChooseCards(t, g, opp.ID, lose)
	if !g.Exile.Contains(lose) || !g.Battlefield.Contains(keep) || !g.Battlefield.Contains(rock) {
		t.Error("only the chosen enchantment should be exiled")
	}
}

// --- Eddymurk Crab -----------------------------------------------------------

func TestBlbDFEddymurkCrabCostsOneLessPerInstantAndSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	advanceToMain(t, g)
	gyCard(me, "Old Instant", "Instant", "{U}", "")
	gyCard(me, "Old Sorcery", "Sorcery", "{U}", "")
	gyCard(me, "Old Bear", "Creature — Bear", "{1}{G}", "")
	card := game.Card{Name: "Eddymurk Crab", TypeLine: blbDFCrabTypeLine, ManaCost: "{5}{U}{U}", OracleID: blbDFCrabOracle, Power: 5, Toughness: 5}
	if err := socB19CastWithPool(g, me, card, "U", 4, game.CastSpellParams{}); err == nil {
		t.Fatal("cast for four mana with two instant and sorcery cards in the graveyard")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	card.InstanceID = uuid.New()
	if err := socB19CastWithPool(g, me, card, "U", 5, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast for {5}{U}{U} minus 2: %v", err)
	}
}

func TestBlbDFEddymurkCrabEntersUntappedOnYourTurnAndTapsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	a := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	b := blbB1Push(g, opp.ID, "Their Other Bear", "Creature — Bear", "", 2, 2)
	crab := castHoldCreature(t, g, "Eddymurk Crab", blbDFCrabTypeLine, blbDFCrabOracle, 5, 5)
	if tappedOnBattlefield(t, g, crab) {
		t.Error("the Crab entered tapped on its controller's turn")
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the tap")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, cardRefs(a, b)); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !tappedOnBattlefield(t, g, a) || !tappedOnBattlefield(t, g, b) {
		t.Error("both target creatures should be tapped")
	}
}

func TestBlbDFEddymurkCrabEntersTappedOnAnotherPlayersTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMainOf(t, g, 1)
	crab := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: crab, Name: "Eddymurk Crab", TypeLine: blbDFCrabTypeLine, OracleID: blbDFCrabOracle,
		Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, crab, game.CastSpellParams{}); err != nil {
		t.Fatalf("flash the Crab in on another player's turn: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(crab) {
		t.Fatal("the Crab did not enter")
	}
	if !tappedOnBattlefield(t, g, crab) {
		t.Error("the Crab entered untapped on another player's turn")
	}
}

// --- Finch Formation ---------------------------------------------------------

func TestBlbDFFinchFormationGivesACreatureYouControlFlying(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	finch := castHoldCreature(t, g, "Finch Formation", "Creature — Bird Scout", blbDFFinchOracle, 2, 2)
	if !hasKeywordOnBattlefield(t, g, finch, "flying") {
		t.Error("the Formation lacks flying")
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
		t.Fatal("an opponent's creature was accepted; the target is a creature you control")
	}
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Error("the target did not gain flying")
	}
}

// --- Flamecache Gecko --------------------------------------------------------

func TestBlbDFFlamecacheGeckoAddsManaOnlyAfterAnOpponentLostLife(t *testing.T) {
	for _, lost := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp, _ := blbB1Seats(g)
		if lost {
			blbDFOpponentLosesLife(t, g, opp.ID)
		}
		castHoldCreature(t, g, "Flamecache Gecko", blbDFGeckoTypeLine, blbDFGeckoOracle, 2, 2)
		passPriorityAroundTable(t, g)
		want := 0
		if lost {
			want = 1
		}
		if b, r := mcPoolCount(me, "B"), mcPoolCount(me, "R"); b != want || r != want {
			t.Errorf("lost=%v: pool has %d {B} and %d {R}, want %d of each", lost, b, r, want)
		}
	}
}

func TestBlbDFFlamecacheGeckoRummages(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	gecko := blbB1Push(g, me.ID, "Flamecache Gecko", blbDFGeckoTypeLine, blbDFGeckoOracle, 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	socB21SetHand(me, 2)
	pitched := me.Hand.Cards[0].InstanceID
	lib := me.Library.Size()
	floatForTest(g, me, "CR")
	if err := g.ActivateCatalogAbility(me.ID, gecko, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{pitched}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pitched) {
		t.Error("the discarded card is not in the graveyard")
	}
	if me.Hand.Size() != 2 || me.Library.Size() != lib-1 {
		t.Errorf("hand %d, library %d -> %d: a card should have been drawn", me.Hand.Size(), lib, me.Library.Size())
	}
}

// --- Flame Lash --------------------------------------------------------------

func TestBlbDFFlameLashDealsFourToAnyTarget(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	life := lifeOf(g, opp.ID)
	castCatalogSpell(t, g, "Flame Lash", "Instant", blbDFFlameLashOracle, b16TargetPlayer(opp.ID))
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != life-4 {
		t.Errorf("the opponent is at %d, want %d", got, life-4)
	}
	big := blbB1Push(g, opp.ID, "Their Big", "Creature — Beast", "", 4, 4)
	castCatalogSpell(t, g, "Flame Lash", "Instant", blbDFFlameLashOracle, cardTarget(big))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("a 4/4 survived 4 damage")
	}
}

// --- Fountainport Bell -------------------------------------------------------

func TestBlbDFFountainportBellPutsABasicLandOnTopThenCycles(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Sol Ring", "Artifact"),
		searchTestLand("Plains", "Basic Land — Plains"),
	)
	bell := castHoldCreature(t, g, "Fountainport Bell", "Artifact", blbDFBellOracle, 0, 0)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Sol Ring") != uuid.Nil {
		t.Error("only basic land cards are legal finds")
	}
	answerSearchNamed(t, g, me.ID, "Plains")
	passPriorityAroundTable(t, g)
	if top := me.Library.Cards[len(me.Library.Cards)-1]; top.Name != "Plains" {
		t.Fatalf("top card %q, want Plains", top.Name)
	}
	hand := me.Hand.Size()
	floatForTest(g, me, "C")
	b16Activate(t, g, me.ID, bell, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bell) {
		t.Error("the Bell was not sacrificed")
	}
	if got := me.Hand.Size(); got != hand+1 || blbDFNamedInHand(me, "Plains") == uuid.Nil {
		t.Errorf("hand %d, want %d with the Plains drawn", got, hand+1)
	}
}

// blbDFNamedInHand is the ID of the first card in p's hand named name.
func blbDFNamedInHand(p *game.Player, name string) uuid.UUID {
	for _, c := range p.Hand.Cards {
		if c.Name == name {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

// --- Frilled Sparkshooter ----------------------------------------------------

func TestBlbDFFrilledSparkshooterEntersWithACounterAfterAnOpponentLostLife(t *testing.T) {
	for _, lost := range []bool{false, true} {
		g := newCatalogGame(t)
		_, opp, _ := blbB1Seats(g)
		if lost {
			blbDFOpponentLosesLife(t, g, opp.ID)
		}
		shooter := castHoldCreature(t, g, "Frilled Sparkshooter", "Creature — Lizard Archer", blbDFSparkshooterOracle, 3, 3)
		want := 3
		if lost {
			want = 4
		}
		if pw, tough := ptOf(t, g, shooter); pw != want || tough != want {
			t.Errorf("lost=%v: the Sparkshooter is %d/%d, want %d/%d", lost, pw, tough, want, want)
		}
		for _, kw := range []string{"reach", "menace"} {
			if !hasKeywordOnBattlefield(t, g, shooter, kw) {
				t.Errorf("the Sparkshooter lacks %s", kw)
			}
		}
	}
}
