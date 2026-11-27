package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Angelic Skirmisher — Creature — Angel {4}{W}{W}, 4/4 (Gatecrash, S82 OLD
// b408, #5565):
//
//	"Flying
//	 At the beginning of each combat, choose first strike, vigilance, or
//	 lifelink. Creatures you control gain that ability until end of turn."
//
// The word is chosen as the trigger resolves (an option pick, so a table waiting
// on it is a restore point), by the Skirmisher's controller, at every player's
// combat. The creatures are the ones its controller controls when the answer
// is given, fixed then (CR 611.2c); a creature that arrives later does not gain it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "29e07d57-0f05-441c-844f-cf1d71102b2d",
		Name:            "Angelic Skirmisher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfEachCombat("Angelic Skirmisher — choose first strike, vigilance, or lifelink; creatures you control gain it until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.PickOption{
						Question: "Angelic Skirmisher — choose first strike, vigilance, or lifelink",
						Options:  []game.ChoiceOption{{Label: "First strike"}, {Label: "Vigilance"}, {Label: "Lifelink"}},
						ThenKey:  angelicSkirmisherChosen,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// angelicSkirmisherChosen gives the chosen ability to the creatures its
// controller controls, until end of turn. An on-disk key: never renamed.
var angelicSkirmisherChosen = effects.OptionPickThen("option-pick/s82-old-b408-angelic-skirmisher", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Option == nil {
		return nil
	}
	keyword := map[string]string{"First strike": "first strike", "Vigilance": "vigilance", "Lifelink": "lifelink"}[r.Option.Label]
	if keyword == "" {
		return nil
	}
	return effects.GrantKeywordUntilEOT{
		Match:    effects.And(effects.Creature(), effects.YouControl()),
		Keywords: []string{keyword},
		Label:    "Angelic Skirmisher — creatures you control gain " + keyword + " until end of turn",
	}.Apply(ctx)
})
