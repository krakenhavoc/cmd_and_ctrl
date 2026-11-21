package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Istvan, Butcher of Eln — Legendary Creature — Human {1}{B}{B}{B}, 1/3
// (Mystery Booster Commander, #5948):
//
//	"Istvan attacks each combat if able.
//	 Prevent all damage that would be dealt to Istvan by creatures.
//	 Whenever a creature an opponent controls dies, put two +1/+1
//	 counters on Istvan."
//
// The attack requirement is AttacksEachCombat (CR 508.1d). The prevention is
// a standing CR 615 replacement like Marble Priest's, for every damage
// event, combat or not, whose source is a creature when it is dealt. The
// trigger is Vincent Valentine's "an opponent's creature dies" condition,
// whose dead creature is read as it last stood (CR 603.10a), with
// CounterOnThis so it carries a footprint.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5e277de5-793a-4364-923a-a55b53c0fa85",
		Name:         "Istvan, Butcher of Eln",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.AttacksEachCombat()},
		Replacements: []game.ReplacementEffect{{
			Watches:    []game.EventKind{game.EventDealDamage},
			Prevention: true, // CR 615.1a
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventDamage || ev.DamageTarget != src.InstanceID {
					return false
				}
				dealer, ok := g.LookupCardForEffect(ev.DamageSource)
				return ok && dealer.IsCreature()
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.Cancel()
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Istvan, Butcher of Eln — prevent all damage that would be dealt to it by creatures",
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.B33OpponentsCreatureDiedWhen,
				"Istvan, Butcher of Eln — put two +1/+1 counters on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 2})),
		},
	})
}
