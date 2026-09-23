package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// activation_timing_view_test.go — the VIEW half of #1208. The client
// used to derive CR 602.5d's and CR 606.3's window itself
// (`client/src/lib/timing.ts`, the last rules derivation in that
// file); the row now carries the engine's own answer, from the one
// read `ActivateCatalogAbility` and `internal/legal` also ask.
//
// Stubs the catalog hook rather than registering a real card: the
// question is about the projection, not about any card, and this
// package must be able to ask it without the catalog.

const viewActivationTimingOracle = "test-view-activation-timing"

func stubViewActivationTimings(t *testing.T, oracleID string, rules []game.ActivationTiming) {
	t.Helper()
	prev := game.CatalogActivationTimings
	game.CatalogActivationTimings = func(id string) []game.ActivationTiming {
		if id == oracleID {
			return rules
		}
		if prev != nil {
			return prev(id)
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogActivationTimings = prev })
}

// seatTimingEquipment puts an Equipment with one sorcery-speed equip
// ability and one instant-speed ability on the battlefield.
func seatTimingEquipment(g *game.Game, owner uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: id,
			Name:       "Test Blade",
			TypeLine:   "Artifact — Equipment",
			OracleID:   "00000000-0000-0000-0000-0000000000eb",
			Owner:      owner,
			Controller: owner,
			ActivatedAbilities: []game.ActivatedAbilityShape{
				{Label: "Equip {1}", Cost: game.AbilityCost{Mana: "{1}"}, SorcerySpeed: true, Equip: true},
				{Label: "{T}: Draw a card.", Cost: game.AbilityCost{Tap: true}},
			},
		})
	})
	return id
}

func abilityRows(t *testing.T, g *game.Game, id uuid.UUID) []ActivatedAbilityView {
	t.Helper()
	v := ViewOfGame(g)
	for _, c := range v.Battlefield.Cards {
		if c.InstanceID == id.String() {
			return c.ActivatedAbilities
		}
	}
	t.Fatalf("card %s missing from the battlefield view", id)
	return nil
}

// timing_closed is the engine's verdict and sorcery_speed is the
// printed clause; the row carries both, and the two differ exactly
// where a per-player statement speaks.
func TestTimingClosedIsStampedFromTheEngineRead(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	blade := seatTimingEquipment(g, me.ID)

	// The active seat's END step: not a main phase, so the sorcery
	// window is shut and the equip row says so.
	advanceTo(t, g, game.StepEnd)
	rows := abilityRows(t, g, blade)
	if len(rows) != 2 {
		t.Fatalf("got %d ability rows, want 2", len(rows))
	}
	if !rows[0].SorcerySpeed || !rows[0].TimingClosed {
		t.Errorf("equip row in an end step: sorcery_speed=%v timing_closed=%v, want both true",
			rows[0].SorcerySpeed, rows[0].TimingClosed)
	}
	if rows[1].TimingClosed {
		t.Error("an instant-speed ability must never carry timing_closed")
	}

	// A Leonin Shikari's clause, and the same row opens — with
	// sorcery_speed still true, because the ability's printed clause
	// has not changed.
	stubViewActivationTimings(t, viewActivationTimingOracle, []game.ActivationTiming{{
		Label:  "You may activate equip abilities any time you could cast an instant.",
		Timing: game.TimingFlash,
		Covers: func(q game.ActivationQuery) bool {
			return q.Ability.Equip && q.Controller == q.Source.Controller
		},
	}})
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: uuid.New(),
			Name:       "Test Shikari",
			TypeLine:   "Creature — Cat Soldier",
			OracleID:   viewActivationTimingOracle,
			Owner:      me.ID,
			Controller: me.ID,
		})
	})

	rows = abilityRows(t, g, blade)
	if rows[0].TimingClosed {
		t.Error("the grant did not clear timing_closed on the equip row")
	}
	if !rows[0].SorcerySpeed {
		t.Error("sorcery_speed is the printed clause and must survive the grant")
	}
}

// A loyalty row gets the same verdict, which is the half the client
// could not read off sorcery_speed: its timing arm is the loyalty
// one, and CR 606.3 is a rule with no printed clause behind it.
func TestTimingClosedOnALoyaltyRow(t *testing.T) {
	g := buildActiveGame(t)
	walker, owner := seatLoyaltyWalker(g, 4)

	advanceTo(t, g, game.StepEnd)
	rows := abilityRows(t, g, walker)
	if len(rows) == 0 {
		t.Fatal("no loyalty rows on the wire")
	}
	if !rows[0].TimingClosed {
		t.Error("a loyalty row in an end step is not activatable (CR 606.3)")
	}

	stubViewActivationTimings(t, viewActivationTimingOracle, []game.ActivationTiming{{
		Label:  "You may activate loyalty abilities of this any time you could cast an instant.",
		Timing: game.TimingFlash,
		Covers: func(q game.ActivationQuery) bool { return q.Ability.Loyalty },
	}})
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: uuid.New(),
			Name:       "Test Emperor",
			TypeLine:   "Enchantment",
			OracleID:   viewActivationTimingOracle,
			Owner:      owner,
			Controller: owner,
		})
	})

	rows = abilityRows(t, g, walker)
	if rows[0].TimingClosed {
		t.Error("the grant did not clear timing_closed on the loyalty row")
	}
}
