package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Underfoot Underdogs — Creature — Goblin Warrior {2}{R}, 1/2 (Tarkir:
// Dragonstorm, #4463):
//
//	"When this creature enters, create a 1/1 red Goblin creature token.
//	 {1}, {T}: Target creature you control with power 2 or less can't be
//	 blocked this turn."
//
// The enters trigger makes the token table's 1/1 red Goblin. The activated
// ability is Deserter's Disciple's without "another": the Underdogs may
// target themselves. Power is checked as the target is chosen and again
// as the ability resolves (CR 608.2b); a creature pumped above 2 in
// response is no longer a legal target and is not made unblockable.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3740b86d-b0e6-4837-b408-e7f62b95c787",
		Name:         "Underfoot Underdogs",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Underfoot Underdogs — create a 1/1 red Goblin creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 red Goblin"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Target creature you control with power 2 or less can't be blocked this turn.",
			Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature you control with power 2 or less", effects.YouControl(), effects.PowerLE(2)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return effects.RestrictUntilEOT{
							Target:       t.ID,
							Restrictions: game.CantBeBlocked,
							Label:        "Underfoot Underdogs — can't be blocked",
						}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
