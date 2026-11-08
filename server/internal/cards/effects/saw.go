package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Saw — Artifact — Equipment {2} (Duskmourn: House of Horror, #4466):
//
//	"Equipped creature gets +2/+0.
//	 Whenever equipped creature attacks, you may sacrifice a permanent
//	 other than that creature or this Equipment. If you do, draw a card.
//	 Equip {2}"
//
// The pump is PumpAttached and the attack trigger is Stormbeacon Blade's
// "whenever equipped creature attacks". The sacrifice is chosen as the
// trigger resolves (maySacrificeOne, Witherbloom Charm's "you may
// sacrifice a permanent. If you do"), over the permanents you control
// then: the Saw itself is left out (another), and so is the creature that
// attacked, read off the trigger's event, even if the Saw has moved to
// another creature since. The card is drawn only for a permanent that was
// really sacrificed (CR 701.21a); with nothing else to sacrifice there is
// no question and no card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d88bfa98-6869-4d64-a9d2-07a308a67fe2",
		Name:         "Saw",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 0),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.IsAttachedTo(ev.CardID)
			}, "Saw — you may sacrifice a permanent other than that creature or this Equipment to draw a card",
				func(g *game.Game, item *game.StackItem) error {
					attacker := NewContext(g, item).Trigger().Event.CardID
					return maySacrificeOne(g, item, "Saw — sacrifice a permanent other than the attacking creature or Saw to draw a card?",
						func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID != attacker },
						true, sawDraw)
				}),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}

// sawDraw is the card, once a permanent has really been sacrificed. Keyed
// (#4290): an on-disk identity, never renamed or reused.
var sawDraw = RegisterSacrificedOne("prompt-run/saw-sacrifice-draw", func(ctx *Context, _ uuid.UUID) error {
	return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
})
