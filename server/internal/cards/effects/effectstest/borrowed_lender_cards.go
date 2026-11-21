package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// borrowed_lender_cards_test.go — the first cards on ADR 0155's seam (its
// PR 2): the battlefield and graveyard lenders. Each test asserts what a
// player gets — who has which ability, who may activate it, what pays for
// it — rather than a label.

const (
	BlMarvinOracle      = "22f0dbbb-9d8d-42bc-8903-fc0b65e00952"
	BlKrajOracle        = "93cfa771-067f-44ed-9e29-6caf698ee9aa"
	BlRobaranOracle     = "1d8c8c71-d569-450d-adc1-1a1b9aea0123"
	BlDranaOracle       = "7abc2a1b-b758-480b-9415-f8f5369ab52f"
	BlRefractorOracle   = "ae3517db-bb96-4c29-91dc-5bd68683dced"
	BlSharkeyOracle     = "93a476c0-c77e-43b0-8695-ada40368a002"
	BlOozeOracle        = "2fdeb920-2e48-4308-9517-743eec219c04"
	BlSafehouseOracle   = "08826de5-a2d7-4732-a5f7-f580507a7844"
	BlThranduilOracle   = "4a5f76e7-40be-4b06-9935-4a3b2672e1c2"
	BlTrazynOracle      = "1ef3b48a-55ef-422e-ac7f-e1c1057eec2c"
	BlLlanowarOracle    = "68954295-54e3-4303-a6bc-fc4547a4e3a3"
	BlEvolvingWildsOrcl = "a75445d3-1303-4bb5-89ad-26ea93fecd48"
)

// blBorrowedMana is a permanent's lent mana rows, in order.
func BlBorrowedMana(t *testing.T, g *game.Game, id uuid.UUID) (idx []int, origins game.AbilityOrigins) {
	t.Helper()
	_, all := game.ManaAbilitiesWithOrigins(GaLayered(t, g, id))
	for i, o := range all {
		if o.Borrowed {
			idx = append(idx, i)
			origins = append(origins, o)
		}
	}
	return idx, origins
}

// blToGraveyard puts a card into a seat's graveyard and drops the layer
// cache, as the graveyard bump (#1117) does for a real move.
func BlToGraveyard(g *game.Game, p *game.Player, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: p.ID, Controller: p.ID})
	g.BumpLayerVersionForTest()
	return id
}

func BlElves(g *game.Game, p *game.Player) uuid.UUID {
	return GaPut(g, p.ID, "Llanowar Elves", "Creature — Elf Druid", BlLlanowarOracle)
}

func BlTapLentMana(t *testing.T, g *game.Game, p *game.Player, card uuid.UUID) {
	t.Helper()
	idx, origins := BlBorrowedMana(t, g, card)
	if len(idx) == 0 {
		t.Fatalf("%s has no lent mana ability", card)
	}
	if err := g.ActivateManaAbility(p.ID, card, idx[0], game.ManaAbilityParams{Ref: origins[0].Ref}); err != nil {
		t.Fatalf("activate the lent mana ability: %v", err)
	}
}
