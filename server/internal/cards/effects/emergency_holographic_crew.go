package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Emergency Holographic Crew — Artifact Creature — Illusion Officer {3}{U},
// 3/4 (Star Trek, #4973):
//
//	"When this creature enters, choose one —
//	 • Return target Spacecraft card from your graveyard to your hand.
//	 • Surveil 2."
//
// Charming Prince's modal enters trigger. The first bullet targets a card
// with the Spacecraft subtype in its controller's graveyard; the second
// surveils.
//
// No simplification.
func init() {
	etb := WhenThisEnters("Emergency Holographic Crew — choose one", func(*game.Game, *game.StackItem) error { return nil })
	etb.Modes = ChooseOne(
		ModeDoing("Return target Spacecraft card from your graveyard to your hand.",
			TargetCardInGraveyard("target Spacecraft card in your graveyard", YouOwn(), HasSubtype("Spacecraft")),
			ReturnTheModesGraveyardTargetToHand),
		ModeDoing("Surveil 2.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
			return Surveil{Player: item.Controller, N: 2}.Apply(ctx)
		}),
	)
	Register(Spec{
		OracleID:     "0b74d7de-8b9e-4a75-870e-230536fa8d79",
		Name:         "Emergency Holographic Crew",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{etb},
	})
}
