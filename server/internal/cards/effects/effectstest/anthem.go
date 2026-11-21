package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// anthem_test.go covers Glorious Anthem (Layer 7c +1/+1 to your
// creatures) end-to-end through the S16 layer engine. The catalog
// blank-import pulls Glorious Anthem's init() into scope; pushing
// it onto the battlefield fires the layer listener which bumps
// layerVersion; the next snapshot resolves the layer, and
// Effective() reads the post-7c power/toughness.

const GloriousAnthemOracle = "e3886fe8-9b76-4613-8891-4ec74657c087"

// pushBattlefieldCardWithTimestamp seeds a card directly on the
// battlefield AND fires EventZoneMove so the layer listener stamps
// EnteredBattlefieldAt + bumps layerVersion. Prefer this over
// raw PushTop for any test that exercises the layer engine.
//
// S18: the listener also sets SummonedThisTurn = true on entry.
// Tests that use this helper are semantically "this card is
// already in play and ready to act," not "this card just ETB'd
// this turn" — so we clear the sickness flag post-stamp. Tests
// specifically exercising summoning sickness should drive ETB
// through the real MoveCardByID path.
func PushBattlefieldCardWithTimestamp(g *game.Game, c game.Card) uuid.UUID {
	g.Battlefield.PushTop(c)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind:    game.EventZoneMove,
			CardID:  c.InstanceID,
			OldZone: game.ZoneHand,
			NewZone: game.ZoneBattlefield,
		})
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == c.InstanceID {
				g.Battlefield.Cards[i].SummonedThisTurn = false
				break
			}
		}
	})
	return c.InstanceID
}

// effectivePower walks the battlefield for the named card and
// returns its post-layer power. Forces a snapshot so the recompute
// runs. Fails the test if the card isn't on the battlefield.
func EffectivePower(t *testing.T, g *game.Game, cardID uuid.UUID) int {
	t.Helper()
	var p int
	var found bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID != cardID {
				continue
			}
			eff := c.Effective()
			p = eff.Power
			found = true
			return
		}
	})
	if !found {
		t.Fatalf("card %s not on battlefield", cardID)
	}
	return p
}

func EffectiveToughness(t *testing.T, g *game.Game, cardID uuid.UUID) int {
	t.Helper()
	var tg int
	var found bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID != cardID {
				continue
			}
			eff := c.Effective()
			tg = eff.Toughness
			found = true
			return
		}
	})
	if !found {
		t.Fatalf("card %s not on battlefield", cardID)
	}
	return tg
}
