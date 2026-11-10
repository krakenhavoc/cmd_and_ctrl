package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rot Farm Mortipede — Creature — Insect {3}{B}, 3/4 (Murders at Karlov
// Manor, #4469):
//
//	"Whenever one or more creature cards leave your graveyard, this
//	 creature gets +1/+0 and gains menace and lifelink until end of turn."
//
// Cyan, Vengeful Samurai's trigger (left_graveyard.go): every route a card
// leaves your graveyard by counts, and cards leaving together are one
// batch (OncePerBatch, CR 603.2c, per the ruling). The Mortipede gets the
// pump and both keywords from one trigger; a Mortipede that has left the
// battlefield gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1f4d3a96-bc30-4362-9c1f-40bf79f78e78",
		Name:         "Rot Farm Mortipede",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(OnAny(leftGraveyardEvents, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return creatureCardLeftYourGraveyard(ev, source, g)
			}, "Rot Farm Mortipede — gets +1/+0 and gains menace and lifelink until end of turn",
				Do(Instruction(func(ctx *Context) error {
					if err := (BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Rot Farm Mortipede — +1/+0"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Target: ctx.Source(), Keywords: []string{"menace", "lifelink"},
						Label: "Rot Farm Mortipede — menace and lifelink"}.Apply(ctx)
				})))),
		},
	})
}
