package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fault Riders — Creature — Human Soldier {2}{R}, 2/2 (Prophecy, S82 OLD batch
// 674, #5617):
//
//	"Sacrifice a land: This creature gets +2/+0 and gains first strike until
//	 end of turn. Activate only once each turn."
//
// The sacrifice is the cost (Spitting Spider's B08SacrificeALand), so it is
// paid even if the ability is countered. "Activate only once each turn" is
// OncePerTurnActivation (CR 602.1b), counted per object from the announcement.
// The ability resolves into Flamespeaker Adept's two effects on the source: the
// pump and a first strike grant, both until end of turn.
//
// No simplification.
func init() {
	const label = "Sacrifice a land: This creature gets +2/+0 and gains first strike until end of turn. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:     "df516046-7f83-4d64-aa0d-d687631d919a",
		Name:         "Fault Riders",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2, Keywords: []string{"first strike"}}},
			Cost:      effects.B08SacrificeALand(),
			Condition: effects.OncePerTurnActivation(label),
			Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				if err := (effects.BoostUntilEOT{Target: ctx.Source(), Power: 2, Label: "Fault Riders — +2/+0"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: ctx.Source(), Keywords: []string{"first strike"},
					Label: "Fault Riders — first strike"}.Apply(ctx)
			})),
		}},
	})
}
