package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Giada, Font of Hope — Legendary Creature — Angel {1}{W}, 2/2
// (Foundations, #4465):
//
//	"Flying, vigilance
//	 Each other Angel you control enters with an additional +1/+1
//	 counter on it for each Angel you already control.
//	 {T}: Add {W}. Spend this mana only to cast an Angel spell."
//
// The counters are Tromell's CR 614 entry replacement narrowed to other
// Angels you control, sized by the Angels its controller already controls
// as the entry is replaced: the newcomer is not on the battlefield yet, so
// it is not counted, and Giada herself is (the ruling), whenever each
// entered. The counters ride the entry event, so Hardened Scales and
// Doubling Season see them. "Angel" is a word of Giada's text (CR 612.2).
// The mana is Eldrazi Temple's shape: {W} spendable only on casting an
// Angel spell, read off the spell as it is paid for (CR 106.6). It is a
// creature, so the {T} ability waits out summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "48e6d3d8-2f27-4017-acdd-40bce8cdbc02",
		Name:            "Giada, Font of Hope",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Replacements:    []game.ReplacementEffect{giadaAngelCounters()},
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			Produced:     "{W}",
			Label:        "{T}: Add {W}. Spend this mana only to cast an Angel spell.",
			Restrictions: []string{game.ManaRestrictCast, game.ManaRestrictSubtype("Angel")},
		}},
	})
}

// giadaAngelCounters is "each other Angel you control enters with an
// additional +1/+1 counter on it for each Angel you already control".
func giadaAngelCounters() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if src == nil || ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || ev.CardID == src.InstanceID {
				return false
			}
			c, ok := g.LookupCardForEffect(ev.CardID)
			return ok && c.IsCreature() && c.Controller == src.Controller && c.HasSubtypeWord("Angel")
		},
		Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
			if n := len(permanentsControlledByMatching(g, src.Controller, And(Creature(), Subtype("Angel")))); n > 0 {
				ev.AddCounterAtETB(game.CounterPlusOne, n)
			}
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: "Giada, Font of Hope — each other Angel you control enters with an additional +1/+1 counter for each Angel you already control",
	}
}
