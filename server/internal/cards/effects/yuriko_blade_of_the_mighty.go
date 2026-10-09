package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Yuriko, Blade of the Mighty — Legendary Creature — Human Samurai
// {3}{W}, 2/3:
//
//	"During combat, players can't cast spells or activate abilities
//	 that aren't mana abilities.
//	 Whenever a creature you control attacks a player alone, it gains
//	 double strike until end of turn."
//
// The first line is a cast restriction and an activation restriction
// that bind every player, Yuriko's controller included, in any of the
// five combat steps (CR 101.2: "can't" wins). Mana abilities are
// exempt from the activation half. Special actions are neither casts
// nor activations and are untouched.
//
// "Attacks a player alone" is CR 506.5: it is the only creature attacking
// at the instant of declaration, and what it attacks is a player — a
// lone attacker aimed at a planeswalker or battle does not trigger. The
// double strike goes to the attacker named by the event, if it is still
// that object on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f89b0143-4b19-4ab1-a96b-356b21e435ef",
		Name:         "Yuriko, Blade of the Mighty",
		Completeness: CompletenessFull,
		CastRestrictions: []game.CastRestriction{{
			Label: "During combat, players can't cast spells.",
			Forbids: func(q game.CastQuery) bool {
				return rfCreatureFCombat(q.Game)
			},
		}},
		ActivationRestrictions: []game.ActivationRestriction{{
			Label: "During combat, players can't activate abilities that aren't mana abilities.",
			Forbids: func(q game.ActivationQuery) bool {
				return !q.Ability.Mana && rfCreatureFCombat(q.Game)
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, rfCreatureFAttackedAPlayerAlone,
				"Yuriko, Blade of the Mighty — the attacking creature gains double strike until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					attacker := ctx.Trigger().Event.CardID
					if !onBattlefield(g, attacker) {
						return nil
					}
					return GrantKeywordUntilEOT{
						Target:   attacker,
						Keywords: []string{"double strike"},
						Label:    "Yuriko, Blade of the Mighty — double strike",
					}.Apply(ctx)
				}),
		},
	})
}
