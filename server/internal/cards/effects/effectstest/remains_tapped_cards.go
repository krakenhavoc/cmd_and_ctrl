package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// remains_tapped_cards_test.go is ADR 0109 Delivery PR 2 (#1894): "for
// as long as you control this and it remains tapped" (a duration with
// two conditions), "for as long as that creature remains tapped", and
// the single "for as long as this remains tapped" cards of owner
// decision 4.

const (
	RtSeasingerOracle     = "13745b92-1e30-4075-9733-5d00da11408c"
	RtShacklesOracle      = "fddac567-c42c-453b-9c82-c3140b374a06"
	RtEvergloveOracle     = "31da3df3-5b79-42d0-9018-76bc73411954"
	RtManaLeechOracle     = "cad05528-917b-4296-9d23-ba9433612709"
	RtKillSwitchOracle    = "0cd5eef6-6830-4c6d-8578-2726900c39a5"
	RtZygonOracle         = "693c071a-d6b5-4a96-8fe3-8c61fa9406bc"
	RtHisokasGuardOracle  = "6136f019-da8f-471e-9ac5-41e535aa8a21"
	RtBlackstaffOracle    = "a4faed58-4b1b-4cab-8a12-4fcd20b320ab"
	RtHelmOracle          = "78f22413-4e61-4191-9d30-7b5d6538446e"
	RtEndoskeletonOracle  = "e9ef7f8c-21ef-45f0-af64-8ffa9f0089d1"
	RtIceFloeOracle       = "cfaaead2-09e8-47cb-9e39-8570b8d8de86"
	RtAquitectsWillOracle = "d02befcb-3e2c-4cfe-a913-9bb648e5bb2c"
	RtMinasMorgulOracle   = "867dbd5a-c3cf-41ce-980b-c9babc6f30f2"
	RtMannequinOracle     = "d9dfdbbb-75b3-43d3-8ed6-371cb98a0124"
	RtMathasOracle        = "1e8d9d0c-185b-4149-9ec9-74a2f2fe8764"
	RtFireheartOracle     = "c3b3e070-28f6-4852-816e-caffe358b6fe"
	RtXolatoyacOracle     = "2c0fc559-857f-4239-97fa-379ced3dfc1f"
)

// rtActivate pays `mana` into the controller's pool and activates the
// source's ability `index`, then resolves it.
func RtActivate(t *testing.T, g *game.Game, controller, source uuid.UUID, index int, targets []game.TargetRef, mana ...string) {
	t.Helper()
	B06AddMana(g.PlayerByIDForEffect(controller), mana...)
	if err := g.ActivateCatalogAbility(controller, source, index, game.ActivateAbilityParams{Targets: targets}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	PassPriorityAroundTable(t, g)
}

func RtUntap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(id) })
}

func RtTapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	return LayeredCard(t, g, id).Tapped
}
