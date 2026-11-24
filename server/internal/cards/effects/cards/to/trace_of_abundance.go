package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Trace of Abundance — Enchantment — Aura {R/W}{G}:
//
//	"Enchant land
//	 Enchanted land has shroud. (It can't be the target of spells or
//	 abilities.)
//	 Whenever enchanted land is tapped for mana, its controller adds an
//	 additional one mana of any color."
//
// Fertile Ground's triggered mana ability (WheneverAttachedTapsForMana: the
// colour is picked inside the trigger, a by-hand tap queues the same five-colour
// pick, and the auto-tapper books the extra mana as one pick) plus ADR 0072's
// layer-6 shroud on the enchanted land (GrantToAttached). Shroud stops the
// Aura's own controller from targeting the land too, so it can't be moved by
// another Aura or bounced by its owner; it does not stop a spell that doesn't
// target. The additional mana goes to the land's controller, whoever owns the
// Aura. Both end when the Aura leaves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "530140f1-e992-4b5e-882c-c15bacc328fa",
		Name:         "Trace of Abundance",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Static:       []game.StaticAbility{effects.GrantToAttached("shroud")},
		ManaTriggers: []game.ManaTrigger{
			effects.WheneverAttachedTapsForMana(
				"Trace of Abundance — add an additional one mana of any color",
				"{W|U|B|R|G}"),
		},
	})
}
