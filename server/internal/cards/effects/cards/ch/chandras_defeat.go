package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chandra's Defeat — Instant {R} (Hour of Devastation, #5534):
//
//	"Chandra's Defeat deals 5 damage to target red creature or red
//	 planeswalker. If that permanent is a Chandra planeswalker, you may
//	 discard a card. If you do, draw a card."
//
// Nissa's Defeat's shape with damage: the target is a red creature or a red
// planeswalker ("red" is a word of the card's text, CR 612.2). Whether it is
// a Chandra planeswalker is read as the spell resolves, after the damage:
// the damage may kill it, and its last existence on the battlefield is
// what counts (CR 608.2h). The loot is the shared "you may discard a card. If
// you do, draw a card" and only a card really discarded draws. A target that
// has gone makes the spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a57ad23b-cff3-45e5-98b3-1b20224aaa19",
		Name:         "Chandra's Defeat",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 5)),
		Targets: effects.TargetPermanent("target red creature or red planeswalker",
			effects.Or(
				effects.And(effects.Creature(), effects.OfColor("R")),
				effects.And(effects.Planeswalker(), effects.OfColor("R")),
			)),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := effects.FirstLegalTarget(ctx)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				if err := (effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 5}).Apply(ctx); err != nil {
					return err
				}
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok || !c.IsPlaneswalker() || !c.HasSubtype(ctx.TextSubtype("Chandra")) {
					return nil
				}
				return effects.TdmMayDiscardThenDrawOne("Chandra's Defeat — you may discard a card; if you do, draw a card")(ctx.Game, ctx.Item)
			}),
		),
	})
}
