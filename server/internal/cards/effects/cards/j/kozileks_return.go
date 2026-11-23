package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kozilek's Return — Instant {2}{R} (Modern Horizons 3 Commander, #5948):
//
//	"Devoid (This card has no color.)
//	 Kozilek's Return deals 2 damage to each creature.
//	 Whenever you cast an Eldrazi creature spell with mana value 7 or
//	 greater, you may exile this card from your graveyard. If you do, this
//	 card deals 5 damage to each creature."
//
// The spell is Pyroclasm's body. The second ability is a trigger that
// watches from the GRAVEYARD (Bloodghast's shape, CR 108.4: "you" is the
// owner while the card sits there), on the cast of an Eldrazi creature spell
// whose mana value is 7 or greater. The "you may" is asked as it resolves
// (CR 603.5); a card that has left the graveyard by then is a new object and
// does nothing (CR 400.7). Only when the exile happened does the card deal 5
// damage to each creature, as the card it was in the graveyard (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6caafe5a-884d-49e8-a62a-7bf609f0ce53",
		Name:            "Kozilek's Return",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatching(ctx, effects.Creature(), 2)
		},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.InGraveyard(effects.On(game.EventCast,
				effects.YouCast(effects.And(effects.Creature(), effects.HasSubtype("Eldrazi"), effects.ManaValueGE(7))),
				"Kozilek's Return — exile it from your graveyard; if you do, it deals 5 damage to each creature",
				func(g *game.Game, item *game.StackItem) error {
					if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
						return nil
					}
					ctx := effects.NewContext(g, item)
					if err := (effects.ExileTarget{Target: item.SourceCardID}).Apply(ctx); err != nil {
						return err
					}
					return effects.DamageEachMatching(ctx, effects.Creature(), 5)
				})), "Kozilek's Return — exile it from your graveyard to deal 5 damage to each creature?"),
		},
	})
}
