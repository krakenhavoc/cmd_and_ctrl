package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "e07b6988-9ee7-4f20-8aa5-7dfa87ff507b",
		Name:         "Shield Mare",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), And(Creature(), OfColor("R")), "red creatures"),
		},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventBecomesTarget},
				func(ev game.Event, source *game.Card, c game.Characteristic, g *game.Game) bool {
					if ev.Kind == game.EventETB {
						return ev.CardID == source.InstanceID
					}
					return b24SelfTargetedByAnOpponentOnce(ev, source, c, g)
				},
				"Shield Mare — you gain 3 life", Do(GainLife{Amount: 3})),
		},
	})
}
