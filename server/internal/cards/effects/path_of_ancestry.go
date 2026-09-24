package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Path of Ancestry — Land (EDHREC rank 14, the highest-ranked card
// this catalog was missing that needed no new machinery):
//
//	"This land enters tapped.
//	 {T}: Add one mana of any color in your commander's color
//	 identity. When that mana is spent to cast a creature spell that
//	 shares a creature type with your commander, scry 1."
//
// The enters-tapped clause is a real CR 614 self-replacement, not an
// OnETB tap — the land is never untapped on the battlefield, the
// distinction the Temple cycle pinned. The mana half is Command
// Tower's identity-narrowed pipe.
//
// The scry is a spend rider (#1547): a triggered ability that fires
// when the token pays for a creature spell and goes on the stack above
// it. "Shares a creature type with your commander" is the rider's
// Condition rather than a tag, because it reads the board — any
// commander the controller owns, wherever it is, partners included
// (sharesCreatureTypeWithYourCommander) — and SharesCreatureType is the
// one place that question is answered, changelings and all.
//
// One declared simplification, weaker than printed: with strict mana
// off the pool is never spent (ADR 0068 §3), so no token pays and the
// scry never happens.
func init() {
	Register(Spec{
		OracleID:     "b473e293-59e3-4e04-acf2-622604aeb25f",
		Name:         "Path of Ancestry",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"With strict mana off, the game doesn't see which mana you spent, so the scry 1 never happens."},
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color in your commander's color identity",
			// The printed text asks for the narrowing (manaPickOptionsFor).
			NarrowToCommanderIdentity: true,
			SpendRiders: []game.ManaSpendRider{WhenManaSpent("Path of Ancestry", game.ManaSpendTrigger{
				Label:     "Path of Ancestry — scry 1",
				Condition: sharesCreatureTypeWithYourCommander,
				Effect:    pathOfAncestryScry,
			}, ManaRestrictCast, ManaRestrictType("Creature"))},
		}},
	})
}

// pathOfAncestryScry is the rider's effect: the player who spent the
// mana scries 1.
func pathOfAncestryScry(g *game.Game, item *game.StackItem) error {
	return Scry{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
}
