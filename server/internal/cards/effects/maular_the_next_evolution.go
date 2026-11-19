package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Maular, the Next Evolution — {5}{G}{G} Legendary Creature — Dinosaur
// Mutant 10/10 (Mystery Booster Commander, S82 GAP batch 3, #5948):
//
//	"Rulebreaker — A deck with this commander can have creature cards
//	 with mana value 7 or greater of any color identity and any basic
//	 land cards.
//	 Whenever a creature you control with mana value 7 or greater
//	 attacks, double its power and toughness until end of turn."
//
// The trigger is one per attacking creature of yours with mana value 7
// or greater (Maular included), doubling the attacker's power and
// toughness as the ability resolves (CR 701.10b). The Rulebreaker
// paragraph is a deckbuilding exception, read by deck validation off the
// oracle text (internal/deck, rulebreaker.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "214f4e0a-eba7-4a08-86b2-d490d6ae1048",
		Name:         "Maular, the Next Evolution",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !attackDeclaredByYou(ev, source.Controller) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.ManaValue() >= 7
			}, "Maular, the Next Evolution — double the attacking creature's power and toughness",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || !onBattlefield(g, item.Trigger.Event.CardID) {
						return nil
					}
					return DoublePowerAndToughnessOfEachUntilEOT(NewContext(g, item), []uuid.UUID{item.Trigger.Event.CardID},
						"Maular, the Next Evolution — double power and toughness")
				}),
		},
	})
}
