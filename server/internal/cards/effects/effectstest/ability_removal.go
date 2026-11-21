package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ability_removal_test.go is the catalog half of CR 613.1f: the three
// Auras that change what a permanent IS. The engine half — the
// CatalogAbilityKey seam, the layer pass's CR 613.6 rule and the
// layer-6 timestamp rule — is pinned in
// game/layer6_ability_removal_test.go and game/layer_dependency_test.go.
//
// The cards are here together because each one proves a different
// consequence of the same machinery, and none of them proves it
// alone:
//
//	Darksteel Mutation       removal that also GRANTS, in one effect
//	Kenrith's Transformation removal plus the first layer-5 colour set
//	Song of the Dryads       removal that makes OTHER attachments
//	                         illegal, which is CR 704.5m/n firing in
//	                         real play rather than only in a fixture

const (
	DarksteelMutationOracle = "05a4f8ff-49da-42af-add5-6248c4b0644b"
	KenrithTransformOracle  = "a492a323-df8a-40fd-bff0-50091baa7700"
	SongOfTheDryadsOracle   = "7c3944fa-7c86-4979-85a9-86196aa94594"
	SolRingOracle           = "6ad8011d-3471-4369-9d68-b264cc027487"
)

// enchant casts the named Aura at `host` and lets it resolve.
func Enchant(t *testing.T, g *game.Game, name, oracle string, host uuid.UUID) uuid.UUID {
	t.Helper()
	id := CastCatalogSpell(t, g, name, AuraTypeLine, oracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: host}})
	PassPriorityAroundTable(t, g)
	return id
}

// settle advances one step so the state-based actions run.
func Settle(t *testing.T, g *game.Game) {
	t.Helper()
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
}

func LayeredCard(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var out game.Card
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				out, found = c, true
				return
			}
		}
	})
	if !found {
		t.Fatalf("card %s not on battlefield", id)
	}
	return out
}

func EffectiveColors(t *testing.T, g *game.Game, id uuid.UUID) []string {
	t.Helper()
	return LayeredCard(t, g, id).Effective().Colors
}
