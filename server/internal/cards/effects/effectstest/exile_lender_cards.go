package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_lender_cards_test.go — ADR 0155 PR 3 (#3201): the borrowers whose
// lenders are in exile. A resolving ability's exile now stamps
// Card.ExiledWith (ExileTarget.Linked), and the "exiled with" borrowers
// read that link. Each test asserts what a player gets: which permanent
// has which ability, and what using it does.

const (
	ElImpostorOracle = "96e25887-e556-423a-8a69-0cd148216f6a"
	ElCrawlerOracle  = "74cbe86b-cc8c-4a5f-93c3-f4eaced50393"
	ElWelderOracle   = "ea2cbf5a-2def-450b-9db8-40a5b6a604d0"
	ElCauldronOracle = "c259e16f-2a44-4552-8678-815f757a02e8"
	ElStewardOracle  = "1bb7a942-c4b9-47a2-9782-3ec3efcdaff5"
	ElForgeOracle    = "a14d89b6-23eb-4337-8dee-93c39b0efbf1"
	ElRexOracle      = "7ad52f9c-48d8-486f-90b7-73ce5cb3b9db"
	ElMairsilOracle  = "ed5fbae0-2dca-44a5-84dc-62674b3a92a6"
	ElJewelOracle    = "c0d2bdd9-1e68-4d3a-83f8-ec2ba440a494"
)

// elRef is the source OBJECT a linked exile names.
func ElRef(t *testing.T, g *game.Game, id uuid.UUID) game.PermissionCardRef {
	t.Helper()
	c := GaLayered(t, g, id)
	return game.PermissionCardRef{ID: id, Epoch: c.ObjectEpoch}
}

// elExiledCard is the card `id` as it sits in exile.
func ElExiledCard(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	for _, c := range g.Exile.Cards {
		if c.InstanceID == id {
			return c
		}
	}
	t.Fatalf("%s is not in exile", id)
	return game.Card{}
}

// elPushExiled puts a card straight into exile, linked to `by` (the zero
// ref for an unlinked card), and drops the layer cache.
func ElPushExiled(g *game.Game, owner uuid.UUID, by game.PermissionCardRef, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: owner, Controller: owner, ExiledWith: by})
	g.BumpLayerVersionForTest()
	return id
}

// elActivate activates a permanent's own ability `index` at one card
// target, then resolves the stack.
func ElActivate(t *testing.T, g *game.Game, p *game.Player, card uuid.UUID, index int, target uuid.UUID) {
	t.Helper()
	params := game.ActivateAbilityParams{}
	if target != uuid.Nil {
		params.Targets = []game.TargetRef{{Kind: game.TargetCard, ID: target}}
	}
	if err := g.ActivateCatalogAbility(p.ID, card, index, params); err != nil {
		t.Fatalf("activate ability %d: %v", index, err)
	}
	PassPriorityAroundTable(t, g)
}

// elToHand moves a card out of exile to its owner's hand, the move a
// DependsOnExile static must notice.
func ElToHand(t *testing.T, g *game.Game, owner, id uuid.UUID) {
	t.Helper()
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneExile}, game.ZoneRef{Kind: game.ZoneHand, Owner: owner}, id); err != nil {
		t.Fatalf("move out of exile: %v", err)
	}
}
