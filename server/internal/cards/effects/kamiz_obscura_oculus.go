package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kamiz, Obscura Oculus — Legendary Creature — Octopus Rogue {1}{W}{U}{B},
// 2/4 (Streets of New Capenna Commander, #5491):
//
//	"Whenever you attack, target attacking creature can't be blocked this
//	 turn. It connives. Then choose another attacking creature with lesser
//	 power. That creature gains double strike until end of turn. (To have a
//	 creature connive, draw a card, then discard a card. If you discarded a
//	 nonland card, put a +1/+1 counter on that creature.)"
//
// "Whenever you attack" is one trigger per declaration (OncePerBatch, Bast's
// shape). Targeting an attacking creature, it first can't be blocked this
// turn, then connives: the draw and discard are the connive keyword's own
// (conniveDrawAndDiscard, so Leader, Super-Genius's replacement applies),
// and the counter lands before anything is chosen. Then the controller
// chooses, as an instruction of its own and not a target, another attacking
// creature whose power is less than the connived creature's power NOW (the
// counter included; its last known power if it has left, CR 608.2h), and
// that creature gains double strike until end of turn. Nothing is chosen
// when no attacker qualifies. If the connive was cancelled by a replacement
// the choice still follows (CR 701.50e: the rest of the sentence runs).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "49da6a78-f3b2-4691-975c-efbb644311f3",
		Name:         "Kamiz, Obscura Oculus",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(Targeting(
				On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclaredByYou(ev, source.Controller)
				}, "Kamiz, Obscura Oculus — target attacking creature can't be blocked this turn and connives; then choose another attacking creature with lesser power, which gains double strike until end of turn",
					kamizResolve),
				TargetCreature("target attacking creature", AttackingCreature()))),
		},
	})
}

func kamizResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	if err := (RestrictUntilEOT{
		Target:       target,
		Restrictions: game.CantBeBlocked,
		Label:        "Kamiz, Obscura Oculus — can't be blocked",
	}).Apply(ctx); err != nil {
		return err
	}
	return g.ConniveForEffect(item.Controller, target, 1, game.KeyedTail{
		Ref:    kamizConniveTail,
		Player: item.Controller,
		Source: item.SourceCardID,
		Carry:  []uuid.UUID{target},
	})
}

// kamizConniveTail runs once the connive's window has settled: a connive
// that happened draws and discards, and either way the choice follows. An
// on-disk key: never renamed, never reused.
var kamizConniveTail = game.RegisterTail("connive/kamiz-obscura-oculus", func(g *game.Game, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	target := o.Tail.Carry[0]
	if o.Tail.N <= 0 {
		return kamizChoose(g, o.Tail.Player, target)
	}
	return conniveDrawAndDiscard(g, o.Tail.Player, target, o.Amount, o.Tail.N,
		"Connive: discard a card", kamizDiscarded, nil)
})

// kamizDiscarded is the counter for a nonland discard, the creature having
// connived, then the choice. An on-disk key: never renamed, never reused.
var kamizDiscarded = game.RegisterDiscardThen("choose-cards/kamiz-obscura-oculus-discarded", func(g *game.Game, r game.DiscardsDone) error {
	creature := r.Source
	if discarded := r.Discarded.By(r.Player); len(discarded) > 0 {
		if card, ok := g.LookupCardForEffect(discarded[0]); ok && !card.IsLand() && onBattlefield(g, creature) {
			if err := g.AddCounterForEffect(creature, game.CounterPlusOne, 1); err != nil {
				return err
			}
		}
	}
	g.ConnivedForEffect(creature, r.Player)
	return kamizChoose(g, r.Player, creature)
})

// kamizChoose asks for another attacking creature with lesser power than
// the connived one.
func kamizChoose(g *game.Game, player, connived uuid.UUID) error {
	power := b43PowerNowOrLastKnown(g, connived)
	var candidates []uuid.UUID
	for _, id := range AttackingCreatures(g) {
		if id == connived {
			continue
		}
		if c, ok := g.LookupCardForEffect(id); ok && c.CurrentPower() < power {
			candidates = append(candidates, id)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   connived,
		Question: "Kamiz, Obscura Oculus — choose another attacking creature with lesser power; it gains double strike until end of turn",
		Cards:    candidates,
		Min:      1,
		Max:      1,
		ThenKey:  kamizDoubleStrike,
		LiveZone: game.ZoneBattlefield,
	})
	return nil
}

// kamizDoubleStrike gives the chosen creature double strike until end of
// turn. An on-disk key: never renamed, never reused.
var kamizDoubleStrike = askedPick("choose-cards/kamiz-obscura-oculus-double-strike",
	func(ctx *Context, r game.CardsChosen) error {
		for _, id := range r.Picked {
			if !onBattlefield(ctx.Game, id) {
				continue
			}
			if err := (GrantKeywordUntilEOT{
				Target:   id,
				Keywords: []string{"double strike"},
				Label:    "Kamiz, Obscura Oculus — double strike",
			}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
