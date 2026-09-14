package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Weapons Manufacturing — Enchantment {1}{R} (EDHREC rank 3342):
//
//	"Whenever a nontoken artifact you control enters, create a
//	 colorless artifact token named Munitions with "When this token
//	 leaves the battlefield, it deals 2 damage to any target.""
//
// Every real artifact brings a Shock that fires when the Munitions
// is sacrificed, bounced or blown up. The entry trigger is
// b31NontokenArtifactYouControlEntered — post-layer types, tokens
// excluded, so the Munitions themselves never chain. The token is a
// colorless artifact named Munitions and nothing else.
//
// The token's printed trigger cannot live on the token: a token
// template carries no triggered abilities and a token has no oracle
// ID for the catalog to key one on. So the Manufacturing carries it
// on the tokens' behalf (the Simulacrum Synthesizer posture): a
// second trigger on the enchantment watches for a Munitions token
// its controller controls leaving the battlefield by any route
// (b31MunitionsYouControlLeft), targets "any target" when it goes
// on the stack, and deals the 2 from the TOKEN that left rather
// than from the Manufacturing — a colorless source, as printed, so
// a red-damage payoff does not see it (b31DamageChosenTargetFrom).
//
// Sandbox simplification, declared and weaker than printed: because
// the enchantment carries the trigger, a Munitions that leaves the
// battlefield after Weapons Manufacturing has left deals no damage,
// and a Munitions an opponent has taken control of deals none
// either. Printed, the token's own ability would fire in both
// cases.
func init() {
	Register(Spec{
		OracleID:     "e48a8160-cffe-4e00-a3e9-a59d7fc7b3a2",
		Name:         "Weapons Manufacturing",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"A Munitions token only deals its 2 damage while Weapons Manufacturing is still on the battlefield and the token is still yours — the enchantment carries the token's ability for it.",
		},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventETB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b31NontokenArtifactYouControlEntered(ev, source, g)
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Weapons Manufacturing — create a Munitions token",
						func(g *game.Game, item *game.StackItem) error {
							return CreateToken{Controller: item.Controller, Template: b31MunitionsToken(), N: 1}.Apply(NewContext(g, item))
						})
				},
			},
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b31MunitionsYouControlLeft(ev, source, g)
				},
				Targets: TargetAny(),
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Munitions — 2 damage to any target", b31DamageChosenTargetFrom(ev.CardID, 2))
				},
			},
		},
	})
}
