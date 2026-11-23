package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aegar, the Freezing Flame — Legendary Creature — Giant Wizard {1}{U}{R},
// 3/3:
//
//	"Whenever a creature or planeswalker an opponent controls is dealt
//	 excess damage, if a Giant, Wizard, or spell you controlled dealt
//	 damage to it this turn, draw a card."
//
// The trigger is one damage event: its ExcessDamage (CR 120.10, judged as
// the damage lands, past lethal for a creature and past loyalty for a
// planeswalker) is above zero and the permanent is an opponent's. The
// intervening "if" (CR 603.4) is read off the turn's damage record, which
// keeps the source as it dealt the damage: a spell on the stack, or a
// permanent that was a Giant or a Wizard (every creature type counts), under
// your control then. It is checked as the damage lands and again as the
// trigger resolves, by which time the permanent may be dead (last-known
// object, CR 608.2h). Aegar itself is a Giant Wizard, so its own damage
// counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "592d847e-e3f3-425d-89f9-01ab533ad581",
		Name:         "Aegar, the Freezing Flame",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondDamagedByGiantWizardOrSpell{}, game.TriggeredAbility{
				Watches: []game.EventKind{game.EventDealDamage},
				Key:     "Aegar, the Freezing Flame — draw a card (an opponent's creature or planeswalker was dealt excess damage)",
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.Kind != game.EventDealDamage || ev.ExcessDamage <= 0 {
						return false
					}
					victim, ok := g.LookupCardForEffect(ev.Target)
					if !ok || victim.Controller == source.Controller || !(victim.IsCreature() || victim.IsPlaneswalker()) {
						return false
					}
					return g.PermanentDealtDamageThisTurnByKindForEffect(ev.Target, source.Controller, true, "Giant", "Wizard")
				},
				Effect: effects.Do(effects.DrawCards{N: 1}),
			}),
		},
	})
}
