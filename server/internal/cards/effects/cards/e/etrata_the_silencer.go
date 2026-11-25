package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Etrata, the Silencer — Legendary Creature — Vampire Assassin {2}{U}{B},
// 3/5 (Guilds of Ravnica):
//
//	"Etrata can't be blocked.
//	 Whenever Etrata deals combat damage to a player, exile target creature
//	 that player controls and put a hit counter on that card. That player
//	 loses the game if they own three or more exiled cards with hit counters
//	 on them. Etrata's owner shuffles Etrata into their library."
//
// "That player" is the one the combat damage named: the target is a creature
// they control (ControlledByThatPlayer). The exile is a continuation (a
// replacement may stop to ask), and the counter goes on the exiled card only
// if it really reached exile, as Mari, the Killing Quill puts hers. The loss
// counts the cards that player OWNS in exile with a hit counter, whoever
// exiled them, and is the whole second sentence (it does not depend on this
// exile having landed). Etrata is shuffled into her owner's library last,
// only while she is still the permanent that dealt the damage (CR 400.7); a
// trigger with no creature to target is removed, so she stays.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2ee87667-ec9d-4ced-b7e6-df3fd18b0ed6",
		Name:         "Etrata, the Silencer",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBeBlocked)},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			Key:     "Etrata, the Silencer — exile target creature that player controls with a hit counter; that player may lose; shuffle Etrata into your library",
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Source == source.InstanceID && effects.CombatDamageToPlayerBy(ev, source.Controller, g)
			},
			Targets: effects.ControlledByThatPlayer(effects.TargetCreature("target creature that player controls")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				target, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				victim := item.Trigger.Event.Target
				return effects.ExileTarget{
					Target:   target,
					ThenTail: effects.WithCarry(ctx.ItemTail(etrataTheSilencerExiled), target, victim),
				}.Apply(ctx)
			},
		}},
	})
}

// etrataTheSilencerExiled is the clause after the exile: the hit counter (when
// the card reached exile), the loss check and the shuffle. An on-disk key.
var etrataTheSilencerExiled = effects.CardTail("etrata-the-silencer/after-exile", func(ctx *effects.Context, o game.TailOutcome) error {
	card := effects.TailCarry0(o)
	victim := uuid.Nil
	if len(o.Tail.Carry) > 1 {
		victim = o.Tail.Carry[1]
	}
	if len(o.IDs) == 1 && card != uuid.Nil {
		if err := (effects.AddCounter{Target: card, Kind: "hit", N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	if victim != uuid.Nil && effects.GapB213HitCountersOwnedBy(ctx.Game, victim) >= 3 {
		if err := (effects.LoseTheGame{Player: victim}).Apply(ctx); err != nil {
			return err
		}
	}
	source := ctx.Source()
	if source == uuid.Nil || ctx.Item == nil || !effects.OnBattlefield(ctx.Game, source) || effects.SourceIsNewObject(ctx.Game, ctx.Item) {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(source)
	if !ok {
		return nil
	}
	return ctx.Game.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{source}, game.TuckOptions{}, game.KeyedTail{Ref: effects.ThenShuffle, Player: c.Owner})
})
