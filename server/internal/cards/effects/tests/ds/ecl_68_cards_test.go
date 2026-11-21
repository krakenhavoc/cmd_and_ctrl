package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ecl_68_cards_test.go — Lorwyn Eclipsed, batch 68 (#4458).

const (
	ecl68CavalryOracle   = "6e04a375-303e-43ba-bca1-4e959711a7f6"
	ecl68GrowthOracle    = "c63402b2-1f8c-4004-82d8-48ea26bfe3fc"
	ecl68SpryOracle      = "04cc9177-1ffc-4ac8-a64d-7995b4f33d23"
	ecl68FarrierOracle   = "ef67c170-68aa-4ada-8cee-261a4dbb1b5d"
	ecl68SprigsOracle    = "91363d71-eef0-48e8-9a68-8f363ff2790e"
	ecl68ChargeOracle    = "a8cdcaa0-8e25-4c17-b5ce-16f72b535123"
	ecl68TrystanOracle   = "961d1128-a022-480c-918e-16ba0abf844b"
	ecl68AimOracle       = "be7e220e-272e-45ca-8265-8d78da2cd2a7"
	ecl68BrawlerOracle   = "5fffd840-1e6e-4f5d-b847-802f9a25eb5c"
	ecl68WildvineOracle  = "6c34674f-856b-4062-87d5-6ed73d9a9855"
	ecl68TrystanBackName = "Trystan, Penitent Culler"
)

func TestEcl68Batch68AreRegisteredFull(t *testing.T) {
	ecl67Registered(t, map[string]string{
		ecl68CavalryOracle:        "Safewright Cavalry",
		ecl68GrowthOracle:         "Shimmerwilds Growth",
		ecl68SpryOracle:           "Spry and Mighty",
		ecl68FarrierOracle:        "Surly Farrier",
		ecl68SprigsOracle:         "Tend the Sprigs",
		ecl68ChargeOracle:         "Thoughtweft Charge",
		ecl68TrystanOracle:        "Trystan, Callous Cultivator",
		ecl68TrystanOracle + "#1": ecl68TrystanBackName,
		ecl68AimOracle:            "Unforgiving Aim",
		ecl68BrawlerOracle:        "Vinebred Brawler",
		ecl68WildvineOracle:       "Wildvine Pummeler",
	})
}

// --- Safewright Cavalry ------------------------------------------------------

func TestEcl68SafewrightCavalryPumpsAnElfAndRefusesAnythingElse(t *testing.T) {
	spec, _ := effects.Lookup(ecl68CavalryOracle)
	if len(spec.BlockRules) != 1 {
		t.Fatalf("%d block rules, want the one-blocker limit", len(spec.BlockRules))
	}
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	cav := effectstest.PushCatalogPermanent(g, me.ID, "Safewright Cavalry", "Creature — Elf Warrior", ecl68CavalryOracle, false)
	elf := effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf Druid", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C", "C", "C", "C", "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, cav, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.B16TargetCard(bear)}); err == nil {
		t.Error("a non-Elf was a legal target")
	}
	effectstest.B16Activate(t, g, me.ID, cav, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(elf)})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, elf); p != 3 || tt != 3 {
		t.Errorf("the Elf is %d/%d, want 3/3", p, tt)
	}
}

// --- Shimmerwilds Growth -----------------------------------------------------

func TestEcl68ShimmerwildsGrowthRecolorsTheLandAndAddsAnExtraMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	forest := effectstest.B31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	aura := effectstest.PushAuraOnLand(t, g, me.ID, "Shimmerwilds Growth", ecl68GrowthOracle, forest)
	g.WithWriteLock(func() { g.QueueColorChoiceForEffect(me.ID, aura, "Shimmerwilds Growth", nil, game.ColorForMana) })
	effectstest.AnswerColor(t, g, me.ID, "U")

	c, _ := effectstest.BattlefieldCard(g, forest)
	if cols := c.EffectiveColors(); len(cols) != 1 || cols[0] != "U" {
		t.Errorf("the land's colors are %v, want blue", cols)
	}
	effectstest.TapForMana(t, g, me.ID, forest)
	if got := effectstest.SortedPool(me); len(got) != 2 || got[0] != "G" || got[1] != "U" {
		t.Errorf("pool = %v, want the Forest's {G} and an additional {U}", got)
	}
}

// --- Spry and Mighty ---------------------------------------------------------

func TestEcl68SpryAndMightyDrawsAndPumpsByThePowerDifference(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Bear", 5, 5)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 1, 1)
	effectstest.SeedLibrary(me, "a", "b", "c", "d")
	effectstest.AdvanceToMain(t, g)
	before := effectstest.HandSizeOf(g, me)
	effectstest.Ecl6Cast(t, g, me, "Spry and Mighty", "Sorcery", ecl68SpryOracle, 0, 0, game.CastSpellParams{})
	effectstest.Ecl6ResolveSpell(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, big, small)
	effectstest.PassPriorityAroundTable(t, g)
	// The difference is 3: three cards (the cast card went in and out).
	if got := effectstest.HandSizeOf(g, me) - before; got != 3 {
		t.Errorf("hand grew by %d, want 3", got)
	}
	if p, tt := effectstest.PtOf(t, g, big); p != 8 || tt != 8 {
		t.Errorf("Big is %d/%d, want 8/8", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, small); p != 5 || tt != 5 {
		t.Errorf("Small is %d/%d, want 5/5", p, tt)
	}
	if p, _ := effectstest.PtOf(t, g, other); p != 1 {
		t.Error("an unchosen creature was pumped")
	}
	if !effectstest.HasEffectiveKeyword(t, g, big, "trample") || !effectstest.HasEffectiveKeyword(t, g, small, "trample") || effectstest.HasEffectiveKeyword(t, g, other, "trample") {
		t.Error("trample went to the wrong creatures")
	}
}

func TestEcl68SpryAndMightyDoesNothingWithFewerThanTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	lone := effectstest.B12Creature(g, me.ID, "Lone", "Creature — Bear", 3, 3)
	effectstest.SeedLibrary(me, "a", "b")
	effectstest.AdvanceToMain(t, g)
	before := effectstest.HandSizeOf(g, me)
	effectstest.Ecl6Cast(t, g, me, "Spry and Mighty", "Sorcery", ecl68SpryOracle, 0, 0, game.CastSpellParams{})
	effectstest.Ecl6ResolveSpell(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a prompt opened with one creature")
	}
	if effectstest.HandSizeOf(g, me) != before {
		t.Error("cards were drawn")
	}
	if p, _ := effectstest.PtOf(t, g, lone); p != 3 {
		t.Error("the lone creature was pumped")
	}
}

// --- Surly Farrier -----------------------------------------------------------

func TestEcl68SurlyFarrierPumpsAtSorcerySpeedOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	farrier := effectstest.PushCatalogPermanent(g, me.ID, "Surly Farrier", "Creature — Kithkin Citizen", ecl68FarrierOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, farrier, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("the target is %d/%d, want 3/3", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("no vigilance")
	}
}

// --- Tend the Sprigs ---------------------------------------------------------

func TestEcl68TendTheSprigsFetchesALandAndMakesATreefolkAtSeven(t *testing.T) {
	for _, tc := range []struct {
		name  string
		lands int
		token int
	}{{"five lands then the fetch makes six", 5, 0}, {"six lands then the fetch makes seven", 6, 1}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, _ := effectstest.Ecl6Seats(g)
			for i := 0; i < tc.lands; i++ {
				effectstest.B31Push(g, me.ID, "Plains", "Basic Land — Plains", "", "", 0, 0)
			}
			g.WithWriteLock(func() {
				me.Library.Cards = nil
				me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
			})
			effectstest.AdvanceToMain(t, g)
			effectstest.Ecl6Cast(t, g, me, "Tend the Sprigs", "Sorcery", ecl68SprigsOracle, 0, 0, game.CastSpellParams{})
			effectstest.Ecl6ResolveSpell(t, g)
			if pc := effectstest.SearchChoiceFor(g, me.ID); pc != nil {
				effectstest.AnswerSearchByID(t, g, me.ID, pc.SearchCards[0])
			}
			effectstest.PassPriorityAroundTable(t, g)
			if n := len(effectstest.BattlefieldIDsNamed(g, "Forest")); n != 1 {
				t.Fatalf("%d Forests on the battlefield, want the fetched one", n)
			}
			if n, _ := effectstest.CountTokensNamed(g, me.ID, "Treefolk"); n != tc.token {
				t.Errorf("%d Treefolk tokens, want %d", n, tc.token)
			}
		})
	}
}

// --- Thoughtweft Charge ------------------------------------------------------

func TestEcl68ThoughtweftChargeDrawsOnlyIfACreatureEnteredThisTurn(t *testing.T) {
	for _, entered := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, _ := effectstest.Ecl6Seats(g)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		effectstest.SeedLibrary(me, "a", "b")
		effectstest.AdvanceToMain(t, g)
		if entered {
			g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 black and green Worm"), 1) })
		}
		before := effectstest.HandSizeOf(g, me)
		effectstest.Ecl6Cast(t, g, me, "Thoughtweft Charge", "Instant", ecl68ChargeOracle, 0, 0,
			game.CastSpellParams{Targets: effectstest.CardRefs(bear)})
		effectstest.PassPriorityAroundTable(t, g)
		if p, tt := effectstest.PtOf(t, g, bear); p != 5 || tt != 5 {
			t.Errorf("entered=%v: the target is %d/%d, want 5/5", entered, p, tt)
		}
		want := before + 0
		if entered {
			want = before + 1
		}
		if got := effectstest.HandSizeOf(g, me); got != want {
			t.Errorf("entered=%v: hand %d, want %d", entered, got, want)
		}
	}
}

// --- Unforgiving Aim ---------------------------------------------------------

func TestEcl68UnforgivingAimModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Ecl6Seats(g)
	flyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	ench := effectstest.B12Permanent(g, opp.ID, "Aura Thing", "Enchantment")
	ground := effectstest.B12Creature(g, opp.ID, "Walker", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	// Mode 0 refuses a ground creature.
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Unforgiving Aim", TypeLine: "Instant", OracleID: ecl68AimOracle, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{0}, Targets: effectstest.CardRefs(ground)}); err == nil {
		t.Error("destroying a creature without flying was allowed")
	}
	effectstest.Ecl6Cast(t, g, me, "Unforgiving Aim", "Instant", ecl68AimOracle, 0, 0,
		game.CastSpellParams{Modes: []int{0}, Targets: effectstest.CardRefs(flyer)})
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.BattlefieldIDsNamed(g, "Bird")) != 0 {
		t.Error("the flyer survived")
	}
	effectstest.Ecl6Cast(t, g, me, "Unforgiving Aim", "Instant", ecl68AimOracle, 0, 0,
		game.CastSpellParams{Modes: []int{1}, Targets: effectstest.CardRefs(ench)})
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.BattlefieldIDsNamed(g, "Aura Thing")) != 0 {
		t.Error("the enchantment survived")
	}
	effectstest.Ecl6Cast(t, g, me, "Unforgiving Aim", "Instant", ecl68AimOracle, 0, 0, game.CastSpellParams{Modes: []int{2}})
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Elf"); n != 1 {
		t.Errorf("%d Elf tokens, want 1", n)
	}
}

// --- Vinebred Brawler --------------------------------------------------------

func TestEcl68VinebredBrawlerMustBeBlockedAndPumpsAnotherElfOnAttack(t *testing.T) {
	spec, _ := effects.Lookup(ecl68BrawlerOracle)
	if len(spec.Static) != 1 {
		t.Fatal("no must-be-blocked static")
	}
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	brawler := effectstest.PushCatalogPermanent(g, me.ID, "Vinebred Brawler", "Creature — Elf Berserker", ecl68BrawlerOracle, false)
	elf := effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf Scout", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(brawler, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no attack trigger target prompt")
	}
	if effectstest.HasID(pick.PickTargetCards, brawler) || effectstest.HasID(pick.PickTargetCards, bear) || !effectstest.HasID(pick.PickTargetCards, elf) {
		t.Fatalf("offered %v: want another Elf only", pick.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, elf)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, elf); p != 3 || tt != 2 {
		t.Errorf("the Elf is %d/%d, want 3/2", p, tt)
	}
}

// --- Wildvine Pummeler -------------------------------------------------------

func TestEcl68WildvinePummelerCostsLessForEachColor(t *testing.T) {
	spec, _ := effects.Lookup(ecl68WildvineOracle)
	if len(spec.SelfCostModifiers) != 1 {
		t.Fatal("no self cost modifier")
	}
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	effectstest.B27Push(g, me.ID, "Red", "Creature — Bear", "", "{1}{R}", 2, 2, "R")
	effectstest.B27Push(g, me.ID, "Blue", "Creature — Bear", "", "{1}{U}", 2, 2, "U")
	effectstest.B27Push(g, me.ID, "Green", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Wildvine Pummeler", TypeLine: "Creature — Giant Berserker",
			OracleID: ecl68WildvineOracle, ManaCost: "{6}{G}", Power: 6, Toughness: 6, Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})
	})
	// Three colors: {3}{G} is enough, and {2}{G} is not.
	effectstest.B06AddMana(me, "G", "C", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast for three mana with three colors among permanents")
	}
	effectstest.B06AddMana(me, "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast for {3}{G} with three colors: %v", err)
	}
}

// --- Trystan, Callous Cultivator // Trystan, Penitent Culler ------------------

func ecl68TrystanRow() cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     "Trystan, Callous Cultivator // Trystan, Penitent Culler",
		Layout:   "transform",
		TypeLine: "Legendary Creature — Elf Druid // Legendary Creature — Elf Warlock",
		OracleID: uuid.MustParse(ecl68TrystanOracle),
		Keywords: []string{"Transform", "Deathtouch"},
		CardFaces: []cards.CardFace{
			{Name: "Trystan, Callous Cultivator", TypeLine: "Legendary Creature — Elf Druid", ManaCost: "{2}{G}", Colors: []string{"G"}, Power: "3", Toughness: "4",
				OracleText: "Deathtouch\nWhenever this creature enters or transforms into Trystan, Callous Cultivator, mill three cards. Then if there is an Elf card in your graveyard, you gain 2 life.\nAt the beginning of your first main phase, you may pay {B}. If you do, transform Trystan."},
			{Name: "Trystan, Penitent Culler", TypeLine: "Legendary Creature — Elf Warlock", Colors: []string{"B"}, Power: "3", Toughness: "4",
				OracleText: "Deathtouch\nWhenever this creature transforms into Trystan, Penitent Culler, mill three cards, then you may exile an Elf card from your graveyard. If you do, each opponent loses 2 life.\nAt the beginning of your first main phase, you may pay {G}. If you do, transform Trystan."},
		},
	}
}

func TestEcl68TrystanFrontMillsThreeAndGainsLifeWithAnElfInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	effectstest.Mm2Library(me, effectstest.Mm2Spell("a"), game.Card{Name: "Elf", TypeLine: "Creature — Elf Scout"}, effectstest.Mm2Spell("c"), effectstest.Mm2Spell("d"))
	life := effectstest.LifeOf(g, me.ID)
	effectstest.ImportAndCast(t, g, ecl68TrystanRow(), me)
	effectstest.MonSettle(t, g)
	if got := effectstest.GraveyardSize(g, me.ID); got != 3 {
		t.Errorf("graveyard %d, want three milled", got)
	}
	if got := effectstest.LifeOf(g, me.ID) - life; got != 2 {
		t.Errorf("gained %d life, want 2 (an Elf card was milled)", got)
	}
}

func TestEcl68TrystanFrontGainsNoLifeWithoutAnElf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	effectstest.Mm2Library(me, effectstest.Mm2Spell("a"), effectstest.Mm2Spell("b"), effectstest.Mm2Spell("c"), effectstest.Mm2Spell("d"))
	life := effectstest.LifeOf(g, me.ID)
	effectstest.ImportAndCast(t, g, ecl68TrystanRow(), me)
	effectstest.MonSettle(t, g)
	if got := effectstest.LifeOf(g, me.ID) - life; got != 0 {
		t.Errorf("gained %d life with no Elf card in the graveyard", got)
	}
}

func TestEcl68TrystanBackMillsThenMayExileAnElfToDrainEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Ecl6Seats(g)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	id := effectstest.ImportToHand(ecl68TrystanRow(), me)
	c, _ := effectstest.CardInZone(me.Hand, id)
	me.Hand.Remove(id)
	effectstest.PushBattlefieldCardWithTimestamp(g, c)
	elf := effectstest.Mm2Library(me, effectstest.Mm2Spell("a"), game.Card{Name: "Elf", TypeLine: "Creature — Elf Scout"}, effectstest.Mm2Spell("c"), effectstest.Mm2Spell("d"))[1]
	life := effectstest.LifeOf(g, opp.ID)
	effectstest.IwTransform(t, g, id)
	effectstest.MonSettle(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, elf)
	effectstest.MonSettle(t, g)
	if got := life - effectstest.LifeOf(g, opp.ID); got != 2 {
		t.Errorf("the opponent lost %d life, want 2", got)
	}
	if zone := g.FindCardZoneForEffect(elf); zone == nil || zone.Kind != game.ZoneExile {
		t.Error("the Elf card was not exiled")
	}
}

func TestEcl68TrystanPaysBToTransformAtFirstMain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	id := effectstest.ImportToHand(ecl68TrystanRow(), me)
	c, _ := effectstest.CardInZone(me.Hand, id)
	me.Hand.Remove(id)
	effectstest.PushBattlefieldCardWithTimestamp(g, c)
	effectstest.Mm2Library(me, effectstest.Mm2Spell("a"), effectstest.Mm2Spell("b"), effectstest.Mm2Spell("c"), effectstest.Mm2Spell("d"), effectstest.Mm2Spell("e"))
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.MonSettle(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.MonSettle(t, g)
	if f := effectstest.IwFace(g, id); f != 1 {
		t.Fatalf("Trystan is on face %d after paying {B}, want the back face", f)
	}
}
