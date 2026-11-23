package effectstest

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// talentWalker seats a planeswalker the catalog prints no abilities for,
// so the only loyalty rows it has are the ones a Talent grants it.
func TalentWalker(g *game.Game, owner uuid.UUID, loyalty int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Blank Walker", TypeLine: "Legendary Planeswalker — Blank",
		Owner: owner, Controller: owner,
		Counters: map[string]int{game.CounterLoyalty: loyalty},
	})
}

// talentOn attaches a Talent to `walker`, with the zone-move event that
// gives its static a timestamp.
func TalentOn(g *game.Game, owner uuid.UUID, name, oracle string, walker uuid.UUID) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Enchantment — Aura",
		OracleID: oracle, Owner: owner, Controller: owner,
		AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: walker},
	})
}

// grantedLoyaltyRow is the index and ref of the walker's granted row
// whose label starts with `prefix`.
func GrantedLoyaltyRow(t *testing.T, g *game.Game, walker uuid.UUID, prefix string) (int, string) {
	t.Helper()
	abs, origins := game.ActivatedAbilitiesWithOrigins(GaLayered(t, g, walker))
	for i, ab := range abs {
		if origins.At(i).Granted() && strings.HasPrefix(ab.Label, prefix) {
			if ab.Cost.Loyalty == nil {
				t.Fatalf("the granted row %q has no loyalty cost", ab.Label)
			}
			return i, origins.Ref(i)
		}
	}
	t.Fatalf("the walker has no granted row starting %q (rows: %d)", prefix, len(abs))
	return -1, ""
}

// talentStackItem is the activated item on the stack from `source`.
func TalentStackItem(g *game.Game, source uuid.UUID) *game.StackItem {
	var out *game.StackItem
	g.WithWriteLock(func() {
		for _, it := range g.StackMeta {
			if it != nil && it.Kind == game.StackItemActivated && it.SourceCardID == source && !it.IsCopy {
				out = it
			}
		}
	})
	return out
}
