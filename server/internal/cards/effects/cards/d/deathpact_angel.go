package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deathpact Angel — Creature — Angel {3}{W}{B}{B}, 5/5 (Gatecrash, #5565):
//
//	"Flying
//	 When this creature dies, create a 1/1 white and black Cleric creature
//	 token. It has "{3}{W}{B}{B}, {T}, Sacrifice this token: Return a card
//	 named Deathpact Angel from your graveyard to the battlefield.""
//
// One death trigger that makes the Cleric; the ability is the token's own
// printed row (PrintedOldB416Cleric), not a grant, and its cost taps and
// sacrifices the token, so the fresh Cleric waits out summoning sickness
// (CR 302.6). The ability returns a card with the name from its controller's
// graveyard: the Angel itself when it is still there, a second copy if
// there are several (the controller picks), nothing when none is left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ad0f50d5-7b40-4fd8-8d13-4a0493398f53",
		Name:            "Deathpact Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Deathpact Angel — create a 1/1 white and black Cleric creature token with the return ability",
				effects.Do(effects.CreateToken{Template: effects.OldB416ClericToken(), N: 1})),
		},
	})
}
