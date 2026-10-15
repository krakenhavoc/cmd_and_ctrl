package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hand_exile_x_costs_test.go — #3175: the two hand-exile costs the cast
// path could not say.
//
//   - Shining Shoal: "You may exile a white card with mana value X from
//     your hand rather than pay this spell's mana cost." An alternative
//     cost that includes X (CR 107.3a / 107.3b): X is announced, the
//     pitched card must match it, and X on the stack is that value.
//   - March of Swirling Mist: "As an additional cost to cast this spell,
//     you may exile any number of blue cards from your hand. This spell
//     costs {2} less to cast for each card exiled this way." Any number,
//     each one {2} of generic off the total (CR 601.2f), {X} included.

const shiningShoalOracle = "9937dae8-e639-46ad-849b-7e7be93bbbad"

func castShiningShoal(t *testing.T, g *game.Game, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	return castWithTapParams(t, g, "Shining Shoal", "Instant — Arcane", "{X}{W}{W}", shiningShoalOracle, params)
}

func shoalPitch(x int, card uuid.UUID, target uuid.UUID) game.CastSpellParams {
	return game.CastSpellParams{
		AlternativeCost: "pitch",
		AltCostIDs:      []uuid.UUID{card},
		XValue:          x,
		Targets:         []game.TargetRef{{Kind: game.TargetPlayer, ID: target}},
		Strict:          true,
	}
}

// The pitch casts the spell for no mana, exiles the white card with the
// spell on the stack, and the spell's X is the card's mana value.
func TestShiningShoalPitchSetsXToTheExiledCardsManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	white := pushHandCardWithManaCost(me, "White Knight", "Creature — Human Knight", "{W}{W}")

	spell, err := castShiningShoal(t, g, shoalPitch(2, white, opp.ID))
	if err != nil {
		t.Fatalf("pitching a mana value 2 white card at X = 2: %v", err)
	}
	if !g.Exile.Contains(white) || me.Hand.Contains(white) {
		t.Fatal("the pitched card was not exiled as the cost")
	}
	item := g.StackMeta[spell]
	if item == nil {
		t.Fatal("Shining Shoal is not on the stack")
	}
	if item.XValue != 2 || item.AltCost != "pitch" {
		t.Errorf("stack item X = %d, alt cost %q; want X = 2 under the pitch", item.XValue, item.AltCost)
	}
}

// The card and X must agree; a card of another colour can't pay; and a
// refused cast moves nothing.
func TestShiningShoalPitchRefusesAMismatchedOrNonWhiteCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	white := pushHandCardWithManaCost(me, "White Knight", "Creature — Human Knight", "{W}{W}")
	blue := pushHandCardWithManaCost(me, "Blue Drake", "Creature — Drake", "{1}{U}")

	for _, tc := range []struct {
		name string
		x    int
		card uuid.UUID
	}{
		{"X above the card's mana value", 3, white},
		{"X below the card's mana value", 1, white},
		{"X = 0, which CR 107.3b does not force here", 0, white},
		{"a blue card of the right mana value", 2, blue},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := castShiningShoal(t, g, shoalPitch(tc.x, tc.card, opp.ID))
			if err == nil {
				t.Fatal("cast accepted a pitch that does not pay \"a white card with mana value X\"")
			}
			if !errors.Is(err, game.ErrInvalidParam) {
				t.Errorf("err = %v, want ErrInvalidParam", err)
			}
			if g.Exile.Contains(white) || g.Exile.Contains(blue) {
				t.Error("a refused cast exiled a card")
			}
		})
	}
}

// The printed cost is untouched: X is announced and paid in mana.
func TestShiningShoalHardCastStillPaysX(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"},
		game.ManaToken{Color: "W"}, game.ManaToken{Color: "W"})
	spell, err := castShiningShoal(t, g, game.CastSpellParams{
		XValue:  3,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		Strict:  true,
	})
	if err != nil {
		t.Fatalf("hard cast at X = 3: %v", err)
	}
	if got := g.StackMeta[spell].XValue; got != 3 {
		t.Errorf("X = %d, want 3", got)
	}
}

// The offer is shown with any white card in hand (X is the caster's to
// choose), and not without one.
func TestShiningShoalOffersThePitchWithAnyWhiteCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushHandCardWithManaCost(me, "Shining Shoal", "Instant — Arcane", "{X}{W}{W}")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == id {
				me.Hand.Cards[i].OracleID = shiningShoalOracle
			}
		}
	})
	offered := func() bool {
		var ok bool
		g.ReadSnapshot(func() {
			for _, c := range me.Hand.Cards {
				if c.InstanceID != id {
					continue
				}
				for _, o := range g.CastOffersForLocked(me.ID, c, game.ZoneHand, nil) {
					if o != nil && o.Key == "pitch" {
						ok = true
					}
				}
			}
		})
		return ok
	}
	if offered() {
		t.Error("the pitch is offered with no white card in hand")
	}
	pushHandCardWithManaCost(me, "White Giant", "Creature — Giant", "{5}{W}")
	if !offered() {
		t.Error("the pitch is not offered with a white card in hand")
	}
}

// --- March of Swirling Mist --------------------------------------------

func castMarch(t *testing.T, g *game.Game, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	return castWithTapParams(t, g, "March of Swirling Mist", "Instant", "{X}{U}", marchOfSwirlingMistOracle, params)
}

// Two blue cards take {4} off: an X = 4 March costs {U}, the cards are
// exiled with the spell on the stack, and X stays 4.
func TestMarchOfSwirlingMistExilesBlueCardsForItsDiscount(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushVoltronCreature(g, opp, "Theirs A")
	b := pushVoltronCreature(g, opp, "Theirs B")
	blue1 := pushHandCardWithManaCost(me, "Blue Drake", "Creature — Drake", "{1}{U}")
	blue2 := pushHandCardWithManaCost(me, "Blue Sphinx", "Creature — Sphinx", "{4}{U}{U}")
	advanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})

	spell, err := castMarch(t, g, game.CastSpellParams{
		XValue:           4,
		Targets:          []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}},
		ExileFromHandIDs: []uuid.UUID{blue1, blue2},
		Strict:           true,
	})
	if err != nil {
		t.Fatalf("X = 4 with two blue cards exiled and {U} floating: %v", err)
	}
	for _, id := range []uuid.UUID{blue1, blue2} {
		if !g.Exile.Contains(id) {
			t.Errorf("%v was not exiled as the cost", id)
		}
	}
	if got := g.StackMeta[spell].XValue; got != 4 {
		t.Errorf("X = %d, want 4: the discount pays for X, it does not change it", got)
	}
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("pool has %d mana left, want the {U} spent", got)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if _, ok := vvPhasedOutCard(g, id); !ok {
			t.Errorf("%v should have phased out", id)
		}
	}
}

// One card is {2} off, not {4}: an X = 4 March with one blue card exiled
// still owes {2}{U}, and a pool of {U} is refused with nothing exiled.
func TestMarchOfSwirlingMistDiscountIsTwoPerCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushVoltronCreature(g, opp, "Theirs")
	blue := pushHandCardWithManaCost(me, "Blue Drake", "Creature — Drake", "{1}{U}")
	advanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})

	_, err := castMarch(t, g, game.CastSpellParams{
		XValue:           4,
		Targets:          []game.TargetRef{{Kind: game.TargetCard, ID: a}},
		ExileFromHandIDs: []uuid.UUID{blue},
		Strict:           true,
	})
	if err == nil {
		t.Fatal("an X = 4 March with one card exiled was cast for {U}")
	}
	if g.Exile.Contains(blue) || !me.Hand.Contains(blue) {
		t.Error("a refused cast exiled the blue card")
	}

	var price game.CastPrice
	var perr error
	var cast game.Card
	g.ReadSnapshot(func() {
		for _, c := range me.Hand.Cards {
			if c.OracleID == marchOfSwirlingMistOracle {
				cast = c
			}
		}
		price, perr = g.PriceCastForEffect(me.ID, cast, game.CastSpellParams{XValue: 4, ExileFromHandIDs: []uuid.UUID{blue}})
	})
	if perr != nil {
		t.Fatalf("price: %v", perr)
	}
	if got := price.Total.GenericWithX(4); got != 2 {
		t.Errorf("generic owed at X = 4 with one card exiled = %d, want 2", got)
	}
}

// Only distinct blue cards in the caster's own hand pay, and a refused
// cast exiles nothing.
func TestMarchOfSwirlingMistRefusesAnIllegalExile(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushVoltronCreature(g, opp, "Theirs")
	blue := pushHandCardWithManaCost(me, "Blue Drake", "Creature — Drake", "{1}{U}")
	red := pushHandCardWithManaCost(me, "Red Ogre", "Creature — Ogre", "{2}{R}")
	theirs := pushHandCardWithManaCost(opp, "Their Drake", "Creature — Drake", "{1}{U}")

	for _, tc := range []struct {
		name string
		ids  []uuid.UUID
	}{
		{"a red card", []uuid.UUID{red}},
		{"the same blue card twice", []uuid.UUID{blue, blue}},
		{"an opponent's blue card", []uuid.UUID{theirs}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := castMarch(t, g, game.CastSpellParams{
				XValue:           1,
				Targets:          []game.TargetRef{{Kind: game.TargetCard, ID: a}},
				ExileFromHandIDs: tc.ids,
			})
			if err == nil {
				t.Fatal("cast accepted an illegal exile")
			}
			if !errors.Is(err, game.ErrInvalidParam) && !errors.Is(err, game.ErrCardNotFound) {
				t.Errorf("err = %v, want a refusal of the payment", err)
			}
			if g.Exile.Contains(blue) || g.Exile.Contains(red) || g.Exile.Contains(theirs) {
				t.Error("a refused cast exiled a card")
			}
		})
	}
}

// exile_from_hand_ids on a spell with no such cost is refused rather than
// ignored.
func TestExileFromHandIDsWithoutTheCostIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	blue := pushHandCardWithManaCost(me, "Blue Drake", "Creature — Drake", "{1}{U}")
	_, err := castShiningShoal(t, g, game.CastSpellParams{
		Targets:          []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		ExileFromHandIDs: []uuid.UUID{blue},
	})
	if err == nil {
		t.Fatal("cast accepted exile_from_hand_ids for a spell with no hand-exile cost")
	}
	if g.Exile.Contains(blue) {
		t.Error("a refused cast exiled the card")
	}
}

// Register refuses the X-bound pitch shapes nothing reads (#3175).
func TestRegisterRefusesAnUnreadAltCostXBound(t *testing.T) {
	white := func() *game.TargetSpec {
		return CardInYourHand("a white card with mana value X from your hand", OfColor("W")).WithManaValueEqualsX()
	}
	mustPanic(t, "does not read X", func() {
		checkAltCostXBound(Spec{Name: "No X Pitch"}, Pitch("pitch", 0, white(), "a white card"))
	})
	mustPanic(t, "more than one card", func() {
		spec := white()
		spec.Min, spec.Max = 2, 2
		checkAltCostXBound(Spec{Name: "Two Card Pitch", XMatters: true}, Pitch("pitch", 0, spec, "two white cards"))
	})
	mustPanic(t, "other than \"mana value X\"", func() {
		spec := CardInYourHand("a white card", OfColor("W"))
		spec.ManaValueAtMostX = true
		checkAltCostXBound(Spec{Name: "At Most Pitch", XMatters: true}, Pitch("pitch", 0, spec, "a white card"))
	})
	mustPanic(t, "only a hand pitch", func() {
		checkAltCostXBound(Spec{Name: "Bounce X", XMatters: true}, game.AlternativeCost{
			Key: "bounce", ReturnToHand: TargetCreature("a creature").WithManaValueEqualsX(),
		})
	})
	// The printed shape passes.
	checkAltCostXBound(Spec{Name: "Shining Shoal", XMatters: true}, Pitch("pitch", 0, white(), "a white card"))
}

// Register refuses the hand-exile cost anywhere but the mandatory slot.
func TestRegisterRefusesAnOptionalHandExileCost(t *testing.T) {
	cost := *ExileAnyNumberFromHandCost("blue cards", "U")
	cost.Optional, cost.Key = true, "kicker"
	mustPanic(t, "exiles cards from hand", func() {
		checkEitherCost(Spec{Name: "Optional March", OptionalCosts: []game.AdditionalCost{cost}})
	})
}
