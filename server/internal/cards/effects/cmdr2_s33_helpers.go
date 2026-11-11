package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2S33YourCommanderAttacked is "whenever your commander attacks": a
// commander its controller owns is declared as an attacker, by that
// controller. An attack on a planeswalker or battle counts, since the
// clause names no player (CR 508.1m, CR 508.3).
func cmdr2S33YourCommanderAttacked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCommander && c.Owner == source.Controller
}

// cmdr2S33CondThisPowerAtLeast is an intervening "if its power is N or
// greater" about the ability's own source creature (CR 603.4): its
// current power while it is on the battlefield, its last-known power once
// it has left (CR 608.2h). Angelic Sell-Sword's attack trigger.
type cmdr2S33CondThisPowerAtLeast struct{ N int }

func (c cmdr2S33CondThisPowerAtLeast) Holds(ctx *Context) bool {
	info, ok := ctx.SourcePermanent()
	return ok && info.Power >= c.N
}

func (c cmdr2S33CondThisPowerAtLeast) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondSource, c.N), c.N > 0
}

// cmdr2S33ThisOrAnotherNontokenCreatureYouControlEntered is "whenever this
// creature or another nontoken creature you control enters": the source
// itself always counts, token or not; any other creature must be a
// nontoken one under the source's controller.
func cmdr2S33ThisOrAnotherNontokenCreatureYouControlEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	if !ok || !c.IsCreature() {
		return false
	}
	return ev.CardID == source.InstanceID || !IsToken(c)
}

// cmdr2S33MisfortuneTellerExile exiles the first target and hangs the
// three-way clause off the exile: the card's creature and land bits are read
// before the move and ride the tail as N (1 creature, 2 land, 3 both).
func cmdr2S33MisfortuneTellerExile(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	ctx := NewContext(g, item)
	target := item.Targets[0].ID
	n := 0
	if c, ok := g.LookupCardForEffect(target); ok {
		if c.IsCreature() {
			n |= 1
		}
		if c.IsLand() {
			n |= 2
		}
	}
	return ExileTarget{Target: target, ThenTail: withN(ctx.ItemTail(cmdr2S33MisfortuneTellerClause), n)}.Apply(ctx)
}

// cmdr2S33MisfortuneTellerClause is "If it was a creature card, create a
// 2/2 black Rogue creature token. If it was a land card, create a Treasure
// token. Otherwise, you gain 3 life." An on-disk key.
var cmdr2S33MisfortuneTellerClause = CardTail("misfortune-teller/exiled-card-clause", func(ctx *Context, o game.TailOutcome) error {
	if !oneWentThisWay(o) {
		return nil
	}
	bits := o.Tail.N
	if bits&1 != 0 {
		if err := (CreateToken{Template: TokenCard("2/2 black Rogue"), N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	if bits&2 != 0 {
		if err := (CreateToken{Template: TreasureToken(), N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	if bits == 0 {
		return GainLife{Amount: 3}.Apply(ctx)
	}
	return nil
})

// cmdr2S33OpponentSacrificedANoncreatureToken is "whenever an opponent
// sacrifices a noncreature token": the sacrifice is announced while the
// permanent is still on the battlefield (EventSacrifice), so it is read as
// it stands there. The sacrificing player is the event's Actor.
func cmdr2S33OpponentSacrificedANoncreatureToken(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventSacrifice || ev.CardID == uuid.Nil || source == nil || ev.Actor == uuid.Nil || ev.Actor == source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && IsToken(c) && !c.IsCreature()
}

// cmdr2S33SacrificerGetsTwoPoison is "that player gets two poison
// counters": the player who sacrificed the token.
func cmdr2S33SacrificerGetsTwoPoison(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	player := item.Trigger.Event.Actor
	if g.PlayerByIDForEffect(player) == nil {
		return nil
	}
	return g.AddPlayerCounterByForEffect(item.Controller, player, game.CounterPoison, 2)
}

// cmdr2S33YouAndTargetOpponentEachCreateATreasure is "you and target
// opponent each create a Treasure token": you first, then the opponent the
// trigger targeted. A target that is no longer legal gets none.
func cmdr2S33YouAndTargetOpponentEachCreateATreasure(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 1}).Apply(ctx); err != nil {
		return err
	}
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
		return nil
	}
	return CreateToken{Controller: item.Targets[0].ID, Template: TreasureToken(), N: 1}.Apply(ctx)
}

// cmdr2S33CounterStash is the stash counter of Glittering Stockpile.
const cmdr2S33CounterStash = "stash"

// cmdr2S33PutAStashCounter is "Put a stash counter on this artifact", the
// rider of Glittering Stockpile's {T} ability: it follows the mana in the
// printed order. A mana ability has no priority window, so a counter
// doubler's order settles now (CR 605.3b), as Empowered Autogenerator's does.
func cmdr2S33PutAStashCounter(g *game.Game, _, source uuid.UUID) error {
	_, err := g.AddCounterMustSettleNowForEffect(source, cmdr2S33CounterStash, 1)
	return err
}

// cmdr2S33StashCounters is X, "the number of stash counters on this
// artifact". The sacrifice is part of the cost, so by the time the mana is
// worked out the artifact has left the battlefield: X is then read from its
// last-known information (CR 608.2h, CR 112.7a).
func cmdr2S33StashCounters(g *game.Game, _, source uuid.UUID) int {
	if c, ok := g.LookupCardForEffect(source); ok && onBattlefield(g, source) {
		return c.Counters[cmdr2S33CounterStash]
	}
	if info, ok := g.LastKnownPermanentForEffect(source); ok {
		return info.Counters[cmdr2S33CounterStash]
	}
	return 0
}

// cmdr2S33OutlawSubtypes is the five creature types the Outlaws of Thunder
// Junction reminder text calls outlaws, "|"-joined for a mana-restriction
// subtype tag (an alternation, game.ManaRestrictAnyType's shape).
const cmdr2S33OutlawSubtypes = "Assassin|Mercenary|Pirate|Rogue|Warlock"

// cmdr2S33YouCastYourFirstOutlawSpellEachTurn is "whenever you cast your
// first outlaw spell each turn": the cast is an outlaw spell of the source's
// controller and no earlier cast this turn by that player was one (the event
// log keeps the order, as Torgal's "first Human" does).
func cmdr2S33YouCastYourFirstOutlawSpellEachTurn(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	outlaw := b40Outlaw()
	if !YouCast(outlaw)(ev, source, lki, g) {
		return false
	}
	for _, prev := range g.EventsThisTurn() {
		if prev.Seq >= ev.Seq {
			break
		}
		if prev.Kind != game.EventCast || prev.Actor != ev.Actor {
			continue
		}
		if c, ok := g.LookupCardForEffect(prev.CardID); ok && outlaw(g, ev.Actor, c) {
			return false
		}
	}
	return true
}
