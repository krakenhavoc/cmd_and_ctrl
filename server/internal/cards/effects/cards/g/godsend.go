package cardsg

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Godsend — Legendary Artifact — Equipment {1}{W}{W} (Foundations
// Jumpstart):
//
//		"Equipped creature gets +3/+3.
//		 Whenever equipped creature blocks or becomes blocked by one or more
//		 creatures, you may exile one of those creatures.
//		 Your opponents can't cast spells with the same name as a card exiled
//		 with Godsend.
//		 Equip {3}"
//
//	  - The trigger is once per block, not per pair (CR 509.3a, 509.3c): the
//	    equipped blocker's first EventBlock (selfBlocksOnce's Amount rule) or
//	    the equipped attacker's one EventBecomesBlocked. "Those creatures" are
//	    read as it resolves: the attackers the equipped creature blocks, or
//	    the creatures blocking it. The choice is a "may" over them, up to one,
//	    and not a target.
//	  - "Exiled with Godsend" is the event-log record keyed by the trigger's
//	    label (b27ExiledWith), so a card another effect exiled isn't in it, a
//	    card that left exile no longer counts, and a Godsend that leaves and
//	    returns remembers nothing (CR 400.7).
//	  - The cast ban is a CastRestriction read live off the record: a static,
//	    so its presence is its duration, and it binds opponents only.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dd91f4f2-ea46-48c5-8a80-0dc4ae5b79c4",
		Name:         "Godsend",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(3, 3)},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventBlock, game.EventBecomesBlocked}, godsendTriggers,
				godsendLabel, func(g *game.Game, item *game.StackItem) error {
					return godsendResolve(effects.NewContext(g, item))
				}),
		},
		CastRestrictions: []game.CastRestriction{{
			Label: "Your opponents can't cast spells with the same name as a card exiled with Godsend.",
			Forbids: func(q game.CastQuery) bool {
				if q.Source.Controller == q.Controller {
					return false
				}
				for _, id := range effects.B27ExiledWith(q.Game, q.Source.InstanceID, godsendLabel) {
					for _, c := range q.Game.Exile.Cards {
						if c.InstanceID == id && strings.EqualFold(c.Name, q.Card.Effective().Name) {
							return true
						}
					}
				}
				return false
			},
		}},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}

// godsendLabel is the trigger's stack label: the key its exiled-with
// record is read by.
const godsendLabel = "Godsend — you may exile one of those creatures"

// godsendTriggers: the equipped creature blocks (once, however many it
// blocks) or becomes blocked (once, however many block it).
func godsendTriggers(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	switch ev.Kind {
	case game.EventBlock:
		return ev.Amount <= 1 && source.IsAttachedTo(ev.CardID)
	case game.EventBecomesBlocked:
		return source.IsAttachedTo(ev.CardID)
	}
	return false
}

// godsendThose is "those creatures": what the equipped creature blocks, or
// what blocks it, still on the battlefield.
func godsendThose(g *game.Game, ev game.Event) []uuid.UUID {
	var out []uuid.UUID
	equipped, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return nil
	}
	if ev.Kind == game.EventBlock {
		onBattlefield := map[uuid.UUID]bool{}
		for _, c := range g.BattlefieldCardsForEffect() {
			onBattlefield[c.InstanceID] = true
		}
		for _, a := range equipped.BlockedAttackers() {
			if onBattlefield[a] {
				out = append(out, a)
			}
		}
		return out
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsBlockingAttacker(ev.CardID) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

func godsendResolve(ctx *effects.Context) error {
	if ctx.Item.Trigger == nil {
		return nil
	}
	those := godsendThose(ctx.Game, ctx.Item.Trigger.Event)
	if len(those) == 0 {
		return nil
	}
	return effects.RevealPick{
		Player:   ctx.Controller(),
		Question: "Godsend — you may exile one of those creatures",
		Cards:    those,
		Min:      0,
		Max:      1,
		Keyed:    godsendExiled,
	}.Apply(ctx)
}

// godsendExiled is the exile after the pick. An on-disk key: never
// renamed, never reused.
var godsendExiled = effects.RegisterRevealPick("choose-cards/godsend-exile", func(ctx *effects.Context, picked, _ []uuid.UUID, _ []int) error {
	for _, id := range picked {
		if err := (effects.ExileTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
