package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bile Blight — Instant {B}{B} (Warhammer 40,000 Commander):
//
//	"Target creature and all other creatures with the same name as that
//	 creature get -3/-3 until end of turn."
//
// The target is a real target (CR 115.1); the others are not, so hexproof
// on one copy changes nothing. The set is the target plus every other
// creature sharing its post-layer name as the spell resolves (CR 611.2c);
// a creature that enters later is not shrunk. With the target illegal the
// spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0dd95516-893d-477e-a38c-3cc37c6e7ae0",
		Name:         "Bile Blight",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -3, -3)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			ctx = ctx.AsGroupMember()
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				for _, id := range effects.S82RprB17SameNameAsTarget(ctx.Game, t.ID) {
					c, ok := ctx.Game.LookupCardForEffect(id)
					if !ok || !c.IsCreature() {
						continue
					}
					if err := (effects.BoostUntilEOT{Target: id, Power: -3, Toughness: -3, Label: "Bile Blight — -3/-3 until end of turn"}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}
			return nil
		},
	})
}
