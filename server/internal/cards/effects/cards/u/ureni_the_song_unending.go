package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ureni, the Song Unending — Legendary Creature — Spirit Dragon
// {5}{G}{U}{R}, 10/10 (#1657):
//
//	"Flying, protection from white and from black
//	 When Ureni enters, it deals X damage divided as you choose among
//	 any number of target creatures and/or planeswalkers your opponents
//	 control, where X is the number of lands you control."
//
// The amount is read off the board: DivideBy(DivideLandsYouControl),
// read as the trigger is put on the stack — the target walk — and
// fixed there. The ruling: "The value of X won't change even if the
// number of lands you control changes after that point." "You cannot
// choose more than X targets" is the divided-clause gate's own rule
// (each target at least 1). A target that leaves in response takes
// nothing and its share is not moved (CR 608.2b).
//
// Protection from white and from black are the ADR 0072 tokens
// Animar, Soul of Elements prints.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1c995c80-3301-409b-822b-9297aa260823",
		Name:            "Ureni, the Song Unending",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "protection from white", "protection from black"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: effects.B06SelfETB,
			Targets: effects.TargetPermanent("any number of target creatures and/or planeswalkers your opponents control",
				effects.And(effects.Or(effects.Creature(), effects.Planeswalker()), effects.OpponentControls())).WithCount(0, 0).
				Dividing(effects.DivideBy(effects.DivideLandsYouControl)),
			Key: "Ureni, the Song Unending — X damage divided among the targets, where X is the number of lands you control",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DealDividedDamage(effects.NewContext(g, item))
			},
		}},
	})
}
