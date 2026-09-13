package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lae'zel, Vlaakith's Champion — Legendary Creature — Gith Warrior
// {2}{W}, 3/3 (EDHREC rank 1486):
//
//	"If you would put one or more counters on a creature or
//	 planeswalker you control or on yourself, put that many plus one
//	 of each of those kinds of counters on that permanent or player
//	 instead.
//	 Choose a Background (You can have a Background as a second
//	 commander.)"
//
// Hardened Scales for every counter kind, on creatures and
// planeswalkers alike. The counter pipeline delivers one placement
// per kind, so "plus one of each of those kinds" is +1 on each
// event; a removal (negative delta) is not a placement and is left
// alone.
//
// "If YOU would put" is narrower than Hardened Scales' passive
// "would be put", and the pipeline carries no actor, so who is
// putting is read the way All Will Be One reads it
// (b13ResolutionInProgressBy): the player whose spell or ability is
// resolving. A placement outside any resolution — a loyalty cost, a
// Saga's lore counter, a permanent entering with counters — is the
// permanent's controller's (CR 606.2, 714.2b), and counts. An
// opponent's effect putting counters on Lae'zel's controller's
// creature does not, as printed; a placement of the controller's own
// that happens to follow an opponent's resolution without a boundary
// between them is missed — weaker, never stronger.
//
// Choose a Background is a deck-construction rule (CR 702.124), not
// an in-game effect; the deck importer's business.
//
// Sandbox simplification, declared: counters put on the CONTROLLER
// (poison, experience, energy) are not increased — player counters
// do not go through the replacement pipeline, so there is nothing
// for the card to replace. Weaker than printed, never stronger.
func init() {
	Register(Spec{
		OracleID:     "c066f921-e349-45d1-8ec3-0955d10bbf19",
		Name:         "Lae'zel, Vlaakith's Champion",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Counters put on you (poison, experience, energy) aren't increased — only counters on your creatures and planeswalkers are."},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventCounterPlaced},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventCounter || ev.CounterDelta <= 0 {
					return false
				}
				target, ok := g.LookupCardForEffect(ev.CounterTarget)
				if !ok || target.Controller != src.Controller {
					return false
				}
				if !target.IsCreature() && !target.IsPlaneswalker() {
					return false
				}
				by := b13ResolutionInProgressBy(g)
				return by == src.Controller || by == uuid.Nil
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.CounterDelta++
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Lae'zel, Vlaakith's Champion: one more counter",
		}},
	})
}
