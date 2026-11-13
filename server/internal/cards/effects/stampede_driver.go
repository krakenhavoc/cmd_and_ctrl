package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stampede Driver — Creature — Human Spellshaper {G}, 1/1
// (Nemesis, #5618):
//
//	"{1}{G}, {T}, Discard a card: Creatures you control get +1/+1 and gain trample until end of turn."
//
// The spellshaper cost (nemSpellshaperCost). "Creatures you control" is
// read once, as the ability resolves (CR 611.2c): a creature that
// arrives later gets neither half. Both end in the cleanup step (CR
// 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3c652a31-3b56-4050-9613-d0481e54dfd7",
		Name:         "Stampede Driver",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}, {T}, Discard a card: Creatures you control get +1/+1 and gain trample until end of turn.",
			Cost:    nemSpellshaperCost("{1}{G}"),
			Purpose: game.Purpose{TeamPump: &game.Pump{Power: 1, Toughness: 1, Keywords: []string{"trample"}}, Answers: game.AnswerPump},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				yours := And(Creature(), YouControl())
				label := "Stampede Driver — +1/+1 and trample until end of turn"
				if err := (BoostUntilEOT{Match: yours, Power: 1, Toughness: 1, Label: label}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Match: yours, Keywords: []string{"trample"}, Label: label}.Apply(ctx)
			},
		}},
	})
}
