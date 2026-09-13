package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soul-Guide Lantern — Artifact {1} (EDHREC rank 1354):
//
//	"When this artifact enters, exile target card from a graveyard.
//	 {T}, Sacrifice this artifact: Exile each opponent's graveyard.
//	 {1}, {T}, Sacrifice this artifact: Draw a card."
//
// One-mana graveyard hate that cycles when nothing needs hating.
// The entry trigger is a real "target card from a graveyard" — any
// graveyard, any card type, chosen when the trigger goes on the
// stack, and dropped with no prompt when every graveyard is empty
// (CR 603.3d). The two activated abilities share the tap-and-
// sacrifice cost; the sweep walks each opponent's pile through the
// body Bojuka Bog uses, and the draw is a draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1b5e6560-ff2e-4475-96cb-63f64c8a86db",
		Name:         "Soul-Guide Lantern",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: TargetCardInGraveyard("target card from a graveyard"),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Soul-Guide Lantern — exile target card from a graveyard",
					func(g *game.Game, item *game.StackItem) error {
						if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
							return nil
						}
						return ExileTarget{Target: item.Targets[0].ID}.Apply(NewContext(g, item))
					})
			},
		}},
		Activated: []ActivatedAbility{
			{
				Label: "{T}, Sacrifice this artifact: Exile each opponent's graveyard.",
				Cost:  Plus(TapCost(), SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					for _, opp := range NewContext(g, item).Opponents() {
						if err := exileGraveyardForEffect(g, item, opp); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Label: "{1}, {T}, Sacrifice this artifact: Draw a card.",
				Cost:  Plus(ManaCost("{1}"), TapCost(), SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
