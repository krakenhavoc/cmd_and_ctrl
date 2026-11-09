package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tyrant's Familiar — Creature — Dragon {5}{R}{R}, 5/5:
//
//	"Flying, haste
//	 Lieutenant — As long as you control your commander, this creature
//	 gets +2/+2 and has "Whenever this creature attacks, it deals 7
//	 damage to target creature defending player controls.""
//
// Lieutenant looks only at YOUR commander on the battlefield under your
// control (b18ControlsYourCommander), so a stolen Familiar checks its new
// controller's commander. The +2/+2 is a layer 7c static. The granted
// trigger is written as the card's own, gated on the same condition when
// the attack is declared; once it has triggered, losing the commander
// does not stop it resolving.
//
// No simplification.
func init() {
	lieutenant := func(_ *game.Card, g *game.Game, source *game.Card) bool {
		return b18ControlsYourCommander(g, source.Controller)
	}
	Register(Spec{
		OracleID:        "63394e24-a456-4cc6-9885-8c62447b050a",
		Name:            "Tyrant's Familiar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Static: []game.StaticAbility{{
			Label:    "Lieutenant — As long as you control your commander, this creature gets +2/+2.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && lieutenant(target, g, source)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power += 2
				c.Toughness += 2
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return ThisAttacked(ev, source, lki, g) && b18ControlsYourCommander(g, source.Controller)
			},
			TargetsFrom: TargetCreatureDefendingPlayerControls,
			Key:         "Tyrant's Familiar — Lieutenant: it deals 7 damage to target creature defending player controls",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				return DealDamage{Source: item.SourceCardID, Target: ts[0].ID, Amount: 7}.Apply(ctx)
			},
		}},
	})
}
