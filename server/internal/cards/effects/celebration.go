package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// celebration.go — "if two or more nonland permanents entered the
// battlefield under your control this turn" (celebration, an ability
// word, CR 207.2c; #5164).
//
// One fact about the player's TURN, read off the per-player turn tally
// (game.PlayerTurnTally.NonlandPermanentsEntered) through
// Game.NonlandPermanentsEnteredThisTurn: each nonland permanent that
// entered under the player's control counts once, whatever its card
// types (an artifact creature is one); tokens count; a land, a land
// creature included, does not; one that has since left the battlefield
// still counts. The source itself counts if it entered this turn.
//
// Readers, one per place the printed text puts the condition:
//
//   - Celebration(g, you) for a spell's or ability's own resolution.
//   - SelfPumpWhileCelebration for "this creature gets +p/+t as long as
//     …", a layer 7c static.
//   - WheneverThisAttacksIfCelebration,
//     AtBeginningOfYourCombatIfCelebration and
//     AtYourEndStepIfCelebration are the trigger shapes: an
//     intervening "if" (CR 603.4), checked as the trigger fires and
//     again as it resolves (InterveningIf with CondCelebration).

// celebrationThreshold is the printed "two or more".
const celebrationThreshold = 2

// Celebration reports whether two or more nonland permanents entered the
// battlefield under `you`'s control this turn. Caller must hold g.mu
// (every effect body and every static's AppliesTo does).
func Celebration(g *game.Game, you uuid.UUID) bool {
	return you != uuid.Nil && g.NonlandPermanentsEnteredThisTurn(you) >= celebrationThreshold
}

// CondCelebration is celebration as an intervening-if Condition, for
// InterveningIf and If: read for the item's controller as it resolves.
// The tally only rises within a turn, so it declares a footprint
// condition the trigger-independence check can argue
// (game.FootprintCondCelebration).
type CondCelebration struct{}

func (CondCelebration) Holds(ctx *Context) bool { return Celebration(ctx.Game, ctx.Controller()) }

func (CondCelebration) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondCelebration, celebrationThreshold), true
}

// YouHaveCelebration is the When for celebration as the trigger fires,
// for composing with AllOf.
func YouHaveCelebration(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return Celebration(g, source.Controller)
}

// SelfPumpWhileCelebration is "Celebration — This creature gets +p/+t as
// long as two or more nonland permanents entered the battlefield under
// your control this turn" (Armory Mice): a layer 7c modification that
// is on while the source's controller has celebration.
func SelfPumpWhileCelebration(power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Label:     "Gets " + signedPT(power, toughness) + " as long as two or more nonland permanents entered the battlefield under your control this turn.",
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: whileYouHave(Celebration, selfCreature),
		Apply:     pumpBy(power, toughness),
	}
}

// WheneverThisAttacksIfCelebration is "Celebration — Whenever ~ attacks,
// if two or more nonland permanents entered the battlefield under your
// control this turn, …" (Ash, Party Crasher).
func WheneverThisAttacksIfCelebration(label string, effect Effect) game.TriggeredAbility {
	return InterveningIf(CondCelebration{}, On(game.EventAttack, AllOf(ThisAttacked, YouHaveCelebration), label, effect))
}

// AtBeginningOfYourCombatIfCelebration is "Celebration — At the beginning
// of combat on your turn, if two or more nonland permanents entered the
// battlefield under your control this turn, …" (Belligerent of the Ball,
// Bespoke Battlegarb). Set Targets on the returned ability for a
// targeted body; with no legal target the trigger is removed (CR 603.3d).
func AtBeginningOfYourCombatIfCelebration(label string, effect Effect) game.TriggeredAbility {
	return InterveningIf(CondCelebration{}, On(game.EventStepBegan, AllOf(StepBegan(game.StepBeginCombat, true), YouHaveCelebration), label, effect))
}

// SelfKeywordWhileCelebration is "Celebration — This creature has
// <keyword> as long as two or more nonland permanents entered the
// battlefield under your control this turn" (Gallant Pie-Wielder's
// double strike, Grand Ball Guest's trample): a layer 6 grant
// (CR 613.1f) that is on while the source's controller has celebration.
func SelfKeywordWhileCelebration(keyword string) game.StaticAbility {
	ab := KeywordGrant(whileYouHave(Celebration, selfCreature), keyword)
	ab.Label = "Has " + keyword + " as long as two or more nonland permanents entered the battlefield under your control this turn."
	return ab
}

// AtYourEndStepIfCelebration is "Celebration — At the beginning of your
// end step, if two or more nonland permanents entered the battlefield
// under your control this turn, …" (Lady of Laughter): the intervening
// "if" is checked as the end step begins and again as the trigger
// resolves (CR 603.4).
func AtYourEndStepIfCelebration(label string, effect Effect) game.TriggeredAbility {
	return InterveningIf(CondCelebration{}, On(game.EventBeginEndStep, AllOf(ByYou, YouHaveCelebration), label, effect))
}
