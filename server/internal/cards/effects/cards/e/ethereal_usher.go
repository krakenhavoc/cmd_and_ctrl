package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ethereal Usher — Creature — Spirit {5}{U}, 2/3 (Ravnica: City of Guilds):
//
//	"{U}, {T}: Target creature can't be blocked this turn.
//	 Transmute {1}{U}{U} ({1}{U}{U}, Discard this card: Search your library for a
//	 card with the same mana value as this card, reveal it, put it into your hand,
//	 then shuffle. Transmute only as a sorcery.)"
//
// Access Tunnel's unblockable grant with no power limit: a restriction on the
// target for the turn (CR 509.1b), skipped if the target left in response
// (CR 608.2b). It can target any creature, the Usher's controller's or not.
// Transmute is the keyword's constructor (CR 702.53a), finding a mana value 6
// card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3ad6e495-d0e6-4b07-8f91-fe057e173975",
		Name:         "Ethereal Usher",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{U}, {T}: Target creature can't be blocked this turn.",
				Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
				Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						return effects.RestrictUntilEOT{
							Target:       t.ID,
							Restrictions: game.CantBeBlocked,
							Label:        "Ethereal Usher — can't be blocked",
						}.Apply(ctx)
					}
					return nil
				},
			},
			effects.Transmute("{1}{U}{U}"),
		},
	})
}
