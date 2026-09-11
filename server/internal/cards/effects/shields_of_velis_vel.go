package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shields of Velis Vel — Kindred Instant — Shapeshifter, {W}:
//
//	"Changeling (This card is every creature type.)
//	 Creatures target player controls get +0/+1 and gain all creature
//	 types until end of turn."
//
// The card that makes "gain all creature types until end of turn" a
// real primitive rather than a slot on a roadmap, and a neat proof of
// the all-zones rule besides: it is an INSTANT with changeling, so a
// Shields of Velis Vel in your graveyard is an Elf, a Goblin and a
// Sliver, and any tribal card that counts cards in graveyards counts
// it. Nothing in this file makes that happen — printedCharacteristic
// does, off the keyword.
//
// The affected set is snapshotted at resolution (CR 611.2c), which
// BoostUntilEOT and GrantKeywordUntilEOT both do for free: a creature
// the targeted player casts afterwards gets neither half.
//
// Two primitives rather than one because the P/T change is layer 7c
// and the type grant is layer 4 — a single entry could not sort into
// both. See AllCreatureTypesGrant for why a TYPE change is carried as
// a keyword marker, and why it still declares layer 4.
func init() {
	Register(Spec{
		OracleID:        "7ad6be4e-5c3c-4633-a641-beb06e4129b9",
		Name:            "Shields of Velis Vel",
		PrintedKeywords: []string{game.KeywordChangeling},
		Targets:         TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			victim := item.Targets[0].ID
			theirCreatures := And(Creature(), controlledBy(victim))
			if err := (BoostUntilEOT{
				Match:     theirCreatures,
				Toughness: 1,
				Label:     "Shields of Velis Vel — +0/+1",
			}).Apply(ctx); err != nil {
				return err
			}
			return GrantAllCreatureTypesUntilEOT{
				Match: theirCreatures,
				Label: "Shields of Velis Vel — all creature types",
			}.Apply(ctx)
		},
	})
}

// controlledBy is "permanents PLAYER controls" for a player named by
// the effect, as distinct from YouControl's "permanents the CASTER
// controls". Shields of Velis Vel targets a player and then talks
// about their creatures, so the caster is the wrong seat to ask.
func controlledBy(player uuid.UUID) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.Controller == player
	}
}
