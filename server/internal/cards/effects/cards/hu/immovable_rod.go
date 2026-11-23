package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Immovable Rod — Artifact {W}:
//
//	"You may choose not to untap this artifact during your untap step.
//	 Whenever this artifact becomes untapped, venture into the dungeon.
//	 {3}{W}, {T}: For as long as this artifact remains tapped, another
//	 target permanent loses all abilities and can't attack or block."
//
// The lock is one ScopedEffect over the target (layer 6 ability loss and
// the attack/block restriction) that lasts while the Rod stays tapped,
// Amber Prison's duration; a Rod untapped in response holds nothing.
// Untapping it later both ends the lock and ventures.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ec664e13-b7c4-4ca3-80d6-55a564d97937",
		Name:         "Immovable Rod",
		Completeness: effects.CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{
			effects.MayChooseNotToUntapSelf("Immovable Rod — you may choose not to untap this artifact"),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventUntapCard, effects.Self, effects.VentureLabel("Immovable Rod"), effects.Venture),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{W}, {T}: For as long as this artifact remains tapped, another target permanent loses all abilities and can't attack or block.",
			Cost:    effects.Plus(effects.ManaCost("{3}{W}"), effects.TapCost()),
			Targets: effects.Another(effects.TargetPermanent("another target permanent")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				d, ok := effects.DurationWhileThisRemainsTapped(ctx)
				if !ok {
					return nil
				}
				return effects.ScopedEffectFor{
					Target:   id,
					Mods:     []game.Mod{game.LoseAllAbilitiesMod(), game.AddRestrictionsMod(game.CantAttackOrBlock)},
					Duration: d,
					Label:    "Immovable Rod — loses all abilities and can't attack or block",
				}.Apply(ctx)
			},
		}},
	})
}
