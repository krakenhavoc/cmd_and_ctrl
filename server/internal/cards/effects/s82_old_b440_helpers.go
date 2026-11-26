package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b440_helpers.go — shared pieces of Dark Ascension batch 440 (#5571).

// S82OldB440CursesAttachedToPlayer counts the Curses on the battlefield
// attached to a player, through the layered view (a permanent that is no
// longer a Curse does not count).
func S82OldB440CursesAttachedToPlayer(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.AttachedTo.Kind == game.TargetPlayer && c.AttachedTo.ID == player && c.HasSubtype("Curse") {
			n++
		}
	}
	return n
}

// S82OldB440SpitefulShadowsReflect is the reflect trigger's body: the damaged
// creature, named by the triggering event, deals the damage it was dealt (the
// batch total) to its controller, read as it was as the damage landed.
func S82OldB440SpitefulShadowsReflect(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ev := item.Trigger.Event
	ctx := NewContext(g, item)
	controller := uuid.Nil
	if c, ok := g.LookupCardForEffect(ev.Target); ok {
		controller = c.Controller
	}
	if controller == uuid.Nil || ev.Amount <= 0 {
		return nil
	}
	return DealDamage{Source: ev.Target, Target: controller, Amount: ev.Amount}.Apply(ctx)
}

// S82OldB440EachOtherPlayerMayCopy is "each other player may copy that spell
// and may choose new targets for the copy they control" (Curse of Echoes): the
// triggering spell is held for last-known information, then every player but
// the one who cast it is asked in turn order from the active player, each
// answer being that player's own copy. `answer` is the registered MayAnswer
// that makes the copy for r.Chooser, given the spell as its Carry.
func S82OldB440EachOtherPlayerMayCopy(answer game.ConfirmThen) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil {
			return nil
		}
		ctx := NewContext(g, item)
		spellID := item.Trigger.Event.CardID
		caster := item.Trigger.Event.Actor
		g.HoldSpellForEffect(spellID)
		n := len(g.Seats)
		for i := 0; i < n; i++ {
			p := g.Seats[(g.Turn.ActiveSeat+i)%n]
			if p == nil || p.Eliminated || p.ID == caster {
				continue
			}
			if err := (MayChoice{
				Player:   p.ID,
				Question: "Curse of Echoes — copy that spell?",
				Then:     answer,
				Carry:    []uuid.UUID{spellID},
			}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}
