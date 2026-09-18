package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Artful Dodge — Sorcery {U} (EDHREC rank 4016):
//
//	"Target creature can't be blocked this turn.
//	 Flashback {U} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// One mana to make a creature unblockable, twice. The Voltron and
// infect decks play it because the second half is what actually
// closes a game: the first copy forces the block-or-die decision, the
// flashback copy makes it moot.
//
// It is in the batch as the smallest card that puts the two S29/S32
// pieces together — a graveyard cast path and a turn-scoped
// restriction — and the two are declared separately for a reason. The
// ZONE is CastableZones (where it may be cast from) and the PRICE is
// the AlternativeCost (what it costs from there); the registry panics
// on a zone-bound cost whose zone was not listed, because such an
// offer could never be claimed.
//
// # "Can't be blocked" is on the ATTACKER
//
// It is a restriction on the DEFENDING player's legal blocks, carried
// on the attacking creature — not evasion that a blocker could
// match. Nothing gets around it: reach, flying, "can block creatures
// with flying", a Wall with any keyword you like. The only answers
// are removing the creature or removing the effect.
//
// The target need not be attacking, or even yours, when the spell
// resolves; the restriction lasts the turn and attaches to whatever
// creature was named.
//
// The flashback copy exiles itself after it resolves, so the card
// gives two uses and no more — that is the keyword's own bookkeeping,
// not this file's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "c174dcbb-03a0-439c-b3d8-ed61bd46dc67",
		Name:             "Artful Dodge",
		Completeness:     CompletenessFull,
		Targets:          TargetCreature("target creature"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{U}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return RestrictUntilEOT{
				Target:       item.Targets[0].ID,
				Restrictions: game.CantBeBlocked,
				Label:        "Artful Dodge — can't be blocked",
			}.Apply(ctx)
		},
	})
}
