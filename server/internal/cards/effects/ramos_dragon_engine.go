package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ramos, Dragon Engine — Legendary Artifact Creature — Dragon {6},
// 4/4 (EDHREC rank 1173):
//
//	"Flying
//	 Whenever you cast a spell, put a +1/+1 counter on Ramos, Dragon
//	 Engine for each of that spell's colors.
//	 Remove five +1/+1 counters from Ramos: Add {W}{W}{U}{U}{B}{B}{R}{R}{G}{G}.
//	 Activate only once each turn."
//
// The five-colour commander that banks counters and cashes them for
// ten mana. Three pieces, all live since #789 gave a MANA ability a
// counter cost:
//
//   - Flying rides PrintedKeywords.
//   - The cast trigger counts THE SPELL'S COLOURS — the card's own
//     colours, not the colours of the mana that paid for it (that is
//     converge's question, and Ramos does not ask it). A colourless
//     spell adds nothing; a five-colour one adds five. The count is
//     read as the trigger is built, while the spell is still on the
//     stack, and captured on the item: by resolution the spell may
//     have been countered or resolved away.
//   - The payout is a mana ability with NO tap cost — Ramos does not
//     tap, and a summoning-sick Ramos can still fire it the turn it
//     lands — whose whole cost is removing five +1/+1 counters.
//
// "Activate only once each turn" is ManaAbilityNotUsedThisTurn, which
// counts this permanent's EventManaAbilityActivated in the turn's own
// slice of the event log. Ramos has exactly one mana ability, so
// per-permanent and per-ability are the same question here; the
// helper says so.
//
// The auto-tapper never plans the payout: the ability costs its source
// neither a {T} nor itself (the demand autoTapAbilityFor makes since
// #1242), so it is not a planned source at all, and ten mana from five
// counters is not a decision a planner should make on the player's
// behalf.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3ed41d2d-211b-4013-8562-8c64d54cc43a",
		Name:            "Ramos, Dragon Engine",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			Key:     "Ramos, Dragon Engine — a +1/+1 counter for each of that spell's colors",
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				// "Whenever YOU cast a spell" — any spell, including
				// a colourless one (it simply adds no counters), but
				// only the controller's. Ramos is on the battlefield
				// when this fires, so it never counts its own cast.
				return source != nil && ev.Actor == source.Controller
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				n := spellColorCount(g, ev.CardID)
				return game.NewTriggeredItem(source,
					"Ramos, Dragon Engine — a +1/+1 counter for each of that spell's colors",
					func(g *game.Game, item *game.StackItem) error {
						if n <= 0 {
							return nil
						}
						return AddCounter{
							Target: item.SourceCardID,
							Kind:   game.CounterPlusOne,
							N:      n,
						}.Apply(NewContext(g, item))
					})
			},
		}},
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{
				RemoveCounters: RemoveCountersFromThis(game.CounterPlusOne, 5).RemoveCounters,
			},
			Produced:  "{W}{W}{U}{U}{B}{B}{R}{R}{G}{G}",
			Label:     "Remove five +1/+1 counters: Add {W}{W}{U}{U}{B}{B}{R}{R}{G}{G}",
			Condition: ManaAbilityNotUsedThisTurn(),
		}},
	})
}

// spellColorCount is how many colours a spell on the stack is, for
// "for each of that spell's colors" (Ramos). Zero for a colourless
// spell and for one that is no longer findable — a spell countered in
// response to the trigger is last-known information (CR 608.2h), and
// the count was taken when the trigger was built.
func spellColorCount(g *game.Game, spellID uuid.UUID) int {
	c, ok := g.LookupCardForEffect(spellID)
	if !ok {
		return 0
	}
	seen := map[string]bool{}
	for _, col := range c.Colors {
		if col != "" && col != "C" {
			seen[col] = true
		}
	}
	return len(seen)
}
