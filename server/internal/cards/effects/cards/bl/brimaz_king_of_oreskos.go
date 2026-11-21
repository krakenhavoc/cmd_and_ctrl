package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brimaz, King of Oreskos — Legendary Creature — Cat Soldier
// {1}{W}{W}, 3/4 (EDHREC rank 3027):
//
//	"Vigilance
//	 Whenever Brimaz attacks, create a 1/1 white Cat Soldier creature
//	 token with vigilance that's attacking.
//	 Whenever Brimaz blocks a creature, create a 1/1 white Cat
//	 Soldier creature token with vigilance that's blocking that
//	 creature."
//
// The Cat king that brings a Cat to every fight. Vigilance rides
// PrintedKeywords.
//
//   - The attack trigger is attackDeclared; the token enters ATTACKING
//     (untapped, vigilance is irrelevant to a token that was never
//     declared) through CreateTokensAttackingYourChoice: the controller
//     chooses what it attacks (CR 508.4), Brimaz's own target asked
//     first, and it is never declared, so it fires no "whenever a
//     creature attacks" trigger (CR 506.3c). It can be blocked and deals
//     its 1 this combat.
//   - The block trigger is b28SelfBlocked (EventBlock with Brimaz in the
//     blocker slot, the attacker in Target) and makes the same token
//     BLOCKING that creature (CR 506.3d): CreateTokensBlockingForEffect
//     puts it onto the battlefield with its block stamped and recorded as
//     announced, so it is never declared as a blocker, fires no block
//     trigger, and deals and takes combat damage with the attacker. An
//     attacker that left combat before the trigger resolved leaves the
//     Cat on the battlefield not blocking.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "49471b65-be9d-4ded-b2e5-dfc81b306b54",
		Name:            "Brimaz, King of Oreskos",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.AttackDeclared(ev, source)
				},
				Key: "Brimaz, King of Oreskos — an attacking 1/1 Cat Soldier with vigilance",
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateTokensAttackingYourChoice{
						Template: effects.TokenCard("1/1 white Cat Soldier with vigilance"),
						N:        1,
						Prefer:   effects.TriggerAttackTarget(item),
						Label:    "Brimaz, King of Oreskos",
					}.Apply(effects.NewContext(g, item))
				},
			},
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.B28SelfBlocked(ev, source)
			}, "Brimaz, King of Oreskos — a 1/1 Cat Soldier with vigilance, blocking that creature",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return g.CreateTokensBlockingForEffect(item.Controller, effects.NewContext(g, item).TextToken(effects.TokenCard("1/1 white Cat Soldier with vigilance")), 1, item.Trigger.Event.Target)
				}),
		},
	})
}
