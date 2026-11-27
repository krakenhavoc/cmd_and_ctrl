package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cathedral Membrane — Artifact Creature — Phyrexian Wall {1}{W/P}, 0/3 (New
// Phyrexia, #5575):
//
//	"({W/P} can be paid with either {W} or 2 life.)
//	 Defender
//	 When this creature dies during combat, it deals 6 damage to each creature
//	 it blocked this combat."
//
// "During combat" is read as the Membrane dies, from the combat step the game is
// in (the intervening condition, CR 603.4, checked as it triggers). The
// creature it blocked is the one its last-known combat state names (the event's
// BlockingTarget, CR 603.10a: leaving the battlefield takes it out of combat),
// and takes 6 damage from the Membrane as it last existed if it is still on the
// battlefield as the trigger resolves. A Membrane that died while blocking
// nothing deals none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0297eeb2-47ec-4f9a-b1ad-4e7286994f00",
		Name:            "Cathedral Membrane",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				return effects.ThisDied(ev, source, ch, g) && effects.InCombatStep(g.Turn.Step) && ev.BlockingTarget != uuid.Nil
			}, "Cathedral Membrane — it deals 6 damage to each creature it blocked this combat",
				func(g *game.Game, item *game.StackItem) error {
					blocked := item.Trigger.Event.BlockingTarget
					if blocked == uuid.Nil || !effects.OnBattlefield(g, blocked) {
						return nil
					}
					if c, ok := g.LookupCardForEffect(blocked); !ok || !c.IsCreature() {
						return nil
					}
					return effects.DealDamage{Source: item.SourceCardID, Target: blocked, Amount: 6}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
