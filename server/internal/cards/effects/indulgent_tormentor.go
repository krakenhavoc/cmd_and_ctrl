package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Indulgent Tormentor — Creature — Demon {3}{B}{B}, 5/3:
//
//	"Flying
//	 At the beginning of your upkeep, draw a card unless target
//	 opponent sacrifices a creature of their choice or pays 3 life."
//
// The trigger targets an opponent when it goes on the stack, and that
// opponent answers a three-way question as it resolves: let you draw,
// sacrifice a creature, or pay 3 life. Letting you draw is the
// first option (the always-available one) and is what a refusal does.
// An opponent with no creature is not offered the sacrifice, and one
// below 3 life is not offered the payment (CR 119.4). The target going
// illegal fizzles the whole trigger, so nothing is drawn (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8b202c63-c961-4590-958f-d17e76610ab5",
		Name:            "Indulgent Tormentor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtYourUpkeep("Indulgent Tormentor — draw a card unless target opponent sacrifices a creature or pays 3 life",
					indulgentTormentorAsk),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}

// The branch labels: the answer is told which option was taken (OptionPicked.Option),
// so the question's branches need no list riding along.
const (
	tormentorSacrificeLabel = "Sacrifice a creature"
	tormentorLifeLabel      = "Pay 3 life"
)

// indulgentTormentorAsk puts the question to the targeted opponent.
func indulgentTormentorAsk(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
		return nil
	}
	victim := item.Targets[0].ID
	p := g.PlayerByIDForEffect(victim)
	if p == nil || p.Eliminated {
		return nil
	}
	ctx := NewContext(g, item)

	options := []game.ChoiceOption{{Label: "Let them draw a card"}}
	creatures := creaturesControlledByPlayer(g, victim)
	if len(creatures) > 0 {
		options = append(options, game.ChoiceOption{Label: tormentorSacrificeLabel})
	}
	if p.Life >= 3 {
		options = append(options, game.ChoiceOption{Label: tormentorLifeLabel, LifeCost: 3})
	}

	return PickOption{
		Player:   victim,
		Question: "Indulgent Tormentor — let them draw a card, sacrifice a creature, or pay 3 life?",
		Options:  options,
		ThenKey:  tormentorAnswer,
	}.Apply(ctx)
}

// tormentorAnswer is the opponent's answer: sacrifice a creature, pay 3 life,
// or (the first option, and what a refusal does) let the controller draw.
// Keyed (#4253): the opponent is the prompt's chooser, the controller its
// first carried ID and the branch the option taken.
var tormentorAnswer = OptionPickThen("option-pick/indulgent-tormentor-answer", func(ctx *Context, r game.OptionPicked) error {
	victim := r.Chooser
	if r.Option != nil {
		switch r.Option.Label {
		case tormentorSacrificeLabel:
			return SacrificeChoice{
				Player:     victim,
				Candidates: creaturesControlledByPlayer(ctx.Game, victim),
				Question:   "Indulgent Tormentor — sacrifice a creature",
			}.Apply(ctx)
		case tormentorLifeLabel:
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), victim, -3)
		}
	}
	return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
})
