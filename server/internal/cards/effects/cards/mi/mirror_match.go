package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mirror Match — Instant {4}{U}{U} (Commander 2015, S82 OLD #5546):
//
//	"Cast this spell only during the declare blockers step.
//	 For each creature attacking you or a planeswalker you control, create a
//	 token that's a copy of that creature and that's blocking that creature.
//	 Exile those tokens at end of combat."
//
// Fog Patch's declare-blockers timing. The attackers are read as the spell
// resolves (an attacker aimed at you, or at a planeswalker you control: Soul
// Snare's reading of CR 506.3), and each gets one token copy of itself, made
// from its copiable values (CR 707.2) and put onto the battlefield blocking it
// (CR 506.3d: Brimaz's way, so nothing that watches blocks sees them). Every
// token, however many, is exiled by one delayed trigger at the beginning of the
// end of combat step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1f0883f8-89c9-4620-9837-72a2909e0599",
		Name:         "Mirror Match",
		Completeness: effects.CompletenessFull,
		CastCondition: func(g *game.Game, _ uuid.UUID, _ game.Card) bool {
			return g.Turn.Step == game.StepDeclareBlockers
		},
		CastConditionLabel: "Cast this spell only during the declare blockers step.",
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var groups []game.TokenGroup
			for _, id := range effects.AttackingCreatures(ctx.Game) {
				c, ok := ctx.Game.LookupCardForEffect(id)
				if !ok || !effects.AttacksYouOrYourPlaneswalker(ctx.Game, c.AttackingTarget, ctx.Controller()) {
					continue
				}
				template, ok := effects.TokenCopyTemplate(ctx.Game, id)
				if !ok {
					continue
				}
				groups = append(groups, game.TokenGroup{Template: effects.CopiedToken(template), Count: 1, Blocking: id})
			}
			if len(groups) == 0 {
				return nil
			}
			return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
				Controller: ctx.Controller(),
				Source:     ctx.Source(),
				Groups:     groups,
			}, ctx.ItemTail(mirrorMatchExileAtEndOfCombat))
		},
	})
}

// mirrorMatchExileAtEndOfCombat is "exile those tokens at end of combat": one
// delayed trigger naming exactly the tokens that landed. An on-disk key: never
// renamed, never reused.
var mirrorMatchExileAtEndOfCombat = effects.CardTail("then/mirror-match-exile-at-end-of-combat", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		At:    game.StepEndCombat,
		Label: "Mirror Match — exile the tokens",
		Cards: o.IDs,
		Body:  effects.ExileListedCardsBody,
	}.Apply(ctx)
})
