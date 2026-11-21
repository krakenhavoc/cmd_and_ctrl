package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Noise Marine — Creature — Astartes Warrior {4}{R}, 3/2 (Warhammer 40,000
// Commander, S82 GAP batch 194, #5948):
//
//	"Cascade (When you cast this spell, exile cards from the top of your library
//	 until you exile a nonland card that costs less. You may cast it without
//	 paying its mana cost. Put the exiled cards on the bottom in a random order.)
//	 Sonic Blaster — When this creature enters, it deals damage equal to the
//	 number of spells you've cast this turn to any target."
//
// Cascade is the keyword's own cast trigger. The enters trigger targets any
// target and counts the spells its controller has cast this turn as it
// resolves, the Marine's own cast and a cascaded spell included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1f7e85cf-afee-4041-9436-1fb69b4ec292",
		Name:            "Noise Marine",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"cascade"},
		Triggered: []game.TriggeredAbility{
			effects.Cascade(),
			effects.Targeting(effects.WhenThisEnters("Noise Marine — damage equal to the number of spells you've cast this turn to any target",
				effects.GapB194SpellsCastThisTurnDamage), effects.TargetAny()),
		},
	})
}
