package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
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
	spec, _ := Lookup(ecl68CavalryOracle)
	if len(spec.BlockRules) != 1 {
		t.Fatalf("%d block rules, want the one-blocker limit", len(spec.BlockRules))
	}
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	cav := pushCatalogPermanent(g, me.ID, "Safewright Cavalry", "Creature — Elf Warrior", ecl68CavalryOracle, false)
	elf := b12Creature(g, me.ID, "Elf", "Creature — Elf Druid", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	b06AddMana(me, "C", "C", "C", "C", "C", "C", "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, cav, 0, game.ActivateAbilityParams{Strict: true, Targets: b16TargetCard(bear)}); err == nil {
		t.Error("a non-Elf was a legal target")
	}
	b16Activate(t, g, me.ID, cav, 0, game.ActivateAbilityParams{Targets: b16TargetCard(elf)})
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, elf); p != 3 || tt != 3 {
		t.Errorf("the Elf is %d/%d, want 3/3", p, tt)
	}
}

// --- Shimmerwilds Growth -----------------------------------------------------

func TestEcl68ShimmerwildsGrowthRecolorsTheLandAndAddsAnExtraMana(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	forest := b31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	aura := pushAuraOnLand(t, g, me.ID, "Shimmerwilds Growth", ecl68GrowthOracle, forest)
	g.WithWriteLock(func() { g.QueueColorChoiceForEffect(me.ID, aura, "Shimmerwilds Growth", nil, game.ColorForMana) })
	answerColor(t, g, me.ID, "U")

	c, _ := battlefieldCard(g, forest)
	if cols := c.EffectiveColors(); len(cols) != 1 || cols[0] != "U" {
		t.Errorf("the land's colors are %v, want blue", cols)
	}
	tapForMana(t, g, me.ID, forest)
	if got := sortedPool(me); len(got) != 2 || got[0] != "G" || got[1] != "U" {
		t.Errorf("pool = %v, want the Forest's {G} and an additional {U}", got)
	}
}

// --- Spry and Mighty ---------------------------------------------------------

func TestEcl68SpryAndMightyDrawsAndPumpsByThePowerDifference(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	big := b12Creature(g, me.ID, "Big", "Creature — Bear", 5, 5)
	small := b12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Other", "Creature — Bear", 1, 1)
	seedLibrary(me, "a", "b", "c", "d")
	advanceToMain(t, g)
	before := handSizeOf(g, me)
	ecl6Cast(t, g, me, "Spry and Mighty", "Sorcery", ecl68SpryOracle, 0, 0, game.CastSpellParams{})
	ecl6ResolveSpell(t, g)
	answerChooseCards(t, g, me.ID, big, small)
	passPriorityAroundTable(t, g)
	// The difference is 3: three cards (the cast card went in and out).
	if got := handSizeOf(g, me) - before; got != 3 {
		t.Errorf("hand grew by %d, want 3", got)
	}
	if p, tt := ptOf(t, g, big); p != 8 || tt != 8 {
		t.Errorf("Big is %d/%d, want 8/8", p, tt)
	}
	if p, tt := ptOf(t, g, small); p != 5 || tt != 5 {
		t.Errorf("Small is %d/%d, want 5/5", p, tt)
	}
	if p, _ := ptOf(t, g, other); p != 1 {
		t.Error("an unchosen creature was pumped")
	}
	if !hasEffectiveKeyword(t, g, big, "trample") || !hasEffectiveKeyword(t, g, small, "trample") || hasEffectiveKeyword(t, g, other, "trample") {
		t.Error("trample went to the wrong creatures")
	}
}

func TestEcl68SpryAndMightyDoesNothingWithFewerThanTwoCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	lone := b12Creature(g, me.ID, "Lone", "Creature — Bear", 3, 3)
	seedLibrary(me, "a", "b")
	advanceToMain(t, g)
	before := handSizeOf(g, me)
	ecl6Cast(t, g, me, "Spry and Mighty", "Sorcery", ecl68SpryOracle, 0, 0, game.CastSpellParams{})
	ecl6ResolveSpell(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a prompt opened with one creature")
	}
	if handSizeOf(g, me) != before {
		t.Error("cards were drawn")
	}
	if p, _ := ptOf(t, g, lone); p != 3 {
		t.Error("the lone creature was pumped")
	}
}

// --- Surly Farrier -----------------------------------------------------------

func TestEcl68SurlyFarrierPumpsAtSorcerySpeedOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	farrier := pushCatalogPermanent(g, me.ID, "Surly Farrier", "Creature — Kithkin Citizen", ecl68FarrierOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, farrier, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bear)})
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("the target is %d/%d, want 3/3", p, tt)
	}
	if !hasEffectiveKeyword(t, g, bear, "vigilance") {
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
			g := newCatalogGame(t)
			me, _ := ecl6Seats(g)
			for i := 0; i < tc.lands; i++ {
				b31Push(g, me.ID, "Plains", "Basic Land — Plains", "", "", 0, 0)
			}
			g.WithWriteLock(func() {
				me.Library.Cards = nil
				me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
			})
			advanceToMain(t, g)
			ecl6Cast(t, g, me, "Tend the Sprigs", "Sorcery", ecl68SprigsOracle, 0, 0, game.CastSpellParams{})
			ecl6ResolveSpell(t, g)
			if pc := searchChoiceFor(g, me.ID); pc != nil {
				answerSearchByID(t, g, me.ID, pc.SearchCards[0])
			}
			passPriorityAroundTable(t, g)
			if n := len(battlefieldIDsNamed(g, "Forest")); n != 1 {
				t.Fatalf("%d Forests on the battlefield, want the fetched one", n)
			}
			if n, _ := countTokensNamed(g, me.ID, "Treefolk"); n != tc.token {
				t.Errorf("%d Treefolk tokens, want %d", n, tc.token)
			}
		})
	}
}

// --- Thoughtweft Charge ------------------------------------------------------

func TestEcl68ThoughtweftChargeDrawsOnlyIfACreatureEnteredThisTurn(t *testing.T) {
	for _, entered := range []bool{false, true} {
		g := newCatalogGame(t)
		me, _ := ecl6Seats(g)
		bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		seedLibrary(me, "a", "b")
		advanceToMain(t, g)
		if entered {
			g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 black and green Worm"), 1) })
		}
		before := handSizeOf(g, me)
		ecl6Cast(t, g, me, "Thoughtweft Charge", "Instant", ecl68ChargeOracle, 0, 0,
			game.CastSpellParams{Targets: cardRefs(bear)})
		passPriorityAroundTable(t, g)
		if p, tt := ptOf(t, g, bear); p != 5 || tt != 5 {
			t.Errorf("entered=%v: the target is %d/%d, want 5/5", entered, p, tt)
		}
		want := before + 0
		if entered {
			want = before + 1
		}
		if got := handSizeOf(g, me); got != want {
			t.Errorf("entered=%v: hand %d, want %d", entered, got, want)
		}
	}
}

// --- Unforgiving Aim ---------------------------------------------------------

func TestEcl68UnforgivingAimModes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ecl6Seats(g)
	flyer := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	ench := b12Permanent(g, opp.ID, "Aura Thing", "Enchantment")
	ground := b12Creature(g, opp.ID, "Walker", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	// Mode 0 refuses a ground creature.
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Unforgiving Aim", TypeLine: "Instant", OracleID: ecl68AimOracle, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{0}, Targets: cardRefs(ground)}); err == nil {
		t.Error("destroying a creature without flying was allowed")
	}
	ecl6Cast(t, g, me, "Unforgiving Aim", "Instant", ecl68AimOracle, 0, 0,
		game.CastSpellParams{Modes: []int{0}, Targets: cardRefs(flyer)})
	passPriorityAroundTable(t, g)
	if len(battlefieldIDsNamed(g, "Bird")) != 0 {
		t.Error("the flyer survived")
	}
	ecl6Cast(t, g, me, "Unforgiving Aim", "Instant", ecl68AimOracle, 0, 0,
		game.CastSpellParams{Modes: []int{1}, Targets: cardRefs(ench)})
	passPriorityAroundTable(t, g)
	if len(battlefieldIDsNamed(g, "Aura Thing")) != 0 {
		t.Error("the enchantment survived")
	}
	ecl6Cast(t, g, me, "Unforgiving Aim", "Instant", ecl68AimOracle, 0, 0, game.CastSpellParams{Modes: []int{2}})
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Elf"); n != 1 {
		t.Errorf("%d Elf tokens, want 1", n)
	}
}

// --- Vinebred Brawler --------------------------------------------------------

func TestEcl68VinebredBrawlerMustBeBlockedAndPumpsAnotherElfOnAttack(t *testing.T) {
	spec, _ := Lookup(ecl68BrawlerOracle)
	if len(spec.Static) != 1 {
		t.Fatal("no must-be-blocked static")
	}
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	brawler := pushCatalogPermanent(g, me.ID, "Vinebred Brawler", "Creature — Elf Berserker", ecl68BrawlerOracle, false)
	elf := b12Creature(g, me.ID, "Elf", "Creature — Elf Scout", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(brawler, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no attack trigger target prompt")
	}
	if hasID(pick.PickTargetCards, brawler) || hasID(pick.PickTargetCards, bear) || !hasID(pick.PickTargetCards, elf) {
		t.Fatalf("offered %v: want another Elf only", pick.PickTargetCards)
	}
	pickCard(t, g, me.ID, elf)
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, elf); p != 3 || tt != 2 {
		t.Errorf("the Elf is %d/%d, want 3/2", p, tt)
	}
}

// --- Wildvine Pummeler -------------------------------------------------------

func TestEcl68WildvinePummelerCostsLessForEachColor(t *testing.T) {
	spec, _ := Lookup(ecl68WildvineOracle)
	if len(spec.SelfCostModifiers) != 1 {
		t.Fatal("no self cost modifier")
	}
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	b27Push(g, me.ID, "Red", "Creature — Bear", "", "{1}{R}", 2, 2, "R")
	b27Push(g, me.ID, "Blue", "Creature — Bear", "", "{1}{U}", 2, 2, "U")
	b27Push(g, me.ID, "Green", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	advanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Wildvine Pummeler", TypeLine: "Creature — Giant Berserker",
			OracleID: ecl68WildvineOracle, ManaCost: "{6}{G}", Power: 6, Toughness: 6, Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})
	})
	// Three colors: {3}{G} is enough, and {2}{G} is not.
	b06AddMana(me, "G", "C", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast for three mana with three colors among permanents")
	}
	b06AddMana(me, "C")
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
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	mm2Library(me, mm2Spell("a"), game.Card{Name: "Elf", TypeLine: "Creature — Elf Scout"}, mm2Spell("c"), mm2Spell("d"))
	life := lifeOf(g, me.ID)
	importAndCast(t, g, ecl68TrystanRow(), me)
	monSettle(t, g)
	if got := graveyardSize(g, me.ID); got != 3 {
		t.Errorf("graveyard %d, want three milled", got)
	}
	if got := lifeOf(g, me.ID) - life; got != 2 {
		t.Errorf("gained %d life, want 2 (an Elf card was milled)", got)
	}
}

func TestEcl68TrystanFrontGainsNoLifeWithoutAnElf(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	mm2Library(me, mm2Spell("a"), mm2Spell("b"), mm2Spell("c"), mm2Spell("d"))
	life := lifeOf(g, me.ID)
	importAndCast(t, g, ecl68TrystanRow(), me)
	monSettle(t, g)
	if got := lifeOf(g, me.ID) - life; got != 0 {
		t.Errorf("gained %d life with no Elf card in the graveyard", got)
	}
}

func TestEcl68TrystanBackMillsThenMayExileAnElfToDrainEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ecl6Seats(g)
	advanceTo(t, g, game.StepPrecombatMain)
	id := importToHand(ecl68TrystanRow(), me)
	c, _ := cardInZone(me.Hand, id)
	me.Hand.Remove(id)
	pushBattlefieldCardWithTimestamp(g, c)
	elf := mm2Library(me, mm2Spell("a"), game.Card{Name: "Elf", TypeLine: "Creature — Elf Scout"}, mm2Spell("c"), mm2Spell("d"))[1]
	life := lifeOf(g, opp.ID)
	iwTransform(t, g, id)
	monSettle(t, g)
	answerChooseCards(t, g, me.ID, elf)
	monSettle(t, g)
	if got := life - lifeOf(g, opp.ID); got != 2 {
		t.Errorf("the opponent lost %d life, want 2", got)
	}
	if zone := g.FindCardZoneForEffect(elf); zone == nil || zone.Kind != game.ZoneExile {
		t.Error("the Elf card was not exiled")
	}
}

func TestEcl68TrystanPaysBToTransformAtFirstMain(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	id := importToHand(ecl68TrystanRow(), me)
	c, _ := cardInZone(me.Hand, id)
	me.Hand.Remove(id)
	pushBattlefieldCardWithTimestamp(g, c)
	mm2Library(me, mm2Spell("a"), mm2Spell("b"), mm2Spell("c"), mm2Spell("d"), mm2Spell("e"))
	advanceTo(t, g, game.StepPrecombatMain)
	monSettle(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	answerPayUnless(t, g, me.ID, true)
	monSettle(t, g)
	if f := iwFace(g, id); f != 1 {
		t.Fatalf("Trystan is on face %d after paying {B}, want the back face", f)
	}
}
