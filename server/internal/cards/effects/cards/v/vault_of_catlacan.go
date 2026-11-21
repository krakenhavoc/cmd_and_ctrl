package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vault of Catlacan — "(Transforms from Storm the Vault.) {T}: Add one
// mana of any color. {T}: Add {U} for each artifact you control."
//
// The BACK face of Storm the Vault, registered under
// "<oracle_id>#1" — game.CatalogKey appends the active face index, so
// the moment ADR 0079's verb flips the permanent this entry is the one
// the engine reads and the front face's trigger pair stops applying.
// Nothing had to be built for that; it is what the "#N" key is for.
//
// The parenthetical is reminder text and needs no code: a back face is
// reached by transforming, and nothing else can put this one onto the
// battlefield (CR 712.11 — a transform card is always cast as its front
// face, which CastableFaces enforces).
//
// Two mana abilities, both plain {T} costs. The second is why this face
// is worth having as a proof: "add {U} for each artifact you control"
// is a count read at ACTIVATION time, which is ProducedFunc rather
// than a fixed Produced string — the same shape Cabal Coffers uses.
// Controlling no artifacts produces no mana and is not an error; the
// land still taps, which is what the card says.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     effects.StormTheVaultOracleID + "#1",
		Name:         "Vault of Catlacan",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "{T}: Add one mana of any color",
			},
			{
				Cost:           effects.ManaAbilityCost{Tap: true},
				ProducedAmount: effects.ProducedPerPermanent("U", effects.MatchArtifact),
				Label:          "{T}: Add {U} for each artifact you control",
			},
		},
	})
}
