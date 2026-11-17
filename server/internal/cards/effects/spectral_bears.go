package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spectral Bears — Creature — Bear Spirit {1}{G}, 3/3 (Homelands, #5753):
//
//	"Whenever this creature attacks, if defending player controls no
//	 black nontoken permanents, it doesn't untap during your next untap
//	 step."
//
// An intervening "if" (CR 603.4), Fear of the Dark's shape. The ruling:
// "The check for black cards is made both at the time you declare
// attackers and as the ability resolves. If there are any at the time
// attackers are declared, it won't trigger at all. If there are any at
// the time it resolves, the ability will do nothing." The defending
// player is the one this attack is aimed at (CR 506.2: the player
// attacked, or the controller of the planeswalker or protector of the
// battle attacked). It is fixed as the ability triggers and carried on
// the item (Params.Player), so it is still the same player when the
// ability resolves, even if the Bears has left combat by then. Black is
// read as the card's text now says it (CR 612.2), and a permanent's
// colour through the layers.
//
// The mark is ADR 0058's one-shot, keyed to the controller's untap step
// ("your", UntapSkip.Player). A Bears that left and came back in response
// is a new object and is not marked (#1432).
//
// No simplification.
func init() {
	const label = "Spectral Bears — doesn't untap during your next untap step"
	t := On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return attackDeclared(ev, source) &&
			!spectralBearsDefenderControlsBlackNontoken(g, g.DefendingPlayerForAttackForEffect(ev.Target), source.TextColor("B"))
	}, label, func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if spectralBearsDefenderControlsBlackNontoken(g, item.Params.Player, ctx.TextColor("B")) {
			return nil
		}
		return DoesntUntapNextUntapStep{
			Targets: []uuid.UUID{ctx.Source()},
			Player:  ctx.Controller(),
			Label:   label,
		}.Apply(ctx)
	})
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Player = g.DefendingPlayerForAttackForEffect(ev.Target)
		return item
	}
	Register(Spec{
		OracleID:     "539899f3-f3f2-4f7a-a6f6-e80d4a39f753",
		Name:         "Spectral Bears",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}

// spectralBearsDefenderControlsBlackNontoken reports whether `player`
// controls a nontoken permanent of the colour `black` (the colour word as
// the source's text reads it). False for no player: an attack aimed at
// nobody has no defending player to control anything.
func spectralBearsDefenderControlsBlackNontoken(g *game.Game, player uuid.UUID, black string) bool {
	if player == uuid.Nil {
		return false
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && !c.IsToken() && c.HasColor(black) {
			return true
		}
	}
	return false
}
