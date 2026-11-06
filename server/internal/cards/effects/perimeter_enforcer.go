package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Perimeter Enforcer — Creature — Human Detective {1}{W}, 1/1 (Murders at
// Karlov Manor, #4469):
//
//	"Flying, lifelink
//	 Whenever another Detective you control enters and whenever a
//	 Detective you control is turned face up, this creature gets +1/+1
//	 until end of turn."
//
// One ability on two event kinds (ADetectiveEnteredOrWasTurnedFaceUp,
// its own entry excluded). The pump is pinned to the Enforcer; one that
// left the battlefield before the trigger resolved gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "be8d3788-709b-4d10-861c-a6f27d4b63a6",
		Name:            "Perimeter Enforcer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventTurnedFaceUp}, ADetectiveEnteredOrWasTurnedFaceUp(false),
				"Perimeter Enforcer — this creature gets +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
						return nil
					}
					return BoostUntilEOT{
						Target:    item.SourceCardID,
						Power:     1,
						Toughness: 1,
						Label:     "Perimeter Enforcer — +1/+1 until end of turn",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
