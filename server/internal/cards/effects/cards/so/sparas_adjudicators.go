package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spara's Adjudicators — Creature — Cat Citizen {2}{G}{W}{U}, 4/4 (Streets
// of New Capenna):
//
//	"When this creature enters, target creature an opponent controls can't
//	 attack or block until your next turn.
//	 {2}, Exile this card from your hand: Target land gains "{T}: Add {G},
//	 {W}, or {U}" until this card is cast from exile. You may cast this
//	 card for as long as it remains exiled."
//
// The restriction is Kaito, Dancing Shadow's: a CantAttackOrBlock scoped
// effect lasting until the controller's next turn begins. The hand ability
// is the cycle's shared one (sncB19ExileForLand).
//
// No simplification.
func init() {
	grant, ability := effects.SncB19ExileForLand("Spara's Adjudicators", "sparas-adjudicators/tap-gwu", "{G|W|U}", "{G}, {W}, or {U}")
	effects.Register(effects.Spec{
		OracleID:     "02718076-4c71-4bf5-988f-e6f94fbf0aef",
		Name:         "Spara's Adjudicators",
		Completeness: effects.CompletenessFull,
		Grants:       []effects.AbilityGrant{grant},
		Activated:    []effects.ActivatedAbility{ability},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: effects.Self,
			Key:       "Spara's Adjudicators — target creature an opponent controls can't attack or block until your next turn",
			Targets:   effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.ScopedEffectFor{
					Target:   id,
					Mods:     []game.Mod{game.AddRestrictionsMod(game.CantAttackOrBlock)},
					Duration: effects.DurationUntilYourNextTurn(ctx, ctx.Controller()),
					Label:    "Spara's Adjudicators — can't attack or block until your next turn",
				}.Apply(ctx)
			},
		}},
	})
}
