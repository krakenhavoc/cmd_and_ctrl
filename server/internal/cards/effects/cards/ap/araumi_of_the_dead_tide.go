package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Araumi of the Dead Tide — Legendary Creature — Merfolk Wizard {1}{U}{B}, 1/4
// (Commander Legends):
//
//	"{T}, Exile cards from your graveyard equal to the number of opponents you
//	 have: Target creature card in your graveyard gains encore until end of
//	 turn. The encore cost is equal to its mana cost. (Exile the creature card
//	 and pay its mana cost: For each opponent, create a token copy that attacks
//	 that opponent this turn if able. They gain haste. Sacrifice them at the
//	 beginning of the next end step. Activate only as a sorcery.)"
//
// Cursecloth Wrappings' grant (ADR 0219) of the shared encore row with its mana
// component read off the card that has it (ItsManaCost), given to the targeted
// graveyard card until end of turn. The exile cost is read as the ability is
// announced: ExileCost.CountOpponents is the number of players still in the
// game other than the controller, so with no opponent left it exiles nothing.
// The cards exiled can't be the target (the cost is paid after targets are
// chosen, CR 601.2), so a target exiled to pay fails to be legal and the grant
// is not made, as printed.
//
// No simplification.
func init() {
	encore := effects.Encore("")
	encore.Label = "Encore — its mana cost, Exile this card from your graveyard: For each opponent, " +
		"create a token that's a copy of it that attacks that opponent this turn if able. They gain haste. " +
		"Sacrifice them at the beginning of the next end step. Activate only as a sorcery."
	encore.Cost = effects.Plus(effects.ItsManaCost(), effects.ExileThis())
	effects.Register(effects.Spec{
		OracleID:     "1a62ba93-153c-4bed-9f6a-ff914df360c1",
		Name:         "Araumi of the Dead Tide",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       araumiEncore,
			Zone:      game.ZoneGraveyard,
			Activated: []effects.ActivatedAbility{encore},
			Text:      "Encore. The encore cost is equal to its mana cost.",
		}},
		Activated: []effects.ActivatedAbility{{
			Label: "{T}, Exile cards from your graveyard equal to the number of opponents you have: Target creature card in your graveyard gains encore until end of turn. The encore cost is equal to its mana cost.",
			Cost: effects.Plus(effects.TapCost(),
				effects.S82OldB083ExileCardsEqualToOpponents("cards from your graveyard equal to the number of opponents you have")),
			Targets: effects.TargetCardInGraveyard("target creature card in your graveyard", effects.Creature(), effects.YouOwn()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if c, ok := g.LookupCardForEffect(t.ID); !ok || c.ManaCost == "" {
						return nil
					}
					g.GrantAbilityToGraveyardCardUntilEndOfTurnForEffect(t.ID, ctx.Source(), ctx.Controller(), araumiEncore)
				}
				return nil
			},
		}},
	})
}

// araumiEncore is the bundle key of the encore Araumi gives. On disk.
const araumiEncore = "araumi-of-the-dead-tide/encore"
