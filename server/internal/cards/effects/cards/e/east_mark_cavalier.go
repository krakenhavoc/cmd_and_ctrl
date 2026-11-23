package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const eastMarkCavalierLabel = "East-Mark Cavalier — destroy the Goblin or Orc it dealt damage to"

// East-Mark Cavalier — Creature — Human Knight {1}{W}, 2/2 (The Lord of the
// Rings: Tales of Middle-earth Commander):
//
//	"Vigilance
//	 Whenever this creature deals damage to a Goblin or Orc, destroy that
//	 creature."
//
// Spider-Slayer's damage watch with two tribes. The damaged creature is
// named as the object it is when the trigger is built, so one that died to
// the damage or came back since is not destroyed (CR 400.7); a Goblin or
// Orc is read through the source's text. Vigilance is printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "884d08bf-7e97-40cc-864e-445ae0bd1646",
		Name:            "East-Mark Cavalier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || ev.Source != source.InstanceID {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.Target)
				if !ok || !c.IsCreature() || !effects.OnBattlefield(g, ev.Target) {
					return false
				}
				return c.HasSubtype(source.TextSubtype("Goblin")) || c.HasSubtype(source.TextSubtype("Orc"))
			},
			Key: eastMarkCavalierLabel,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, eastMarkCavalierLabel)
				item.Params.Object = effects.DamagedCreatureRef(g, ev.Target)
				return item
			},
			Effect: effects.DestroyDamagedCreature,
		}},
	})
}
