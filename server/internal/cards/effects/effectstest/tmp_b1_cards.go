package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmpB1Push seeds a permanent on the battlefield, catalog row or not,
// untapped and not summoning sick.
func TmpB1Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int, colors ...string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Colors: colors, Owner: owner, Controller: owner,
	})
}

// tmpB1Shadow seeds a vanilla creature with shadow.
func TmpB1Shadow(g *game.Game, owner uuid.UUID, name string, colors ...string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Spirit",
		Power: 2, Toughness: 2, Keywords: []string{"shadow"}, Colors: colors, Owner: owner, Controller: owner,
	})
}

// --- buyback ------------------------------------------------------------------------------

// tmpB1BuybackCases is each buyback spell cast twice: without buyback it
// ends in the graveyard, with it back in its owner's hand.
var TmpB1BuybackCases = []struct {
	Name     string
	Optional []int
	WantHand bool
}{
	{"without buyback", nil, false},
	{"with buyback", []int{0}, true},
}

// tmpB1AssertBuyback checks where the resolved spell went.
func TmpB1AssertBuyback(t *testing.T, caster *game.Player, id uuid.UUID, wantHand bool) {
	t.Helper()
	inHand, inGrave := caster.Hand.Contains(id), caster.Graveyard.Contains(id)
	if inHand != wantHand || inGrave == wantHand {
		t.Errorf("after resolution: hand=%v graveyard=%v, want hand=%v", inHand, inGrave, wantHand)
	}
}
