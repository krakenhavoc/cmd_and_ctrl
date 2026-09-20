package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Enduring Curiosity — Enchantment Creature — Cat Glimmer, {2}{U}{U},
// 4/3 (#321, #1107):
//
//	"Flash
//	 Whenever a creature you control deals combat damage to a player,
//	 draw a card.
//	 When Enduring Curiosity dies, if it was a creature, return it to
//	 the battlefield under its owner's control. It's an enchantment.
//	 (It's not a creature.)"
//
// Three ordinary-looking clauses, none of which turned out to be:
//
//   - Flash is PrintedKeywords, same as any other card.
//   - The draw is "A creature", not "one or more creatures" — Old
//     Gnawbone's shape, not Keeper of Fables' — so no OncePerBatch:
//     three attackers connecting is three draws.
//   - The dies half is the reason this file exists. "If it was a
//     creature" is CR 603.4's intervening-if, checked once against
//     the CR 603.10 last-known-information the harvester hands
//     AppliesTo as `lki`, not re-checked at resolution — it is a fact
//     about the object's PAST characteristics, not a live board
//     condition, so there is nothing left to re-check by the time the
//     trigger would resolve.
//
// "It's an enchantment. (It's not a creature.)" is a CR 613.3 type
// change with no stated duration but an implicit one: it is about
// THIS RETURN, so it lasts for as long as the permanent that came
// back stays on the battlefield (CR 611.2b), exactly Sower of
// Temptation's shape one layer over (2 there, 4 here). It is built
// with the same `StaticForDuration` + `DurationWhileSourceRemains`
// pair Sower uses, pinned to the returned permanent's own instance ID
// so it cannot leak onto some other Enduring Curiosity at the table or
// onto a fresh hard-cast of this one.
//
// The effect removes ONLY "Creature" from Types and leaves Subtypes
// untouched — the printed text doesn't say "loses all other types"
// (contrast Song of the Dryads' SetsBasicLandType, which does), so Cat
// and Glimmer stay on the type line same as any noncreature
// permanent's leftover creature types. If it dies again later, it
// dies as an enchantment: `lki.Types` no longer has "Creature", the
// intervening-if is false, and the ability simply doesn't trigger a
// second time — which is also the printed card's own answer.
//
// Closes #321: the card was never in the catalog (the bug's own
// triage comment confirms this — "no file in
// server/internal/cards/effects/, zero grep hits"), so there was no
// trigger to misfire. Both halves it flagged as blocked (the
// combat-damage trigger, and layer 4 not yet reading Effective()) have
// since shipped; this is the card being written now that both are
// real.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9d2460c3-8eeb-4f35-b6f6-748c478664c7",
		Name:            "Enduring Curiosity",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g)
			}, "Enduring Curiosity — draw a card", Do(DrawCards{N: 1})),
			On(game.EventLTB, func(ev game.Event, source *game.Card, lki game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source) && keywordSliceContains(lki.Types, "Creature")
			}, "Enduring Curiosity — return it to the battlefield; it's an enchantment, not a creature",
				enduringCuriosityReturnAsEnchantment),
		},
	})
}

// enduringCuriosityReturnAsEnchantment is the dies trigger's Effect: a
// package-level func, not a closure, so it captures nothing — undo
// restores a cloned game and this has to resolve against whichever one
// it is handed. item.SourceCardID already carries the instance to
// move (NewTriggeredItem stamps it from the ability's source), and the
// returned permanent keeps that SAME instance ID: CR 400.7's "new
// object" is tracked here by the (InstanceID, EnteredBattlefieldAt)
// pair the duration model already reads (see
// sameObjectOnBattlefieldLocked in game/duration.go), not by minting a
// fresh one on every zone change.
func enduringCuriosityReturnAsEnchantment(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id := item.SourceCardID
	if zone := g.FindCardZoneForEffect(id); zone == nil || zone.Kind != game.ZoneGraveyard {
		// CR 608.2b: something moved the card out of the graveyard in
		// response — exiled it, put it back on top of a library, a
		// second effect reanimated it first — and the trigger still
		// resolves; it just has nothing left to bring back.
		return nil
	}
	if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
		return err
	}
	d, ok := DurationWhileSourceRemains(ctx, id)
	if !ok {
		// It didn't actually take the battlefield — a replacement or a
		// state-based action swept it away the instant it entered —
		// so there is nothing left to keep from being a creature.
		return nil
	}
	return StaticForDuration{
		Ability: game.StaticAbility{
			Layer: game.Layer4Type,
			AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
				return target.InstanceID == id
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				kept := c.Types[:0]
				for _, t := range c.Types {
					if t != "Creature" {
						kept = append(kept, t)
					}
				}
				c.Types = kept
			},
		},
		Duration: d,
		Label:    "Enduring Curiosity — it's an enchantment (it's not a creature)",
	}.Apply(ctx)
}
