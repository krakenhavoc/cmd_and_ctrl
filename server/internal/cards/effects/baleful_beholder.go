package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Baleful Beholder — Creature — Beholder {4}{B}{B}, 6/5 (AFR):
//
//	"When this creature enters, choose one —
//	 • Antimagic Cone — Each opponent sacrifices an enchantment of their choice.
//	 • Fear Ray — Creatures you control gain menace until end of turn."
func init() {
	beholder := WhenThisEnters("Baleful Beholder — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	beholder.Modes = ChooseOne(
		ModeDoing("Antimagic Cone — Each opponent sacrifices an enchantment of their choice.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return EachPlayerSacrifices{ExceptController: true, Match: Enchantment(), Label: "an enchantment"}.Apply(ctx)
			}),
		ModeDoing("Fear Ray — Creatures you control gain menace until end of turn.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return GrantKeywordUntilEOT{Match: And(Creature(), YouControl()), Keywords: []string{"menace"},
					Label: "Baleful Beholder — Fear Ray: menace until end of turn"}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:     "e7645ac4-6e05-426a-a62d-a33b30c33102",
		Name:         "Baleful Beholder",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{beholder},
	})
}
