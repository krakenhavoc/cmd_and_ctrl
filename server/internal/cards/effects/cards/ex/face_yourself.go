package cardsex

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Face Yourself — Sorcery {5}{R}{R}:
//
//	"For each creature target player controls, create a token that's a
//	 copy of that creature, except it has haste and 'At the beginning of
//	 the end step, if you don't control a planeswalker, sacrifice this
//	 token.'"
//
// The creatures are read once as the spell resolves, then each is copied
// under the caster's control (the copies' "you" is the caster). Haste
// rides the token's printed keywords (TokenCopyGainsHaste). The
// sacrifice clause is gained as part of the copying process (CR 707.9a),
// so it is a copiable value of each token: a named bundle on its
// GrantedAbilities (TokenCopyGainsAbility), which a copy of the token
// copies too, and which a token that loses all its abilities loses. It
// fires at EVERY end
// step, not only the caster's, and checks "if you don't control a
// planeswalker" both when the step begins and again on resolution
// (CR 603.4), so a planeswalker arriving in response saves the token.
// The 2026 Oracle update reads "sacrifice this token" where the card
// printed "this creature" (#5949): either way it names the token itself,
// so the sacrifice did not change.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3c0fe71c-bee1-41d3-aeb5-1723028eedde",
		Name:         "Face Yourself",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Grants: []effects.AbilityGrant{{
			Key: FaceYourselfSacrifice,
			Triggered: []game.TriggeredAbility{
				effects.On(game.EventStepBegan, effects.AllOf(effects.StepBegan(game.StepEnd, false), controllerHasNoPlaneswalker),
					"Face Yourself — sacrifice this token unless you control a planeswalker",
					faceYourselfSacrificeUnlessWalker),
			},
			Text: "At the beginning of the end step, if you don't control a planeswalker, sacrifice this token.",
		}},
		// The sacrifice rider waits for the prompts the copies ask (a CR 616 order on each token;
		// CR 608.2c, #4155). The event log's position before the copies is filed for it.
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := effects.FirstLegalTarget(ctx)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				var originals []uuid.UUID
				for _, c := range effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.ControlledBy(t.ID))) {
					originals = append(originals, c.InstanceID)
				}
				ctx.RecordStep(faceYourselfRecord, nil, int(effects.B25LastEventSeq(ctx.Game)))
				// One creation (#4594): the copies are created, and enter, together.
				return effects.CreateTokenCopiesTogether{Controller: ctx.Controller(), Copies: originals, Except: faceYourselfException}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				r := ctx.StepRecord(faceYourselfRecord)
				if len(r.Ns) == 0 {
					return nil
				}
				// Before the grant was a copiable value, this step gave each
				// token a layer-6 grant. It stays for a table restored between
				// the two steps of the old resolution, whose tokens were made
				// without the copiable grant; a token that has it is skipped.
				for _, id := range effects.B27TokensCreatedByAfter(ctx.Game, ctx.Controller(), uint64(r.N())) {
					if tok, ok := ctx.Game.LookupCardForEffect(id); ok && slices.Contains(tok.GrantedAbilities, game.GrantKey(FaceYourselfSacrifice)) {
						continue
					}
					if err := (effects.GrantAbilitiesFor{
						Target:   id,
						Keys:     []string{FaceYourselfSacrifice},
						Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), id),
						Label:    "Face Yourself — the copy is sacrificed at end step without a planeswalker",
					}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
		),
	})
}

// faceYourselfSacrifice is the key of the granted sacrifice bundle.
const FaceYourselfSacrifice = "face-yourself/sacrifice"

// faceYourselfException is "except it has haste and '…'" (CR 707.9a).
func faceYourselfException(t *game.Card) {
	effects.TokenCopyGainsHaste(t)
	effects.TokenCopyGainsAbility(t, FaceYourselfSacrifice)
}

// faceYourselfRecord files the event log's position before the copies are made.
const faceYourselfRecord = "face-yourself/cursor"

// controllerHasNoPlaneswalker is the intervening "if you don't control a
// planeswalker", "you" being the host creature's controller.
func controllerHasNoPlaneswalker(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return !effects.YouControlAPlaneswalker(g, source.Controller)
}

// faceYourselfSacrificeUnlessWalker re-checks the condition on resolution
// (CR 603.4) before sacrificing the host.
func faceYourselfSacrificeUnlessWalker(g *game.Game, item *game.StackItem) error {
	if effects.YouControlAPlaneswalker(g, item.Controller) {
		return nil
	}
	return effects.SacrificePermanent{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
}
