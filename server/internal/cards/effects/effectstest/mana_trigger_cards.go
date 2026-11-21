package effectstest

import (
	"sort"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mana_trigger_cards_test.go — the catalog half of #763: the cards
// CR 605.1b's triggered mana abilities unblocked. The engine rules
// themselves are pinned in game/mana_trigger_test.go; these are the
// printed cards, end to end, through the real registry.

const (
	WildGrowthOracle    = "706ae742-1807-44b7-a4fa-f2e26f61519a"
	OvergrowthOracle    = "e6ccebf4-f4f0-404f-a634-4d751ef9c8aa"
	UtopiaSprawlOracle  = "00d8efa6-a2d9-4249-8da7-b45173675329"
	FertileGroundOracle = "cf14d4e5-5965-45ad-97f7-26facf2884b5"
	ManaFlareOracle     = "97159138-c34b-416e-b079-5c952383a243"
)

// pushAuraOnLand seeds an Aura already attached to `host`, the shape
// the cast path leaves behind once the Aura resolves.
func PushAuraOnLand(t *testing.T, g *game.Game, owner uuid.UUID, name, oracle string, host uuid.UUID) uuid.UUID {
	t.Helper()
	id := PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Enchantment — Aura", OracleID: oracle,
		Owner: owner, Controller: owner,
	})
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(id, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("AttachForEffect %s: %v", name, err)
		}
	})
	return id
}

// tapForMana activates a permanent's first mana ability.
func TapForMana(t *testing.T, g *game.Game, player, card uuid.UUID) {
	t.Helper()
	if err := g.ActivateManaAbility(player, card, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
}

// sortedPool is poolColors sorted, for an assertion that does not care
// which order the tokens arrived in.
func SortedPool(p *game.Player) []string {
	out := PoolColors(p)
	sort.Strings(out)
	return out
}
