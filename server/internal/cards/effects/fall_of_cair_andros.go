package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fall of Cair Andros — Enchantment {2}{R}:
//
//	"Whenever a creature an opponent controls is dealt excess
//	 noncombat damage, amass Orcs X, where X is that excess damage.
//	 {7}{R}: This enchantment deals 7 damage to target creature."
//
// Magmatic Galleon's excess-damage reading (excess_damage.go) turned
// into a number: the trigger is harvested as the damage lands, so the
// excess is computed there and carried on the stack item's Params
// (Kaervek's shape), and the effect amasses that many. One trigger per
// creature dealt excess damage, not one per batch, because the text
// names each creature ("a creature ... is dealt"). The activated
// ability is a plain targeted 7 damage.
//
// Sandbox simplification, declared: damage from a source with
// deathtouch is not treated as excess beyond one point (CR 120.4a), so
// the X read for it can be lower than printed, never higher.
func init() {
	const label = "Fall of Cair Andros — amass Orcs X, where X is that excess damage"
	Register(Spec{
		OracleID:     "33c0a8c0-d1d9-4b03-869d-65dab8ace3df",
		Name:         "Fall of Cair Andros",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Noncombat damage from a source with deathtouch only counts as excess past the creature's own lethal damage, so you may amass fewer Orcs than printed."},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return excessNoncombatDamageToOpponentCreature(ev, source, g) > 0
			},
			Key: label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				item.Params.Amount = excessNoncombatDamageToOpponentCreature(ev, source, g)
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Amass{Subtype: "Orc", N: item.Params.Amount}.Apply(NewContext(g, item))
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{7}{R}: This enchantment deals 7 damage to target creature.",
			Cost:    ManaCost("{7}{R}"),
			Targets: TargetCreature("target creature"),
			Purpose: ForTargets(DamageToTarget(0, 7)),
			Effect:  DealDamageToTheTarget(7),
		}},
	})
}
