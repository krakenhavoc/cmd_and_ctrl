package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// one_cast_per_card_type_test.go — #2167: a cast permission spent once
// per card type. Muldrotha, the Gravetide (a standing graveyard
// permission) and Aminatou's Augury (a stored one over the cards it
// exiled).

const (
	MuldrothaOracle       = "e4625704-1d52-44e4-804f-2f45644d76ac"
	AminatousAuguryOracle = "c9160997-0305-47f5-9a2f-77588d167da0"
)

// muldrothaTable is seat 0's precombat main with Muldrotha on the
// battlefield under seat 0.
func MuldrothaTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := NewCatalogGame(t)
	AdvanceToMain(t, g)
	me := g.Seats[0]
	m := PushCatalogPermanent(g, me.ID, "Muldrotha, the Gravetide", "Legendary Creature — Elemental Avatar", MuldrothaOracle, false)
	return g, me, m
}

// ptGraveCard puts a card with a type line and a {0} cost into its
// owner's graveyard, known to them.
func PtGraveCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	cost := "{0}"
	if typeLineIsLand(typeLine) {
		cost = ""
	}
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		Power: 1, Toughness: 1, Owner: p.ID, Controller: p.ID, KnownBy: map[uuid.UUID]bool{p.ID: true},
	})
	return id
}

func typeLineIsLand(typeLine string) bool {
	return (game.Card{TypeLine: typeLine}).IsLand()
}

func MuldrothaCast(g *game.Game, p *game.Player, id uuid.UUID, permType string) error {
	return g.CastSpell(p.ID, id, game.CastSpellParams{FromZone: "graveyard", PermissionType: permType})
}

// resolveAll empties the stack so the next sorcery-speed cast is open.
func ResolveAllPT(t *testing.T, g *game.Game) {
	t.Helper()
	PassPriorityAroundTable(t, g)
}
