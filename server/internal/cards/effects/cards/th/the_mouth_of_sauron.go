package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Mouth of Sauron — Legendary Creature — Human Advisor {3}{U}{B}, 3/4
// (The Lord of the Rings: Tales of Middle-earth):
//
//	"When The Mouth of Sauron enters, target player mills three cards.
//	 Then amass Orcs X, where X is the number of instant and sorcery cards
//	 in that player's graveyard."
//
// Syphon Flesh-style sequencing: the mill finishes first (keyed, CR 608.2c),
// so the milled cards are counted; X is read as the amass begins, over the
// instants and sorceries in that player's graveyard then, not the number
// milled. The amass is Orcs X for the Mouth's controller whoever was
// targeted (CR 701.47); X may be 0, which still finds or makes the Army
// (CR 701.47b). A target that is no longer legal as the ability resolves
// makes it do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49e3396a-a125-4cca-b02a-4a96102bcb16",
		Name:         "The Mouth of Sauron",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("The Mouth of Sauron — target player mills three cards, then amass Orcs X",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						t, ok := effects.FirstLegalTarget(ctx)
						if !ok || t.Kind != game.TargetPlayer {
							return nil
						}
						return g.MillToZoneKeyedMovedForEffect(t.ID, 3, game.ZoneGraveyard, ctx.ItemTail(mouthOfSauronMilled))
					}),
				effects.TargetPlayer("target player")),
		},
	})
}

// mouthOfSauronMilled is "then amass Orcs X, where X is the number of instant
// and sorcery cards in that player's graveyard", after the mill has finished.
// The player is the ability's target, read from the carried item. An on-disk
// key: never renamed, never reused.
var mouthOfSauronMilled = effects.CardTail("then/the-mouth-of-sauron-milled", func(ctx *effects.Context, _ game.TailOutcome) error {
	t, ok := effects.FirstLegalTarget(ctx)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	p := ctx.Game.PlayerByIDForEffect(t.ID)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	x := 0
	for _, c := range p.Graveyard.Cards {
		if c.IsInstant() || c.IsSorcery() {
			x++
		}
	}
	return effects.Amass{Subtype: "Orc", N: x}.Apply(ctx)
})
