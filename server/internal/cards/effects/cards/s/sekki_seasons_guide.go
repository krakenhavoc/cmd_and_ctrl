package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sekki, Seasons' Guide — Legendary Creature — Spirit {5}{G}{G}{G}, 0/0
// (Saviors of Kamigawa, #5601):
//
//	"Sekki enters with eight +1/+1 counters on it.
//	 If damage would be dealt to Sekki, prevent that damage, remove that many
//	 +1/+1 counters from Sekki, and create that many 1/1 colorless Spirit
//	 creature tokens.
//	 Sacrifice eight Spirits: Return this card from your graveyard to the
//	 battlefield."
//
// The counters are a self-replacement (CR 614.1c). The damage replacement is
// Protean Hydra's prevention (CR 615) with a follow-up that removes the
// counters (up to what Sekki has) and makes a Spirit for every point of damage
// prevented. The recursion is an activated ability from the graveyard
// (CR 602.2) whose cost is eight Spirits you control, each sacrificed; Sekki is
// not on the battlefield, so it is not one of them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1b92501c-80fe-4ae7-834e-72f57ba41a6a",
		Name:         "Sekki, Seasons' Guide",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters("+1/+1", 8, "Sekki, Seasons' Guide: enters with eight +1/+1 counters"),
			effects.PreventDamageDealtTo(effects.PreventionStatic{
				To:    effects.ToThisCreature,
				Then:  effects.OldB592SekkiSpiritsBody,
				Label: "Sekki, Seasons' Guide — prevent damage to it, remove that many +1/+1 counters and create that many Spirits",
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice eight Spirits: Return this card from your graveyard to the battlefield.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.SacrificeN(8, "eight Spirits", effects.HasSubtype("Spirit")),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.ReturnThisFromGraveyardToBattlefield,
		}},
	})
}
