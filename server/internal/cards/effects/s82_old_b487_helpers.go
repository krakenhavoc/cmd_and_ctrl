package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b487_helpers.go — shared shapes of S82 OLD batch 487 (#5582):
// Worldwake's Allies and quest enchantments.

// S82OldB487AlliesYouControl counts the Allies (any permanent with the
// subtype, as the layers leave it) `controller` controls, "Ally" read through
// `source`'s text so a text-changing effect reaches it (CR 612.2). It is
// ZenB2AlliesYouControl for a caller with no Context, a mana ability's amount.
// Caller holds g.mu.
func S82OldB487AlliesYouControl(g *game.Game, controller uuid.UUID, source *game.Card) int {
	ally := "Ally"
	if source != nil {
		ally = source.TextSubtype("Ally")
	}
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.HasSubtype(ally) {
			n++
		}
	}
	return n
}

// S82OldB487QuestCounters is how many quest counters the permanent `id` has:
// the live count while it is on the battlefield, else the count it had as it
// left (last-known information, CR 603.10a / CR 608.2h), which is what an
// intervening "if this enchantment has N or more quest counters" reads.
func S82OldB487QuestCounters(g *game.Game, id uuid.UUID) int {
	if c, ok := g.LookupCardForEffect(id); ok && g.Battlefield.Contains(id) {
		return c.Counters["quest"]
	}
	return g.LastKnownCountersForEffect(id)["quest"]
}

// S82OldB487QuestAnthem is a quest enchantment's "As long as this enchantment
// has N or more quest counters on it, creatures you control get +P/+T": a
// layer 7c static whose AppliesTo reads the enchantment's own counters, so the
// last counter switches it on as it is put there (Beastmaster Ascension's shape).
func S82OldB487QuestAnthem(label string, n, power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Layer:    game.Layer7PT,
		Label:    label,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsCreature() && target.Controller == source.Controller && source.Counters["quest"] >= n
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power += power
			c.Toughness += toughness
		},
	}
}
