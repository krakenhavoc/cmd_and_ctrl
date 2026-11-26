package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Virtus's Maneuver — Sorcery {2}{B} (Battlebond):
//
//	"For each player, choose friend or foe. Each friend returns a creature
//	 card from their graveyard to their hand. Each foe sacrifices a creature
//	 of their choice."
//
// Pir's Whim's friend-or-foe question; each friend picks a creature card in
// their own graveyard, each foe a creature they control to sacrifice. A player
// with nothing to pick does nothing.
//
// No simplification.
func init() {
	effects.RegisterOldB206FriendFoe(effects.OldB206FriendFoe{
		Name: "Virtus's Maneuver",
		Friend: func(ctx *effects.Context, who uuid.UUID) error {
			effects.OldB206PickGraveyardToHand(ctx, who, "Virtus's Maneuver — return a creature card to your hand",
				func(c game.Card) bool { return c.IsCreature() })
			return nil
		},
		Foe: func(ctx *effects.Context, who uuid.UUID) error {
			effects.OldB206PickToSacrifice(ctx, who, "Virtus's Maneuver — sacrifice a creature", effects.Creature())
			return nil
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "1c2d6b55-653a-41fd-8a5e-6463a992640b",
		Name:         "Virtus's Maneuver",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.OldB206ChooseFriendOrFoe(ctx, "Virtus's Maneuver")
		},
	})
}
