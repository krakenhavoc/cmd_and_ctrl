package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mardu Ascendancy — Enchantment {R}{W}{B} (Khans of Tarkir):
//
//	"Whenever a nontoken creature you control attacks, create a 1/1 red Goblin
//	 creature token that's tapped and attacking.
//	 Sacrifice this enchantment: Creatures you control get +0/+3 until end of
//	 turn."
//
// The trigger fires once per nontoken attacker (Beastmaster Ascension's
// per-creature attack watch), and the Goblin is Battle Cry Goblin's: it enters
// tapped and attacking the player the creature attacked, and was never declared
// as an attacker (CR 508.4), so it does not trigger this again. The sacrifice
// is a cost; the creatures you control are those on the battlefield as it
// resolves (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aa1de5ce-ddd4-4c1f-8c52-0311f4419e7b",
		Name:         "Mardu Ascendancy",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.AttackDeclaredByYou(ev, source.Controller) {
					return false
				}
				attacker, ok := g.LookupCardForEffect(ev.CardID)
				return ok && !attacker.IsToken()
			}, "Mardu Ascendancy — create a 1/1 red Goblin creature token that's tapped and attacking",
				effects.OldB351TappedAttackingTokenLikeTheAttacker("1/1 red Goblin")),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this enchantment: Creatures you control get +0/+3 until end of turn.",
			Purpose: game.Purpose{TeamPump: &game.Pump{Toughness: 3}, Answers: game.AnswerPump},
			Cost:    effects.SacrificeThis(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Toughness: 3, Label: "Mardu Ascendancy — +0/+3"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
