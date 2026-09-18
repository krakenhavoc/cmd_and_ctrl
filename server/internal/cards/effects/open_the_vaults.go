package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
// The permanents arrive as ordinary entries, all in one sweep, so
// enters-the-battlefield triggers fire and entry replacements apply.
//
// # Declared simplification (weaker than printed): Auras stay put
//
// The parenthesis on the card is doing real work: an Aura that comes
// back has to arrive ATTACHED to something legal, chosen by its
// controller, and only remains in the graveyard when there is nothing
// it could enchant. Putting an Aura onto the battlefield attached to
// a chosen permanent is machinery the engine does not have yet
// (`aura-put-onto-battlefield`), and an Aura returned WITHOUT it would
// enter attached to nothing and be put straight back into the
// graveyard by the state-based sweep — a trip through the battlefield
// the printed card never takes, which would wrongly fire enters- and
// dies-triggers.
//
// So Aura cards are left in their graveyards, unconditionally. That
// is strictly less than the card does: an Aura that had nothing to
// enchant stays put either way, and one that did simply is not
// returned. Nothing here lets a deck do something paper would not.
//
// Every other artifact and enchantment — equipment, mana rocks,
// Sagas, enchantment creatures — comes back.
func init() {
	Register(Spec{
		OracleID:     "1e9c473e-bd65-4e1f-b2ab-cac58dc581c9",
		Name:         "Open the Vaults",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Aura cards are left in their graveyards — only the other artifacts and enchantments come back.",
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var ids []uuid.UUID
			for _, p := range ctx.Game.Seats {
				if p == nil || p.Graveyard == nil {
					continue
				}
				for _, c := range p.Graveyard.Cards {
					if c.IsAura() {
						continue
					}
					if c.IsArtifact() || c.IsEnchantment() {
						ids = append(ids, c.InstanceID)
					}
				}
			}
			for _, id := range ids {
				// Controller left zero: "under their owners' control"
				// is exactly ReturnFromGraveyard's default, and
				// naming the caster here is the mistake the field's
				// doc comment exists to prevent.
				if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
