package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Yellowjacket, Heartless Marauder — {1}{B} Legendary Creature — Human
// Rogue Villain 1/2 (Marvel Super Heroes, #4455):
//
//	"Flying
//	 Whenever another Villain you control enters, Yellowjacket gets +1/+0
//	 and gains lifelink until end of turn."
//
// Baron Strucker's entering-Villain test (a changeling counts). Each Villain
// is its own trigger, so two entering together give +2/+0. A Yellowjacket
// who has left the battlefield, or come back as a new object (CR 400.7),
// gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0a364879-31bd-4760-833d-165cde1e96c6",
		Name:            "Yellowjacket, Heartless Marauder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return mshB2AnotherVillainEntered(ev, source, g)
			}, "Yellowjacket, Heartless Marauder — gets +1/+0 and gains lifelink until end of turn", yellowjacketBoost),
		},
	})
}

func yellowjacketBoost(g *game.Game, item *game.StackItem) error {
	if sourceIsNewObject(g, item) || !onBattlefield(g, item.SourceCardID) {
		return nil
	}
	ctx := NewContext(g, item)
	if err := (BoostUntilEOT{Target: item.SourceCardID, Power: 1, Label: "Yellowjacket — +1/+0"}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"lifelink"}, Label: "Yellowjacket — lifelink"}.Apply(ctx)
}
