package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// mana_color_upfront_test.go — #1443: the colour of a pipe slot named
// BEFORE the source is tapped. The view publishes the colours each
// picking slot would offer (`color_options`), the activation takes the
// answer up front (ManaAbilityParams.Colors), a legal one is produced
// with no mana_pick queued, an illegal one is refused with nothing
// tapped, and without it the two-step activation is unchanged.

const (
	UpfrontBirdsOracle        = "d3a0b660-358c-41bd-9cd2-41fbf3491b1a"
	UpfrontCommandTowerOracle = "0895c9b7-ae7d-4bb3-af17-3b75deb50a25"
	UpfrontArcaneSignetOracle = "0bc7f093-bef0-4f1a-852c-4b75ebf54838"
	UpfrontMysticGateOracle   = "e9f5feb2-2c1a-46ce-885a-4f378d7d10af"
)

// upfrontManaRow is the `idx`-th mana ability row the controller sees
// on the wire for `id`.
func UpfrontManaRow(t *testing.T, g *game.Game, viewer, id uuid.UUID, idx int) protocol.ManaAbilityView {
	t.Helper()
	KnownToTable(g, id)
	view := protocol.ViewOfGameFor(g, viewer.String())
	for _, c := range view.Battlefield.Cards {
		if c.InstanceID != id.String() {
			continue
		}
		for _, m := range c.ManaAbilities {
			if m.Index == idx {
				return m
			}
		}
	}
	t.Fatalf("no mana ability %d on the wire for %v", idx, id)
	return protocol.ManaAbilityView{}
}
