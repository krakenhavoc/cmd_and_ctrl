package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// aether_free_casts_test.go — Aetherflux Conduit and Aetherworks Marvel
// (#2110): the last two cards of the "cast a spell for free as an ability
// resolves" row, on the #3270 pick and grant.

const (
	AetherfluxConduitOracle = "775617cd-7264-4365-94c4-ffae4eb9ec8b"
	AetherworksMarvelOracle = "470ec98e-120c-446d-ade6-7d80aa836b55"
)

// --- Aetherworks Marvel --------------------------------------------------

func AetherworksTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	AdvanceToMain(t, g)
	marvel := PushCatalogPermanent(g, me.ID, "Aetherworks Marvel", "Legendary Artifact", AetherworksMarvelOracle, false)
	return g, me, marvel
}

// aetherworksLibrary stacks six cards on top, top first: two lands, three
// spells and one more spell; and a distinct deep card below them.
func AetherworksLibrary(me *game.Player) (spells []uuid.UUID, lands []uuid.UUID, deep uuid.UUID) {
	deep = SunbirdLibrary(me, game.Card{Name: "Deep Card", TypeLine: "Basic Land — Plains"})[0]
	ids := SunbirdLibrary(me,
		game.Card{Name: "Top Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Cheap Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Huge Wurm", TypeLine: "Creature — Wurm", ManaCost: "{7}{G}", Power: 8, Toughness: 8},
		game.Card{Name: "Top Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Big Sorcery", TypeLine: "Sorcery", ManaCost: "{6}{R}"},
		game.Card{Name: "Cheap Elk", TypeLine: "Creature — Elk", ManaCost: "{G}"},
	)
	return []uuid.UUID{ids[1], ids[2], ids[4], ids[5]}, []uuid.UUID{ids[0], ids[3]}, deep
}

func AetherworksLook(t *testing.T, g *game.Game, me *game.Player, marvel uuid.UUID) *game.PendingChoice {
	t.Helper()
	SetEnergy(t, g, me, 7)
	if err := g.ActivateCatalogAbility(me.ID, marvel, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	PassPriorityUntilChoice(t, g)
	// #4250: the cast is offered inside the ability's resolution.
	return MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast)
}
