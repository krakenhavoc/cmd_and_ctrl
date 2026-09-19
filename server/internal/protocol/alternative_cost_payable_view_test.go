package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// alternative_cost_payable_view_test.go — #695. The view shows an
// alternative-cost offer only when the server would accept it.
//
// The S28 comment in viewOfAlternativeCosts already said so — "a
// greyed-out button the server would reject is worse than no button"
// — and checked only the offer's Condition. The life half was never
// added, so Force of Will at 0 life and Snuff Out at 3 were listed,
// selectable, and refused with ErrInvalidParam the moment they were
// chosen.

// withAltCosts stubs the alternative-cost catalog hook for one oracle
// ID.
func withAltCosts(t *testing.T, oracle string, costs ...game.AlternativeCost) {
	t.Helper()
	prev := game.CatalogAlternativeCosts
	game.CatalogAlternativeCosts = func(id string) []game.AlternativeCost {
		if id == oracle {
			return costs
		}
		if prev != nil {
			return prev(id)
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogAlternativeCosts = prev })
}

// offerKeys lists the offer keys the view stamped on a hand card.
func offerKeys(t *testing.T, g *game.Game, seat *game.Player, id uuid.UUID) []string {
	t.Helper()
	v := ViewOfGameFor(g, seat.ID.String())
	card := cardInZone(v.Seats[0].Hand, id)
	if card == nil {
		t.Fatalf("the hand card is missing from its owner's view")
	}
	keys := make([]string, 0, len(card.AlternativeCosts))
	for _, offer := range card.AlternativeCosts {
		keys = append(keys, offer.Key)
	}
	return keys
}

// CR 119.4 at the view: absent below N, present at exactly N.
func TestAlternativeCostOfferHiddenBelowItsLifeComponent(t *testing.T) {
	const oracle = "test-view-snuff-out"
	g := buildActiveGame(t)
	me := g.Seats[0]
	withAltCosts(t, oracle, game.AlternativeCost{
		Key: "pay_life", Label: "Pay 4 life rather than pay this spell's mana cost", Life: 4,
	})
	spell := game.NewCard("Test Snuff Out", me.ID)
	spell.TypeLine = "Instant"
	spell.ManaCost = "{3}{B}"
	spell.OracleID = oracle
	spell.KnownBy = map[uuid.UUID]bool{me.ID: true}
	me.Hand.PushTop(spell)

	me.Life = 20
	if got := offerKeys(t, g, me, spell.InstanceID); len(got) != 1 || got[0] != "pay_life" {
		t.Fatalf("at 20 life, offers = %v, want the pay_life offer", got)
	}
	// Exactly the payment is still an offer: CR 119.4 lets a player
	// pay life down to zero, and the state-based action that follows
	// is the player's business, not the view's.
	me.Life = 4
	if got := offerKeys(t, g, me, spell.InstanceID); len(got) != 1 || got[0] != "pay_life" {
		t.Errorf("at exactly 4 life, offers = %v, want the pay_life offer", got)
	}
	me.Life = 3
	if got := offerKeys(t, g, me, spell.InstanceID); len(got) != 0 {
		t.Errorf("at 3 life, offers = %v, want none — the server would refuse this cast", got)
	}
	me.Life = 0
	if got := offerKeys(t, g, me, spell.InstanceID); len(got) != 0 {
		t.Errorf("at 0 life, offers = %v, want none", got)
	}
}

// Force of Will's shape: a life component AND a card component, and
// either one unpayable takes the offer off the table.
func TestAlternativeCostOfferHiddenWithNothingToPitch(t *testing.T) {
	const oracle = "test-view-force-of-will"
	g := buildActiveGame(t)
	me := g.Seats[0]
	withAltCosts(t, oracle, game.AlternativeCost{
		Key: "pitch", Label: "Pay 1 life, exile a blue card from your hand", Life: 1,
		PayLabel: "a blue card",
		ExileFromHand: &game.TargetSpec{
			Zones: []game.ZoneKind{game.ZoneHand},
			CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
				for _, col := range c.Colors {
					if col == "U" {
						return true
					}
				}
				return false
			},
		},
	})
	known := map[uuid.UUID]bool{me.ID: true}
	// Empty the opening hand so the only cards in it are the ones
	// this test puts there.
	me.Hand.Cards = nil
	spell := game.NewCard("Test Force of Will", me.ID)
	spell.TypeLine = "Instant"
	spell.ManaCost = "{3}{U}{U}"
	spell.OracleID = oracle
	spell.Colors = []string{"U"}
	spell.KnownBy = known
	me.Hand.PushTop(spell)

	// The spell is blue and is the only card in hand — and CR 601.2a
	// has it on the stack before the cost is paid, so it cannot pitch
	// itself.
	if got := offerKeys(t, g, me, spell.InstanceID); len(got) != 0 {
		t.Errorf("with nothing but the spell in hand, offers = %v, want none", got)
	}
	red := game.NewCard("Test Bolt", me.ID)
	red.TypeLine = "Instant"
	red.Colors = []string{"R"}
	red.KnownBy = known
	me.Hand.PushTop(red)
	if got := offerKeys(t, g, me, spell.InstanceID); len(got) != 0 {
		t.Errorf("with only a red card to pitch, offers = %v, want none", got)
	}
	blue := game.NewCard("Test Brainstorm", me.ID)
	blue.TypeLine = "Instant"
	blue.Colors = []string{"U"}
	blue.KnownBy = known
	me.Hand.PushTop(blue)
	if got := offerKeys(t, g, me, spell.InstanceID); len(got) != 1 || got[0] != "pitch" {
		t.Fatalf("with a blue card in hand, offers = %v, want the pitch offer", got)
	}
	// One point of life is a real cost.
	me.Life = 0
	if got := offerKeys(t, g, me, spell.InstanceID); len(got) != 0 {
		t.Errorf("at 0 life, offers = %v, want none — the 1-life half is unpayable", got)
	}
}

// The filter reaches a SHARED zone too. #978 moved the per-card offer
// body into stampCastOffers so exile and the library top are stamped
// beside hand, command and the graveyard; the payability predicate
// lives in viewOfAlternativeCosts, which is the one function both of
// stampCastOffers' offer stamps — the granted one and the printed one
// — route through, so every zone is filtered by construction rather
// than by four call sites remembering to.
//
// Bolas's Citadel is the grant that makes this observable: "you may
// play the top card of your library, paying life equal to its mana
// value rather than its mana cost" (CR 118.9 + CR 119.4), synthesised
// by CastPermission.LifeEqualToManaValue. A five-drop under it is a
// five-life offer.
func TestGrantedOfferInASharedZoneIsFilteredByItsLifeComponent(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]

	spell := game.NewCard("Test Five Drop", me.ID)
	spell.TypeLine = "Sorcery"
	spell.ManaCost = "{4}{B}"
	id := exileWithGrant(t, g, spell, game.CastPermission{
		Player:               me.ID,
		AltCostKey:           "citadel",
		Label:                "Pay life equal to its mana value",
		LifeEqualToManaValue: true,
	})

	offersAt := func(life int) []string {
		me.Life = life
		card := cardInZone(ViewOfGameFor(g, me.ID.String()).Exile, id)
		if card == nil {
			t.Fatalf("the exiled card is missing from the holder's view")
		}
		keys := make([]string, 0, len(card.AlternativeCosts))
		for _, offer := range card.AlternativeCosts {
			keys = append(keys, offer.Key)
		}
		return keys
	}

	if got := offersAt(20); len(got) != 1 || got[0] != "citadel" {
		t.Fatalf("at 20 life, offers = %v, want the citadel offer", got)
	}
	// Exactly the mana value is payable: CR 119.4 allows paying down
	// to zero.
	if got := offersAt(5); len(got) != 1 || got[0] != "citadel" {
		t.Errorf("at exactly 5 life, offers = %v, want the citadel offer", got)
	}
	if got := offersAt(4); len(got) != 0 {
		t.Errorf("at 4 life, offers = %v, want none — the server would refuse this cast", got)
	}
}
