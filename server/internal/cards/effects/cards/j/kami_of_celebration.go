package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kami of Celebration — Creature — Spirit {4}{R}, 3/3 (Kamigawa: Neon
// Dynasty):
//
//	"Whenever a modified creature you control attacks, exile the top card of
//	 your library. You may play that card this turn. (Equipment, Auras you
//	 control, and counters are modifications.)
//	 Whenever you cast a spell from exile, put a +1/+1 counter on target
//	 creature you control."
//
// The first ability triggers once for each modified attacker (CR 700.9,
// b07IsModified, read as the attack is declared) and is Kellan's impulse
// exile. The second is Fire Lord Zuko's read of a cast from exile
// (EventCast stamps the origin zone), so the card the first ability exiled,
// or any other spell cast from exile, grows a creature of yours.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ecfabd06-c059-4f3d-9e5f-587364033cac",
		Name:         "Kami of Celebration",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.S82NeoB11ModifiedCreatureYouControlAttacked, "Kami of Celebration — exile the top card of your library; you may play that card this turn",
				func(g *game.Game, item *game.StackItem) error {
					_, err := effects.B12ImpulseExileForTurn(g, item, 1)
					return err
				}),
			effects.Targeting(effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller && ev.OldZone == game.ZoneExile
			}, "Kami of Celebration — put a +1/+1 counter on target creature you control", effects.B36CounterOnChosenAnimal),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
