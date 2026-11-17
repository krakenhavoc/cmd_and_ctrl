package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Saproling Burst — Enchantment {4}{G}:
//
//	"Fading 7 (This enchantment enters with seven fade counters on it. At
//	 the beginning of your upkeep, remove a fade counter from it. If you
//	 can't, sacrifice it.)
//	 Remove a fade counter from this enchantment: Create a green Saproling
//	 creature token. It has "This token's power and toughness are each
//	 equal to the number of fade counters on Saproling Burst."
//	 When this enchantment leaves the battlefield, destroy all tokens
//	 created with this enchantment. They can't be regenerated."
//
// Fading is Fading/FadingUpkeep (fading.go, CR 702.32a); the ability's
// cost is a fade counter (RemoveAFadeCounterFromThis, CR 602.2b).
//
// Each Saproling is linked to the Saproling Burst OBJECT that made it
// (game.Card.CreatedWith, CR 607.1d), Tombstone Stairwell's link. Its
// size is a layer 7a characteristic-defining ability (CR 604.3, 613.4a)
// that reads the fade counters on that object while it is still on the
// battlefield as the same object, and is 0 otherwise. So every Saproling
// shrinks as the Burst spends counters, a Saproling made by a Burst that
// has left is a 0/0, and a Burst that left and came back is a new object
// whose counters the old Saprolings do not read (CR 400.7). The leaves
// trigger destroys the linked tokens whoever controls them, ignoring
// regeneration (CR 701.19c), as the Stairwell's does; by then they are
// usually 0/0s the state-based actions have already put away.
//
// No simplification.
const saprolingBurstOracleID = "7b31ecf8-5266-4376-ace0-8b873706c4ea"

func init() {
	Register(Spec{
		OracleID:     saprolingBurstOracleID,
		Name:         "Saproling Burst",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{Fading(7)},
		Triggered: []game.TriggeredAbility{
			FadingUpkeep(),
			On(game.EventLTB, Self,
				"Saproling Burst — destroy all tokens created with it; they can't be regenerated",
				tombstoneStairwellDestroy),
		},
		Activated: []ActivatedAbility{{
			Label:   "Remove a fade counter from this enchantment: Create a green Saproling creature token. It has \"This token's power and toughness are each equal to the number of fade counters on Saproling Burst.\"",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
			Cost:    RemoveAFadeCounterFromThis(),
			Effect:  saprolingBurstCreate,
		}},
	})
}

// saprolingBurstCreate makes one Saproling linked to the Burst object
// the ability came from, on the battlefield or gone.
func saprolingBurstCreate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     item.SourceCardID,
		Groups: []game.TokenGroup{{
			Template:    ctx.TextToken(tokenFromCatalog(printedSaprolingBurstToken)),
			Count:       1,
			CreatedWith: game.PermissionCardRef(ref),
		}},
	}, game.KeyedTail{})
}

// printedSaprolingBurstToken is Saproling Burst's Saproling. Its slug is
// an on-disk identity: never renamed, never reused.
func printedSaprolingBurstToken() tokenTemplate {
	const text = "This token's power and toughness are each equal to the number of fade counters on Saproling Burst."
	return tokenTemplate{
		Slug: "saproling-burst-saproling",
		Card: game.Card{
			Name:     "Saproling",
			TypeLine: "Token Creature — Saproling",
			Colors:   []string{"G"},
		},
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			Label:     text,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				if link := source.CreatedWith; link.ID != uuid.Nil {
					for _, p := range g.BattlefieldCardsForEffect() {
						if p.InstanceID == link.ID && p.ObjectEpoch == link.Epoch {
							n = p.Counters[game.CounterFade]
							break
						}
					}
				}
				c.Power, c.Toughness = n, n
			},
		}},
		Text: text,
	}
}
