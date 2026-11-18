package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	grant, ability := sncB19ExileForLand("Spara's Adjudicators", "sparas-adjudicators/tap-gwu", "{G|W|U}", "{G}, {W}, or {U}")
	Register(Spec{
		OracleID:     "02718076-4c71-4bf5-988f-e6f94fbf0aef",
		Name:         "Spara's Adjudicators",
		Completeness: CompletenessFull,
		Grants:       []AbilityGrant{grant},
		Activated:    []ActivatedAbility{ability},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: Self,
			Key:       "Spara's Adjudicators — target creature an opponent controls can't attack or block until your next turn",
			Targets:   TargetCreature("target creature an opponent controls", OpponentControls()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return ScopedEffectFor{
					Target:   id,
					Mods:     []game.Mod{game.AddRestrictionsMod(game.CantAttackOrBlock)},
					Duration: DurationUntilYourNextTurn(ctx, ctx.Controller()),
					Label:    "Spara's Adjudicators — can't attack or block until your next turn",
				}.Apply(ctx)
			},
		}},
	})
}
