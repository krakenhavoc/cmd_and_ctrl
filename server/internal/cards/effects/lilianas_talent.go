package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Liliana's Talent — Enchantment — Aura {B}{B}:
//
//	"Enchant planeswalker
//	 Enchanted planeswalker has "[−8]: Put all creature cards from all
//	 graveyards onto the battlefield under your control."
//	 Whenever a creature deals damage to enchanted planeswalker, destroy
//	 that creature."
//
// The −8 is a loyalty ability granted to the enchanted planeswalker
// (ADR 0109 §2): the walker pays it (CR 606.6) and it shares the
// walker's one loyalty activation a turn (CR 606.3). "Under your
// control" is the ability's controller. The cards are read before
// any moves and enter together, Rise of the Dark Realms' sentence.
//
// The trigger is No Mercy's with "enchanted planeswalker" for "you":
// one trigger per damaging creature, combat or not; the creature is
// destroyed only if it is still on the battlefield, and it is not
// targeted.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8c55039d-1303-420c-8547-ebdffb63bf89",
		Name:         "Liliana's Talent",
		Completeness: CompletenessFull,
		Targets:      EnchantPlaneswalker(),
		Grants: []AbilityGrant{{
			Key: lilianasTalentGrant,
			Activated: []ActivatedAbility{{
				Label: "−8: Put all creature cards from all graveyards onto the battlefield under your control.",
				Cost:  LoyaltyCost(-8),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return ReturnFromGraveyardTogether{
						Targets:    allGraveyardsCardIDs(ctx, game.Card.IsCreature),
						Controller: ctx.Controller(),
					}.Apply(ctx)
				},
			}},
			Text: "[−8]: Put all creature cards from all graveyards onto the battlefield under your control.",
		}},
		Static: []game.StaticAbility{GrantAbilitiesToAttached(lilianasTalentGrant)},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, ACreatureDealtDamageToEnchanted,
				"Liliana's Talent — destroy the creature that damaged enchanted planeswalker",
				destroyTheCreatureThatDealtTheDamage),
		},
	})
}

const lilianasTalentGrant = "lilianas-talent/reanimate"
