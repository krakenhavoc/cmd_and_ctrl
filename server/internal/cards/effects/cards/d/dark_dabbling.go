package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dark Dabbling — Instant {2}{B} (Magic Origins, #5549):
//
//	"Regenerate target creature. Draw a card.
//	 Spell mastery — If there are two or more instant and/or sorcery cards in
//	 your graveyard, also regenerate each other creature you control."
//
// Refresh's regeneration shield and draw, then, with spell mastery read as
// the spell resolves, one shield on each creature the caster controls other
// than the target (CR 701.19a): the target gets exactly one shield, not two.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "342daae1-f7f0-4bce-94bc-e9bdce53e48e",
		Name:         "Dark Dabbling",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerSurvive, Draws: 1},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.Regenerate{Target: target}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.DrawCards{N: 1}).Apply(ctx); err != nil {
				return err
			}
			if !effects.SpellMastery(ctx.Game, ctx.Controller()) {
				return nil
			}
			var others []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == ctx.Controller() && c.IsCreature() && c.InstanceID != target {
					others = append(others, c.InstanceID)
				}
			}
			for _, id := range others {
				if err := (effects.Regenerate{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
