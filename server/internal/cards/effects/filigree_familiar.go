package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Filigree Familiar — 2/2 Artifact Creature — Fox for {3}:
//
//	"When Filigree Familiar enters the battlefield, you gain 2 life.
//	When Filigree Familiar dies, draw a card.
//	{2}, Sacrifice Filigree Familiar: Add one mana of any color."
//
// S19 sub-PR 4 implements only the dies half ("draw a card") — a
// mandatory LTB trigger, the simplest counterpart to Solemn's
// optional dies-draw. The ETB-lifegain and sacrifice-for-mana
// halves are deferred (ETB-lifegain to a later batch, the
// activated sac-mana ability to S15's cost model follow-ups);
// splitting a multi-ability card across sprints matches how
// Solemn's two halves landed separately.
//
// cardDied gates the trigger to graveyard-only: a bounced or
// exiled Familiar does not draw.
func init() {
	Register(Spec{
		OracleID: "b544f690-e4bf-4a5b-984d-9256518fd574",
		Name:     "Filigree Familiar",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				ctx := NewContext(g, nil)
				_ = DrawCards{Player: source.Controller, N: 1}.Apply(ctx)
				return nil
			},
		}},
	})
}
