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
//   - Empty graveyard at fire time → trigger no-ops silently.
//
// OnETB is dropped — the listener now owns ETB dispatch for this
// card. Cards still using OnETB stay on the direct-call path.
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
				ctx := NewContext(g, nil)
				_ = ReturnFromGraveyard{Target: top, Dest: game.ZoneHand}.Apply(ctx)
				return nil
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Eternal Witness — return top of graveyard to hand?",
			},
		}},
	})
}
