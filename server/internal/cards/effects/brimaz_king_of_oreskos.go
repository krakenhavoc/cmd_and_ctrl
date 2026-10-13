package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "49471b65-be9d-4ded-b2e5-dfc81b306b54",
		Name:            "Brimaz, King of Oreskos",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclared(ev, source)
				},
				Key: "Brimaz, King of Oreskos — an attacking 1/1 Cat Soldier with vigilance",
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateTokensAttackingYourChoice{
						Template: TokenCard("1/1 white Cat Soldier with vigilance"),
						N:        1,
						Prefer:   triggerAttackTarget(item),
						Label:    "Brimaz, King of Oreskos",
					}.Apply(NewContext(g, item))
				},
			},
			On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b28SelfBlocked(ev, source)
			}, "Brimaz, King of Oreskos — a 1/1 Cat Soldier with vigilance, blocking that creature",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return g.CreateTokensBlockingForEffect(item.Controller, TokenCard("1/1 white Cat Soldier with vigilance"), 1, item.Trigger.Event.Target)
				}),
		},
	})
}
