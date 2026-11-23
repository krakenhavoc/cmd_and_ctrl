package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guardian Naga // Banishing Coils — an adventure card (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	Guardian Naga {5}{W}{W}, Creature — Snake 5/6:
//	"Vigilance
//	 During your turn, prevent all damage that would be dealt to this
//	 creature."
//
//	Banishing Coils {2}{W}, Instant — Adventure:
//	"Exile target artifact or enchantment."
//
// The adventure layout is Brazen Borrower's. The prevention is a plain CR
// 615.1a replacement over damage dealt to this creature (Glittering Lion),
// applied only while its controller is the active player, which is read as
// the damage would be dealt. The Adventure is Forsake the Worldly's exile.
//
// No simplification.
func init() {
	const oracleID = "67dec976-bcf5-4995-8da1-cd570862d3cf"

	effects.Register(effects.Spec{
		OracleID:        oracleID,
		Name:            "Guardian Naga",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Replacements: []game.ReplacementEffect{
			effects.PreventAllDamageDealtTo(effects.PreventionStatic{
				To: effects.ToThisCreature,
				While: func(g *game.Game, src *game.Card) bool {
					return effects.IsActivePlayer(g, src.Controller)
				},
				Label: "Guardian Naga — during your turn, prevent all damage that would be dealt to this creature",
			}),
		},
	})

	effects.Register(effects.Spec{
		OracleID:     oracleID + "#1",
		Name:         "Banishing Coils",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ExileFirstTarget(ctx.Game, item)
		},
	})
}
