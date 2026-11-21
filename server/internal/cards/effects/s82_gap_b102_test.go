package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB102YennettOracle  = "4eec6838-3565-407f-8941-ce9421402f0b"
	gapB102ComposerOracle = "6b5d1b0a-544c-41e6-bc32-c2b0c9bb4576"
	gapB102HatcheryOracle = "6227fb65-d7ef-4bc6-a076-9aa10941632c"
	gapB102HeraldOracle   = "9ef24a79-fade-4c38-b63c-22983206b4ad"
)

func TestGapB102CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB102YennettOracle:  "Yennett, Cryptic Sovereign",
		gapB102ComposerOracle: "Composer of Spring",
		gapB102HatcheryOracle: "Hatchery Sliver",
		gapB102HeraldOracle:   "Herald of the Pantheon",
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

// yennettAttacks puts Yennett on the active player's side with `top` on the
// library and attacks with it, leaving the trigger resolving.
func gapB102YennettAttacks(t *testing.T, top game.Card) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	yennett := b12Push(g, me.ID, "Yennett, Cryptic Sovereign", "Legendary Creature — Sphinx", gapB102YennettOracle, 3, 5)
	topID := seedLibraryTop(me, top.Name, top.TypeLine, top.ManaCost)
	for i := range me.Library.Cards {
		if me.Library.Cards[i].InstanceID == topID {
			me.Library.Cards[i].Power, me.Library.Cards[i].Toughness = top.Power, top.Toughness
		}
	}
	declareAttack(t, g, g.Seats[(g.Turn.ActiveSeat+1)%4].ID, yennett)
	passPriorityAroundTable(t, g)
	return g, me, yennett, topID
}

func TestGapB102YennettCastsAnOddCardFreeAndDrawsNothing(t *testing.T) {
	g, me, _, top := gapB102YennettAttacks(t, game.Card{Name: "Three Drop", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Power: 3, Toughness: 3})
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, top)
	hand := me.Hand.Size()
	if err := castThroughResolutionCast(t, g, me.ID, top, game.CastSpellParams{}); err != nil {
		t.Fatalf("casting the revealed card: %v", err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, top) != game.ZoneBattlefield {
		t.Fatalf("the cast creature is in %v", b12ZoneOf(g, top))
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d: a card was drawn although the revealed card was cast", hand, me.Hand.Size())
	}
}

func TestGapB102YennettDrawsTheCardWhenYouDeclineAnOddCast(t *testing.T) {
	g, me, _, top := gapB102YennettAttacks(t, game.Card{Name: "Three Drop", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Power: 3, Toughness: 3})
	hand := me.Hand.Size()
	declineResolutionCast(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(top) || me.Hand.Size() != hand+1 {
		t.Errorf("declining did not draw the revealed card (hand %d -> %d)", hand, me.Hand.Size())
	}
}

func TestGapB102YennettNeverOffersAnEvenCardOrALand(t *testing.T) {
	for _, top := range []game.Card{
		{Name: "Two Drop", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		{Name: "Free Spell", TypeLine: "Instant", ManaCost: "{0}"},
		{Name: "Forest", TypeLine: "Basic Land — Forest"},
	} {
		g, me, _, id := gapB102YennettAttacks(t, top)
		if resolutionCastOfferFor(g, me.ID) != nil {
			t.Errorf("%s: a cast was offered", top.Name)
		}
		if !me.Hand.Contains(id) {
			t.Errorf("%s: it was not drawn", top.Name)
		}
	}
}

// Composer of Spring: an enchantment entering puts a land from hand tapped;
// a creature card is not offered until six enchantments, and an
// opponent's enchantment triggers nothing.
func TestGapB102ComposerPutsALandWhenAnEnchantmentEnters(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Composer of Spring", "Creature — Satyr Bard", gapB102ComposerOracle, 1, 3)
	land := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	bear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Hand Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Test Enchantment", "Enchantment", "", nil)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt after an enchantment entered")
	}
	if slices.Contains(pick.ChooseCards, bear) {
		t.Fatal("a creature card was offered with fewer than six enchantments")
	}
	answerChooseCards(t, g, me.ID, land)
	if !g.Battlefield.Contains(land) || !c2b16IsTapped(g, land) {
		t.Error("the land is not on the battlefield tapped")
	}
	if g.LandsPlayedThisTurnFor(me.ID) != 0 {
		t.Error("the put used the land drop")
	}
}

func TestGapB102ComposerOffersACreatureWithSixEnchantments(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Composer of Spring", "Creature — Satyr Bard", gapB102ComposerOracle, 1, 3)
	for i := 0; i < 5; i++ {
		b12Permanent(g, me.ID, "Old Enchantment", "Enchantment")
	}
	bear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Hand Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Test Enchantment", "Enchantment", "", nil)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil || !slices.Contains(pick.ChooseCards, bear) {
		t.Fatal("with six enchantments the creature card was not offered")
	}
	answerChooseCards(t, g, me.ID, bear)
	if !g.Battlefield.Contains(bear) || !c2b16IsTapped(g, bear) {
		t.Error("the creature is not on the battlefield tapped")
	}
}

func TestGapB102ComposerIgnoresAnOpponentsEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	b12Push(g, me.ID, "Composer of Spring", "Creature — Satyr Bard", gapB102ComposerOracle, 1, 3)
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	b12Push(g, opp.ID, "Their Enchantment", "Enchantment", "", 0, 0)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("an opponent's enchantment triggered Composer of Spring")
	}
}

// Herald of the Pantheon: enchantment spells cost {1} less and gain a life
// when cast; other spells neither.
func TestGapB102HeraldDiscountsAndGainsLifeForEnchantmentSpells(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Herald of the Pantheon", "Creature — Centaur Shaman", gapB102HeraldOracle, 2, 2)
	ench := trekB39Hand(g, me, "Test Enchantment", "Enchantment", "{1}{W}", "", 0, 0)
	rock := trekB39Hand(g, me, "Test Rock", "Artifact", "{1}{W}", "", 0, 0)
	life := me.Life
	floatMana(t, g, me, "{W}")
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("an artifact spell got the discount")
	}
	if err := g.CastSpell(me.ID, ench, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("an enchantment for one mana less: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want +1", life, me.Life)
	}
	if !g.Battlefield.Contains(ench) {
		t.Error("the enchantment did not resolve")
	}
}

// Hatchery Sliver: its own replicate {1}{G}, and replicate on every Sliver
// spell at the spell's own mana cost, for me only.
func TestGapB102HatcheryGivesSliverSpellsReplicateAtTheirManaCost(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Hatchery Sliver", "Creature — Sliver", gapB102HatcheryOracle, 2, 2)
	sliver := trekB39Hand(g, me, "Test Sliver", "Creature — Sliver", "{1}{B}", "", 1, 1)
	bear := trekB39Hand(g, me, "Test Bear", "Creature — Bear", "{1}{B}", "", 2, 2)
	floatMana(t, g, me, "{B}{B}{B}{C}{C}{C}")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}}); err == nil {
		t.Fatal("a non-Sliver spell was offered replicate")
	}
	// {1}{B} plus two replicates of {1}{B}: three castings' worth of mana.
	if err := g.CastSpell(me.ID, sliver, game.CastSpellParams{Strict: true, OptionalCosts: []int{0, 0}}); err != nil {
		t.Fatalf("a Sliver spell with replicate twice: %v", err)
	}
	if left := me.ManaPool.Total(); left != 0 {
		t.Errorf("%d mana left after paying {1}{B} three times", left)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Test Sliver"); n != 3 {
		t.Fatalf("%d Test Slivers on the battlefield, want the spell and two token copies", n)
	}
}

func TestGapB102HatcheryOnlyGrantsItsControllersSpells(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	b12Push(g, opp.ID, "Hatchery Sliver", "Creature — Sliver", gapB102HatcheryOracle, 2, 2)
	sliver := trekB39Hand(g, me, "Test Sliver", "Creature — Sliver", "{1}{B}", "", 1, 1)
	floatMana(t, g, me, "{B}{B}{B}{C}{C}{C}")
	if err := g.CastSpell(me.ID, sliver, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}}); err == nil {
		t.Fatal("an opponent's Hatchery Sliver gave my Sliver spell replicate")
	}
}

func TestGapB102HatcheryReplicatesItself(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	hatchery := trekB39Hand(g, me, "Hatchery Sliver", "Creature — Sliver", "{1}{G}", gapB102HatcheryOracle, 2, 2)
	floatMana(t, g, me, "{G}{G}{C}{C}")
	if err := g.CastSpell(me.ID, hatchery, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("Hatchery Sliver with its own replicate: %v", err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Hatchery Sliver"); n != 2 {
		t.Fatalf("%d Hatchery Slivers, want the card and one token copy", n)
	}
}
