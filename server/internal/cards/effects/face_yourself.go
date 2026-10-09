package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Face Yourself — Sorcery {5}{R}{R}:
//
//	"For each creature target player controls, create a token that's a
//	 copy of that creature, except it has haste and 'At the beginning of
//	 the end step, if you don't control a planeswalker, sacrifice this
//	 creature.'"
//
// The creatures are read once as the spell resolves, then each is copied
// under the caster's control (the copies' "you" is the caster). Haste
// rides the token's printed keywords (TokenCopyGainsHaste). The
// sacrifice clause is a granted trigger (ADR 0093): a catalog bundle
// pinned to each new token with an indefinite duration, so a copy of a
// copy does not inherit it (a granted ability is not copiable, CR 707.2)
// and a token that leaves is gone with its grant. It fires at EVERY end
// step, not only the caster's, and checks "if you don't control a
// planeswalker" both when the step begins and again on resolution
// (CR 603.4), so a planeswalker arriving in response saves the token.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "3c0fe71c-bee1-41d3-aeb5-1723028eedde",
		Name:         "Face Yourself",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		Grants: []AbilityGrant{{
			Key: faceYourselfSacrifice,
			Triggered: []game.TriggeredAbility{
				On(game.EventStepBegan, AllOf(StepBegan(game.StepEnd, false), controllerHasNoPlaneswalker),
					"Face Yourself — sacrifice this creature unless you control a planeswalker",
					faceYourselfSacrificeUnlessWalker),
			},
			Text: "At the beginning of the end step, if you don't control a planeswalker, sacrifice this creature.",
		}},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := firstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetPlayer {
				return nil
			}
			var originals []uuid.UUID
			for _, c := range MatchingBattlefield(ctx, And(Creature(), ControlledBy(t.ID))) {
				originals = append(originals, c.InstanceID)
			}
			cursor := b25LastEventSeq(ctx.Game)
			for _, id := range originals {
				if err := (CreateTokenCopy{Controller: item.Controller, Copy: id, N: 1, Except: TokenCopyGainsHaste}).Apply(ctx); err != nil {
					return err
				}
			}
			for _, id := range b27TokensCreatedByAfter(ctx.Game, item.Controller, cursor) {
				if err := (GrantAbilitiesFor{
					Target:   id,
					Keys:     []string{faceYourselfSacrifice},
					Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), id),
					Label:    "Face Yourself — the copy is sacrificed at end step without a planeswalker",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

// faceYourselfSacrifice is the key of the granted sacrifice bundle.
const faceYourselfSacrifice = "face-yourself/sacrifice"

// controllerHasNoPlaneswalker is the intervening "if you don't control a
// planeswalker", "you" being the host creature's controller.
func controllerHasNoPlaneswalker(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return !youControlAPlaneswalker(g, source.Controller)
}

// faceYourselfSacrificeUnlessWalker re-checks the condition on resolution
// (CR 603.4) before sacrificing the host.
func faceYourselfSacrificeUnlessWalker(g *game.Game, item *game.StackItem) error {
	if youControlAPlaneswalker(g, item.Controller) {
		return nil
	}
	return SacrificePermanent{Target: item.SourceCardID}.Apply(NewContext(g, item))
}
