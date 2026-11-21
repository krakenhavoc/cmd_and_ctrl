package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shield Mare — Creature — Horse {1}{W}{W}, 2/3 (Foundations Jumpstart):
//
//	"This creature can't be blocked by red creatures.
//	 When this creature enters or becomes the target of a spell or ability
//	 an opponent controls, you gain 3 life."
//
// The evasion is a block rule. The second sentence is one printed trigger
// with two conditions: the engine emits EventBecomesTarget once per target
// slot (CR 115.3), so b24SelfTargetedByAnOpponentOnce lets only the first
// slot naming Shield Mare on one spell or ability trigger it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e07b6988-9ee7-4f20-8aa5-7dfa87ff507b",
		Name:         "Shield Mare",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.And(effects.Creature(), effects.OfColor("R")), "red creatures"),
		},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventETB, game.EventBecomesTarget},
				func(ev game.Event, source *game.Card, c game.Characteristic, g *game.Game) bool {
					if ev.Kind == game.EventETB {
						return ev.CardID == source.InstanceID
					}
					return effects.B24SelfTargetedByAnOpponentOnce(ev, source, c, g)
				},
				"Shield Mare — you gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
		},
	})
}
