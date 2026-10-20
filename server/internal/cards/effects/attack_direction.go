package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attack_direction.go — the card-facing half of ADR 0206 (#4197):
// "each player may attack only the nearest opponent in the chosen
// direction and planeswalkers controlled by that opponent". The engine
// half (left and right at the table, the nearest opponent, the one
// reader every attack-legality caller shares) is
// game/attack_direction.go.
//
// A permanent's static (Pramikon, Sky Rampart):
//
//	AsEnters: ChooseLeftOrRightAsEnters("Pramikon, Sky Rampart"),
//	AttackOnlyNearestOpponentInChosenDirection: true,
//
// A resolved ability with a duration (Teyo, Geometric Tactician's −2):
//
//	var teyoDirection = AttackDirectionUntilYourNextTurnThen("option-pick/teyo-geometric-tactician-direction", "Teyo, Geometric Tactician")
//	Effect: ChooseLeftOrRightThenEachPlayerAttacksOnlyThatWayUntilYourNextTurn("Teyo, Geometric Tactician", teyoDirection),

// leftOrRightQuestion is the prompt header: the card's name, the
// choice, and what each answer means at this table, since a screen has
// no left and right of its own.
func leftOrRightQuestion(name string) string {
	return name + " — choose left or right. Each player may attack only the nearest opponent that way " +
		"and their planeswalkers (left is the next player in turn order, right the player before)"
}

// leftOrRightOptions are the two answers in printed order. The first is
// the one a bot with no opinion takes.
func leftOrRightOptions() []string {
	return []string{game.AttackDirectionLeftWord, game.AttackDirectionRightWord}
}

// ChooseLeftOrRightAsEnters builds the `Spec.AsEnters` for "As this
// permanent enters, choose left or right" (CR 614.12). The answer lands
// on the permanent's Card.ChosenOption as "Left" or "Right", which
// Spec.AttackOnlyNearestOpponentInChosenDirection reads. Until it is
// answered the permanent restricts nothing.
func ChooseLeftOrRightAsEnters(name string) func(*game.Card, *Context) error {
	question := leftOrRightQuestion(name)
	return func(card *game.Card, ctx *Context) error {
		ctx.Game.QueueChooseOptionAsEntersForEffect(card.Controller, card.InstanceID, question, leftOrRightOptions())
		return nil
	}
}

// ChooseLeftOrRightThenEachPlayerAttacksOnlyThatWayUntilYourNextTurn is
// "Choose left or right. Until your next turn, each player may attack
// only the nearest opponent in the last chosen direction and
// planeswalkers controlled by that opponent." The controller chooses as
// the ability resolves; `then` is the card's registered answer
// (AttackDirectionUntilYourNextTurnThen), so a table waiting on it is
// still a restore point. "The last chosen direction" is the one chosen
// here (see the engine file).
func ChooseLeftOrRightThenEachPlayerAttacksOnlyThatWayUntilYourNextTurn(name string, then game.OptionPickThen) func(*game.Game, *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		options := make([]game.ChoiceOption, 0, 2)
		for _, w := range leftOrRightOptions() {
			options = append(options, game.ChoiceOption{Label: w})
		}
		return PickOption{
			Question: leftOrRightQuestion(name),
			Options:  options,
			ThenKey:  then,
		}.Apply(NewContext(g, item))
	}
}

// AttackDirectionUntilYourNextTurnThen registers the answer to that
// question for one card, as a package-level var in the card's file:
// the chosen direction becomes a grant on every player, labelled with
// `name` for the refusal sentence, until the next turn of the
// ability's controller. The key is an on-disk identity
// ("option-pick/<card>-direction"): never renamed, never reused.
func AttackDirectionUntilYourNextTurnThen(key, name string) game.OptionPickThen {
	return OptionPickThen(key, func(ctx *Context, r game.OptionPicked) error {
		if r.Option == nil {
			return nil
		}
		dir := game.AttackDirectionOfWord(r.Option.Label)
		if dir == "" {
			return nil
		}
		ctx.Game.GrantAttackDirectionForEffect(dir, name, r.Source, ctx.Game.UntilYourNextTurnDuration(ctx.Controller()))
		return nil
	})
}
