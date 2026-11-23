package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bogardan Dragonheart — Creature — Human Shaman {2}{R}, 2/2 (Jumpstart
// 2022, #5948):
//
//	"Sacrifice another creature: Until end of turn, this creature becomes a
//	 Dragon with base power and toughness 4/4, flying, and haste."
//
// One until-end-of-turn record on the Dragonheart with the mods at one
// timestamp, as Wild Shape writes them: base 4/4 at layer 7b (counters and
// pumps still apply on top), the creature type Dragon replacing its others
// (CR 205.1a; its card types stay) and flying and haste at layer 6. It acts
// only while the Dragonheart is still the permanent that activated it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3f0dc347-adac-428b-8f04-858799e91243",
		Name:         "Bogardan Dragonheart",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice another creature: Until end of turn, this creature becomes a Dragon with base power and toughness 4/4, flying, and haste.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
			Cost:    effects.SacrificeAnotherN(1, "another creature", effects.Creature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				return effects.UntilEndOfTurn(ctx, item.SourceCardID, nil, "Bogardan Dragonheart — becomes a 4/4 Dragon with flying and haste",
					game.SetBasePowerMod(4), game.SetBaseToughnessMod(4), game.SetCreatureTypesMod(ctx.TextSubtype("Dragon")),
					game.AddKeywordsMod("flying", "haste"))
			},
		}},
	})
}
