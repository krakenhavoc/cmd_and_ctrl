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
// PrintedKeywords. The attack trigger is attackDeclared and the token
// enters ATTACKING the same player Brimaz was declared against,
// through CreateTokensAttackingForEffect (Parhelion II's path): put
// onto the battlefield attacking, never declared, so it fires no
// "whenever a creature attacks" trigger of its own (CR 508.4), it
// can be blocked, and it deals its 1 this combat. The block trigger
// is b28SelfBlocked (EventBlock with Brimaz in the blocker slot) and
// makes the same token.
//
// DECLARED SIMPLIFICATION, weaker than printed, on the block half:
// the token Brimaz makes when it blocks is created but is NOT
// blocking. Two things in the way, both engine-side: there is no
// enter-the-battlefield-blocking path (CreateTokensAttackingForEffect
// has no blocking twin, and a block is a player declaration the
// engine validates), and the block declaration is locked in when it
// is complete (#830), so a token that arrived after it could not
// join the declaration even if there were such a path. The token
// arrives as an ordinary untapped Cat
// Soldier. Never stronger: the printed token would have blocked and
// this one merely exists.
//
// Two smaller postures, both as printed: an attack at a planeswalker
// or battle (S27) makes the token untapped and not attacking, since
// the attacking-token path takes a player; and the choice of what
// the token attacks is not the controller's — it attacks what
// Brimaz attacks, which the printed card also says.
func init() {
	Register(Spec{
		OracleID:        "49471b65-be9d-4ded-b2e5-dfc81b306b54",
		Name:            "Brimaz, King of Oreskos",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"The Cat Soldier Brimaz makes when it blocks isn't blocking — it arrives after combat damage as an ordinary untapped token."},
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclared(ev, source)
				},
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					defender := ev.Target
					return game.NewTriggeredItem(source, "Brimaz, King of Oreskos — an attacking 1/1 Cat Soldier with vigilance",
						func(g *game.Game, item *game.StackItem) error {
							return g.CreateTokensAttackingForEffect(item.Controller, TokenCard("1/1 white Cat Soldier with vigilance"), 1, defender)
						})
				},
			},
			On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b28SelfBlocked(ev, source)
			}, "Brimaz, King of Oreskos — a 1/1 Cat Soldier with vigilance", Do(CreateToken{Template: TokenCard("1/1 white Cat Soldier with vigilance"), N: 1})),
		},
	})
}
