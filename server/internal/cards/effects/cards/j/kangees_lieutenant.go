package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kangee's Lieutenant — Creature — Bird Soldier {2}{W}, 1/1 (Starter
// Commander Decks):
//
//	"Flying
//	 Whenever this creature attacks, attacking creatures with flying get
//	 +1/+1 until end of turn.
//	 Encore {5}{W} (...)"
//
// The pump is Warmonger Hellkite's: every attacking creature that has flying
// as the trigger resolves, whoever controls it, the Lieutenant included. A
// creature that gains flying or attacks later in the turn is not pumped (CR
// 611.2c). Encore is the shared keyword; each token copy has the trigger too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5a040363-dba5-49d2-9756-e14af61cf818",
		Name:            "Kangee's Lieutenant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []effects.ActivatedAbility{effects.Encore("{5}{W}")},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Kangee's Lieutenant — attacking creatures with flying get +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{
						Match:     effects.And(effects.AttackingCreature(), effects.HasKeyword("flying")),
						Power:     1,
						Toughness: 1,
						Label:     "Kangee's Lieutenant — attacking creatures with flying get +1/+1",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
