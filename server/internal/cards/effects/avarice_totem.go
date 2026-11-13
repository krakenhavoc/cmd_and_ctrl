package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Avarice Totem — Artifact {1} (Fifth Dawn, #5604):
//
//	"{5}: Exchange control of this artifact and target nonland
//	 permanent."
//
// Gilded Drake's exchange as an activated ability. The exchange is all
// or nothing (CR 701.12a): if the Totem has left the battlefield by the
// time the ability resolves, nothing changes hands. Whoever controls the
// Totem afterwards may activate it in turn, which is the printed card's
// back-and-forth.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c4698dcc-baa0-44c5-a283-22f9dfe150f5",
		Name:         "Avarice Totem",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{5}: Exchange control of this artifact and target nonland permanent.",
			Cost:    ManaCost("{5}"),
			Targets: TargetPermanent("target nonland permanent", Nonland()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return ExchangeControl{A: item.SourceCardID, B: id, Label: "Avarice Totem — exchange control"}.Apply(ctx)
			},
		}},
	})
}
