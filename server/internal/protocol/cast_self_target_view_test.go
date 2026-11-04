package protocol

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cast_self_target_view_test.go — #4869, CR 601.2a. The legal-target
// list the client's picker reads for a flashback card in the graveyard
// leaves the card itself out: by the time CR 601.2c announces targets
// the card is on the stack, not in the graveyard its clause names.
func TestGraveyardCastLegalTargetsOmitTheCardItself(t *testing.T) {
	g := buildActiveGame(t)
	castWindowOpen(t, g)
	me := g.Seats[0]
	const oracle = "test-view-4869"

	prevZones := game.CatalogCastableZones
	game.CatalogCastableZones = func(id string) []game.ZoneKind {
		if id == oracle {
			return []game.ZoneKind{game.ZoneGraveyard}
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogCastableZones = prevZones })
	prevAlts := game.CatalogAlternativeCosts
	game.CatalogAlternativeCosts = func(id string) []game.AlternativeCost {
		if id != oracle {
			return nil
		}
		return []game.AlternativeCost{{Key: "flashback", Label: "Flashback {2}{R}", ManaCost: "{2}{R}", FromZone: game.ZoneGraveyard}}
	}
	t.Cleanup(func() { game.CatalogAlternativeCosts = prevAlts })
	prevSpec := game.CatalogTargetSpec
	game.CatalogTargetSpec = func(id string) *game.TargetSpec {
		if id != oracle {
			return nil
		}
		return &game.TargetSpec{
			Mode: "card_in_graveyard", Label: "target card in a graveyard",
			Zones: []game.ZoneKind{game.ZoneGraveyard}, Min: 1, Max: 1,
		}
	}
	t.Cleanup(func() { game.CatalogTargetSpec = prevSpec })

	seen := map[uuid.UUID]bool{me.ID: true, g.Seats[1].ID: true}
	spell := game.NewCard("Test Journey", me.ID)
	spell.TypeLine = "Instant"
	spell.OracleID = oracle
	spell.KnownBy = seen
	me.Graveyard.PushTop(spell)
	other := game.NewCard("Other Card", me.ID)
	other.TypeLine = "Sorcery"
	other.KnownBy = seen
	me.Graveyard.PushTop(other)

	v := ViewOfGameFor(g, me.ID.String())
	var cv *CardView
	for i := range v.Seats[0].Graveyard.Cards {
		if v.Seats[0].Graveyard.Cards[i].InstanceID == spell.InstanceID.String() {
			cv = &v.Seats[0].Graveyard.Cards[i]
		}
	}
	if cv == nil || cv.LegalTargets == nil {
		t.Fatalf("the graveyard flashback card carries no legal targets: %+v", cv)
	}
	if slices.Contains(cv.LegalTargets.Cards, spell.InstanceID.String()) {
		t.Errorf("the picker offers the card being cast as its own target: %v", cv.LegalTargets.Cards)
	}
	if !slices.Contains(cv.LegalTargets.Cards, other.InstanceID.String()) {
		t.Errorf("the picker lost the other graveyard card: %v", cv.LegalTargets.Cards)
	}
}
