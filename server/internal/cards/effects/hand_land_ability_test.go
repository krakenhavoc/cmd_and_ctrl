package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// hand_land_ability_test.go — #3280. Owner report from live play: a
// Raffine's Tower in hand could not be cycled. The server half is pinned
// here: once the turn's land drop is spent, the land is no longer a
// playable card but its cycling is still an offered activation, carried
// on the card's `zone_abilities` row and in the legal-move list, and the
// activation is one the engine accepts. (The client half — a left-click
// on such a card opens the ability menu — is in
// client/src/lib/handLandAbility.render.test.ts.)
func TestSpentLandDropStillOffersAHandLandsAbility(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, mana string
		activate                     bool
	}{
		// Cycling.
		{"Raffine's Tower", "Land — Plains Island Swamp", raffinesTowerOracle, "{C}{C}{C}", true},
		// Basic landcycling.
		{"Ash Barrens", "Land", "58257464-278e-45fa-8e0b-bcd9a7500bc1", "{C}", false},
		// Channel.
		{"Sokenzan, Crucible of Defiance", "Legendary Land", "c5ee72d5-3a9e-4fe5-8802-3286ee612055", "{C}{C}{C}{R}", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			id := pushCatalogHandCard(me, tc.name, tc.typeLine, tc.oracle)
			if err := g.AddManaForEffect(me.ID, uuid.Nil, tc.mana); err != nil {
				t.Fatal(err)
			}
			// The land drop is spent.
			if g.LandsPlayedThisTurn == nil {
				g.LandsPlayedThisTurn = map[uuid.UUID]int{}
			}
			g.LandsPlayedThisTurn[me.ID] = 1

			offered := false
			for _, m := range legal.EnumerateFor(g, me.ID) {
				if m.Source != id {
					continue
				}
				if m.Type == legal.TypeCastSpell {
					t.Errorf("the land is offered as a play with the land drop spent: %q", m.Label)
				}
				if m.Type == legal.TypeActivateAbility {
					offered = true
				}
			}
			if !offered {
				t.Fatal("no hand activation enumerated for the land")
			}

			v := protocol.FilterViewFor(protocol.ViewOfGame(g), me.ID.String())
			found := false
			for _, seat := range v.Seats {
				for _, c := range seat.Hand.Cards {
					if c.InstanceID != id.String() {
						continue
					}
					found = true
					if len(c.ZoneAbilities) != 1 || !c.ZoneAbilities[0].DiscardSelf {
						t.Errorf("zone_abilities = %+v, want the one discard-self row", c.ZoneAbilities)
					}
				}
			}
			if !found {
				t.Fatal("the land is missing from the hand view")
			}

			if !tc.activate {
				return
			}
			if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
				t.Fatalf("activating with the land drop spent: %v", err)
			}
			if !me.Graveyard.Contains(id) {
				t.Error("the discarded land is not in the graveyard")
			}
		})
	}
}
