package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Chain Veil — Legendary Artifact {4}:
//
//	"At the beginning of your end step, if you didn't activate a
//	 loyalty ability of a planeswalker this turn, you lose 2 life.
//	 {4}, {T}: For each planeswalker you control, you may activate one
//	 of its loyalty abilities once this turn as though none of its
//	 loyalty abilities have been activated this turn."
//
// The end-step trigger has an intervening "if" (CR 603.4), checked as
// the step begins and again on resolution, against the turn's
// activation events (youActivatedALoyaltyAbilityThisTurn): a walker
// that has since left still counts (the 2014-07-18 ruling), and a
// granted loyalty row counts like a printed one. Only a planeswalker
// has loyalty abilities to activate, so "of a planeswalker" adds
// nothing to the read.
//
// The activated ability is Game.GrantExtraLoyaltyActivationForEffect
// (#4191, ADR 0145 amendment 2026-10-21). It changes the rules for the
// rest of the turn, so it is the activating player's and covers a
// planeswalker that comes under their control after it resolves; each
// resolution adds one more (an untapped Veil activated again makes
// three activations); it stacks with Oath of Teferi's and Urza's
// "twice"; and the timing of each activation is still a loyalty
// ability's, the controller's main phase with an empty stack.
//
// No simplification. Guffster's Party deck request (#4184).
func init() {
	const name = "The Chain Veil"
	Register(Spec{
		OracleID:     "fb88fb3d-6a85-4bb1-b316-7f759b45a0fa",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(chainVeilNoLoyaltyActivation{}, On(game.EventBeginEndStep,
				AllOf(ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return !youActivatedALoyaltyAbilityThisTurn(g, source.Controller)
				}),
				name+" — you lose 2 life", Do(GainLife{Amount: -2}))),
		},
		Activated: []ActivatedAbility{{
			Label: "{4}, {T}: For each planeswalker you control, you may activate one of its loyalty abilities once this turn as though none of its loyalty abilities have been activated this turn.",
			Cost:  Plus(ManaCost("{4}"), TapCost()),
			// The bot activates it in its own main phase, with
			// planeswalkers to use it (aiseat/heuristic/loyalty_grant.go).
			// It answers nothing on the stack (ADR 0142).
			Purpose: game.Purpose{Answers: game.AnswerValue, ExtraLoyaltyActivations: 1},
			Effect: func(g *game.Game, item *game.StackItem) error {
				g.GrantExtraLoyaltyActivationForEffect(item.Controller)
				return nil
			},
		}},
	})
}

// chainVeilNoLoyaltyActivation is the intervening "if you didn't
// activate a loyalty ability of a planeswalker this turn", re-read on
// resolution.
type chainVeilNoLoyaltyActivation struct{}

func (chainVeilNoLoyaltyActivation) Holds(ctx *Context) bool {
	return !youActivatedALoyaltyAbilityThisTurn(ctx.Game, ctx.Controller())
}

func (chainVeilNoLoyaltyActivation) condition() (game.FootprintStep, bool) {
	return game.FootprintStep{}, false
}
