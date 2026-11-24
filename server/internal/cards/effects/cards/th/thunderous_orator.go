package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// thunderousOratorKeywords are the eight abilities the Orator may gain, in
// printed order.
var thunderousOratorKeywords = []string{
	"flying", "first strike", "double strike", "deathtouch", "indestructible", "lifelink", "menace", "trample",
}

// Thunderous Orator — {1}{W} Creature — Kor Wizard 2/2 (Strixhaven: School of
// Mages, #5502):
//
//	"Vigilance
//	 Whenever this creature attacks, it gains flying until end of turn if you
//	 control a creature with flying. The same is true for first strike, double
//	 strike, deathtouch, indestructible, lifelink, menace, and trample."
//
// Odric, Lunarch Marshal's test on one creature: the eight checks are made
// together as the ability resolves, against the keywords your creatures have
// then through the layered view, the Orator included. A keyword the Orator
// gains from this ability does not make another one shared (CR 608.2c). "It"
// is the creature that attacked (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7716ad72-9e96-4915-bc1e-e12b4d0f6f9c",
		Name:            "Thunderous Orator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Thunderous Orator — it gains each keyword a creature you control has until end of turn",
				effects.ThisStillHere(func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					creatures := effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.ControlledBy(item.Controller)))
					var shared []string
					for _, kw := range thunderousOratorKeywords {
						for i := range creatures {
							if game.HasKeyword(&creatures[i], kw) {
								shared = append(shared, kw)
								break
							}
						}
					}
					if len(shared) == 0 {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: shared,
						Label: "Thunderous Orator — shared keywords until end of turn"}.Apply(ctx)
				})),
		},
	})
}
