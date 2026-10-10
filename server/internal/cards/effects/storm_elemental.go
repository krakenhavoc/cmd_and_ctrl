package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Storm Elemental — Creature — Elemental {5}{U}, 3/4:
//
//	"Flying
//	 {U}, Exile the top card of your library: Tap target creature with
//	 flying.
//	 {U}, Exile the top card of your library: If the exiled card is a
//	 snow land, this creature gets +1/+1 until end of turn."
//
// ADR 0109 §7 (#1902): "Exile the top N cards of your library" is a
// cost with nothing to choose. A library of fewer than N cards can't pay
// it (CR 118.3), and it is paid after every other cost (CR 601.2h).
// "The exiled card" is read off the payment record (Context.Exiled,
// CR 400.7j) as the second ability resolves: a land with the snow
// supertype (CR 205.4g) pumps the Elemental, anything else does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f1fee486-660e-4356-896e-671e8b675ad8",
		Name:            "Storm Elemental",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{
			{
				Label:   "{U}, Exile the top card of your library: Tap target creature with flying.",
				Cost:    Plus(ManaCost("{U}"), ExileTopOfLibrary(1)),
				Targets: TargetCreature("target creature with flying", HasKeyword("flying")),
				Effect:  tapTheFirstLegalTarget,
			},
			{
				Label:   "{U}, Exile the top card of your library: If the exiled card is a snow land, this creature gets +1/+1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    Plus(ManaCost("{U}"), ExileTopOfLibrary(1)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					for _, id := range NewContext(g, item).Exiled() {
						if c, ok := g.LookupCardForEffect(id); ok && c.IsLand() && c.HasSupertype("snow") {
							return thisGetsUntilEndOfTurn(1, 1, "Storm Elemental — +1/+1")(g, item)
						}
					}
					return nil
				},
			},
		},
	})
}
