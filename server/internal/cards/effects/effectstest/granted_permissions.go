package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// granted_permissions_test.go — S42 / ADR 0066. The game package pins
// the model (cast_permission_test.go there); this file pins the eight
// cards end to end against the REAL catalog, which is the half a hook
// stub cannot check: that the card file actually declared the
// permission, that the price it names is the price the cast path
// charges, and that the window really closes.

const (
	SnapcasterOracle   = "2bb2eda7-3b38-4c56-870f-c3218a1056f5"
	PastInFlamesOracle = "37a18736-5fe2-4897-809b-013497bdd890"
	GrimLockerOracle   = "29931e18-9dac-46fe-b9a4-72d838d79882"
	BreachOracle       = "27e0948b-9916-473b-8d8c-a51bdfbc7457"
	CitadelOracle      = "2bd111bb-ce02-414c-b5b7-e0e037d8d96b"
	OracleMulDayaID    = "6c5b02eb-7829-436a-8555-fea200e4b67f"
	CourserOracle      = "46779609-4fa7-4fd2-b5b4-7d4d749339e6"
	RealmwalkerOracle  = "b81eaa2f-0554-41c6-bdf6-d1cb73b8f56f"
)

// pickTriggerTarget answers the pick_target prompt a targeted trigger
// queues (CR 603.3d) with one card.
func PickTriggerTarget(t *testing.T, g *game.Game, chooser, card uuid.UUID) {
	t.Helper()
	pick := LatestPickTarget(g, chooser)
	if pick == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
	if err := g.ResolvePickTarget(pick.ID, chooser, game.TargetRef{Kind: game.TargetCard, ID: card}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}

// grantedPermissionOn is the one read every test here makes: does an
// effect let `seat` cast this card out of the zone it is sitting in?
func GrantedPermissionOn(g *game.Game, seat, card uuid.UUID, zone game.ZoneKind) *game.CastPermission {
	var out *game.CastPermission
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(card)
		if !ok {
			return
		}
		out = g.CastPermissionForLocked(seat, c, zone)
	})
	return out
}

// seedTopOfLibrary replaces the seat's library with one named card on
// top, so a library-top test knows exactly what is there.
func SeedTopOfLibrary(p *game.Player, c game.Card) uuid.UUID {
	if c.InstanceID == uuid.Nil {
		c.InstanceID = uuid.New()
	}
	c.Owner, c.Controller = p.ID, p.ID
	p.Library.Cards = nil
	p.Library.PushTop(c)
	return c.InstanceID
}
