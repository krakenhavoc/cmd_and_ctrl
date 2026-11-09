package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jund — Plane — Alara (Commander Masters, #4870):
//
//	"Whenever a player casts a creature spell that's black, red, or green,
//	 it gains devour 5. (As that creature enters, you may sacrifice any
//	 number of creatures. It enters with five times that many +1/+1
//	 counters on it.)
//	 Whenever chaos ensues, create two 1/1 red Goblin creature tokens."
//
// A plane (ADR 0216 §6). The first ability triggers for any player's creature
// spell of one of the three colors (a multicolored one counts, the ruling) and
// gives that spell the keyword "devour 5" for as long as it is that object
// (the spell, then the permanent it becomes, CR 400.7a), like any "that spell
// gains" effect. Devour is an as-enters replacement (CR 614.12) with a choice
// inside it, so the entering creature is what carries it: the plane's
// replacement applies to a creature entering with the granted keyword, asks its
// controller which creatures to sacrifice, and enters it with five +1/+1
// counters for each. A creature that has devour itself works separately for
// each instance, and each creature is sacrificed for only one (the ruling).
// The chaos ability (CR 311.7) makes two Goblins for the planar controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9aa94de8-b114-4479-a771-c7bc0297be58",
		Name:         "Jund",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{jundDevour()},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				spell, ok := g.LookupCardForEffect(ev.CardID)
				if !ok || !spell.HasCardType("creature") {
					return false
				}
				for _, c := range spell.EffectiveColors() {
					if c == "B" || c == "R" || c == "G" {
						return true
					}
				}
				return false
			}, "Jund — that spell gains devour 5", func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil {
					return nil
				}
				g.GrantKeywordsToSpellForEffect(item.SourceCardID, item.Trigger.Event.CardID, []string{jundDevourKeyword},
					game.IndefiniteDuration(), "Jund — devour 5")
				return nil
			}),
			WheneverChaosEnsues("Jund — create two 1/1 red Goblin creature tokens",
				Do(CreateToken{Template: RedGoblinToken(), N: 2})),
		},
	})
}

// jundDevourKeyword is the keyword token a spell Jund gave devour carries.
// An on-disk identity: never renamed, never reused.
const jundDevourKeyword = "devour 5"

// jundDevour is devour 5 for a creature entering with the granted keyword:
// the devour replacement (devourOver) made non-self, since its source is the
// plane and not the entering card.
func jundDevour() game.ReplacementEffect {
	r := devourOver("Jund", 5, "creatures", func(c game.Card) bool { return c.IsCreature() })
	r.SelfReplacement = false
	r.EntryOnly = false
	r.AppliesTo = func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
		if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || ev.CardID == uuid.Nil {
			return false
		}
		entering, ok := g.LookupCardForEffect(ev.CardID)
		return ok && entering.IsCreature() && game.HasKeyword(&entering, jundDevourKeyword)
	}
	return r
}
