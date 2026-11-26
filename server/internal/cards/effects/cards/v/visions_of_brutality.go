package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Visions of Brutality — {1}{B} Enchantment — Aura (Oath of the Gatewatch,
// S82 OLD b295):
//
//	"Devoid (This card has no color.)
//	 Enchant creature
//	 Enchanted creature can't block.
//	 Whenever enchanted creature deals damage, its controller loses that much
//	 life."
//
// A restriction on the host (Arco-Flagellant's bit) and Vampiric Link's damage
// trigger turned on the host's own controller, who is read as the creature last
// had them if it has left the battlefield (CR 608.2h). Combat or not, every
// damage event counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "66553485-4e9d-4049-9e80-3ac903f8127a",
		Name:            "Visions of Brutality",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantBlock),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Amount > 0 && source.IsAttachedTo(ev.Source)
			},
			Key:    "Visions of Brutality — its controller loses that much life",
			Effect: effects.OldB295DamagedCreatureControllerLosesLife,
		}},
	})
}
