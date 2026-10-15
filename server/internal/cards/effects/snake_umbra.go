package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Snake Umbra — Enchantment — Aura {2}{G} (CR 702.89, #3654):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 and has "Whenever this creature deals
//	 damage to an opponent, you may draw a card."
//	 Umbra armor (If enchanted creature would be destroyed, instead remove
//	 all damage from it and destroy this Aura.)"
//
// The quoted trigger is a granted bundle (ADR 0093), so "this creature"
// is the enchanted creature and "you" is its controller; "an opponent"
// is measured against that controller. Any damage counts, as Curiosity's
// does. The "may" is an optional prompt. No simplification.
func init() {
	Register(Spec{
		OracleID:     snakeUmbraOracleID,
		Name:         "Snake Umbra",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []AbilityGrant{{
			Key: snakeUmbraDraw,
			Triggered: []game.TriggeredAbility{
				Optional(On(game.EventDealDamage, snakeUmbraDealtDamageToOpponent,
					"Snake Umbra — draw a card", Do(DrawCards{N: 1})),
					"Snake Umbra — draw a card?"),
			},
			Text: "Whenever this creature deals damage to an opponent, you may draw a card.",
		}},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantAbilitiesToAttached(snakeUmbraDraw),
		},
		Replacements: []game.ReplacementEffect{UmbraArmor()},
	})
}

// snakeUmbraDealtDamageToOpponent reads the granted trigger's condition
// for the creature that carries it: that creature dealt damage to a
// player who is not its controller.
func snakeUmbraDealtDamageToOpponent(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || ev.Source != source.InstanceID {
		return false
	}
	p := g.PlayerByIDForEffect(ev.Target)
	return p != nil && p.ID != source.Controller
}

const (
	snakeUmbraOracleID = "70ed2052-f03c-461d-b32a-8835c97c559f"
	snakeUmbraDraw     = "snake-umbra/draw"
)
