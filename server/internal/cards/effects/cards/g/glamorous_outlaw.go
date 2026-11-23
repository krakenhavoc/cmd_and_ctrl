package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glamorous Outlaw — Creature — Vampire Rogue {3}{U}{B}{R}, 4/5 (Streets
// of New Capenna):
//
//	"When this creature enters, it deals 2 damage to each opponent and you
//	 scry 2.
//	 {2}, Exile this card from your hand: Target land gains "{T}: Add {U},
//	 {B}, or {R}" until this card is cast from exile. You may cast this
//	 card for as long as it remains exiled."
//
// The damage is dealt by the creature itself, then the scry is the
// controller's. The hand ability is the cycle's shared one
// (sncB19ExileForLand).
//
// No simplification.
func init() {
	grant, ability := effects.SncB19ExileForLand("Glamorous Outlaw", "glamorous-outlaw/tap-ubr", "{U|B|R}", "{U}, {B}, or {R}")
	effects.Register(effects.Spec{
		OracleID:     "cd4f7912-81c0-44dd-87c2-a98e955fe5a8",
		Name:         "Glamorous Outlaw",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 2, DamageEachOpponent: 2},
		Grants:       []effects.AbilityGrant{grant},
		Activated:    []effects.ActivatedAbility{ability},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Glamorous Outlaw — 2 damage to each opponent and scry 2",
				effects.Do(effects.DamageEachOpponent{N: 2}, effects.Scry{N: 2})),
		},
	})
}
