package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cait, Cage Brawler — Legendary Creature — Human Warrior {R}{G}, 1/1
// (Fallout, #5948):
//
//	"During your turn, Cait has indestructible.
//	 Whenever Cait attacks, you and defending player each draw a card, then
//	 discard a card. Put two +1/+1 counters on Cait if you discarded the card
//	 with the greatest mana value among those cards or tied for greatest."
//
// The indestructible is a layer 6 self-grant that is on while its controller
// is the active player. The attack trigger pins the defending player as the
// attack is declared (a planeswalker's controller, CR 506.3d: the player
// being attacked or whose planeswalker or battle is). Both players draw
// first, the attacker then the defender (CR 101.4), and then discard in a
// keyed two-seat run, each choosing their own card. "Those cards" are the two
// cards actually discarded: the continuation compares their mana values
// (a card with no mana cost, a land, has 0) and puts two +1/+1 counters on
// Cait only if the attacker's card is not smaller than the defender's. A
// player who discarded nothing (an empty hand) has no card among them; if
// the attacker discarded nothing there is no counter. Cait gets her counters
// only if she is still the same object on the battlefield.
//
// No simplification.
func init() {
	label := "Cait, Cage Brawler — you and defending player each draw a card, then discard a card; two +1/+1 counters if you discarded the greatest mana value"
	attack := game.TriggeredAbility{
		Watches:   []game.EventKind{game.EventAttack},
		AppliesTo: effects.ThisAttacked,
		Key:       label,
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Params.Player = effects.B17DefendingPlayer(g, ev)
			return item
		},
		Effect: caitAttack,
	}
	ind := effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return effects.SelfCreature(target, g, source) && effects.IsActivePlayer(g, source.Controller)
	}, "indestructible")
	ind.Label = "During your turn, Cait has indestructible."
	effects.Register(effects.Spec{
		OracleID:     "f24319a1-0662-44be-9027-11e7bc020e1d",
		Name:         "Cait, Cage Brawler",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{ind},
		Triggered:    []game.TriggeredAbility{attack},
	})
}

// caitAttack draws for the attacker and the defender, then asks each to
// discard, the rest riding the run (caitDiscarded) with Cait and the
// attacker as its carry.
func caitAttack(g *game.Game, item *game.StackItem) error {
	players := []uuid.UUID{item.Controller}
	if def := item.Params.Player; def != uuid.Nil && def != item.Controller && g.PlayerByIDForEffect(def) != nil {
		players = append(players, def)
	}
	for _, p := range players {
		if err := g.DrawNForEffect(p, 1); err != nil {
			return err
		}
	}
	return g.PlayersDiscardThenKeyedForEffect(players, game.DiscardPrompt{
		Source:   item.SourceCardID,
		N:        1,
		Question: "Cait, Cage Brawler — discard a card",
	}, caitDiscarded, []uuid.UUID{item.SourceCardID, item.Controller})
}

// caitDiscarded is the rest of Cait's attack trigger, once both discards are
// done: two +1/+1 counters on Cait if the attacker's discarded card has the
// greatest mana value among the cards discarded, or ties for it. An on-disk
// key: never renamed, never reused.
var caitDiscarded = game.RegisterPromptRunThen("prompt-run/cait-cage-brawler-discarded", func(g *game.Game, r game.RunDone) error {
	if len(r.Carry) < 2 {
		return nil
	}
	cait, me := r.Carry[0], r.Carry[1]
	mine := r.Discards().By(me)
	if len(mine) == 0 {
		return nil
	}
	valueOf := func(id uuid.UUID) int {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			return 0
		}
		mv, _ := g.ManaValueForEffect(c)
		return mv
	}
	best := valueOf(mine[0])
	for _, e := range r.Landed {
		if e.Seat == me {
			continue
		}
		for _, id := range e.Cards {
			if valueOf(id) > best {
				return nil
			}
		}
	}
	if !effects.OnBattlefield(g, cait) {
		return nil
	}
	return effects.AddCounter{Target: cait, Kind: game.CounterPlusOne, N: 2}.Apply(effects.KeyedContext(g, me, cait))
})
