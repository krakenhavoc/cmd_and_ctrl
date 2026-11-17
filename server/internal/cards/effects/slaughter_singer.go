package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slaughter Singer — Creature — Phyrexian Cleric {G}{W}, 2/2
// (Phyrexia: All Will Be One, #5479):
//
//	"Toxic 2
//	 Whenever another creature you control with toxic attacks, it gets
//	 +1/+1 until end of turn."
//
// One trigger per attacker. "With toxic" is part of the trigger event,
// read as the creature is declared as an attacker.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1c832491-6b78-49c6-baf2-63f17847f5f5",
		Name:            "Slaughter Singer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"toxic 2"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID != source.InstanceID && attackDeclaredByYou(ev, source.Controller) &&
					s82OneB05HasToxic(g, ev.CardID)
			}, "Slaughter Singer — the attacking creature gets +1/+1",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					id := item.Trigger.Event.CardID
					if !onBattlefield(g, id) {
						return nil
					}
					return BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Slaughter Singer — +1/+1"}.Apply(NewContext(g, item))
				}),
		},
	})
}
