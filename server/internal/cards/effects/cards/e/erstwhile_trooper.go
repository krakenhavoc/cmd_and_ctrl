package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Erstwhile Trooper — Creature — Zombie Soldier {1}{B}{G}, 2/2 (Guilds of
// Ravnica, #5525):
//
//	"Discard a creature card: This creature gets +2/+2 and gains trample until
//	 end of turn. Activate only once each turn."
//
// Vampire Hounds' discard cost and Mindful Biomancer's once-a-turn gate, with a
// trample grant beside the pump. Both effects are on the Trooper itself, and a
// Trooper that has left and come back is a new object that gets neither
// (CR 400.7).
//
// No simplification.
func init() {
	const label = "Discard a creature card: This creature gets +2/+2 and gains trample until end of turn. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:     "a1640d9a-367f-4b90-bd4d-18f2c9244ccf",
		Name:         "Erstwhile Trooper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2, Toughness: 2, Keywords: []string{"trample"}}},
			Cost:      effects.DiscardCardsMatching(1, "a creature card", effects.MatchCreature),
			Condition: effects.OncePerTurnActivation(label),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.BoostUntilEOT{Target: item.SourceCardID, Power: 2, Toughness: 2, Label: "Erstwhile Trooper — +2/+2 until end of turn"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"trample"},
					Label: "Erstwhile Trooper — trample"}.Apply(ctx)
			},
		}},
	})
}
