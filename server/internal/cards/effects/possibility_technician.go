package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Possibility Technician — Creature — Kavu Artificer {2}{R}, 3/3:
//
//	"Whenever this creature or another Kavu you control enters, exile
//	 the top card of your library. For as long as that card remains
//	 exiled, you may play it if you control a Kavu.
//	 Warp {1}{R} (You may cast this card from your hand for its warp
//	 cost. Exile this creature at the beginning of the next end step,
//	 then you may cast it from exile on a later turn.)"
//
// A conditional exile play (#3192, ADR 0066's second 2026-10-13
// amendment): the permission lasts for as long as the card remains
// exiled and is live only while its holder controls a Kavu
// (game.CondControlSubtype, read each time the card is asked about, so
// losing the last Kavu shuts it and getting another opens it again).
// The trigger is one ability watching the Technician's own entry and
// every other Kavu's. Warp is the shared alternative cost, paid from
// hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "ebbf84f7-d142-4de5-bfc1-ea86f082d733",
		Name:             "Possibility Technician",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Warp("{1}{R}")},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && (ev.CardID == source.InstanceID || c.HasSubtype("Kavu"))
			}, "Possibility Technician — exile the top card of your library", func(g *game.Game, item *game.StackItem) error {
				return ExileTopWithPermission{
					From: item.Controller, GrantTo: item.Controller, N: 1,
					WhileExiled: true,
					Condition:   &game.PermissionCondition{Key: game.CondControlSubtype, Subtype: "Kavu"},
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
