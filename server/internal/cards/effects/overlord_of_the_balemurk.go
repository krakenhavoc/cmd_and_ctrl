package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Overlord of the Balemurk — Enchantment Creature — Avatar Horror
// {3}{B}{B}, 5/5:
//
//	"Impending 5—{1}{B} (If you cast this spell for its impending
//	 cost, it enters with five time counters and isn't a creature
//	 until the last is removed. At the beginning of your end step,
//	 remove a time counter from it.)
//	 Whenever this permanent enters or attacks, mill four cards, then
//	 you may return a non-Avatar creature card or a planeswalker card
//	 from your graveyard to your hand."
//
// The first card in the tree with impending, and the reason
// impending.go exists — the keyword ships as a shared constructor, so
// the other nine Overlords and the rest of the cycle are one line
// each. Register(Impending(5, "{1}{B}", Spec{…})) attaches all three
// halves of CR 702.176 at once; see impending.go for why they cannot
// be declared separately.
//
// The trigger is one printed ability with two conditions, so it is one
// WhenThisEntersOrAttacks watching EventETB and EventAttack — Sun
// Titan's shape. Both halves work under either price, and the two
// clauses need no special-casing to agree: a permanent with time
// counters is not a creature, so it cannot be declared as an
// attacker, and while the countdown runs only the enters half can
// fire. Cast for {1}{B} the Overlord is an enchantment that mills
// four on arrival and wakes up as a 5/5 five end steps later; cast
// for {3}{B}{B} it is a 5/5 that mills four now and again every
// time it attacks.
//
// "From YOUR GRAVEYARD", not "from among them" — the return ranges
// over the whole pile, the four cards just milled included, which is
// what makes the mill and the return one card rather than two. It is
// still written inside the mill's continuation, because the four
// cards have to have landed before the prompt can offer them (#893).
//
// "Non-Avatar creature card OR planeswalker card" is one predicate,
// and the Avatar exclusion applies only to the creature half: an
// Avatar planeswalker card would be a legal pick. That is the
// printed reading, and it is also the one that stops the Overlord
// returning itself — it is an Avatar — which is the clause's whole
// purpose.
func init() {
	Register(Impending(5, "{1}{B}", Spec{
		OracleID:     "652e87af-7cf0-407f-9fb7-1a630fb8dd47",
		Name:         "Overlord of the Balemurk",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks(
				"Overlord of the Balemurk — mill four cards, then you may return a creature or planeswalker card",
				overlordOfTheBalemurkMill),
		},
	}))
}

// overlordOfTheBalemurkMill is the enters-or-attacks trigger: mill
// four, then offer the graveyard.
func overlordOfTheBalemurkMill(g *game.Game, item *game.StackItem) error {
	return MillToZone{
		N: 4,
		Then: func(ctx *Context, _ []uuid.UUID) error {
			return mayTakeOneFromYourGraveyard(ctx, nonAvatarCreatureOrPlaneswalker(),
				"Overlord of the Balemurk — you may return a non-Avatar creature card or a planeswalker card from your graveyard to your hand")
		},
	}.Apply(NewContext(g, item))
}

// nonAvatarCreatureOrPlaneswalker is the return clause's filter.
func nonAvatarCreatureOrPlaneswalker() CardPredicate {
	return Or(And(Creature(), Not(OfSubtype("Avatar"))), Planeswalker())
}
