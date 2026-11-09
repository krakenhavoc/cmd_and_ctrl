package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dream Spoilers — Creature — Faerie Warlock {3}{B}, 3/3 (Wilds of
// Eldraine, #4471):
//
//	"Flying
//	 Whenever you cast a spell during an opponent's turn, up to one target
//	 creature an opponent controls gets -1/-1 until end of turn."
//
// The trigger watches EventCast for the controller's own spells and reads
// the turn as the cast happens: it is an opponent's turn when the active
// player is not the controller. The -1/-1 is applied to the one chosen
// target still legal as the trigger resolves; with no target chosen, or
// none left, nothing happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "07c3e91c-3593-4797-b7be-ef4fb2913cef",
		Name:            "Dream Spoilers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventCast,
				func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					return YouCast(nil)(ev, source, lki, g) && !IsYourTurn(g, source.Controller)
				},
				"Dream Spoilers — up to one target creature an opponent controls gets -1/-1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return BoostUntilEOT{Target: id, Power: -1, Toughness: -1,
						Label: "Dream Spoilers — -1/-1 until end of turn"}.Apply(ctx)
				}),
				UpToOneTargetCreature("up to one target creature an opponent controls", OpponentControls())),
		},
	})
}
