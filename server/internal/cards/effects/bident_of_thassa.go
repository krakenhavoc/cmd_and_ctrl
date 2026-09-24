package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bident of Thassa — Legendary Enchantment Artifact for {2}{U}{U}:
//
//	"Whenever a creature you control deals combat damage to a player,
//	you may draw a card.
//	{1}{U}, {T}: Creatures your opponents control attack this turn if
//	able."
//
// S19 sub-PR 7: the first combat-damage trigger. EventDealDamage
// carries Combat=true for damage dealt in the combat damage step;
// combatDamageToPlayerBy checks the source is a creature the
// Bident's controller controls and the target is a player. One
// event per attacker per player, so three unblocked attackers queue
// three "you may draw" prompts — as in paper.
//
// Re-audited for #1565: the activated ability's cost ({1}{U}, {T})
// is ordinary machinery (ManaCost + TapCost), so that is not what
// blocks the second line. What is missing is the EFFECT: "creatures
// your opponents control attack this turn if able" is a forced-attack
// restriction, and the engine's restriction vocabulary
// (game.Restriction — CantAttack, CantBlock, …) has no MUST-attack
// bit. Goad is the closest existing case and its own doc says the
// same thing: SetGoaded is "sandbox-only: the must-attack-and-
// not-the-goader constraint is not enforced". Until a forced-attack
// primitive exists (for goad and for this), the second line stays out.
func init() {
	Register(Spec{
		OracleID:     "e1afaef7-9fa3-4662-a95f-adfb0da9fd11",
		Name:         "Bident of Thassa",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The \"{1}{U}, {T}: Creatures your opponents control attack this turn if able\" ability isn't implemented — only the combat-damage draw trigger works."},
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g)
			}, "Bident of Thassa — draw a card", Do(DrawCards{N: 1})), "Bident of Thassa — draw a card?"),
		},
	})
}
