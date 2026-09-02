package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eternal Witness — 2/1 Human Shaman for {1}{G}{G}:
//
//	"When Eternal Witness enters the battlefield, you may return
//	target card from your graveyard to your hand."
//
// S19 sub-PR 3 migrates the S14 OnETB direct-call to the
// Triggered slot. Behavior parity:
//
//   - OptionalPrompt drives the "you may" gate (S14 treated may as
//     do; S19 surfaces the prompt — the controller can decline).
//   - Sandbox auto-pick: most-recently-added card in the
//     controller's graveyard (top of pile), matching S14's
//     simplification. A real target picker lands with S20.
//   - The pick is stamped onto item.Targets when the trigger is
//     put on the stack. If the card has left the graveyard by the
//     time the trigger resolves (someone exiled it in response),
//     ReturnFromGraveyard returns ErrCardNotFound, which surfaces
//     as an effect_error breadcrumb rather than wedging the stack.
//   - Empty graveyard at fire time → a "yes" builds nothing and
//     the trigger never reaches the stack.
func init() {
	Register(Spec{
		OracleID: "30b24e8e-3b0e-4d8e-90f3-f66eb7c1858c",
		Name:     "Eternal Witness",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				controller := g.PlayerByIDForEffect(source.Controller)
				if controller == nil || controller.Graveyard.Size() == 0 {
					return nil
				}
				top := controller.Graveyard.Cards[controller.Graveyard.Size()-1].InstanceID
				item := game.NewTriggeredItem(source, "Eternal Witness — return target card to hand",
					func(g *game.Game, item *game.StackItem) error {
						if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
							return nil
						}
						return ReturnFromGraveyard{Target: item.Targets[0].ID, Dest: game.ZoneHand}.Apply(NewContext(g, item))
					})
				item.Targets = []game.TargetRef{{Kind: game.TargetCard, ID: top}}
				return item
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Eternal Witness — return top of graveyard to hand?",
			},
			// Warn the chooser when the graveyard is empty — "Yes"
			// would pass without effect otherwise.
			HasLegalTarget: func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				controller := g.PlayerByIDForEffect(source.Controller)
				return controller != nil && controller.Graveyard.Size() > 0
			},
		}},
	})
}
