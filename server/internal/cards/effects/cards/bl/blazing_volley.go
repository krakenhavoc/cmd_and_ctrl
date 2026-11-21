package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blazing Volley — Sorcery {R} (EDHREC rank 4329):
//
//	"Blazing Volley deals 1 damage to each creature your opponents
//	 control."
//
// One mana that answers three players' token boards at once, which is
// a Commander card in a way it never was in a two-player game. The
// asymmetry — "your opponents control" — is the whole reason it beats
// Pyroclasm in a deck that is itself going wide.
//
// Mass damage is deliberately not one of the four mass ZONE-CHANGE
// primitives: nothing leaves the battlefield here. Each creature takes
// its 1, and the ones that took lethal die later, when the
// state-based-action sweep runs at the next boundary. That is why
// indestructible does not save a 1/1 from this and a damage-prevention
// shield does, and why a creature with a +1/+1 counter shrugs it off.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3656338d-ca08-465b-b09a-d8d8d1196eec",
		Name:         "Blazing Volley",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, OpponentsOnly: true}},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.OpponentControls()), 1)
		},
	})
}
