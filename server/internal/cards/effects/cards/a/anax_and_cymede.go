package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Anax and Cymede — Legendary Creature — Human Soldier {1}{R}{W}, 3/2 (Theros,
// S82 OLD b388):
//
//	"First strike, vigilance
//	 Heroic — Whenever you cast a spell that targets Anax and Cymede, creatures
//	 you control get +1/+1 and gain trample until end of turn."
//
// One until-end-of-turn record over the creatures the caster controls as the
// trigger resolves (CR 611.2c): +1/+1 (layer 7c) and trample (layer 6). A
// creature that enters later gets neither.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9c4219db-1a7a-4d61-bb8b-7998cfed1faf",
		Name:            "Anax and Cymede",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Heroic("Anax and Cymede — creatures you control get +1/+1 and gain trample until end of turn (heroic)",
				func(g *game.Game, item *game.StackItem) error {
					return effects.UntilEndOfTurn(effects.NewContext(g, item), uuid.Nil,
						effects.And(effects.Creature(), effects.YouControl()),
						"Anax and Cymede — creatures you control get +1/+1 and gain trample",
						game.ModifyPTMod(1, 1), game.AddKeywordsMod("trample"))
				}),
		},
	})
}
