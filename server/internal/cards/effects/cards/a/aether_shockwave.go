package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aether Shockwave — Instant {3}{W} (Saviors of Kamigawa, #5601):
//
//	"Choose one —
//	 • Tap all Spirits.
//	 • Tap all non-Spirit creatures."
//
// Riptide's group tap: the set is read as the spell resolves, whoever
// controls it, through the permanents' current types (a Spirit is a
// permanent with the Spirit creature type, CR 205.3m). Each is tapped as a
// member of the group. Nothing is targeted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e9db79fd-a5b6-4d59-9176-e713e2e9c708",
		Name:         "Aether Shockwave",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Tap all Spirits.", nil, aetherShockwaveTap(effects.HasSubtype("Spirit"))),
			effects.ModeDoing("Tap all non-Spirit creatures.", nil, aetherShockwaveTap(effects.And(effects.Creature(), effects.Not(effects.HasSubtype("Spirit"))))),
		),
	})
}

// aetherShockwaveTap is one bullet: tap every untapped permanent the filter
// accepts, read once as the mode resolves.
func aetherShockwaveTap(match effects.CardPredicate) func(*game.StackItem, *effects.Context, int) error {
	return func(_ *game.StackItem, ctx *effects.Context, _ int) error {
		var ids []uuid.UUID
		for _, c := range effects.MatchingBattlefield(ctx, match) {
			if !c.Tapped {
				ids = append(ids, c.InstanceID)
			}
		}
		for _, id := range ids {
			if err := (effects.TapTarget{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
				return err
			}
		}
		return nil
	}
}
