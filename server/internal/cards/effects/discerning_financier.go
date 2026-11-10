package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Discerning Financier — Creature — Human Noble {2}{W}, 2/3 (Wilds of
// Eldraine, #4471):
//
//	"At the beginning of your upkeep, if an opponent controls more lands
//	 than you, create a Treasure token.
//	 {2}{W}: Choose another player. That player gains control of target
//	 Treasure you control. You draw a card."
//
// The upkeep trigger's intervening "if" (CR 603.4) is Loyal Warhound's
// count, checked as the upkeep begins and again as the trigger resolves.
//
// The activated ability targets the Treasure as it is activated and
// chooses the player as it resolves (it is "choose", not "target", so
// nothing can respond to the pick): ChoosePlayer over every seat but you.
// The chosen player gains control for good (CR 611.2a, no stated
// duration). If the Treasure is no longer a legal target, the ability's
// only target is illegal and it does not resolve (CR 608.2b): no control
// changes and no card is drawn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "734a5186-caa6-4811-b6d7-91b7b32fc403",
		Name:         "Discerning Financier",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep,
				AllOf(ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return anOpponentHasMoreLands(g, source.Controller)
				}),
				"Discerning Financier — create a Treasure token",
				func(g *game.Game, item *game.StackItem) error {
					if !anOpponentHasMoreLands(g, item.Controller) {
						return nil
					}
					return CreateToken{Template: TreasureToken(), N: 1}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{W}: Choose another player. That player gains control of target Treasure you control. You draw a card.",
			Cost:    ManaCost("{2}{W}"),
			Targets: TargetPermanent("target Treasure you control", Subtype("Treasure"), YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return ChoosePlayer{
					Except:   []uuid.UUID{ctx.Controller()},
					Question: "Discerning Financier — choose another player to gain control of the Treasure",
					ThenKey:  discerningFinancierChosen,
				}.Apply(ctx)
			},
		}},
	})
}

// discerningFinancierChosen is the ability's second half once the player is
// named: they gain control of the Treasure, then you draw. With nobody to
// name (every other player has left) nothing changes hands and you still
// draw. An on-disk key: never renamed, never reused.
var discerningFinancierChosen = ChoosePlayerThen("option-pick/discerning-financier-chosen", func(ctx *Context, chosen uuid.UUID, _ []uuid.UUID) error {
	if chosen != uuid.Nil {
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			if err := (GainControl{Target: t.ID, Controller: chosen, Duration: game.IndefiniteDuration(), Label: "Discerning Financier — gain control"}).Apply(ctx); err != nil {
				return err
			}
			break
		}
	}
	return DrawCards{N: 1}.Apply(ctx)
})
