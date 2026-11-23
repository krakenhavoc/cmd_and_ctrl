package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Destined Warrior — Legendary Creature — Human Warrior {1}{W}{U}{B},
// 3/3 (Final Fantasy, #4462):
//
//	"First strike, vigilance, menace
//	 Cleric, Rogue, Warrior, and Wizard spells you cast cost {1} less to
//	 cast.
//	 At the beginning of combat on your turn, creatures you control get
//	 +1/+0 until end of turn. If you have a full party, creatures you
//	 control get +3/+0 until end of turn instead."
//
// The discount is Ballyrush Banneret's over four creature types: one
// predicate, so a spell of two of the types is discounted once. The pump
// is decided as the trigger resolves (a full party is four creatures,
// CR 700.8c; fin_b9_helpers.go) and affects the creatures you control
// then (CR 611.2c), the Warrior included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "393a5d32-fc8e-422b-a957-59e545a54cb6",
		Name:            "The Destined Warrior",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "vigilance", "menace"},
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Cleric, Rogue, Warrior, and Wizard spells you cast cost {1} less to cast.",
				effects.YourSpell(), finB9SpellOfAPartyType),
		},
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfYourCombat("The Destined Warrior — creatures you control get +1/+0 until end of turn, or +3/+0 with a full party",
				effects.Do(
					effects.If{Cond: effects.FinB9FullParty, Then: []effects.Applier{finB9PumpTeam(3)}},
					effects.If{Cond: effects.FinB9NoFullParty, Then: []effects.Applier{finB9PumpTeam(1)}},
				)),
		},
	})
}

// finB9SpellOfAPartyType is "Cleric, Rogue, Warrior, and Wizard spells": a
// spell with any of the four creature types.
func finB9SpellOfAPartyType(q game.CostQuery) bool {
	for _, t := range effects.FinB9PartyTypes {
		if q.Card.HasSubtype(q.Source.TextSubtype(t)) {
			return true
		}
	}
	return false
}

// finB9PumpTeam is "creatures you control get +N/+0 until end of turn".
func finB9PumpTeam(n int) effects.Applier {
	return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: n, Label: "The Destined Warrior"}
}
