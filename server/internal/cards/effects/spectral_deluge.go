package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spectral Deluge — Sorcery {4}{U}{U} (EDHREC rank 4267):
//
//	"Return each creature your opponents control with toughness X or
//	 less to its owner's hand, where X is the number of Islands you
//	 control.
//	 Foretell {1}{U}{U} (During your turn, you may pay {2} and exile
//	 this card from your hand face down. Cast it on a later turn for
//	 its foretell cost.)"
//
// A one-sided Evacuation for a blue deck that is mostly Islands, which
// by turn six is most of the opposing board. It is the asymmetry that
// makes it worth six mana: your own creatures stay, so this is a wipe
// you cast into your own board.
//
// X is read at RESOLUTION, inside the predicate, not when the spell
// was cast — an opponent who blows up an Island in response really
// does shrink the sweep. The count is the land TYPE, so a Tropical
// Island and a land something turned into an Island both count.
//
// Toughness is the CURRENT toughness, post-layers, so an anthem or a
// +1/+1 counter can lift a creature out of the sweep and a −1/−1
// counter can drop one into it. Counting toughness rather than power
// is the printed text and is why this misses a Ball Lightning and
// catches a Wall.
//
// The set is snapshotted before anything moves and returned as ONE
// simultaneous event, so a Blood Artist caught in it — it is not,
// since the sweep is opponents-only — and any "whenever one or more
// permanents leave" watcher sees the whole batch at once.
//
// Declared simplification, weaker than printed (#259): FORETELL is not
// implemented (#658). The Deluge can only be cast from hand for its
// full {4}{U}{U}; the "pay {2} now, {1}{U}{U} later" split, and the
// information hiding that comes with the face-down exile, are not
// available.
func init() {
	Register(Spec{
		OracleID:     "bb9bb65e-504f-4edd-8d47-d0efa03e7d17",
		Name:         "Spectral Deluge",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Foretell is not available — the Deluge can only be cast from your hand for its full {4}{U}{U}, never split across two turns as a face-down exiled card.",
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return BounceAllMatching{
				Match: And(Creature(), OpponentControls(), b41ToughnessAtMostYourIslands()),
			}.Apply(ctx)
		},
	})
}

// b41ToughnessAtMostYourIslands is the Deluge's "with toughness X or
// less, where X is the number of Islands you control".
//
// The Island count is recomputed per candidate rather than hoisted,
// which is a handful of extra walks over a battlefield and buys the
// predicate the property every predicate in targets.go has: it is a
// pure function of (game, caster, card) and can be composed with And
// without a closure capturing a stale number.
//
// Kept beside the card because it is one card's sentence.
func b41ToughnessAtMostYourIslands() CardPredicate {
	return func(g *game.Game, caster uuid.UUID, c game.Card) bool {
		return c.CurrentToughness() <= b08LandsWithSubtypeControlled(g, caster, "Island")
	}
}
