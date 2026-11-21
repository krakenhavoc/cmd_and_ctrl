package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "0a364879-31bd-4760-833d-165cde1e96c6",
		Name:            "Yellowjacket, Heartless Marauder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.MshB2AnotherVillainEntered(ev, source, g)
			}, "Yellowjacket, Heartless Marauder — gets +1/+0 and gains lifelink until end of turn", yellowjacketBoost),
		},
	})
}

func yellowjacketBoost(g *game.Game, item *game.StackItem) error {
	if effects.SourceIsNewObject(g, item) || !effects.OnBattlefield(g, item.SourceCardID) {
		return nil
	}
	ctx := effects.NewContext(g, item)
	if err := (effects.BoostUntilEOT{Target: item.SourceCardID, Power: 1, Label: "Yellowjacket — +1/+0"}).Apply(ctx); err != nil {
		return err
	}
	return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"lifelink"}, Label: "Yellowjacket — lifelink"}.Apply(ctx)
}
