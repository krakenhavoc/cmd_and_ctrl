package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glóin the Mighty // Easy Pickings — an Adventure card (CR 715, The
// Hobbit, #4454), oracle 6505a53b…:
//
//	Glóin the Mighty — Legendary Creature — Dwarf Warrior {3}{R}, 4/3
//	  "At the beginning of your first main phase, add {R}{R}."
//	Easy Pickings — Sorcery — Adventure {2}{R}
//	  "Easy Pickings deals 1 damage to each creature your opponents
//	   control. (Then exile this card. You may cast the creature later
//	   from exile.)"
//
// Glóin's trigger is Hulking Raptor's: EventBeginPrecombatMain is the
// first main phase of an ordinary turn, and the mana empties at the end
// of the phase like any other (CR 106.4). Easy Pickings is Boiling
// Earth's first half, the same sweep over the opponents' creatures.
//
// No simplification.

// gloinTheMightyOracleID is shared by both faces of the card.
const GloinTheMightyOracleID = "6505a53b-00db-44c4-8eaa-09a1a01451a4"

func init() {
	effects.Register(effects.Spec{
		OracleID:     GloinTheMightyOracleID,
		Name:         "Glóin the Mighty",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourPrecombatMain("Glóin the Mighty — add {R}{R}", func(g *game.Game, item *game.StackItem) error {
				return g.AddManaForEffect(item.Controller, item.SourceCardID, "{R}{R}")
			}),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     GloinTheMightyOracleID + "#1",
		Name:         "Easy Pickings",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, OpponentsOnly: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.OpponentControls()), 1)
		},
	})
}
