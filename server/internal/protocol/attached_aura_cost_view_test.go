package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #1945: the sacrifice picker for "Sacrifice an Aura attached to this
// creature" lists the Auras on the paying creature and no others — the
// same set the engine accepts and the enumerator pays from.
func TestSacrificePickerListsOnlyAurasAttachedToTheSource(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	push := func(c game.Card) uuid.UUID {
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = owner, owner
		g.Battlefield.PushTop(c)
		return c.InstanceID
	}
	troll := push(game.Card{Name: "Faunsbane Troll", TypeLine: "Creature — Troll", OracleID: "b707c131-de13-4d4d-839d-b9f47d62f090", Power: 4, Toughness: 4})
	bear := push(game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	role := func(host uuid.UUID) uuid.UUID {
		r := effects.RoleToken(effects.RoleMonster)
		r.AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: host}
		r.AttachedAt = 1
		return push(r)
	}
	own, stray := role(troll), role(bear)
	g.BumpLayerVersionForTest()

	o := vehicleView(t, g, troll).ActivatedAbilities[0].SacrificeOptions
	if o == nil || len(o.Cards) != 1 || o.Cards[0] != own.String() {
		t.Errorf("picker cards = %+v, want only the Troll's Role %s (not %s)", o, own, stray)
	}
}
