package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scourge of the Throne — Creature — Dragon {4}{R}{R}, 5/5:
//
//	"Flying
//	 Dethrone (Whenever this creature attacks the player with the most
//	 life or tied for most life, put a +1/+1 counter on it.)
//	 Whenever this creature attacks for the first time each turn, if
//	 it's attacking the player with the most life or tied for most life,
//	 untap all attacking creatures. After this phase, there is an
//	 additional combat phase."
//
// Dethrone checks the defender once, as the attack is declared (CR
// 702.105a); an attack at a planeswalker or battle never triggers it,
// whoever leads (CR 508.3e). The second ability has an intervening "if"
// (CR 603.4): the creature must be attacking the life leader when it is
// declared and still be attacking that player as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "06496a05-7f1a-402c-9594-304cd82bfefd",
		Name:            "Scourge of the Throne",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.PreacherAttackedTheLifeLeader,
				"Scourge of the Throne — dethrone: put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.AttacksForTheFirstTimeThisTurn(ev, source, lki, g) && effects.PreacherAttackedTheLifeLeader(ev, source, lki, g)
			}, "Scourge of the Throne — untap all attacking creatures, additional combat",
				func(g *game.Game, item *game.StackItem) error {
					src, ok := g.LookupCardForEffect(item.SourceCardID)
					if !ok || src.AttackingTarget == uuid.Nil ||
						g.ClassifyAttackTargetForEffect(src.AttackingTarget) != game.AttackTargetPlayer ||
						!effects.PlayerHasTheMostLife(g, src.AttackingTarget) {
						return nil
					}
					return effects.Do(effects.UntapAttackingCreatures{}, effects.ExtraCombatAfterThisPhase())(g, item)
				}),
		},
	})
}
