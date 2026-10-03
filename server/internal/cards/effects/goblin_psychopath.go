package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Psychopath — Creature — Goblin Mutant {3}{R}, 5/5:
//
//	"Whenever this creature attacks or blocks, flip a coin. If you lose the
//	 flip, the next time it would deal combat damage this turn, it deals
//	 that damage to you instead."
//
// ADR 0108 §9 (#1905): on a lost flip (CR 705.2), a "next time"
// redirection of the Psychopath's combat damage, to anything, to the
// trigger's controller. One that has left the battlefield by then is
// gone (CR 400.7): nothing is redirected.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "9369f131-3356-4926-9a9f-88a636305bd0",
		Name:         "Goblin Psychopath",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclared(ev, source) || selfBlocksOnce(ev, source)
			}, "Goblin Psychopath — flip a coin", goblinPsychopathFlip),
		},
	})
}

// goblinPsychopathFlip flips; a lost flip redirects the creature's next
// combat damage to its controller.
func goblinPsychopathFlip(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	controller, source := ctx.Controller(), ctx.Source()
	var ref game.ObjectRef
	var zone game.ZoneKind
	ok := !sourceIsNewObject(g, item) && onBattlefield(g, source)
	if ok {
		ref, zone, ok = g.DamageSourceRefLocked(source)
	}
	g.FlipCoinForEffect(game.CoinFlipSpec{
		Flipper:  controller,
		Source:   source,
		Question: "Goblin Psychopath — call the coin flip",
		Then: func(g *game.Game, result game.CoinFlipResult) error {
			if !ok || (len(result.Won) > 0 && result.Won[0]) {
				return nil
			}
			g.RedirectDamageThisTurnForEffect(game.DamageRedirection{
				EffectSource: source,
				Controller:   controller,
				Source:       ref,
				SourceZone:   zone,
				CombatOnly:   true,
				Next:         true,
				To:           controller,
				Label:        "Goblin Psychopath — its next combat damage is dealt to you",
			})
			return nil
		},
	})
	return nil
}
