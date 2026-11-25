package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Linvala, Shield of Sea Gate — Legendary Creature — Angel Wizard {1}{W}{U},
// 3/3 (Zendikar Rising, #5508):
//
//	"Flying
//	 At the beginning of combat on your turn, if you have a full party,
//	 choose target nonland permanent an opponent controls. Until your next
//	 turn, it can't attack or block, and its activated abilities can't be
//	 activated.
//	 Sacrifice Linvala: Choose hexproof or indestructible. Creatures you
//	 control gain that ability until end of turn."
//
// The trigger's "if you have a full party" is an intervening if (CR 603.4,
// CR 700.8c): checked as combat begins, so no target is chosen without one,
// and again as it resolves. Its effect is exactly detain (CR 701.35a), which
// also stops mana and loyalty abilities, until the controller's next turn.
// Linvala's party is counted live, so she is the Wizard if nothing else is.
//
// The sacrifice is a cost, so the choice is made as the ability resolves and
// does not need Linvala: every creature the controller has then gains the
// chosen keyword until end of turn (CR 611.2c), those that enter later do not.
//
// No simplification.
func init() {
	const detainLabel = "Linvala, Shield of Sea Gate — until your next turn, it can't attack or block, and its activated abilities can't be activated"
	effects.Register(effects.Spec{
		OracleID:        "afd8ff3e-1d57-4fc5-91dc-1a8400c98787",
		Name:            "Linvala, Shield of Sea Gate",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.InterveningIfYou("you have a full party", effects.FinB9HasFullParty,
					effects.AtBeginningOfYourCombat("Linvala, Shield of Sea Gate — choose target nonland permanent an opponent controls; until your next turn it can't attack or block and its activated abilities can't be activated",
						effects.DetainEachLegalTarget(detainLabel))),
				effects.TargetPermanent("target nonland permanent an opponent controls", effects.Nonland(), effects.OpponentControls())),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice Linvala: Choose hexproof or indestructible. Creatures you control gain that ability until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.SacrificeThis(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.PickOption{
					Question: "Linvala, Shield of Sea Gate — choose hexproof or indestructible",
					Options:  []game.ChoiceOption{{Label: "Hexproof"}, {Label: "Indestructible"}},
					ThenKey:  linvalaGain,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

// linvalaGain is the chosen keyword going to the creatures the controller
// has as the ability finishes resolving. An on-disk key.
var linvalaGain = effects.OptionPickThen("option-pick/linvala-shield-of-sea-gate-keyword", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Option == nil {
		return nil
	}
	kw := "hexproof"
	if r.Index == 1 {
		kw = "indestructible"
	}
	return effects.GrantKeywordUntilEOT{
		Match:    effects.And(effects.Creature(), effects.YouControl()),
		Keywords: []string{kw},
		Label:    "Linvala, Shield of Sea Gate — " + kw + " until end of turn",
	}.Apply(ctx)
})
