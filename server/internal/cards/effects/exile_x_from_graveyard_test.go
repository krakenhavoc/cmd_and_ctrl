package effects

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_x_from_graveyard_test.go — #1948, ADR 0189's 2026-10-16
// amendment: "exile X cards from your graveyard" as a spell's additional
// cost (the X is the one announced with the cast), and a fixed count on a
// branch of an either/or cost (Soaring Stoneglider).

const (
	skeletalScryingOracle   = "561c68d8-b727-48d8-a832-2e42d1345fac"
	harvestPyreOracle       = "00e36265-0cd8-4084-adf5-d1ed44f0d689"
	chillHauntingOracle     = "ebe61471-e236-4ff8-9416-151970b5efd6"
	hauntingMiseryOracle    = "8db6651b-c368-46fd-b9c2-69a4d6562246"
	soaringStonegliderOracl = "62d18d2b-fa68-49cd-b62d-803728ddfa8b"
)

func TestExileXFromGraveyardCardsAreFull(t *testing.T) {
	for _, id := range []string{skeletalScryingOracle, harvestPyreOracle, chillHauntingOracle, hauntingMiseryOracle, soaringStonegliderOracl} {
		spec, ok := Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: Completeness = %v, Caveats = %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// exileXFuel puts n cards of `typeLine` into p's graveyard.
func exileXFuel(p *game.Player, typeLine string, n int) []uuid.UUID {
	ids := make([]uuid.UUID, n)
	for i := range ids {
		ids[i] = uuid.New()
		p.Graveyard.PushTop(game.Card{InstanceID: ids[i], Name: "Fuel", TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	}
	return ids
}

// Skeletal Scrying: X in the mana cost and in the exile are one number,
// and the spell reads it: two cards exiled, two drawn, two life lost.
func TestSkeletalScryingExilesXDrawsXLosesX(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fuel := exileXFuel(me, "Sorcery", 3)
	for range 4 {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Draw", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	life := me.Life

	spell, err := castWithTapParams(t, g, "Skeletal Scrying", "Instant", "{X}{B}", skeletalScryingOracle,
		game.CastSpellParams{XValue: 2, ExileIDs: fuel[:2]})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	for _, id := range fuel[:2] {
		if !g.Exile.Contains(id) {
			t.Errorf("named card %s was not exiled with the cast", id)
		}
	}
	if !me.Graveyard.Contains(fuel[2]) {
		t.Error("the card not named left the graveyard")
	}
	var exiled []uuid.UUID
	g.WithWriteLock(func() { exiled = NewContext(g, g.StackMeta[spell]).Exiled() })
	if !slices.Equal(exiled, fuel[:2]) {
		t.Errorf("ctx.Exiled() = %v, want the two named cards", exiled)
	}
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand = %d, want %d (two drawn)", got, hand+2)
	}
	if got := me.Life; got != life-2 {
		t.Errorf("life = %d, want %d", got, life-2)
	}
}

// The count must be exactly the announced X, and X can't be more than the
// graveyard can pay (CR 118.3). A refused cast moves nothing.
func TestExileXCostDemandsExactlyXAndRefusesAnXTheGraveyardCantPay(t *testing.T) {
	for name, tc := range map[string]struct {
		x     int
		named int
	}{
		"fewer than X named":    {x: 2, named: 1},
		"more than X named":     {x: 1, named: 2},
		"X above the graveyard": {x: 3, named: 2},
		"cards named at X = 0":  {x: 0, named: 1},
	} {
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			bear := b12Creature(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, "Bear", "Creature — Bear", 2, 2)
			fuel := exileXFuel(me, "Sorcery", 2)
			spell, err := castWithTapParams(t, g, "Harvest Pyre", "Instant", "{1}{R}", harvestPyreOracle,
				game.CastSpellParams{XValue: tc.x, ExileIDs: fuel[:tc.named],
					Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
			if !errors.Is(err, game.ErrInvalidParam) {
				t.Fatalf("cast = %v, want ErrInvalidParam", err)
			}
			if g.Stack.Contains(spell) || g.Exile.Size() != 0 || me.Graveyard.Size() != 2 {
				t.Errorf("a refused cast moved cards: stack %v, exile %d, graveyard %d",
					g.Stack.Contains(spell), g.Exile.Size(), me.Graveyard.Size())
			}
		})
	}
}

// Harvest Pyre: X lives only in the additional cost, and it is the damage.
func TestHarvestPyreDealsXForXExiledCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	small := b12Creature(g, opp.ID, "Three Toughness", "Creature — Bear", 3, 3)
	big := b12Creature(g, opp.ID, "Four Toughness", "Creature — Bear", 4, 4)
	fuel := exileXFuel(me, "Land", 3)

	if _, err := castWithTapParams(t, g, "Harvest Pyre", "Instant", "{1}{R}", harvestPyreOracle,
		game.CastSpellParams{XValue: 3, ExileIDs: fuel,
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: small}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("3 damage did not kill a 3/3")
	}

	more := exileXFuel(me, "Land", 3)
	if _, err := castWithTapParams(t, g, "Harvest Pyre", "Instant", "{1}{R}", harvestPyreOracle,
		game.CastSpellParams{XValue: 3, ExileIDs: more,
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big}}}); err != nil {
		t.Fatalf("second cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(big) {
		t.Error("3 damage killed a 4/4")
	}
}

// X = 0 is a legal announcement that names nothing and does nothing.
func TestHarvestPyreAtXZeroExilesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, "Bear", "Creature — Bear", 1, 1)
	exileXFuel(me, "Land", 2)
	if _, err := castWithTapParams(t, g, "Harvest Pyre", "Instant", "{1}{R}", harvestPyreOracle,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("cast at X = 0: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Exile.Size() != 0 || !g.Battlefield.Contains(bear) {
		t.Errorf("X = 0 exiled %d cards or dealt damage", g.Exile.Size())
	}
}

// Chill Haunting exiles creature cards only, and the target gets -X/-X.
func TestChillHauntingExilesXCreatureCardsForMinusXMinusX(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 3, 3)
	creatures := exileXFuel(me, "Creature — Zombie", 2)
	lands := exileXFuel(me, "Land", 1)
	target := []game.TargetRef{{Kind: game.TargetCard, ID: bear}}

	if _, err := castWithTapParams(t, g, "Chill Haunting", "Instant", "{1}{B}", chillHauntingOracle,
		game.CastSpellParams{XValue: 2, ExileIDs: []uuid.UUID{creatures[0], lands[0]}, Targets: target}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a land card paid an 'X creature cards' cost: %v", err)
	}
	if _, err := castWithTapParams(t, g, "Chill Haunting", "Instant", "{1}{B}", chillHauntingOracle,
		game.CastSpellParams{XValue: 3, ExileIDs: append(append([]uuid.UUID(nil), creatures...), lands...), Targets: target}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("X = 3 with two creature cards was accepted: %v", err)
	}
	if _, err := castWithTapParams(t, g, "Chill Haunting", "Instant", "{1}{B}", chillHauntingOracle,
		game.CastSpellParams{XValue: 2, ExileIDs: creatures, Targets: target}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, bear)
	if !ok {
		t.Fatal("the 3/3 died to -2/-2")
	}
	if c.CurrentPower() != 1 || c.CurrentToughness() != 1 {
		t.Errorf("Bear is %d/%d, want 1/1", c.CurrentPower(), c.CurrentToughness())
	}
	if !me.Graveyard.Contains(lands[0]) {
		t.Error("the land card was exiled")
	}
}

// Haunting Misery deals X to a player.
func TestHauntingMiseryDealsXToTargetPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	creatures := exileXFuel(me, "Creature — Spirit", 4)
	life := opp.Life
	if _, err := castWithTapParams(t, g, "Haunting Misery", "Sorcery", "{1}{B}{B}", hauntingMiseryOracle,
		game.CastSpellParams{XValue: 4, ExileIDs: creatures,
			Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("life = %d, want %d", opp.Life, life-4)
	}
}

// Soaring Stoneglider: either branch pays, and the exile branch is out
// with fewer than two graveyard cards (CR 118.3).
func TestSoaringStonegliderPaysEitherBranch(t *testing.T) {
	exile, mana := 0, 1
	t.Run("exile two cards", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		fuel := exileXFuel(me, "Sorcery", 3)
		spell, err := castWithTapParams(t, g, "Soaring Stoneglider", "Creature — Elephant Cleric", "{2}{W}", soaringStonegliderOracl,
			game.CastSpellParams{CostBranch: &exile, ExileIDs: fuel[:2]})
		if err != nil {
			t.Fatalf("cast: %v", err)
		}
		if !g.Exile.Contains(fuel[0]) || !g.Exile.Contains(fuel[1]) || !me.Graveyard.Contains(fuel[2]) {
			t.Error("the exile branch did not exile exactly the two named cards")
		}
		passPriorityAroundTable(t, g)
		if _, ok := battlefieldCard(g, spell); !ok {
			t.Error("Stoneglider did not resolve")
		}
	})
	t.Run("a graveyard of one leaves only the mana branch", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		fuel := exileXFuel(me, "Sorcery", 1)
		card := game.Card{InstanceID: uuid.New(), OracleID: soaringStonegliderOracl, Name: "Soaring Stoneglider", Owner: me.ID}
		var exilePayable, manaPayable bool
		g.WithWriteLock(func() {
			exilePayable = g.AdditionalCostBranchPayableLocked(me.ID, card, exile)
			manaPayable = g.AdditionalCostBranchPayableLocked(me.ID, card, mana)
		})
		if exilePayable || !manaPayable {
			t.Errorf("payable: exile %v, mana %v; want only the mana branch", exilePayable, manaPayable)
		}
		if _, err := castWithTapParams(t, g, "Soaring Stoneglider", "Creature — Elephant Cleric", "{2}{W}", soaringStonegliderOracl,
			game.CastSpellParams{CostBranch: &exile, ExileIDs: fuel}); !errors.Is(err, game.ErrInvalidParam) {
			t.Fatalf("one card paid 'exile two cards': %v", err)
		}
		if _, err := castWithTapParams(t, g, "Soaring Stoneglider", "Creature — Elephant Cleric", "{2}{W}", soaringStonegliderOracl,
			game.CastSpellParams{CostBranch: &mana}); err != nil {
			t.Fatalf("the mana branch: %v", err)
		}
		if !me.Graveyard.Contains(fuel[0]) {
			t.Error("the mana branch exiled a card")
		}
	})
}
