package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_bns_b21_helpers.go — the shared pieces of set BNS batch 21 (#5929).

// bnsB21ClueTapland is the Ravnica: Clue Edition common dual cycle
// (Conservatory, Dining Room, Hall, Kitchen, Library, Lounge, Secret
// Passage, Study):
//
//	"This land enters tapped.
//	 {T}: Add {A} or {B}.
//	 {4}, {T}: Investigate."
//
// Enters-tapped is the self-replacement every tapland uses, the dual is the
// one-row pipe every two-colour land uses, and investigating is Magnifying
// Glass's ability: a CR 602 ability with a mana-and-tap cost that creates a
// Clue token (CR 701.16a). Each member has its own file and registers
// through here.
func bnsB21ClueTapland(oracleID, name, a, b string) Spec {
	return Spec{
		OracleID:      oracleID,
		Name:          name,
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility(a, b)},
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}: Investigate.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1},
			Cost:    Plus(ManaCost("{4}"), TapCost()),
			Effect:  Do(Investigate{N: 1}),
		}},
	}
}

// bnsB21GreatestManaValueLeft is "the greatest mana value among those
// cards" for a batch trigger over cards that left the controller's
// graveyard (Amzu, Swarm's Hunger): the greatest mana value among the cards
// the batch's zone-move and cast events took out of a graveyard they own,
// read off the card where it went. Zero when there are none.
func bnsB21GreatestManaValueLeft(g *game.Game, item *game.StackItem) int {
	if item.Trigger == nil {
		return 0
	}
	batch := item.Trigger.Event.Batch
	greatest := 0
	for _, ev := range g.EventsThisTurn() {
		if ev.Batch != batch {
			continue
		}
		if c, ok := cardThatLeftAGraveyard(ev, g); ok && c.Owner == item.Controller {
			greatest = max(greatest, c.ManaValue())
		}
	}
	return greatest
}
