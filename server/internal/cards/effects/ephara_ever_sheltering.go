package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ephara, Ever-Sheltering — Legendary Enchantment Creature — God 4/4, the
// BACK FACE of Invasion of Theros (face 1):
//
//	"Ephara has lifelink and indestructible as long as you control at
//	 least three other enchantments.
//	 Whenever another enchantment you control enters, draw a card."
//
// Registered under "<oracle_id>#1". The two keywords are two layer 6
// grants sharing one condition: three OTHER enchantments, so Ephara
// itself (an enchantment creature) is never counted.
//
// No simplification.
func init() {
	grant := func(keyword string) game.StaticAbility {
		ab := KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
			return selfCreature(target, g, source) && otherEnchantmentsYouControl(g, source) >= 3
		}, keyword)
		ab.Label = "Has " + keyword + " as long as you control at least three other enchantments."
		return ab
	}
	Register(Spec{
		OracleID:     invasionOfTherosOracleID + "#1",
		Name:         "Ephara, Ever-Sheltering",
		Purpose:      game.Purpose{Draws: 1},
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{grant("lifelink"), grant("indestructible")},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.CardID == source.InstanceID {
					return false
				}
				c, ok := enteredUnderYourControl(ev, source, g, true)
				return ok && c.IsEnchantment()
			}, "Ephara, Ever-Sheltering — draw a card", Do(DrawCards{N: 1})),
		},
	})
}

// otherEnchantmentsYouControl counts the enchantments source's controller
// controls other than source, read through the layered view.
func otherEnchantmentsYouControl(g *game.Game, source *game.Card) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != source.InstanceID && c.Controller == source.Controller && c.IsEnchantment() {
			n++
		}
	}
	return n
}
