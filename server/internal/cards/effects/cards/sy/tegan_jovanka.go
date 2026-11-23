package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tegan Jovanka — Legendary Creature — Human {2}{W}, 2/2 (Doctor Who, #5948):
//
//	"Brave Heart — Whenever you attack, target attacking historic creature gets
//	 +1/+1 and gains indestructible until end of turn. (Artifacts, legendaries,
//	 and Sagas are historic.)
//	 Doctor's companion"
//
// One trigger per attack declaration however many creatures attack (CR 603.2c,
// as Adeline): the target is an attacking historic creature of any controller
// (CR 700.6), judged on announcement and again as it resolves; it gets +1/+1
// and indestructible until end of turn. Doctor's companion is a
// deck-construction rule; "Brave Heart" is an ability word (CR 207.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ac46f435-5b25-4b46-b89f-175bca81b9ad",
		Name:         "Tegan Jovanka",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.AttackDeclaredByYou(ev, source.Controller)
				}, "Tegan Jovanka — target attacking historic creature gets +1/+1 and gains indestructible until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						if err := (effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Tegan Jovanka — +1/+1"}).Apply(ctx); err != nil {
							return err
						}
						return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"indestructible"}, Label: "Tegan Jovanka — indestructible"}.Apply(ctx)
					})),
				effects.TargetCreature("target attacking historic creature", func(g *game.Game, _ uuid.UUID, c game.Card) bool {
					return c.AttackingTarget != uuid.Nil && effects.B09IsHistoric(c)
				})),
		},
	})
}
