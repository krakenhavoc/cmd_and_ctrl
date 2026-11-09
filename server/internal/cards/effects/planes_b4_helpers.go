package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// planes_b4_helpers.go holds the shapes planes batches 4 and 5 share
// (ADR 0216 slice 4, #4870). Append-only.

// CondNoMonarch is "if there is no monarch" (Paliano), the intervening
// "if" of a trigger. It reads the same state CondMonarch does, the
// crown, so it declares the same footprint read.
type CondNoMonarch struct{}

func (CondNoMonarch) Holds(ctx *Context) bool { return ctx.Game.Monarch == uuid.Nil }

func (CondNoMonarch) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondMonarch, 0), true
}

// onlyIfThereIsNoMonarch is the trigger-time half of "if there is no
// monarch" (CR 603.4): the ability does not trigger while someone holds the
// crown. InterveningIf supplies the resolution-time half.
func onlyIfThereIsNoMonarch(t game.TriggeredAbility) game.TriggeredAbility {
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return g.Monarch == uuid.Nil && (when == nil || when(ev, source, lki, g))
	}
	return t
}

// AnyPermanentCountersDoubled is "If an effect would put one or more
// counters on a permanent, it puts twice that many of those counters on
// that permanent instead" (Selesnya Loft Gardens): Doubling Season's
// counter half with the "you control" clause dropped, so it doubles every
// permanent's, whoever controls it. The same exclusions (CR 614.1): a
// removal, combat damage's wither and infect counters, a Saga's lore
// counter and a paid cost are not "an effect" putting counters.
func AnyPermanentCountersDoubled(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventCounterPlaced},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
			if ev.Kind != game.RepEventCounter || ev.CounterDelta <= 0 {
				return false
			}
			if ev.CounterFromCombatDamage || ev.CounterFromCost || ev.CounterFromTurnBasedAction {
				return false
			}
			_, ok := g.LookupCardForEffect(ev.CounterTarget)
			return ok
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.CounterDelta *= 2
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}

// TurnManaTriggerAnyLand is "until end of turn, whenever you tap a land
// for mana, add one mana of any type that land produced" (Selesnya Loft
// Gardens' chaos ability): a triggered mana ability (CR 605.1b) set up for
// the turn, for any land. Two resolutions are two records, so it is
// cumulative, per the ruling.
type TurnManaTriggerAnyLand struct {
	Label string
}

func (t TurnManaTriggerAnyLand) Apply(ctx *Context) error {
	ctx.Game.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:            ctx.Controller(),
		SourceCardID:          ctx.Source(),
		Label:                 t.Label,
		ManaAddsProduced:      true,
		Repeats:               true,
		ManaTapControllerOnly: true,
	})
	return nil
}

// BlitzForCreatureSpellsFromHand is The Caldaia's static: "Creature spells
// you cast from your hand have blitz {3}." A granted blitz offer (CR
// 702.152a) over creature spells cast from the hand, whose clause is blitz's
// own (game.AltCostKeyGrantedBlitz).
func BlitzForCreatureSpellsFromHand(cost string) game.GrantedAlternativeCost {
	return game.GrantedAlternativeCost{
		Offer: game.AlternativeCost{
			Key:      game.AltCostKeyGrantedBlitz,
			Label:    "Blitz " + cost,
			ManaCost: cost,
			Blitz:    true,
		},
		Zones:  []game.ZoneKind{game.ZoneHand},
		Spells: game.PermissionFilter{CreatureOnly: true},
	}
}

// revealUntilToHandRestOnBottom is "reveal cards from the top of your
// library until you reveal a <match> card. Put that card into your hand and
// the rest on the bottom of your library in a random order" (New Argive's
// chaos ability). A library with no such card reveals itself and every card
// goes to the bottom in a random order: "that card" does not exist.
func revealUntilToHandRestOnBottom(ctx *Context, match func(game.Card) bool, reason string) error {
	player := ctx.Controller()
	run, hit := revealUntil(ctx, player, match, reason)
	var rest []uuid.UUID
	for _, id := range run {
		if id != hit {
			rest = append(rest, id)
		}
	}
	if hit == uuid.Nil {
		return ctx.Game.PutOnBottomInRandomOrderForEffect(player, game.ZoneLibrary, rest)
	}
	return ctx.Game.TakeFromLibraryToHandKeyedForEffect(player, []uuid.UUID{hit}, withCarry(ctx.ItemTail(revealedRestOnBottomInRandomOrder), rest...))
}

// revealedRestOnBottomInRandomOrder is the rest of a reveal-until put into
// the hand: the cards still in the library go to the bottom in a random
// order. An on-disk key: never renamed, never reused.
var revealedRestOnBottomInRandomOrder = CardTail("planes/reveal-until-to-hand-rest-on-bottom", func(ctx *Context, o game.TailOutcome) error {
	return ctx.Game.PutOnBottomInRandomOrderForEffect(ctx.Controller(), game.ZoneLibrary, cardsStillInALibrary(ctx.Game, o.Tail.Carry))
})
