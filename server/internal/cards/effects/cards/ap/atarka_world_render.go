package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Atarka, World Render — Legendary Creature — Dragon {5}{R}{G}, 6/4:
//
//	"Flying, trample
//	 Whenever a Dragon you control attacks, it gains double strike until
//	 end of turn."
//
// "A Dragon you control", the attacker itself included, so Atarka triggers
// off her own attack (b14DragonYouControlAttacked, Utvara Hellkite's
// condition). "It" is the attacking Dragon, read off the trigger's event
// at resolution; one that has left the battlefield gains nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c0b3bbce-977c-42a9-afcb-dabdfc717c97",
		Name:            "Atarka, World Render",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B14DragonYouControlAttacked(ev, source, g)
			},
			Key: "Atarka, World Render — the attacking Dragon gains double strike until end of turn",
			Effect: func(g *game.Game, item *game.StackItem) error {
				attacker := item.Trigger.Event.CardID
				if z := g.FindCardZoneForEffect(attacker); z == nil || z.Kind != game.ZoneBattlefield {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: attacker, Keywords: []string{"double strike"}, Label: "Atarka, World Render — double strike"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
