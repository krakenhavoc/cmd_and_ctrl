package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Retaliation — Enchantment {2}{G} (Urza's Saga, #5624):
//
//	"Creatures you control have "Whenever this creature becomes blocked by a
//	 creature, this creature gets +1/+1 until end of turn.""
//
// The trigger is a granted bundle (ADR 0093): each creature you control has
// its own, once per blocker, like Pygmy Troll's.
//
// No simplification.
func init() {
	const label = "Retaliation — this creature gets +1/+1 until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "51571fa8-fd52-4439-8331-4eb4f651e764",
		Name:         "Retaliation",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: retaliationGrant,
			Triggered: []game.TriggeredAbility{
				effects.On(game.EventBlock, effects.ViciousBattleragerBlocked, label, effects.ThisGetsUntilEndOfTurn(1, 1, label)),
			},
			Text: "Whenever this creature becomes blocked by a creature, this creature gets +1/+1 until end of turn.",
		}},
		Static: []game.StaticAbility{effects.Labelled(effects.GrantAbilities(effects.B16CreaturesYouControl, retaliationGrant),
			"Creatures you control have \"Whenever this creature becomes blocked by a creature, this creature gets +1/+1 until end of turn.\"")},
	})
}

const retaliationGrant = "retaliation/becomes-blocked-pump"
