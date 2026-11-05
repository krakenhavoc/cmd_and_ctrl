package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Donatello, the Brains — Legendary Creature — Mutant Ninja Turtle {2}{U},
// 2/4 (Teenage Mutant Ninja Turtles Commander, #5011):
//
//	"If one or more tokens would be created under your control, those
//	 tokens plus a Mutagen token are created instead.
//	 Partner—Character select"
//
// The first line is Peregrin Took's CR 614 token-creation replacement
// (the CR 701.7b window) with the printed Mutagen (CR 111.10v) in place of
// the Food: it applies once per creation instruction under Donatello's
// controller, whoever controls the spell or ability making the tokens, and
// appends one Mutagen group. The rulings say the extra token gets nothing
// the instruction's tokens were created WITH as abilities, but anything
// else the instruction specifies applies to it too: so the Mutagen copies
// the instruction's entry clause (tapped, entering counters), what it is
// created attacking and its "created with" link, and not its granted
// keywords. A creation-wide "attacking" or end-of-combat exile already
// covers every group, the Mutagen's included.
//
// Partner—Character select is a deck-construction rule the deck importer
// reads (CR 702.124i).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1ba2f763-5895-4663-b38f-95591ebb2ef4",
		Name:         "Donatello, the Brains",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventTokenCreated},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventCreateTokens && ev.TokenCount() > 0 && ev.TokenController == src.Controller
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.TokenGroups = append(ev.TokenGroups, donatelloMutagenGroup(ev.TokenGroups))
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Donatello, the Brains: plus a Mutagen token",
		}},
	})
}

// donatelloMutagenGroup is the one Mutagen the replacement adds, carrying
// the creation's own clauses from the first group that makes anything.
// Granted keywords are abilities and stay behind (the card's ruling); an
// Aura host and a "blocking" clause are about the token that names them.
func donatelloMutagenGroup(groups []game.TokenGroup) game.TokenGroup {
	g := game.TokenGroup{Template: ruleDefinedToken(MutagenToken()), Count: 1}
	for _, from := range groups {
		if from.Count <= 0 {
			continue
		}
		g.Entry.Tapped = from.Entry.Tapped
		if len(from.Entry.Counters) > 0 {
			g.Entry.Counters = make(map[string]int, len(from.Entry.Counters))
			for k, n := range from.Entry.Counters {
				g.Entry.Counters[k] = n
			}
		}
		g.Attacking = from.Attacking
		g.AttackOpponent = from.AttackOpponent
		g.CreatedWith = from.CreatedWith
		break
	}
	return g
}
