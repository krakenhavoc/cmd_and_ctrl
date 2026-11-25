package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// copy_for_each_could_target.go — CR 707.10d for any permanent or player
// (#6159): "Copy that spell or ability for each other permanent or player
// the spell or ability could target. Each copy targets a different one of
// those permanents and players" (Radiant Performer).
//
//	707.10d Some effects copy a spell or ability for each player or object
//	it "could target." The copies are put onto the stack with those
//	targets in the order of their controller's choice. If the spell or
//	ability has more than one target, each of its targets must be the
//	same player or object. If that player or object isn't a legal target
//	for each instance of the word "target," a copy isn't created for that
//	player or object.
//
// The copies themselves are game.CopySpellRetargetedEachForEffect and
// game.CopyAbilityRetargetedEachForEffect (Mirrorwing Dragon, Agrus Kos),
// which take a player for `to` as readily as a permanent. The order is
// the controller's, asked one copy at a time with a keyed option pick
// whose options are the permanents (shown as cards) and the players
// still to place, because the card-order prompt cannot hold a player.
// The prompt is plain data (the carry), so a table paused on it is a
// restore point.

// ItemTargetsOnlyASinglePermanentOrPlayer is "spell or ability that
// targets only a single permanent or player": every target of the item is
// one and the same player, or one and the same permanent on the
// battlefield.
func ItemTargetsOnlyASinglePermanentOrPlayer() StackItemPredicate {
	return func(g *game.Game, _ uuid.UUID, item *game.StackItem) bool {
		sole, ok := g.StackItemSoleTargetForEffect(item.ID)
		if !ok || targetsSomethingElse(item.AllTargets(), sole.ID) {
			return false
		}
		if sole.Kind == game.TargetPlayer {
			return true
		}
		z := g.FindCardZoneForEffect(sole.ID)
		return z != nil && z.Kind == game.ZoneBattlefield
	}
}

// targetsSomethingElse reports whether any player or card target in refs
// names something other than id (a spliced text's target, CR 702.47d).
func targetsSomethingElse(refs []game.TargetRef, id uuid.UUID) bool {
	for _, t := range refs {
		if (t.Kind == game.TargetCard || t.Kind == game.TargetPlayer) && t.ID != id {
			return true
		}
	}
	return false
}

// CopyForEachOtherPermanentOrPlayer copies the spell or ability itemID,
// under ctx's controller, for each other permanent or player it could
// target instead of its single target, each copy targeting a different one
// of them, in the order the controller chooses (CR 707.10d). `name` heads
// the order prompt. An item that has left the stack, or no longer targets
// a single permanent or player, makes no copy.
func CopyForEachOtherPermanentOrPlayer(ctx *Context, itemID uuid.UUID, name string) {
	g := ctx.Game
	item := g.StackItemForEffect(itemID)
	if item == nil {
		return
	}
	sole, ok := g.StackItemSoleTargetForEffect(itemID)
	if !ok {
		return
	}
	controller := ctx.Controller()
	could := func(to uuid.UUID) bool {
		if item.Kind == game.StackItemSpell {
			return g.SpellCouldTargetInsteadForEffect(itemID, controller, sole.ID, to)
		}
		return g.AbilityCouldTargetInsteadForEffect(itemID, controller, sole.ID, to)
	}
	var tos []uuid.UUID
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated && p.ID != sole.ID && could(p.ID) {
			tos = append(tos, p.ID)
		}
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != sole.ID && could(c.InstanceID) {
			tos = append(tos, c.InstanceID)
		}
	}
	if len(tos) == 0 {
		return
	}
	carry := append([]uuid.UUID{itemID, sole.ID}, tos...)
	copyForEachNextOrLast(g, controller, ctx.Source(), name, carry, 0)
}

// copyForEachNextOrLast makes the copies once at most one target is left
// to place, and otherwise asks which goes next. carry is the item, its
// single target, then the targets in the chosen order followed by those
// still to place; placed is how many are in order.
func copyForEachNextOrLast(g *game.Game, controller, source uuid.UUID, name string, carry []uuid.UUID, placed int) {
	tos := carry[2:]
	if len(tos)-placed <= 1 {
		itemID, from := carry[0], carry[1]
		if it := g.StackItemForEffect(itemID); it != nil && it.Kind != game.StackItemSpell {
			g.CopyAbilityRetargetedEachForEffect(itemID, controller, from, tos)
			return
		}
		g.CopySpellRetargetedEachForEffect(itemID, controller, from, tos)
		return
	}
	var opts []game.ChoiceOption
	for _, id := range tos[placed:] {
		if p := g.PlayerByIDForEffect(id); p != nil {
			opts = append(opts, game.ChoiceOption{Label: p.Name, Player: id})
			continue
		}
		label := "a permanent"
		if c, ok := g.LookupCardForEffect(id); ok {
			label = c.Name
		}
		opts = append(opts, game.ChoiceOption{Label: label, Cards: []uuid.UUID{id}})
	}
	g.QueueOptionPickForEffect(game.OptionPickPrompt{
		Chooser:  controller,
		Source:   source,
		Question: name + " — choose the target of the next copy to go on the stack (the first goes on first and resolves last)",
		Options:  opts,
		ThenKey:  copyForEachOrderThen,
		Carry:    carry,
		Words:    []string{name},
		Nums:     []int{placed},
	})
}

// copyForEachOrderThen places the chosen target next and asks again, or
// makes the copies. Unanswered, the chooser has left the game, and so
// would the copies they controlled (CR 800.4a).
//
// It is set by the registration below rather than by its own initializer,
// because the continuation queues the next question under the same key and
// a self-referencing initializer is an initialization cycle.
var copyForEachOrderThen game.OptionPickThen

var _ = registerCopyForEachOrder()

func registerCopyForEachOrder() bool {
	copyForEachOrderThen = game.RegisterOptionPickThen("option-pick/copy-for-each-could-target-order", copyForEachPlaceNext)
	return true
}

// copyForEachPlaceNext is the continuation's body.
func copyForEachPlaceNext(g *game.Game, r game.OptionPicked) error {
	if r.Option == nil || len(r.Carry) < 3 || len(r.Nums) != 1 || len(r.Words) != 1 {
		return nil
	}
	chosen := r.Option.Player
	if len(r.Option.Cards) == 1 {
		chosen = r.Option.Cards[0]
	}
	placed := r.Nums[0]
	carry := append([]uuid.UUID(nil), r.Carry...)
	tos := carry[2:]
	for i := placed; i < len(tos); i++ {
		if tos[i] == chosen {
			tos[placed], tos[i] = tos[i], tos[placed]
			copyForEachNextOrLast(g, r.Chooser, r.Source, r.Words[0], carry, placed+1)
			return nil
		}
	}
	return nil
}
