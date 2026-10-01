package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Lord of Pain — Legendary Creature — Human Assassin {3}{B}{R}, 5/5:
//
//	"Menace
//	 Your opponents can't gain life.
//	 Whenever a player casts their first spell each turn, choose another
//	 target player. The Lord of Pain deals damage equal to that spell's
//	 mana value to the chosen player."
//
// Menace rides PrintedKeywords; "Your opponents can't gain life" is
// ADR 0107 §5's battlefield static (CR 119.7, #1880).
//
// "Their first spell each turn" is the caster's per-turn cast tally,
// which the cast path bumps before EventCast, so a total of exactly one
// is the first (Maelstrom Nexus's clock). Any player's first spell,
// the controller's included.
//
// "Another target player" is a player other than the CASTER — the
// player the event names — not other than the Lord's controller, so the
// clause is built per trigger from the cast event (TargetsFrom, which
// reads only the event and nothing the board can change). The damage is
// the spell's mana value read as the trigger resolves, with its
// announced X while it is still on the stack (triggeringSpellManaValue).
//
// No simplification.
func init() {
	t := On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
		return ev.Actor != uuid.Nil && g.CastTallyFor(ev.Actor).Total == 1
	}, "The Lord of Pain — damage equal to that spell's mana value to another target player", theLordOfPainDamage)
	t.TargetsFrom = theLordOfPainTargets
	Register(Spec{
		OracleID:        "b194f277-5045-4391-8b48-d555a6d657a7",
		Name:            "The Lord of Pain",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		CantGainLife:    OpponentsCantGainLife(),
		Triggered:       []game.TriggeredAbility{t},
	})
}

// theLordOfPainTargets is "another target player": any player but the
// one who cast the spell.
func theLordOfPainTargets(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
	caster := tc.Event.Actor
	return TargetPlayer("another target player", func(_ *game.Game, _ uuid.UUID, p *game.Player) bool {
		return p.ID != caster
	})
}

// theLordOfPainDamage deals the spell's mana value to the chosen player,
// if that player is still a legal target (CR 608.2b).
func theLordOfPainDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	amount := triggeringSpellManaValue(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: amount}.Apply(ctx)
		}
	}
	return nil
}
