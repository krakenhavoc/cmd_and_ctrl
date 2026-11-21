package testsu

import (
	"slices"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// usg_batch1_cards_test.go — Urza's Saga, batch 1 (A–C, #5624). Each test
// plays the real catalog entry and asserts what its text says.

const (
	usgB1AbyssalHorrorOracle   = "6b60422a-765b-4168-a37b-68109103385a"
	usgB1AngelicPageOracle     = "f7f76abc-8656-4e97-9b7f-7f89ca3e6d93"
	usgB1EnchantressOracle     = "11f17f85-ca97-4551-838f-7cb32f0e5f10"
	usgB1SwineOracle           = "80e08ee2-4272-4783-9a0d-90079600b594"
	usgB1TreefolkOracle        = "a3090daf-26d9-4ba5-9fa3-42bd128e229d"
	usgB1BlastedOracle         = "9c8007ac-4b3d-4444-93e9-f583185e5d81"
	usgB1BreachOracle          = "3c412ae1-3499-41fa-b3c4-70131c9bf0bd"
	usgB1CacklingFiendOracle   = "2029954b-6fa7-40d7-bb19-d7534c62be5d"
	usgB1CatalogOracle         = "965139fb-2e58-4ef2-b639-3ceaa2995c00"
	usgB1ClearOracle           = "d0783be9-e518-431a-8e71-121610da50e1"
	usgB1CongregateOracle      = "d139294c-36bb-4a1e-800d-50c4a4ae9877"
	usgB1CoralMerfolkOracle    = "4ed27607-21a8-4bc3-997e-6d2242313f6d"
	usgB1CrazedSkirgeOracle    = "72a53d9e-0645-4024-a90f-73b05c2284c6"
	usgB1CongregatePerCreature = 2
)

func TestUsgB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		usgB1AbyssalHorrorOracle: "Abyssal Horror",
		usgB1AngelicPageOracle:   "Angelic Page",
		usgB1EnchantressOracle:   "Argothian Enchantress",
		usgB1SwineOracle:         "Argothian Swine",
		usgB1TreefolkOracle:      "Blanchwood Treefolk",
		usgB1BlastedOracle:       "Blasted Landscape",
		usgB1BreachOracle:        "Breach",
		usgB1CacklingFiendOracle: "Cackling Fiend",
		usgB1CatalogOracle:       "Catalog",
		usgB1ClearOracle:         "Clear",
		usgB1CongregateOracle:    "Congregate",
		usgB1CoralMerfolkOracle:  "Coral Merfolk",
		usgB1CrazedSkirgeOracle:  "Crazed Skirge",
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

// usgB1CastCreature casts a catalog creature with its printed power and
// toughness from the active seat's hand in a main phase and passes
// priority until the stack is empty or a prompt stops the table.
func usgB1CastCreature(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) *game.Player {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.DftABCast(t, g, me, game.Card{Name: name, TypeLine: typeLine, OracleID: oracle, Power: power, Toughness: toughness})
	return me
}

// --- Abyssal Horror ----------------------------------------------------------------

func TestUsgB1AbyssalHorrorMakesTheTargetPlayerDiscardTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := usgB1CastCreature(t, g, "Abyssal Horror", "Creature — Horror", usgB1AbyssalHorrorOracle, 2, 2)
	opp := g.Seats[1]
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.DiscardOwed(g, opp.ID); n != 2 {
		t.Fatalf("the target owes %d discards, want 2", n)
	}
	for _, p := range []*game.Player{me, g.Seats[2], g.Seats[3]} {
		if n := effectstest.DiscardOwed(g, p.ID); n != 0 {
			t.Errorf("%s owes %d discards; only the target discards", p.Name, n)
		}
	}
	before := opp.Hand.Size()
	effectstest.DiscardFromHand(t, g, opp.ID)
	if opp.Hand.Size() != before-2 {
		t.Errorf("hand %d → %d, want two cards discarded", before, opp.Hand.Size())
	}
	if !effectstest.HasKeywordOf(g, effectstest.FindBattlefieldByName(g, "Abyssal Horror"), "flying") {
		t.Error("Abyssal Horror has no flying")
	}
}

// --- Angelic Page --------------------------------------------------------------------

func TestUsgB1AngelicPagePumpsOnlyAnAttackingOrBlockingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	page := effectstest.PushCatalogPermanent(g, me.ID, "Angelic Page", "Creature — Angel Spirit", usgB1AngelicPageOracle, false)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Bear", 2, 2)
	if !effectstest.HasKeywordOf(g, page, "flying") {
		t.Error("Angelic Page has no flying")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	if err := effectstest.HobB4Activate(t, g, me.ID, page, 0, "", effectstest.CardRefs(idle)); err == nil {
		t.Fatal("targeted a creature that is neither attacking nor blocking")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, page, 0, "", effectstest.CardRefs(attacker)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, attacker); p != 3 || tough != 3 {
		t.Errorf("the attacker is %d/%d, want 3/3", p, tough)
	}
	if !effectstest.B13Tapped(t, g, page) {
		t.Error("Angelic Page did not tap")
	}
}

// --- Argothian Enchantress -------------------------------------------------------------

func TestUsgB1ArgothianEnchantressDrawsOnYourEnchantmentSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Argothian Enchantress", "Creature — Human Druid", usgB1EnchantressOracle, false)
	effectstest.B24CastNoncreature(t, g, "Plain Anthem", "Enchantment", "{1}{W}")
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("an enchantment spell: hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}
	effectstest.B24CastNoncreature(t, g, "Plain Bolt", "Instant", "{R}")
	hand = me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("an instant drew a card: hand %d → %d", hand, me.Hand.Size())
	}
}

func TestUsgB1ArgothianEnchantressHasShroud(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	enchantress := effectstest.PushCatalogPermanent(g, me.ID, "Argothian Enchantress", "Creature — Human Druid", usgB1EnchantressOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Breach", "Instant", usgB1BreachOracle, effectstest.CardRefs(enchantress)); err == nil {
		t.Fatal("its own controller targeted a creature with shroud")
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Breach", "Instant", usgB1BreachOracle, effectstest.CardRefs(bear)); err != nil {
		t.Fatalf("control: Breach could not target a creature without shroud: %v", err)
	}
}

// --- Argothian Swine, Blanchwood Treefolk, Coral Merfolk, Crazed Skirge ----------------

func TestUsgB1BodiesResolveWithTheirPrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		keywords               []string
	}{
		{"Argothian Swine", "Creature — Boar", usgB1SwineOracle, 3, 3, []string{"trample"}},
		{"Blanchwood Treefolk", "Creature — Treefolk", usgB1TreefolkOracle, 4, 5, nil},
		{"Coral Merfolk", "Creature — Merfolk", usgB1CoralMerfolkOracle, 2, 1, nil},
		{"Crazed Skirge", "Creature — Phyrexian Imp", usgB1CrazedSkirgeOracle, 2, 2, []string{"flying", "haste"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			usgB1CastCreature(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			id := effectstest.FindBattlefieldByName(g, tc.name)
			if !g.Battlefield.Contains(id) {
				t.Fatalf("%s did not enter", tc.name)
			}
			if p, tough := effectstest.EffectivePTOf(t, g, id); p != tc.power || tough != tc.toughness {
				t.Errorf("%s is %d/%d, want %d/%d", tc.name, p, tough, tc.power, tc.toughness)
			}
			for _, kw := range tc.keywords {
				if !effectstest.HasKeywordOf(g, id, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
			if len(g.PendingChoices) != 0 || g.Stack.Size() != 0 {
				t.Errorf("%s left something behind: choices %v, stack %d", tc.name, g.PendingChoices, g.Stack.Size())
			}
		})
	}
}

func TestUsgB1CrazedSkirgeAttacksTheTurnItArrives(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	usgB1CastCreature(t, g, "Crazed Skirge", "Creature — Phyrexian Imp", usgB1CrazedSkirgeOracle, 2, 2)
	skirge := effectstest.FindBattlefieldByName(g, "Crazed Skirge")
	usgB1CastCreature(t, g, "Coral Merfolk", "Creature — Merfolk", usgB1CoralMerfolkOracle, 2, 1)
	merfolk := effectstest.FindBattlefieldByName(g, "Coral Merfolk")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	opp := g.Seats[1]
	if err := g.DeclareAttacker(merfolk, opp.ID); err == nil {
		t.Error("control: Coral Merfolk attacked the turn it arrived")
	}
	if err := g.DeclareAttacker(skirge, opp.ID); err != nil {
		t.Errorf("Crazed Skirge could not attack the turn it arrived: %v", err)
	}
}

// --- Blasted Landscape -------------------------------------------------------------------

func TestUsgB1BlastedLandscapeTapsForColourlessAndCycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	land := effectstest.B12PlayFromHand(t, g, "Blasted Landscape", "Land", usgB1BlastedOracle, game.CastSpellParams{})
	if effectstest.B13Tapped(t, g, land) {
		t.Error("Blasted Landscape entered tapped")
	}
	if got := effectstest.LtProduced(t, g, land); !slices.Equal(got, []string{"{C}"}) {
		t.Errorf("Blasted Landscape adds %v, want {C}", got)
	}

	g = effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	hand := me.Hand.Size()
	id, _ := effectstest.CycleFromHand(t, g, "Blasted Landscape", "Land", usgB1BlastedOracle, "{C}{C}")
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id) {
		t.Error("the cycled card is not in the graveyard")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want the cycled card replaced by one draw", hand, me.Hand.Size())
	}
}

// --- Breach --------------------------------------------------------------------------------

func TestUsgB1BreachGivesPlusTwoZeroAndFear(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Breach", "Instant", usgB1BreachOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the target is %d/%d, want 4/2", p, tough)
	}
	if !effectstest.HasKeywordOf(g, bear, "fear") {
		t.Error("the target did not gain fear")
	}
}

// --- Cackling Fiend --------------------------------------------------------------------------

func TestUsgB1CacklingFiendMakesEachOpponentDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := usgB1CastCreature(t, g, "Cackling Fiend", "Creature — Zombie", usgB1CacklingFiendOracle, 2, 1)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := 1
		if p.ID == me.ID {
			want = 0
		}
		if n := effectstest.DiscardOwed(g, p.ID); n != want {
			t.Errorf("%s owes %d discards, want %d", p.Name, n, want)
		}
	}
}

// --- Catalog ----------------------------------------------------------------------------------

func TestUsgB1CatalogDrawsTwoThenDiscardsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Catalog", "Instant", usgB1CatalogOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.DiscardOwed(g, me.ID); n != 1 {
		t.Fatalf("the caster owes %d discards, want 1", n)
	}
	if me.Hand.Size() != hand+2 {
		t.Fatalf("before the discard: hand %d → %d, want two cards drawn", hand, me.Hand.Size())
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand+1 {
		t.Errorf("after the discard: hand %d → %d, want one card net", hand, me.Hand.Size())
	}
}

// --- Clear --------------------------------------------------------------------------------------

func TestUsgB1ClearDestroysTargetEnchantmentOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Clear", "Instant", usgB1ClearOracle, effectstest.CardRefs(bear)); err == nil {
		t.Fatal("Clear targeted a creature")
	}
	g = effectstest.NewCatalogGame(t)
	opp = g.Seats[1]
	anthem := effectstest.B12Permanent(g, opp.ID, "Their Anthem", "Enchantment")
	effectstest.CastCatalogSpell(t, g, "Clear", "Instant", usgB1ClearOracle, effectstest.CardRefs(anthem))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(anthem) || !opp.Graveyard.Contains(anthem) {
		t.Error("the enchantment was not destroyed")
	}
}

func TestUsgB1ClearCycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	hand := me.Hand.Size()
	effectstest.CycleFromHand(t, g, "Clear", "Instant", usgB1ClearOracle, "{C}{C}")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want the cycled card replaced by one draw", hand, me.Hand.Size())
	}
}

// --- Congregate ------------------------------------------------------------------------------------

func TestUsgB1CongregateGainsTwoPerCreatureOnTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, g.Seats[2].ID, "Another Bear", "Creature — Bear", 2, 2)
	effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	life, myLife := opp.Life, me.Life
	effectstest.CastCatalogSpell(t, g, "Congregate", "Instant", usgB1CongregateOracle, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if want := life + 3*usgB1CongregatePerCreature; opp.Life != want {
		t.Errorf("the target's life %d, want %d: three creatures, every controller's", opp.Life, want)
	}
	if me.Life != myLife {
		t.Errorf("the caster gained life: %d → %d", myLife, me.Life)
	}
}
