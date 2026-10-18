package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Moonlight Hunt — Instant {1}{G}:
//
//	"Choose target creature you don't control. Each creature you control
//	 that's a Wolf or a Werewolf deals damage equal to its power to that
//	 creature."
//
// The many-sourced Bite Down. The Wolves and Werewolves you control are
// fixed as the spell resolves, each deals damage equal to its own power
// read then, and the damage is one simultaneous event dealt BY those
// creatures (deathtouch and lifelink carry, a prevention shield on one
// applies to it alone). With the target gone nothing is dealt (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f38a6cb4-79dd-4c12-9896-8c6655e275f8",
		Name:         "Moonlight Hunt",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you don't control", OpponentControls()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			legal := ctx.LegalTargets()
			if len(legal) == 0 || legal[0].Kind != game.TargetCard {
				return nil
			}
			victim := legal[0].ID
			type bite struct {
				source uuid.UUID
				power  int
			}
			var bites []bite
			for _, c := range wolvesAndWerewolvesOf(ctx, item.Controller) {
				if p := c.CurrentPower(); p > 0 {
					bites = append(bites, bite{c.InstanceID, p})
				}
			}
			if len(bites) == 0 {
				return nil
			}
			return ctx.Game.DamageInstanceForEffect(func() error {
				for _, b := range bites {
					if err := (DealDamage{Source: b.source, Target: victim, Amount: b.power}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			})
		},
	})
}

// wolvesAndWerewolvesOf lists the creatures `player` controls that are a
// Wolf or a Werewolf, as they stand now ("each creature you control
// that's a Wolf or a Werewolf": Moonlight Hunt, Hollowhenge Overlord).
// The types are words of the resolving object's text (CR 612.2).
func wolvesAndWerewolvesOf(ctx *Context, player uuid.UUID) []game.Card {
	wolf, werewolf := ctx.TextSubtype("Wolf"), ctx.TextSubtype("Werewolf")
	var out []game.Card
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() && (c.HasSubtype(wolf) || c.HasSubtype(werewolf)) {
			out = append(out, c)
		}
	}
	return out
}
