package cardsx

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zevlor, Elturel Exile — Legendary Creature — Tiefling Warrior {1}{U}{B}{R},
// 4/2 (Commander Legends: Battle for Baldur's Gate, S82 GAP2 b05 #5948,
// #6094):
//
//	"Haste
//	 {2}, {T}: When you next cast an instant or sorcery spell that targets
//	 only a single opponent or a single permanent an opponent controls this
//	 turn, for each other opponent, choose that player or a permanent they
//	 control, copy that spell, and the copy targets the chosen player or
//	 permanent."
//
// The ability sets up a CR 603.7 delayed trigger on its controller's next
// cast this turn of an instant or sorcery spell whose targets are one player
// or object only (CR 115.9c), that being an opponent or a permanent an
// opponent controls; a cast that does not match neither fires nor uses it
// up. As it resolves, the opponent the spell was aimed at is excluded, and
// for each other opponent in turn the controller chooses that player or one
// of their permanents (a keyed option pick); the copy is then made with
// every target being the chosen one (CR 707.10e: if that one isn't a legal
// target for each instance of "target", the copy isn't created). The spell
// may have left the stack: it is copied as it last existed (CR 608.2h).
//
// No simplification.
func init() {
	zevlorChosen = effects.OptionPickThen("option-pick/zevlor-chosen", zevlorChosenBody)
	effects.Register(effects.Spec{
		OracleID:        "0c85a577-db82-4a36-bc42-49644eba1cf2",
		Name:            "Zevlor, Elturel Exile",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: When you next cast an instant or sorcery spell that targets only a single opponent or a single permanent an opponent controls this turn, for each other opponent, choose that player or a permanent they control, copy that spell, and the copy targets the chosen player or permanent.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				return effects.DelayedOnEvent{
					Label:     "Zevlor, Elturel Exile — copy that spell for each other opponent",
					On:        []game.EventKind{game.EventCast},
					Condition: zevlorCondition,
					Body:      zevlorBody,
				}.Apply(ctx)
			})),
		}},
	})
}

var (
	// zevlorCondition is the next instant or sorcery cast by the trigger's
	// controller that targets only a single opponent or a single permanent
	// an opponent controls. An on-disk key.
	zevlorCondition = game.DelayedCondition("zevlor/next-single-opponent-spell",
		func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
			if ev.Kind != game.EventCast || ev.Actor != dt.Controller {
				return false
			}
			c, ok := g.LookupCardForEffect(ev.CardID)
			if !ok || (!c.IsInstant() && !c.IsSorcery()) {
				return false
			}
			sole, ok := g.StackItemSoleTargetForEffect(ev.CardID)
			if !ok {
				return false
			}
			opp := zevlorTargetedOpponent(g, sole)
			return opp != uuid.Nil && opp != dt.Controller
		})

	// zevlorBody starts the per-opponent choices. An on-disk key.
	zevlorBody = game.SimpleDelayedBody("zevlor/copy-for-each-other-opponent", zevlorCopies)

	// zevlorChosen copies the spell at the chosen player or permanent and
	// asks the next opponent's question. An on-disk key, registered in init
	// because the chain asks it again.
	zevlorChosen game.OptionPickThen
)

func zevlorChosenBody(ctx *effects.Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 {
		return nil
	}
	spell, rest := r.Carry[0], r.Carry[1:]
	if o := r.Option; o != nil {
		switch {
		case o.Player != uuid.Nil:
			ctx.Game.CopySpellWithNewTargetForEffect(spell, ctx.Controller(), game.TargetRef{Kind: game.TargetPlayer, ID: o.Player})
		case len(o.Cards) == 1:
			ctx.Game.CopySpellWithNewTargetForEffect(spell, ctx.Controller(), game.TargetRef{Kind: game.TargetCard, ID: o.Cards[0]})
		}
	}
	return zevlorAsk(ctx, spell, rest)
}

// zevlorTargetedOpponent is the player a sole target names: the player
// itself, or the controller of a permanent on the battlefield. uuid.Nil for
// anything else (a card in a graveyard, a spell, a player who has left).
func zevlorTargetedOpponent(g *game.Game, t game.TargetRef) uuid.UUID {
	switch t.Kind {
	case game.TargetPlayer:
		if p := g.PlayerByIDForEffect(t.ID); p != nil && !p.Eliminated {
			return p.ID
		}
	case game.TargetCard:
		if !effects.OnBattlefield(g, t.ID) {
			return uuid.Nil
		}
		if c, ok := g.LookupCardForEffect(t.ID); ok {
			return c.Controller
		}
	}
	return uuid.Nil
}

func zevlorCopies(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	spell := ctx.Trigger().Event.CardID
	if spell == uuid.Nil {
		return nil
	}
	var aimed uuid.UUID
	if sole, ok := g.SpellSoleTargetForEffect(spell); ok {
		aimed = zevlorTargetedOpponent(g, sole)
		if aimed == uuid.Nil && sole.Kind == game.TargetPlayer {
			aimed = sole.ID
		}
	}
	var others []uuid.UUID
	for _, opp := range ctx.Opponents() {
		if opp != aimed {
			others = append(others, opp)
		}
	}
	return zevlorAsk(ctx, spell, others)
}

// zevlorAsk asks the controller, for the first of `opponents`, to choose
// that player or a permanent they control, carrying the spell and the rest.
func zevlorAsk(ctx *effects.Context, spell uuid.UUID, opponents []uuid.UUID) error {
	for len(opponents) > 0 {
		opp, rest := opponents[0], opponents[1:]
		p := ctx.Game.PlayerByIDForEffect(opp)
		if p == nil || p.Eliminated {
			opponents = rest
			continue
		}
		options := []game.ChoiceOption{{Label: p.Name, Player: opp}}
		for _, c := range ctx.Game.Battlefield.Cards {
			if c.Controller == opp {
				options = append(options, game.ChoiceOption{Label: c.Name, Cards: []uuid.UUID{c.InstanceID}})
			}
		}
		return effects.PickOption{
			Question: "Zevlor — choose " + p.Name + " or a permanent they control; a copy of that spell targets it",
			Options:  options,
			ThenKey:  zevlorChosen,
			Carry:    append([]uuid.UUID{spell}, rest...),
		}.Apply(ctx)
	}
	return nil
}
