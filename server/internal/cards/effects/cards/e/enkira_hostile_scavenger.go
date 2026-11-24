package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Enkira, Hostile Scavenger — Legendary Creature — Human Warrior {3}{B}{G},
// 3/3 (Secret Lair Drop):
//
//	"When Enkira, Hostile Scavenger enters, create two Walker tokens. (They're
//	 2/2 black Zombie creatures.)
//	 As long as Enkira is equipped, it must be blocked if able.
//	 Whenever Enkira and at least two Zombies attack, Enkira gains
//	 indestructible until end of turn."
//
// The block requirement is a static that applies to Enkira only while some
// Equipment is attached to her, read live on every recompute
// (B41SourceIsEquipped). The attack trigger is battalion-shaped: Enkira is
// declared as an attacker and at least two other creatures its controller
// controls that are Zombies are attacking in the same declaration.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "558fd7a9-9b39-4c82-b909-5febf55dc010",
		Name:         "Enkira, Hostile Scavenger",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		Static: []game.StaticAbility{
			func() game.StaticAbility {
				s := effects.BlockRequirementWhere(game.BlockRequirementMustBeBlocked, func(target *game.Card, g *game.Game, source *game.Card) bool {
					return target.InstanceID == source.InstanceID && effects.B41SourceIsEquipped(g, source)
				})
				s.Label = "As long as Enkira is equipped, it must be blocked if able."
				return s
			}(),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Enkira, Hostile Scavenger — create two Walker tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 black Walker"), N: 2})),
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.AttackDeclared(ev, source) {
					return false
				}
				n := 0
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == source.Controller && c.IsCreature() && c.InstanceID != source.InstanceID &&
						c.AttackingTarget != uuid.Nil && c.HasSubtype(source.TextSubtype("Zombie")) {
						n++
					}
				}
				return n >= 2
			}, "Enkira, Hostile Scavenger — Enkira gains indestructible until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"indestructible"},
						Label: "Enkira, Hostile Scavenger — indestructible"}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
