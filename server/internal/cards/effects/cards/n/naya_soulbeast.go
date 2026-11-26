package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Naya Soulbeast — Creature — Beast {6}{G}{G}, 0/0 (Commander 2013, S82 older
// sets):
//
//	"When you cast this spell, each player reveals the top card of their
//	 library. This creature enters with X +1/+1 counters on it, where X is the
//	 total mana value of all cards revealed this way.
//	 Trample"
//
// The cast trigger (Abundant Maw's WhenYouCastThisSpell) reveals the top card
// of every library in turn order and writes the total mana value to the
// spell's stack item (CastValues), where the Soulbeast's own entry
// replacement (CR 614.1c, 614.12) reads it as it enters, Gluttonous
// Hellkite's shape. A Soulbeast that is countered, or that is no longer on
// the stack when the trigger resolves, enters with none (it is a 0/0 and
// dies, CR 704.5f). The revealed cards stay where they are.
//
// No simplification.
const nayaSoulbeastKey = "revealed-mana-value"

func nayaSoulbeastCast(g *game.Game, item *game.StackItem) error {
	spell := g.StackItemForEffect(item.SourceCardID)
	total := 0
	for _, seat := range effects.ApnapPlayers(g) {
		for _, id := range g.RevealTopOfLibraryForEffect(seat, item.SourceCardID, 1, "Naya Soulbeast — revealed from the top of the library") {
			if c, ok := g.LookupCardForEffect(id); ok {
				total += c.ManaValue()
			}
		}
	}
	if spell == nil {
		return nil
	}
	if spell.CastValues == nil {
		spell.CastValues = map[string]int{}
	}
	spell.CastValues[nayaSoulbeastKey] = total
	return nil
}

func init() {
	const name = "Naya Soulbeast"
	effects.Register(effects.Spec{
		OracleID:        "5ea0c608-2c56-4889-a5d3-d435df515950",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenYouCastThisSpell(name+" — each player reveals the top card of their library; this creature enters with X +1/+1 counters, X the total mana value", nayaSoulbeastCast),
		},
		Replacements: []game.ReplacementEffect{{
			Label:     name + ": enters with X +1/+1 counters, where X is the total mana value of the cards revealed",
			EntryOnly: true,
			Watches:   []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventMove && ev.NewZone == game.ZoneBattlefield &&
					src != nil && ev.CardID == src.InstanceID
			},
			Replace: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) error {
				_, item, ok := g.ResolvingSpellForEffect(ev.CardID)
				if !ok {
					return nil
				}
				if n, ok := item.CastValue(nayaSoulbeastKey); ok && n > 0 {
					ev.AddCounterAtETB(game.CounterPlusOne, n)
				}
				return nil
			},
		}},
	})
}
