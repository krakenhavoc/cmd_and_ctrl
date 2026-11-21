package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const (
	watcherDrawLabel     = "The Watcher in the Water — create a 1/1 blue Tentacle creature token"
	watcherTentacleLabel = "The Watcher in the Water — untap up to one target Kraken and put a stun counter on up to one target nonland permanent"
)

// The Watcher in the Water — Legendary Creature — Kraken {3}{U}{U}, 9/9 (The
// Lord of the Rings: Tales of Middle-earth Commander):
//
//	"The Watcher in the Water enters tapped with nine stun counters on it.
//	 Whenever you draw a card during an opponent's turn, create a 1/1 blue
//	 Tentacle creature token.
//	 Whenever a Tentacle you control dies, untap up to one target Kraken and
//	 put a stun counter on up to one target nonland permanent."
//
// Baloth Prime's enters-tapped-with-counters replacement (nine stun
// counters). The draw watch reads whose turn it is as each card is drawn,
// once per card. The death watch is a "died under your control" check on a
// Tentacle (read as it last stood), with two independent "up to one"
// clauses: any Kraken, whoever controls it, and any nonland permanent. A
// target no longer legal is skipped (CR 608.2b).
//
// No simplification.
func init() {
	death := effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		dead, ok := effects.DiedCreature(ev, g)
		return ok && effects.LeftUnderControlOf(ev, dead) == source.Controller && dead.HasSubtype(source.TextSubtype("Tentacle"))
	}, watcherTentacleLabel, watcherTentacleDeath)
	death.Targets = effects.Clauses(
		effects.TargetCreature("up to one target Kraken", effects.HasSubtype("Kraken")).WithCount(0, 1),
		effects.TargetPermanent("up to one target nonland permanent", effects.Nonland()).WithCount(0, 1),
	)
	effects.Register(effects.Spec{
		OracleID:     "a9ddb3dc-f289-46a1-af17-3a8523a26b0d",
		Name:         "The Watcher in the Water",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Replacements: []game.ReplacementEffect{effects.SelfEntersTappedWithCounters(game.CounterStun, 9)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventDrawCard && ev.Actor == source.Controller && effects.GapB127IsOpponentsTurn(g, source.Controller)
			}, watcherDrawLabel, effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 blue Tentacle"), N: 1})),
			death,
		},
	})
}

func watcherTentacleDeath(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if kraken, ok := ctx.ClauseTarget(0); ok {
		if err := (effects.UntapTarget{Target: kraken.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	if perm, ok := ctx.ClauseTarget(1); ok {
		if err := (effects.AddCounter{Target: perm.ID, Kind: game.CounterStun, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
