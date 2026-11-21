package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tap_cost_test.go — S22: convoke and waterbend, the two names for
// "tap permanents you control to help pay for this".
//
// The tests that matter here are the ones about the COST, not about
// the cards: what the tapping buys, what it is allowed to tap, and
// what happens when a caster asks for more than the cost is worth.
// The pure cost arithmetic (the colour matching, the waterbend
// split) is tested next to the code in game/tap_cost_test.go.

const (
	TheWanderingRescuerOracle = "b8ef65df-f8e7-44e3-9864-9c127232a2b6"
	WanderingRescuerCost      = "{3}{W}{W}"
	restorationCost           = "{U}{U}"
	tapCostProbeOracle        = "8b2d1f04-6d1e-4a0f-9f2a-7f1e6d2c5a01"
)

// castWithTapParams is castCatalogSpell with the printed mana cost and
// the full announce-time params exposed — X, tap IDs, the strict
// gate — because every test below turns on one of them.
func CastWithTapParams(t *testing.T, g *game.Game, name, typeLine, manaCost, oracleID string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		ManaCost:   manaCost,
		OracleID:   oracleID,
		Owner:      active.ID,
		Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return id, g.CastSpell(active.ID, id, params)
}

func CastRestoration(t *testing.T, g *game.Game, params game.CastSpellParams) error {
	t.Helper()
	_, err := CastWithTapParams(t, g, "Waterbender's Restoration", "Instant — Lesson",
		restorationCost, waterbendersRestorationOracle, params)
	return err
}

// pushTapCostPermanent seeds one untapped permanent under the given
// seat's control with the supplied type line and colours.
func PushTapCostPermanent(g *game.Game, owner uuid.UUID, name, typeLine string, colors ...string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   tapCostProbeOracle,
		Colors:     colors,
		Power:      1,
		Toughness:  1,
		Owner:      owner,
		Controller: owner,
	})
	return id
}

func PushTapCostSoldiers(g *game.Game, owner uuid.UUID, n int) []uuid.UUID {
	ids := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		ids = append(ids, PushTapCostPermanent(g, owner, "Soldier", "Creature — Soldier", "W"))
	}
	return ids
}

func TapCostTapped(g *game.Game, id uuid.UUID) bool {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.Tapped
		}
	}
	return false
}
