package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Amplifire — Creature — Elemental {2}{R}{R}, 1/1 (Ravnica Allegiance):
//
//	"At the beginning of your upkeep, reveal cards from the top of your library
//	 until you reveal a creature card. Until your next turn, this creature's base
//	 power becomes twice that card's power and its base toughness becomes twice that
//	 card's toughness. Put the revealed cards on the bottom of your library in a
//	 random order."
//
// The base values are set in layer 7b (CR 613.4b) until the controller's next turn,
// read from the revealed card as it is in the library (a "*" card counts what its
// own characteristic-defining ability says, CR 604.3). A library with no creature
// card changes nothing, and every revealed card still goes to the bottom, the
// creature card included, in a random order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c6861c50-d336-4352-a76d-ec22a86ec09e",
		Name:         "Amplifire",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Amplifire — reveal until a creature card, double its base power and toughness", amplifireUpkeep),
		},
	})
}

func amplifireUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	run, hit := effects.RevealUntil(ctx, ctx.Controller(), func(c game.Card) bool { return c.IsCreature() }, "Amplifire")
	if hit != uuid.Nil {
		if c, ok := g.LookupCardForEffect(hit); ok && effects.SourceIsStillThisPermanent(g, item) {
			if err := (effects.ScopedEffectFor{
				Target:   item.SourceCardID,
				Mods:     game.SetBasePTMods(2*g.PowerAnywhereForEffect(c), 2*g.ToughnessAnywhereForEffect(c)),
				Duration: effects.DurationUntilYourNextTurn(ctx, ctx.Controller()),
				Label:    "Amplifire — base power and toughness twice the revealed creature card's",
			}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return g.PutOnBottomInRandomOrderForEffect(ctx.Controller(), game.ZoneLibrary, run)
}
