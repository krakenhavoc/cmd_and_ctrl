package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Longshot, Rebel Bowman — Legendary Creature — Human Rebel Ally {3}{R},
// 3/3 (Avatar: The Last Airbender Eternal, #5075):
//
//	"Reach
//	 Noncreature spells you cast cost {1} less to cast.
//	 Whenever you cast a noncreature spell, Longshot deals 2 damage to each
//	 opponent."
//
// Reach is a printed keyword. The tax cut is Lyse Hext's generic {1} off
// noncreature spells you cast, applied after any increase (CR 601.2f, the
// ruling); the mana value is unchanged. The trigger is Black Waltz No. 3's:
// the spell is read off the stack, and the damage's source is Longshot.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f93c3468-5b1d-4fa7-a281-1009ea0f2062",
		Name:            "Longshot, Rebel Bowman",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Noncreature spells you cast cost {1} less to cast.",
				effects.YourSpell(), effects.NoncreatureSpell()),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Noncreature(), "Longshot, Rebel Bowman — 2 damage to each opponent", effects.Do(effects.DamageEachOpponent{N: 2})),
		},
	})
}
