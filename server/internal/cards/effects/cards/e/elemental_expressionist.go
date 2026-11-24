package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elemental Expressionist — Creature — Orc Wizard {U/R}{U/R}{U/R}{U/R}, 4/4
// (Strixhaven: School of Mages, #5502):
//
//	"Magecraft — Whenever you cast or copy an instant or sorcery spell,
//	 choose target creature you control. Until end of turn, it gains "If this
//	 creature would leave the battlefield, exile it instead of putting it
//	 anywhere else" and "When this creature is put into exile, create a 4/4
//	 blue and red Elemental creature token." (Each instance of that ability
//	 triggers separately.)"
//
// Two effects on the chosen creature until end of turn (CR 611.2c): the
// exile-instead replacement (Whip of Erebos's ModExileInsteadOfLeaving, the
// creature's own "would leave" clause, CR 614.6) and a granted triggered
// ability (ADR 0093, Feign Death's shape) that fires from last-known
// information when the creature is put into exile from the battlefield
// (CR 603.10a). The trigger is the creature's, so its controller gets the
// 4/4 blue and red Elemental token. Granting it twice gives two instances,
// each of which triggers separately, as the printed reminder says.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dfc45c58-821c-42f3-ab48-05b912748bdf",
		Name:         "Elemental Expressionist",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Grants: []effects.AbilityGrant{{
			Key: elementalExpressionistToken,
			Triggered: []game.TriggeredAbility{
				effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Kind == game.EventLTB && ev.CardID == source.InstanceID && ev.NewZone == game.ZoneExile
				}, "Elemental Expressionist — create a 4/4 blue and red Elemental creature token",
					effects.Do(effects.CreateToken{Template: effects.TokenCard("4/4 blue and red Elemental"), N: 1})),
			},
			Text: "When this creature is put into exile, create a 4/4 blue and red Elemental creature token.",
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Magecraft("Elemental Expressionist — until end of turn, target creature you control is exiled instead of leaving and makes a 4/4 Elemental when exiled (magecraft)",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind != game.TargetCard {
								continue
							}
							if err := effects.UntilEndOfTurn(ctx, t.ID, nil, "Elemental Expressionist — exile it instead of leaving the battlefield",
								game.Mod{Kind: game.ModExileInsteadOfLeaving}); err != nil {
								return err
							}
							return effects.GrantAbilitiesFor{
								Target: t.ID,
								Keys:   []string{elementalExpressionistToken},
								Label:  "Elemental Expressionist — when it is put into exile, create a 4/4 Elemental",
							}.Apply(ctx)
						}
						return nil
					}),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}

// elementalExpressionistToken is the granted bundle's key. An on-disk
// identity: never renamed, never reused.
const elementalExpressionistToken = "elemental-expressionist/exile-token"
