package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ability_removal_view_test.go — the wire half of CR 613.1f. The
// client's right-click menu is built from CardView.activated_abilities
// and CardView.mana_abilities, so a permanent whose abilities were
// removed must ship both lists empty or the menu offers a control the
// server will refuse. No client change was needed for this: both
// fields are projected from game.ActivatedAbilitiesForCard and
// game.ManaAbilitiesForCard, which is where the removal is checked.

// silencedPermanent seeds a permanent carrying both an intrinsic
// activated ability and an intrinsic mana ability, plus a layer-6
// removal aimed at it, and returns its ID.
func silencedPermanent(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	owner := g.Seats[0].ID
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Clue",
		TypeLine:   "Artifact — Clue",
		// stampActivatedAbilities skips a card with no oracle ID, so
		// the fixture needs one even though the abilities below are
		// instance-carried rather than catalogued.
		OracleID:   "00000000-0000-0000-0000-0000000000cc",
		Owner:      owner,
		Controller: owner,
		ActivatedAbilities: []game.ActivatedAbilityShape{
			{Label: "{2}, Sacrifice this: Draw a card", Cost: game.AbilityCost{Mana: "{2}", SacrificeSelf: true}},
		},
		ManaAbilities: []game.ManaAbilityShape{
			{TapCost: true, Produced: "{C}", Label: "Add {C}"},
		},
	})
	return id
}

func cardViewByID(t *testing.T, g *game.Game, id uuid.UUID) CardView {
	t.Helper()
	for _, v := range ViewOfGame(g).Battlefield.Cards {
		if v.InstanceID == id.String() {
			return v
		}
	}
	t.Fatalf("card %s not in the battlefield view", id)
	return CardView{}
}

func TestCardViewHidesAbilitiesOfASilencedPermanent(t *testing.T) {
	g := buildActiveGame(t)
	id := silencedPermanent(t, g)

	before := cardViewByID(t, g, id)
	if len(before.ActivatedAbilities) != 1 || len(before.ManaAbilities) != 1 {
		t.Fatalf("fixture is wrong: the menu should have two entries before the removal, got %+v", before)
	}

	prev := game.CatalogStaticAbilities
	game.CatalogStaticAbilities = func(oracleID string) []game.StaticAbility {
		if oracleID != "view-silencer" {
			return nil
		}
		return []game.StaticAbility{{
			Layer:            game.Layer6Ability,
			RemovesAbilities: true,
			AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
				return target.InstanceID == id
			},
		}}
	}
	t.Cleanup(func() { game.CatalogStaticAbilities = prev })

	owner := g.Seats[0].ID
	g.Battlefield.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Silencer", TypeLine: "Enchantment — Aura",
		OracleID: "view-silencer", Owner: owner, Controller: owner,
	})
	g.BumpLayerVersionForTest()

	after := cardViewByID(t, g, id)
	if len(after.ActivatedAbilities) != 0 {
		t.Errorf("the context menu still offers %d activated abilities: %+v",
			len(after.ActivatedAbilities), after.ActivatedAbilities)
	}
	if len(after.ManaAbilities) != 0 {
		t.Errorf("the context menu still offers %d mana abilities: %+v",
			len(after.ManaAbilities), after.ManaAbilities)
	}
}
