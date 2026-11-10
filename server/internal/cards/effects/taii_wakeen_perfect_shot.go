package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Taii Wakeen, Perfect Shot — Legendary Creature — Human Mercenary
// {R}{W}, 2/3 (Outlaws of Thunder Junction, #4468):
//
//	"Whenever a source you control deals noncombat damage to a creature
//	 equal to that creature's toughness, draw a card.
//	 {X}, {T}: If a source you control would deal noncombat damage to a
//	 permanent or player this turn, it deals that much damage plus X
//	 instead."
//
// THE TRIGGER. One per damage event, Niv-Mizzet, Visionary's shape
// (b34DamageSourceController reads "a source you control" off the event:
// a permanent, or the spell still on the stack). The comparison is the
// damage dealt against the creature's toughness as the damage is dealt
// (the trigger is harvested as the event is emitted, before any wither
// counters land and before state-based actions), and damage already
// marked on it does not matter (the first ruling). Combat damage never
// counts.
//
// THE ACTIVATION. A turn-long replacement that outlives Taii, so it is a
// scoped record (game.AddDamageForEffect, ModAddDamage, #4468) rather
// than a static on the card: every source its controller controls as
// the damage would be dealt, noncombat damage only, to any permanent or
// player, plus X. The extra damage is dealt by the same source (the
// second ruling), since only the amount changes. Beside another damage
// modifier the affected player orders them (CR 616.1), and when all of
// the damage is prevented there is nothing to add to (the third ruling).
// With X = 0 nothing is registered: plus 0 changes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "17327d50-be12-423d-9d37-d6bae519f50c",
		Name:         "Taii Wakeen, Perfect Shot",
		Completeness: CompletenessFull,
		XMatters:     true,
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, taiiNoncombatDamageEqualToToughness,
				"Taii Wakeen, Perfect Shot — draw a card", Do(DrawCards{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label: "{X}, {T}: If a source you control would deal noncombat damage to a permanent or player this turn, it deals that much damage plus X instead.",
			// It answers nothing by itself: it changes the damage a
			// later source deals (ADR 0142's value).
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{X}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				g.AddDamageForEffect(game.DamageAddition{
					EffectSource:  ctx.Source(),
					Controller:    ctx.Controller(),
					Amount:        ctx.X(),
					Sources:       game.DamageSourcesYours,
					Recipients:    game.DamageRecipientsAny,
					NoncombatOnly: true,
					Label:         "Taii Wakeen, Perfect Shot",
				})
				return nil
			},
		}},
	})
}

// taiiNoncombatDamageEqualToToughness is "a source you control deals
// noncombat damage to a creature equal to that creature's toughness".
func taiiNoncombatDamageEqualToToughness(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Combat || ev.Amount <= 0 {
		return false
	}
	controller, ok := b34DamageSourceController(ev, g)
	if !ok || controller != source.Controller {
		return false
	}
	if !b15OnBattlefield(g, ev.Target) {
		return false
	}
	victim, ok := g.LookupCardForEffect(ev.Target)
	return ok && victim.IsCreature() && ev.Amount == victim.CurrentToughness()
}
