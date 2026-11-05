package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jefferies Tube — Artifact {1} (Star Trek Commander, #4974):
//
//	"{2}, {T}: Put two charge counters on target artifact.
//	 {3}, {T}: Target creature you control can't be blocked this turn.
//	 {4}, {T}: Draw a card. You may put a land card from your hand onto
//	 the battlefield tapped."
//
// Three ordinary activated abilities. The charge counters go on through
// the CR 614 counter window, on any artifact (a Spacecraft stations with
// them). The unblockability is Thassa, God of the Sea's restriction until
// end of turn. The third ability draws first and then offers the land
// put (H.E.R.B.I.E. Scout Unit's sequence), so the card just drawn can be
// the land.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2ca270f1-a8a2-438f-9321-2c653165f3e4",
		Name:         "Jefferies Tube",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{2}, {T}: Put two charge counters on target artifact.",
				Cost:    Plus(ManaCost("{2}"), TapCost()),
				Targets: TargetPermanent("target artifact", Artifact()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return AddCounter{Target: id, Kind: game.CounterCharge, N: 2}.Apply(ctx)
				},
			},
			{
				Label:   "{3}, {T}: Target creature you control can't be blocked this turn.",
				Purpose: ForTargets(CantBeBlockedUntilEOT(0, 0, 0)),
				Cost:    Plus(ManaCost("{3}"), TapCost()),
				Targets: TargetCreature("target creature you control", YouControl()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return RestrictUntilEOT{
						Target:       id,
						Restrictions: game.CantBeBlocked,
						Label:        "Jefferies Tube — can't be blocked",
					}.Apply(ctx)
				},
			},
			{
				Label:   "{4}, {T}: Draw a card. You may put a land card from your hand onto the battlefield tapped.",
				Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{4}"), TapCost()),
				Effect: Do(DrawCards{N: 1}, PutFromHandOntoBattlefield{
					Match:    Land(),
					Tapped:   true,
					Optional: true,
					Label:    "Jefferies Tube — you may put a land card from your hand onto the battlefield tapped",
				}),
			},
		},
	})
}
