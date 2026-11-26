package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Agadeem Occultist — Creature — Human Shaman Ally {2}{B}, 0/2 (Worldwake, S82 OLD
// b487):
//
//	"{T}: Put target creature card from an opponent's graveyard onto the
//	 battlefield under your control if its mana value is less than or equal to
//	 the number of Allies you control."
//
// Puppeteer Clique's reanimation, gated by the condition the card prints: the
// creature card is a legal target whatever its mana value (CR 608.2b), and as the
// ability resolves it is put onto the battlefield only when that value is at most
// the number of Allies you control then, the Occultist included while it is
// still there. A card whose cost the engine cannot read is never put.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9c462cd8-36f9-4e39-adf3-310ad4d76051",
		Name:         "Agadeem Occultist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Put target creature card from an opponent's graveyard onto the battlefield under your control if its mana value is less than or equal to the number of Allies you control.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCardInGraveyard("target creature card from an opponent's graveyard", effects.Creature(), effects.Not(effects.YouOwn())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				legal := ctx.LegalTargets()
				if len(legal) == 0 {
					return nil
				}
				card, ok := g.LookupCardForEffect(legal[0].ID)
				if !ok {
					return nil
				}
				if mv := card.ManaValue(); mv > effects.ZenB2AlliesYouControl(ctx) {
					return nil
				}
				return effects.ReturnFromGraveyard{Target: legal[0].ID, Dest: game.ZoneBattlefield, Controller: item.Controller}.Apply(ctx)
			},
		}},
	})
}
