package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Open the Vaults — Sorcery {4}{W}{W} (EDHREC rank 3996):
//
//	"Return all artifact and enchantment cards from all graveyards to
//	 the battlefield under their owners' control. (Auras with nothing
//	 to enchant remain in graveyards.)"
//
// A symmetrical mass reanimation for the half of the board that
// enchantress and artifact decks care about. It rebuilds a board a
// wrath took apart — and rebuilds everyone else's too, which is the
// tension: the deck that plays it is the deck whose graveyard is
// fullest.
//
// It is in the batch as the widest "from ALL graveyards" clause in
// the catalog. Three words do the work:
//
//   - "all graveyards" — every seat's, not just the caster's. Rise of
//     the Dark Realms is the only other card that reads every pile.
//   - "under their OWNERS' control" — nothing changes hands. An
//     opponent's Sol Ring comes back to the opponent. This is the
//     opposite of Rise of the Dark Realms, which takes everything,
//     and it is why Open the Vaults is a rebuild rather than a theft.
//   - "artifact and enchantment CARDS" — a token that died is gone
//     (CR 111.7) and is not among them.
//
// The permanents arrive as one entry (#1867), every seat's together,
// so entry replacements apply and every enters-the-battlefield trigger
// sees the whole batch (CR 603.6a).
//
// # Auras
//
// The parenthesis on the card is doing real work: an Aura that comes
// back arrives ATTACHED to something legal, chosen by its controller
// (its owner) as it enters, and only remains in the graveyard when
// there is nothing it could enchant (CR 303.4f, 303.4g; #3158). Each
// Aura asks its own question, among the objects and players already
// there — not another permanent this spell is returning.
func init() {
	Register(Spec{
		OracleID:     "1e9c473e-bd65-4e1f-b2ab-cac58dc581c9",
		Name:         "Open the Vaults",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			// Controller left zero: "under their owners' control" is
			// exactly the default, and naming the caster here is the
			// mistake the field's doc comment exists to prevent.
			return ReturnFromGraveyardTogether{
				Targets: allGraveyardsCardIDs(ctx, func(c game.Card) bool {
					return c.IsArtifact() || c.IsEnchantment()
				}),
			}.Apply(ctx)
		},
	})
}
