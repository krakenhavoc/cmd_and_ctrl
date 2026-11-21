package testss82v

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	vowCrawlingInfestation = "a42262b8-8a26-4570-a6ee-24b239d095fc"
	vowVilespawnSpider     = "3e969073-76cc-4a2c-9abd-3a4094488fe5"
	vowHiveheartShaman     = "c3780529-d47e-48dc-86f6-c1a282c3b656"
	vowOldRutstein         = "8f5ffb0b-0585-4a37-8048-eda53990d1dd"
	vowCartographersSurvey = "e05029b4-552d-4691-a5c4-885284fb1636"
	vowCultivatorColossus  = "4d8505d3-c8a1-4850-95d4-5ddb8ad49770"
	vowGloriousSunrise     = "844c1024-db80-4287-9770-3b10e53d0fc8"
	vowNaturesEmbrace      = "9ab101c2-215b-4736-873a-4a1ef6666b6d"
	vowReclusiveTaxiderm   = "7a0168c5-e258-4c50-96aa-95deb2198e6f"
	vowAncientLumberknot   = "d714f1f6-539e-4a2a-b738-226854d7ff09"
	vowSpikedRipsaw        = "660ff9fa-1e9b-4697-a3f5-b12238703161"
	vowShelteringBoughs    = "b6bde4eb-15f0-4e6e-9d25-459a2f75ccf2"
)

func TestVowB11AllCardsRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		vowCrawlingInfestation: "Crawling Infestation", vowVilespawnSpider: "Vilespawn Spider",
		vowHiveheartShaman: "Hiveheart Shaman", vowOldRutstein: "Old Rutstein",
		vowCartographersSurvey: "Cartographer's Survey", vowCultivatorColossus: "Cultivator Colossus",
		vowGloriousSunrise: "Glorious Sunrise", vowNaturesEmbrace: "Nature's Embrace",
		vowReclusiveTaxiderm: "Reclusive Taxidermist", vowAncientLumberknot: "Ancient Lumberknot",
		vowSpikedRipsaw: "Spiked Ripsaw", vowShelteringBoughs: "Sheltering Boughs",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

// --- Ancient Lumberknot ---

// Lumberknot (1/4) and a 2/5 Wall deal 4 and 5; a 4/4 still deals 4.
func TestVowB11AncientLumberknotAssignsByToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knot := effectstest.B12Push(g, me.ID, "Ancient Lumberknot", "Creature — Treefolk", vowAncientLumberknot, 1, 4)
	wall := effectstest.PushVanillaCreature(g, me.ID, "Stout Wall", 2, 5)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Big Bear", 4, 4)
	before := opp.Life
	effectstest.AttackWith(t, g, opp.ID, knot, wall, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if want := before - (4 + 5 + 4); opp.Life != want {
		t.Errorf("opponent at %d, want %d", opp.Life, want)
	}
}

// "You control": an opponent's 1/6 still deals 1.
func TestVowB11AncientLumberknotLeavesOpposingCreaturesAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Ancient Lumberknot", "Creature — Treefolk", vowAncientLumberknot, 1, 4)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Wall", 1, 6)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	before := me.Life
	effectstest.AttackWith(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before-1 {
		t.Errorf("an opponent's 1/6 still deals 1: %d → %d", before, me.Life)
	}
}

// --- Crawling Infestation ---

func vowMill(t *testing.T, g *game.Game, p *game.Player, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if _, err := g.MillToZoneForEffect(p.ID, n, game.ZoneGraveyard); err != nil {
			t.Fatalf("mill: %v", err)
		}
	})
	effectstest.Mm2Settle(t, g)
}

func TestVowB11CrawlingInfestationOneInsectEachTurnForCreatureCardsOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Crawling Infestation", OracleID: vowCrawlingInfestation, TypeLine: "Enchantment"})
	effectstest.Mm2Library(me, effectstest.Mm2Spell("A"), effectstest.Mm2Land("Forest"), effectstest.Mm2CreatureCard("B", 2), effectstest.Mm2CreatureCard("C", 2), effectstest.Mm2CreatureCard("D", 2),
		effectstest.Mm2Spell("F1"), effectstest.Mm2Spell("F2"), effectstest.Mm2Spell("F3"), effectstest.Mm2Spell("F4"))
	vowMill(t, g, me, 2) // a spell and a land: nothing
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Insect"); got != 0 {
		t.Fatalf("%d Insects after a spell and a land, want 0", got)
	}
	vowMill(t, g, me, 2) // two creature cards together: one Insect
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Insect"); got != 1 {
		t.Fatalf("%d Insects after two creature cards milled together, want 1", got)
	}
	vowMill(t, g, me, 1) // a third creature card the same turn: still one
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Insect"); got != 1 {
		t.Errorf("%d Insects after another creature card the same turn, want 1", got)
	}
	// The next of my turns, a creature card reaches the graveyard from hand.
	effectstest.AdvanceDecliningTriggerPrompts(t, g, 1)
	effectstest.AdvanceDecliningTriggerPrompts(t, g, 0)
	effectstest.CmdrB17ToGraveyard(t, g, me, effectstest.Mm2CreatureCard("Discarded", 1))
	effectstest.Mm2Settle(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Insect"); got != 2 {
		t.Errorf("%d Insects after a creature card reached the graveyard on a later turn of mine, want 2", got)
	}
}

func TestVowB11CrawlingInfestationIgnoresAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Crawling Infestation", OracleID: vowCrawlingInfestation, TypeLine: "Enchantment"})
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.CmdrB17ToGraveyard(t, g, me, effectstest.Mm2CreatureCard("Discarded", 1))
	effectstest.Mm2Settle(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Insect"); got != 0 {
		t.Errorf("%d Insects on an opponent's turn, want 0", got)
	}
}

func TestVowB11CrawlingInfestationUpkeepMillIsOptional(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Crawling Infestation", OracleID: vowCrawlingInfestation, TypeLine: "Enchantment"})
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	before := me.Library.Size()
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceTriggerPrompt, me.ID)
	if prompt == nil {
		t.Fatalf("no 'you may mill' question: %+v", g.PendingChoices)
	}
	if err := g.ResolveTriggerPrompt(prompt.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - me.Library.Size(); got < 2 {
		t.Errorf("milled %d, want 2", got)
	}
}

// --- Reclusive Taxidermist ---

func TestVowB11ReclusiveTaxidermistGrowsWithFourCreatureCardsAndMakesMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tax := effectstest.B12Push(g, me.ID, "Reclusive Taxidermist", "Creature — Human Druid", vowReclusiveTaxiderm, 1, 2)
	for i := 0; i < 3; i++ {
		effectstest.CmdrB17ToGraveyard(t, g, me, effectstest.Mm2CreatureCard("Dead", 1))
	}
	effectstest.CmdrB17ToGraveyard(t, g, me, effectstest.Mm2Spell("Not a creature"))
	if p, tt := effectstest.EffectivePower(t, g, tax), effectstest.EffectiveToughness(t, g, tax); p != 1 || tt != 2 {
		t.Errorf("three creature cards: %d/%d, want 1/2", p, tt)
	}
	effectstest.CmdrB17ToGraveyard(t, g, me, effectstest.Mm2CreatureCard("Dead", 1))
	if p, tt := effectstest.EffectivePower(t, g, tax), effectstest.EffectiveToughness(t, g, tax); p != 4 || tt != 4 {
		t.Errorf("four creature cards: %d/%d, want 4/4", p, tt)
	}
	if err := g.ActivateManaAbility(me.ID, tax, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"U"}) {
		t.Errorf("pool %v, want [U]", got)
	}
}

// --- Sheltering Boughs ---

func TestVowB11ShelteringBoughsDrawsAndGivesPlusOnePlusThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	handBefore := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Sheltering Boughs", "Enchantment — Aura", vowShelteringBoughs,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || tt != 5 {
		t.Errorf("enchanted Bear is %d/%d, want 3/5", p, tt)
	}
	// The Aura left the hand (cast) and one card came in.
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand %d → %d, want one drawn card", handBefore, got)
	}
}

// --- Nature's Embrace ---

func TestVowB11NaturesEmbracePumpsACreatureButGivesItNoMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Nature's Embrace", "Enchantment — Aura", vowNaturesEmbrace,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tt != 4 {
		t.Errorf("enchanted Bear is %d/%d, want 4/4", p, tt)
	}
	if err := g.ActivateManaAbility(me.ID, bear, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err == nil {
		t.Error("an enchanted creature that is not a land must not get a mana ability")
	}
}

func TestVowB11NaturesEmbraceGivesALandTwoManaOfOneColorAndNoPump(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Wastes", TypeLine: "Basic Land"})
	effectstest.CastCatalogSpell(t, g, "Nature's Embrace", "Enchantment — Aura", vowNaturesEmbrace,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}})
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("granted mana ability: %v", err)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"R", "R"}) {
		t.Errorf("pool %v, want two red", got)
	}
	if c, ok := effectstest.BattlefieldCardByID(g, land); !ok || c.IsCreature() {
		t.Error("the land became a creature")
	}
}

// --- Cartographer's Survey ---

func vowSurveyLibrary(me *game.Player) (lands []uuid.UUID, rest []uuid.UUID) {
	ids := effectstest.Mm2Library(me,
		effectstest.Mm2Land("Forest"), effectstest.Mm2Spell("S1"), effectstest.Mm2Land("Forest"), effectstest.Mm2Spell("S2"), effectstest.Mm2Land("Forest"), effectstest.Mm2Spell("S3"), effectstest.Mm2Spell("S4"),
		effectstest.Mm2Spell("Deep"))
	return []uuid.UUID{ids[0], ids[2], ids[4]}, []uuid.UUID{ids[1], ids[3], ids[5], ids[6]}
}

func TestVowB11CartographersSurveyPutsUpToTwoLandsTappedAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lands, rest := vowSurveyLibrary(me)
	effectstest.CastCatalogSpell(t, g, "Cartographer's Survey", "Sorcery", vowCartographersSurvey, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("no pick prompt")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, lands); err == nil {
		t.Fatal("three lands were accepted: up to two")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, lands[:2]); err != nil {
		t.Fatalf("pick two lands: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range lands[:2] {
		c, ok := effectstest.BattlefieldCardByID(g, id)
		if !ok || !c.Tapped {
			t.Errorf("land %s: on battlefield %v tapped %v, want a tapped land", id, ok, c.Tapped)
		}
	}
	for _, id := range append(rest, lands[2]) {
		if z := effectstest.Mm2Zone(g, id); z != game.ZoneLibrary {
			t.Errorf("card %s in %v, want the library", id, z)
		}
	}
	// The unlooked card was on top of the rest: the seven others went below it.
	if n := me.Library.Size(); n != 6 {
		t.Errorf("library %d, want 6", n)
	}
}

func TestVowB11CartographersSurveyMayTakeNoLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lands, _ := vowSurveyLibrary(me)
	effectstest.CastCatalogSpell(t, g, "Cartographer's Survey", "Sorcery", vowCartographersSurvey, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("no pick prompt")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range lands {
		if z := effectstest.Mm2Zone(g, id); z != game.ZoneLibrary {
			t.Errorf("land %s in %v, want the library", id, z)
		}
	}
}

// --- Cultivator Colossus ---

func vowLandCard(name string) game.Card {
	return game.Card{Name: name, TypeLine: "Basic Land — Forest"}
}

func TestVowB11CultivatorColossusSizeIsItsLandsAndTramples(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushLands(g, me.ID, 4)
	col := effectstest.B12Push(g, me.ID, "Cultivator Colossus", "Creature — Plant Beast", vowCultivatorColossus, 0, 0)
	if p, tt := effectstest.EffectivePower(t, g, col), effectstest.EffectiveToughness(t, g, col); p != 4 || tt != 4 {
		t.Errorf("with 4 lands: %d/%d, want 4/4", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, col, "trample") {
		t.Error("no trample")
	}
}

// Two lands in hand: each is put tapped and draws a card, then the loop asks again
// and ends when no land is left to put.
func TestVowB11CultivatorColossusPutsLandsDrawingEachTimeUntilNoneLeft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	l1 := effectstest.PutInHand(me, vowLandCard("Forest"))
	l2 := effectstest.PutInHand(me, vowLandCard("Forest"))
	handBefore := me.Hand.Size() // two lands
	effectstest.CastCatalogSpell(t, g, "Cultivator Colossus", "Creature — Plant Beast", vowCultivatorColossus, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i, id := range []uuid.UUID{l1, l2} {
		pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
		if pick == nil {
			t.Fatalf("pass %d: no prompt", i+1)
		}
		if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{id}); err != nil {
			t.Fatalf("pass %d: %v", i+1, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if c, ok := effectstest.BattlefieldCardByID(g, id); !ok || !c.Tapped {
			t.Errorf("pass %d: land on battlefield %v, tapped %v; want tapped", i+1, ok, c.Tapped)
		}
	}
	if pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID); pick != nil {
		t.Error("asked a third time with no land card in hand")
	}
	// Two cards were drawn; the lands left the hand.
	if got := me.Hand.Size(); got != handBefore-2+2 {
		t.Errorf("hand %d, want %d", got, handBefore)
	}
}

func TestVowB11CultivatorColossusDecliningEndsTheLoopWithoutDrawing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PutInHand(me, vowLandCard("Forest"))
	handBefore := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Cultivator Colossus", "Creature — Plant Beast", vowCultivatorColossus, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("no prompt")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handBefore {
		t.Errorf("hand %d, want %d (no draw)", got, handBefore)
	}
}

// --- Old Rutstein ---

func vowRutsteinMills(t *testing.T, top game.Card) (g *game.Game, me *game.Player) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	effectstest.Mm2Library(me, top, effectstest.Mm2Spell("Filler"))
	effectstest.CastCatalogSpell(t, g, "Old Rutstein", "Legendary Creature — Human Peasant", vowOldRutstein, nil)
	effectstest.PassPriorityAroundTable(t, g)
	return g, me
}

func TestVowB11OldRutsteinLandMakesATreasure(t *testing.T) {
	g, me := vowRutsteinMills(t, effectstest.Mm2Land("Forest"))
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); got != 1 {
		t.Errorf("%d Treasures, want 1", got)
	}
	if effectstest.CountBattlefieldNamed(g, me.ID, "Insect")+effectstest.CountBattlefieldNamed(g, me.ID, "Blood") != 0 {
		t.Error("a land also made an Insect or Blood")
	}
}

func TestVowB11OldRutsteinCreatureMakesAnInsect(t *testing.T) {
	g, me := vowRutsteinMills(t, effectstest.Mm2CreatureCard("Bear", 2))
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Insect"); got != 1 {
		t.Errorf("%d Insects, want 1", got)
	}
	if effectstest.CountBattlefieldNamed(g, me.ID, "Treasure")+effectstest.CountBattlefieldNamed(g, me.ID, "Blood") != 0 {
		t.Error("a creature also made a Treasure or Blood")
	}
}

func TestVowB11OldRutsteinNoncreatureNonlandMakesBlood(t *testing.T) {
	g, me := vowRutsteinMills(t, effectstest.Mm2Spell("Shock"))
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Blood"); got != 1 {
		t.Errorf("%d Blood tokens, want 1", got)
	}
	if effectstest.CountBattlefieldNamed(g, me.ID, "Treasure")+effectstest.CountBattlefieldNamed(g, me.ID, "Insect") != 0 {
		t.Error("a sorcery also made a Treasure or Insect")
	}
}

func TestVowB11OldRutsteinLandCreatureMakesBothAndMillsAgainEachUpkeep(t *testing.T) {
	g, me := vowRutsteinMills(t, game.Card{Name: "Dryad Arbor", TypeLine: "Land Creature — Forest Dryad", Power: 1, Toughness: 1})
	if effectstest.CountBattlefieldNamed(g, me.ID, "Treasure") != 1 || effectstest.CountBattlefieldNamed(g, me.ID, "Insect") != 1 {
		t.Errorf("a land creature should make a Treasure and an Insect")
	}
	// Upkeep: the next card (the Filler sorcery) makes Blood.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Blood"); got != 1 {
		t.Errorf("%d Blood tokens after the upkeep mill, want 1", got)
	}
}

// --- Vilespawn Spider ---

func TestVowB11VilespawnSpiderMillsAtUpkeepAndTradesForInsects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spider := effectstest.B12Push(g, me.ID, "Vilespawn Spider", "Creature — Spider", vowVilespawnSpider, 2, 3)
	// A reach keyword and the upkeep mill.
	if !effectstest.HasEffectiveKeyword(t, g, spider, "reach") {
		t.Error("no reach")
	}
	effectstest.CmdrB17ToGraveyard(t, g, me, effectstest.Mm2CreatureCard("A", 1))
	effectstest.CmdrB17ToGraveyard(t, g, me, effectstest.Mm2CreatureCard("B", 1))
	effectstest.CmdrB17ToGraveyard(t, g, me, effectstest.Mm2Spell("Not a creature"))
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	for _, c := range []string{"G", "U", "C", "C"} {
		me.ManaPool.AddMana(game.ManaToken{Color: c})
	}
	if err := g.ActivateCatalogAbility(me.ID, spider, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCardByID(g, spider); ok && c.InstanceID == spider {
		t.Error("the Spider was not sacrificed")
	}
	// Two creature cards plus the Spider itself.
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Insect"); got != 3 {
		t.Errorf("%d Insects, want 3 (two creature cards and the Spider)", got)
	}
}

func TestVowB11VilespawnSpiderOnlyAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spider := effectstest.B12Push(g, me.ID, "Vilespawn Spider", "Creature — Spider", vowVilespawnSpider, 2, 3)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	for _, c := range []string{"G", "U", "C", "C"} {
		me.ManaPool.AddMana(game.ManaToken{Color: c})
	}
	if err := g.ActivateCatalogAbility(me.ID, spider, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated outside a main phase")
	}
}

func TestVowB11VilespawnSpiderMillsAtYourUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Vilespawn Spider", "Creature — Spider", vowVilespawnSpider, 2, 3)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	before := me.Library.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - me.Library.Size(); got != 1 {
		t.Errorf("milled %d, want 1", got)
	}
}

// --- Hiveheart Shaman ---

func TestVowB11HiveheartShamanFetchesABasicSharingNoTypeAndOnlyThat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	shaman := effectstest.B12Push(g, me.ID, "Hiveheart Shaman", "Creature — Human Shaman", vowHiveheartShaman, 3, 5)
	forest := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	plains := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	nonbasic := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Hall", TypeLine: "Land — Plains"})
	vowDeclareAttack(t, g, opp.ID, shaman)
	search := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, me.ID)
	if search == nil {
		t.Fatal("no search prompt")
	}
	if err := g.BeginSearchLibrary(search.ID, me.ID); err != nil {
		t.Fatalf("begin search: %v", err)
	}
	for _, bad := range []uuid.UUID{forest, nonbasic} {
		if err := g.ResolveSearchLibrary(search.ID, me.ID, []uuid.UUID{bad}); err == nil {
			t.Errorf("accepted %s, which shares a type or isn't basic", bad)
		}
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, []uuid.UUID{plains}); err != nil {
		t.Fatalf("search: %v", err)
	}
	if c, ok := effectstest.BattlefieldCardByID(g, plains); !ok || c.Tapped {
		t.Errorf("the Plains should be on the battlefield untapped (found %v, tapped %v)", ok, c.Tapped)
	}
}

func TestVowB11HiveheartShamanMakesAnInsectWithCountersEqualToBasicLandTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Wastes", TypeLine: "Basic Land"})
	shaman := effectstest.B12Push(g, me.ID, "Hiveheart Shaman", "Creature — Human Shaman", vowHiveheartShaman, 3, 5)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	for _, c := range []string{"G", "C", "C", "C", "C", "C"} {
		me.ManaPool.AddMana(game.ManaToken{Color: c})
	}
	if err := g.ActivateCatalogAbility(me.ID, shaman, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var insect uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Insect" {
			insect = c.InstanceID
		}
	}
	if insect == uuid.Nil {
		t.Fatal("no Insect token")
	}
	if n := effectstest.CountersOn(g, insect, game.CounterPlusOne); n != 2 {
		t.Errorf("%d counters, want 2 (Forest and Island)", n)
	}
	if p := vowCurrentPower(g, insect); p != 3 {
		t.Errorf("Insect power %d, want 3", p)
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	if err := g.ActivateCatalogAbility(me.ID, shaman, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated outside a main phase")
	}
}

// --- Glorious Sunrise ---

func vowSunriseMode(t *testing.T, mode int) (g *game.Game, me *game.Player) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	effectstest.B12Push(g, me.ID, "Glorious Sunrise", "Enchantment", vowGloriousSunrise, 0, 0)
	return g, me
}

func TestVowB11GloriousSunrisePumpAndTrample(t *testing.T) {
	g, me := vowSunriseMode(t, 0)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, g.Seats[1].ID, "Their Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no mode prompt")
	}
	effectstest.AnswerMode(t, g, c, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("Bear %d/%d, want 3/3", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("no trample")
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 {
		t.Errorf("an opposing creature got the pump: %d", p)
	}
}

func TestVowB11GloriousSunriseLandGainsGGG(t *testing.T) {
	g, me := vowSunriseMode(t, 1)
	land := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no mode prompt")
	}
	effectstest.AnswerMode(t, g, c, me.ID, 1)
	effectstest.PickCard(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	// A basic Forest taps for {G}; with the grant it also has a {G}{G}{G} ability.
	var got []string
	for idx := 0; idx < 3 && len(got) != 3; idx++ {
		me.ManaPool = game.ManaPool{}
		g.WithWriteLock(func() { effectstest.FindBattlefieldCardByID(g, land).Tapped = false })
		if err := g.ActivateManaAbility(me.ID, land, idx, game.ManaAbilityParams{}); err != nil {
			continue
		}
		got = effectstest.PoolColors(me)
	}
	if !reflect.DeepEqual(got, []string{"G", "G", "G"}) {
		t.Errorf("no {T}: Add {G}{G}{G} ability found; last pool %v", got)
	}
}

func TestVowB11GloriousSunriseDrawNeedsAPowerThreeCreature(t *testing.T) {
	g, me := vowSunriseMode(t, 2)
	effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	before := me.Hand.Size()
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no mode prompt")
	}
	effectstest.AnswerMode(t, g, c, me.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before {
		t.Errorf("drew with no power-3 creature: hand %d → %d", before, got)
	}

	g2, me2 := vowSunriseMode(t, 2)
	effectstest.PushVanillaCreature(g2, me2.ID, "Brute", 3, 3)
	effectstest.AdvanceTo(t, g2, game.StepBeginCombat)
	before = me2.Hand.Size()
	c = effectstest.ModePickChoiceFor(g2, me2.ID)
	if c == nil {
		t.Fatal("no mode prompt")
	}
	effectstest.AnswerMode(t, g2, c, me2.ID, 2)
	effectstest.PassPriorityAroundTable(t, g2)
	if got := me2.Hand.Size(); got != before+1 {
		t.Errorf("hand %d → %d, want one card drawn", before, got)
	}
}

func TestVowB11GloriousSunriseGainsThreeLifeOnlyOnYourTurn(t *testing.T) {
	g, me := vowSunriseMode(t, 3)
	before := me.Life
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no mode prompt")
	}
	effectstest.AnswerMode(t, g, c, me.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+3 {
		t.Errorf("life %d → %d, want +3", before, me.Life)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if c := effectstest.ModePickChoiceFor(g, me.ID); c != nil {
		t.Error("triggered on another player's combat")
	}
}

// --- Spiked Ripsaw ---

func vowRipsawSetup(t *testing.T) (g *game.Game, me, opp *game.Player, bear uuid.UUID, forest uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	bear = effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	saw := effectstest.B12Push(g, me.ID, "Spiked Ripsaw", "Artifact — Equipment", vowSpikedRipsaw, 0, 0)
	effectstest.NcdAttach(t, g, saw, bear)
	forest = effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	return
}

func TestVowB11SpikedRipsawPumpsAndSacrificingAForestGivesTrample(t *testing.T) {
	g, me, opp, bear, forest := vowRipsawSetup(t)
	if p, tt := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 5 || tt != 5 {
		t.Fatalf("equipped Bear %d/%d, want 5/5", p, tt)
	}
	vowDeclareAttack(t, g, opp.ID, bear)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatal("not asked to sacrifice a Forest")
	}
	if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{forest}); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCardByID(g, forest); ok {
		t.Error("the Forest is still on the battlefield")
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the attacker did not gain trample")
	}
}

func TestVowB11SpikedRipsawDecliningLeavesTheForestAndNoTrample(t *testing.T) {
	g, me, opp, bear, forest := vowRipsawSetup(t)
	vowDeclareAttack(t, g, opp.ID, bear)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatal("not asked")
	}
	if err := g.ResolveOwnPermanents(c.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCardByID(g, forest); !ok {
		t.Error("the Forest was sacrificed")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("trample without sacrificing")
	}
}

// vowDeclareAttack declares the attackers and locks the declaration in, so
// "whenever attacks" triggers go to the stack, without moving on to damage.
func vowDeclareAttack(t *testing.T, g *game.Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range attackers {
		if err := g.DeclareAttacker(id, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
}

// vowCurrentPower reads the live layered power of a battlefield card.
func vowCurrentPower(g *game.Game, id uuid.UUID) int {
	p := 0
	g.ReadSnapshot(func() { p = effectstest.FindBattlefieldCardByID(g, id).CurrentPower() })
	return p
}
