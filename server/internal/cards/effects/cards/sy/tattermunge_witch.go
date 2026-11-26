package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tattermunge Witch — Creature — Goblin Shaman {1}{R/G} 2/1:
//
//	"{R}{G}: Each blocked creature gets +1/+0 and gains trample until end of
//	 turn."
//
// The blocked creatures (attackers with a blocker declared, any player's) as
// the ability resolves are the ones affected (CR 611.2c, CR 509.1h).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b1696989-7f00-41ec-82a4-fd10aed7a7f9",
		Name:         "Tattermunge Witch",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}{G}: Each blocked creature gets +1/+0 and gains trample until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:    effects.ManaCost("{R}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				match := effects.And(effects.Creature(), effects.BlockedCreature())
				if err := (effects.BoostUntilEOT{Match: match, Power: 1, Label: "Tattermunge Witch — blocked creatures get +1/+0 until end of turn"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Match: match, Keywords: []string{"trample"},
					Label: "Tattermunge Witch — blocked creatures gain trample until end of turn"}.Apply(ctx)
			},
		}},
	})
}
