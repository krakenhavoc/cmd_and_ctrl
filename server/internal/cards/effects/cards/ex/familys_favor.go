package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Family's Favor — Enchantment {2}{G} (Streets of New Capenna
// Commander):
//
//	"Whenever you attack, put a shield counter on target attacking
//	 creature. Until end of turn, it gains "Whenever this creature
//	 deals combat damage to a player, remove a shield counter from it.
//	 If you do, draw a card." (If a creature with a shield counter on it
//	 would be dealt damage or destroyed, remove a shield counter from
//	 it instead.)"
//
// "Whenever you attack" is one trigger per declaration (OncePerBatch),
// the attacker chosen as it goes on the stack. The granted trigger is an
// ADR 0093 duration grant that lasts the turn; its "if you do" is real:
// a creature whose counter is already gone draws nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7f5bebfd-faf3-485f-b906-aa96d621d499",
		Name:         "Family's Favor",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: familysFavorGrant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisDealsCombatDamageToAPlayer("Family's Favor — remove a shield counter from this creature; if you do, draw a card",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						self, ok := ctx.SourcePermanent()
						if !ok || self.Counters[game.CounterShield] <= 0 {
							return nil
						}
						if err := (effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterShield, N: -1}).Apply(ctx); err != nil {
							return err
						}
						return effects.DrawCards{N: 1}.Apply(ctx)
					}),
			},
			Text: "Whenever this creature deals combat damage to a player, remove a shield counter from it. If you do, draw a card.",
		}},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.Targeting(
				effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.AttackDeclaredByYou(ev, source.Controller)
				}, "Family's Favor — put a shield counter on target attacking creature; until end of turn it gains a draw trigger",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						if err := (effects.AddCounter{Target: id, Kind: game.CounterShield, N: 1}).Apply(ctx); err != nil {
							return err
						}
						return effects.GrantAbilitiesFor{
							Target: id,
							Keys:   []string{familysFavorGrant},
							Label:  "Family's Favor — removes a shield counter to draw on combat damage to a player",
						}.Apply(ctx)
					}),
				effects.TargetCreature("target attacking creature", effects.AttackingCreature()))),
		},
	})
}

const familysFavorGrant = "familys-favor/remove-shield-draw"
