package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unholy Annex // Ritual Chamber — Enchantment — Room (ADR 0103):
//
//	Unholy Annex {2}{B}: "At the beginning of your end step, draw a
//	card. If you control a Demon, each opponent loses 2 life and you
//	gain 2 life. Otherwise, you lose 2 life."
//	Ritual Chamber {3}{B}{B}: "When you unlock this door, create a 6/6
//	black Demon creature token with flying."
//
// The Demon check is at resolution, after the draw, as printed — so a
// Demon made by Ritual Chamber in the same end step's earlier trigger
// counts.
func init() {
	Register(Room(RoomSpec{
		OracleID:     "bd388ad9-a47b-4b0b-b94a-8e4343cd3de5",
		Name:         "Unholy Annex // Ritual Chamber",
		Completeness: CompletenessFull,
		Left: Door{Triggered: []game.TriggeredAbility{
			AtYourEndStep("Unholy Annex — draw a card; with a Demon drain 2, otherwise lose 2", unholyAnnexEndStep),
		}},
		Right: Door{Triggered: []game.TriggeredAbility{
			WhenYouUnlockThisDoor(game.DoorRight, "Ritual Chamber — create a 6/6 black Demon creature token with flying",
				Do(CreateToken{Template: TokenCard("6/6 black Demon with flying"), N: 1})),
		}},
	}))
}

func unholyAnnexEndStep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (DrawCards{N: 1}).Apply(ctx); err != nil {
		return err
	}
	if ControlsA("Demon")(g, item.Controller) {
		return b40DrainEachOpponent(2)(g, item)
	}
	return g.ChangePlayerLifeForEffect(ctx.Source(), item.Controller, -2)
}
