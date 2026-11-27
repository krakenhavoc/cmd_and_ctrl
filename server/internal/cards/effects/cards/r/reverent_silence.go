package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reverent Silence — Sorcery {3}{G} (Nemesis, S82 OLD batch 679, #5618):
//
//	"If you control a Forest, rather than pay this spell's mana cost, you may
//	 have each other player gain 6 life.
//	 Destroy all enchantments."
//
// The alternative cost is the #6214 seam: each player other than the caster
// gains 6 life as the spell is cast (CR 118.9, 601.2f), the caster none, and
// no mana is paid. It is available only while the caster controls a Forest (a
// land with the subtype, CR 205.3i) and some other player can gain life. The
// effect is Wrath of Gods' shape: every enchantment on the battlefield, anyone's,
// is destroyed at once (regeneration and indestructible apply).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e670d1c9-5e28-4df8-aad9-d08d19dc74e9",
		Name:         "Reverent Silence",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepEnchantments, How: game.SweepDestroy}},
		AlternativeCosts: []game.AlternativeCost{{
			Key:                      "reverent-silence",
			Label:                    "If you control a Forest, have each other player gain 6 life rather than pay this spell's mana cost",
			EachOtherPlayerGainsLife: 6,
			Condition:                effects.OldB679YouControlAForest,
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.Enchantment()}.Apply(ctx)
		},
	})
}
