package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cerulean Wisps — Instant {U} (EDHREC rank 4519):
//
//	"Target creature becomes blue until end of turn. Untap that
//	 creature.
//	 Draw a card."
//
// A free untapper. One blue mana replaces itself and untaps a
// creature, which is the whole card in a deck built on a tap
// ability: a second activation of a Kiora's Follower, a Seedborn
// Muse impression for one turn, an extra block after an attack. The
// colour change is the rider — it turns on an Oona's Prowler's
// pitch, it dodges a colour-based protection, and in a devotion or
// "blue creatures you control" deck it is not nothing.
//
// "Becomes blue" is a layer 5 colour change that REPLACES the
// creature's colours rather than adding to them (CR 105.3): a green
// creature that becomes blue is blue and no longer green. That is
// the same layer SetAttachedColors uses for Song of the Dryads, with
// a turn-scoped duration instead of an attachment.
//
// The affected creature is SNAPSHOTTED as the spell resolves, with
// its battlefield-entry stamp, so a creature flickered in response is
// a new object and is not recoloured (CR 400.7, CR 611.2c) — the
// same guard BoostUntilEOT applies.
//
// Order matters and is printed: the colour change happens, then the
// untap, then the draw. A target that left in response fizzles the
// whole spell including the cantrip (CR 608.2b) — this is a
// single-target spell with no other legal target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a6605c50-558e-417c-8c75-6c45b06d6e13",
		Name:         "Cerulean Wisps",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if set := eotSnapshot(ctx, target, nil); set != nil {
				if err := (StaticUntilEOT{
					Label: "Cerulean Wisps — that creature becomes blue",
					Ability: game.StaticAbility{
						Layer:     game.Layer5Color,
						AppliesTo: set.appliesTo(),
						Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
							c.Colors = []string{"U"}
						},
					},
				}).Apply(ctx); err != nil {
					return err
				}
			}
			if err := (UntapTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
