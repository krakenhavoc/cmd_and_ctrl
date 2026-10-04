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

// The branch keys, parallel to the option list.
const (
	tormentorDraw      = "draw"
	tormentorSacrifice = "sacrifice"
	tormentorLife      = "life"
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
	controller := item.Controller

	options := []game.ChoiceOption{{Label: "Let them draw a card"}}
	branches := []string{tormentorDraw}
	creatures := creaturesControlledByPlayer(g, victim)
	if len(creatures) > 0 {
		options = append(options, game.ChoiceOption{Label: "Sacrifice a creature"})
		branches = append(branches, tormentorSacrifice)
	}
	if p.Life >= 3 {
		options = append(options, game.ChoiceOption{Label: "Pay 3 life", LifeCost: 3})
		branches = append(branches, tormentorLife)
	}

	return PickOption{
		Player:   victim,
		Question: "Indulgent Tormentor — let them draw a card, sacrifice a creature, or pay 3 life?",
		Options:  options,
		Then: func(ctx *Context, index int) error {
			draw := func(ctx *Context) error {
				return DrawCards{Player: controller, N: 1}.Apply(ctx)
			}
			if index < 0 || index >= len(branches) {
				return draw(ctx)
			}
			switch branches[index] {
			case tormentorSacrifice:
				return SacrificeChoice{
					Player:     victim,
					Candidates: creaturesControlledByPlayer(ctx.Game, victim),
					Question:   "Indulgent Tormentor — sacrifice a creature",
				}.Apply(ctx)
			case tormentorLife:
				return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), victim, -3)
			default:
				return draw(ctx)
			}
		},
	}.Apply(ctx)
}
