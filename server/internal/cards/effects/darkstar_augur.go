package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Darkstar Augur — Creature — Bat Warlock {2}{B}, 2/3:
//
//	"Offspring {B}
//	 Flying
//	 At the beginning of your upkeep, reveal the top card of your
//	 library and put that card into your hand. You lose life equal to
//	 its mana value."
//
// Dark Confidant on a flying body, with a second copy of itself
// available for one more black mana. In a deck that treats life as a
// resource the upkeep flip is the engine and the life is the price;
// with the offspring token out, the price is paid twice a turn.
//
// # Three printed lines, three different mechanisms
//
//   - Offspring is an OPTIONAL ADDITIONAL COST plus an entry trigger
//     gated on it having been paid. Both halves come from
//     offspring.go — see there for why the token cannot make a token
//     of its own.
//   - Flying is a printed keyword string; the engine's combat code
//     reads it.
//   - The upkeep flip is the whole of Dark Confidant's ability.
//
// # Reveal, then move, then lose — and the order is the card
//
// REVEALING IS NOT DRAWING. Nothing here emits EventDrawCard, so a
// draw payoff sitting next to the Augur does not fire on the upkeep
// flip and a draw replacement does not see it.
//
// The reveal happens FROM THE LIBRARY, before anything moves, so the
// whole table learns what the Augur just charged its controller for.
// The move then goes through the named library-to-hand door with a
// continuation (#952): "return it to its owner's hand" happening to
// move a card out of a library is an accident of the zone router, and
// the continuation is what makes the life loss wait for the card to
// actually land.
//
// MANA VALUE IS READ BEFORE THE MOVE, because that is the last moment
// the card is guaranteed findable. The number is the same in either
// zone (CR 202.3e — {X} is zero anywhere but the stack), so this is
// about not losing the card, not about picking a zone to read it in.
//
// You LOSE the life; you do not pay it. It is not a cost, it cannot
// be declined, and it goes through the life-change path so a
// life-loss watcher sees it. A LAND FLIP COSTS NOTHING, and an empty
// library reveals nothing and is not an error — revealing is not
// drawing, so bottoming out this way does not set up the CR 704.5b
// loss.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "67d2a021-a042-4e5f-b6e4-39cb35514794",
		Name:            "Darkstar Augur",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		OptionalCosts:   []game.AdditionalCost{Offspring("{B}")},
		Triggered: []game.TriggeredAbility{
			OffspringToken("Darkstar Augur"),
			AtYourUpkeep(darkstarAugurFlipLabel, darkstarAugurFlip),
		},
	})
}

// darkstarAugurFlipLabel is the upkeep ability's stack label.
const darkstarAugurFlipLabel = "Darkstar Augur — reveal the top card of your library and put it into your hand"

// darkstarAugurFlip is the upkeep trigger: reveal one off the top,
// read what it costs, put it in hand, then lose that much life.
func darkstarAugurFlip(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := item.Controller
	revealed := g.RevealTopOfLibraryForEffect(player, ctx.Source(), 1, darkstarAugurFlipLabel)
	if len(revealed) == 0 {
		// An empty library. Nothing was revealed, so there is nothing
		// to put into hand and nothing to pay for.
		return nil
	}
	life := 0
	if c, ok := g.LookupCardForEffect(revealed[0]); ok {
		life = c.ManaValue()
	}
	source := item.SourceCardID
	return TakeFromLibraryToHand{
		Player: player,
		Cards:  revealed,
		All:    true,
		Label:  darkstarAugurFlipLabel,
		Then: func(g *game.Game, _ TakeFromLibraryResult) error {
			if life == 0 {
				// A land, or a card with no mana cost. The clause
				// still happened; it just cost nothing, and asking
				// the engine for a life change of nothing would emit
				// an event that a life-loss watcher would answer.
				return nil
			}
			return g.ChangePlayerLifeForEffect(source, player, -life)
		},
	}.Apply(ctx)
}
