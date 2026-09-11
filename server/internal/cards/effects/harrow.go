package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Harrow — Instant for {2}{G}:
//
//	"As an additional cost to cast this spell, sacrifice a land."
//	"Search your library for up to two basic land cards, put them
//	 onto the battlefield, then shuffle."
//
// Instant-speed ramp that fixes two colours and is land-drop
// independent. Net zero lands, so it is played for the fixing, for
// the landfall triggers (two lands entering at once), and for
// sacrificing a land that has already done its job — a cracked
// fetchland's replacement, a Karoo you want back.
//
// The batch-02 triage (#295) filed Harrow under "cost modification".
// That is not what the card needs: "as an additional cost, sacrifice
// a land" is CR 601.2f, an ADDITIONAL cost, and Spec.AdditionalCost
// has carried a sacrifice clause since the S21 pass — SacrificeCost
// in additional_cost.go. Nothing about the mana cost is modified.
//
// Paying the cost with the spell already on the stack is the whole
// point of doing it this way: the land dies while Harrow is on the
// stack, so a landfall-adjacent or dies-trigger payoff goes ABOVE
// Harrow and resolves first, exactly as in paper. Sacrificing in
// OnResolve would invert that and would hand the land back when
// Harrow is countered.
//
// The lands enter UNTAPPED. Harrow is one of the few fetch effects
// with no "tapped" clause, which is most of why it costs three and
// is an instant — so TappedOnEntry stays false. The fetched land's
// OWN enters-tapped replacement still runs (#263), which is correct:
// a basic has none.
//
// "Up to two basic land cards" — the searcher picks, because the
// search chooser queues a prompt whenever a library holds more
// matches than Limit, which a real deck always does for basics.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:       "705509e9-a034-4a5a-9c65-66f58748b8a2",
		Name:           "Harrow",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a land", Land()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:    item.Controller,
				Predicate: IsBasicLand,
				Dest:      game.ZoneBattlefield,
				Limit:     2,
				Shuffle:   true,
				Reason:    "Harrow — put up to two basic lands onto the battlefield",
			}.Apply(ctx)
		},
	})
}
