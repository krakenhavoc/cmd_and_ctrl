package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Corruption of Towashi — Enchantment {4}{U} (March of the Machine):
//
//	"When this enchantment enters, incubate 4.
//	 Whenever a permanent you control transforms or a permanent you
//	 control enters transformed, you may draw a card. Do this only once
//	 each turn."
//
// The transform half is complete: any permanent of yours turning over
// (an Incubator paid for, a daybound permanent changing side) triggers
// it. "Enters transformed" is not: no engine event marks a permanent
// that entered already turned over, so that half never fires.
//
// "Do this only once each turn" is read as the trigger firing once per
// turn, per object. Printed, a trigger whose draw was already done still
// goes on the stack and does nothing; the only observable difference is
// a first trigger that is countered, which here also uses up the turn.
// Weaker than printed.
func init() {
	const label = "Corruption of Towashi — you may draw a card"
	Register(Spec{
		OracleID:     "242c39b4-5ef0-469a-a852-738f79d93bd6",
		Name:         "Corruption of Towashi",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"A permanent entering transformed doesn't trigger the draw — only a permanent turning over does.",
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Corruption of Towashi — incubate 4", Do(Incubate{N: 4})),
			Optional(On(game.EventTransform, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && !b11TriggeredThisTurn(g, source.InstanceID, label)
			}, label, Do(DrawCards{N: 1})), "Corruption of Towashi — draw a card?"),
		},
	})
}
