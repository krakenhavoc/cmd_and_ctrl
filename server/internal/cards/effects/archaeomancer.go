package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Archaeomancer — Creature — Human Wizard, {2}{U}{U}, 1/2
// (EDHREC rank 959):
//
//	"When this creature enters, return target instant or sorcery card
//	 from your graveyard to your hand."
//
// The blue Eternal Witness, and the flicker deck's engine — every
// blink is a spell back. One targeted ETB trigger over the graveyard
// clause; mandatory, so with an instant or sorcery in the graveyard
// the controller picks one, and with none the trigger is removed
// without a prompt (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a91a3266-cadd-47a0-9b20-160307f14c07",
		Name:         "Archaeomancer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets:   TargetCardInGraveyard("target instant or sorcery card in your graveyard", Or(Instant(), Sorcery()), YouOwn()),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Archaeomancer — return an instant or sorcery card to hand",
					func(g *game.Game, item *game.StackItem) error {
						if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
							return nil
						}
						return ReturnFromGraveyard{Target: item.Targets[0].ID, Dest: game.ZoneHand}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
