package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_b23_helpers.go — helpers for the +1/+1 counter commander cards of
// cmdr2 batch 23 (#5368).

// cmdr2B23HasPlusOneCounter reports whether the card `id` is on the
// battlefield with at least one +1/+1 counter, read now.
func cmdr2B23HasPlusOneCounter(g *game.Game, id uuid.UUID) bool {
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
		return false
	}
	c, ok := g.LookupCardForEffect(id)
	return ok && c.Counters[game.CounterPlusOne] > 0
}

// cmdr2B23OtherWizardsEnterWithACounter is Sage of Fables' CR 614
// entry replacement: "each other Wizard creature you control enters
// with an additional +1/+1 counter on it" (Grumgully's shape).
func cmdr2B23OtherWizardsEnterWithACounter(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || src == nil || ev.CardID == src.InstanceID {
				return false
			}
			entering, ok := g.LookupCardForEffect(ev.CardID)
			return ok && entering.IsCreature() && entering.Controller == src.Controller && entering.HasSubtype("Wizard")
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.AddCounterAtETB(game.CounterPlusOne, 1)
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}

// cmdr2B23AnotherCreatureWithPlusOne is "you control another creature
// with a +1/+1 counter on it": a creature `you` control other than `self`.
func cmdr2B23AnotherCreatureWithPlusOne(g *game.Game, you, self uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != self && c.IsCreature() && c.Controller == you && c.Counters[game.CounterPlusOne] > 0 {
			return true
		}
	}
	return false
}

// cmdr2B23CreatureYouControlWithPlusOne is the CardPredicate for "a
// creature you control with a +1/+1 counter on it", read live.
func cmdr2B23CreatureYouControlWithPlusOne(_ *game.Game, you uuid.UUID, c game.Card) bool {
	return c.IsCreature() && c.Controller == you && c.Counters[game.CounterPlusOne] > 0
}

// cmdr2B23GrowthCounter is Simic Ascendancy's counter kind.
const cmdr2B23GrowthCounter = "growth"
