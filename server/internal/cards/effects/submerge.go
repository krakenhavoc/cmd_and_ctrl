package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Submerge — Instant {4}{U} (EDHREC rank 4309):
//
//	"If an opponent controls a Forest and you control an Island, you
//	 may cast this spell without paying its mana cost.
//	 Put target creature on top of its owner's library."
//
// The blue half of the Nemesis free-spell cycle, and the one whose
// condition a Commander table meets constantly: somebody is playing
// green, and you are playing blue. Five mana is the price you never
// pay.
//
// The condition is the card, and it is asymmetric in a way worth
// spelling out. "An OPPONENT controls a Forest" — not you, and not any
// player — so a green deck cannot cast its own Submerge for free off
// its own lands. "You control an Island" is yours. Both read the land
// TYPE off effective subtypes, so a Tropical Island satisfies both
// halves at once for two different players, and a land something
// turned into a Forest counts.
//
// The condition is checked in the VIEW as well as at announce, so a
// player who cannot meet it is never shown the offer rather than being
// shown one the server would reject.
//
// Putting a creature on top of its owner's library is the removal that
// beats indestructible, a death trigger and a graveyard recursion deck
// all at once, and costs the victim their next draw. A token tucked
// this way ceases to exist (CR 111.8).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "99427ebe-c00d-4206-84ca-9764f6e952c6",
		Name:         "Submerge",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{{
			Key:       "free",
			Label:     "Cast without paying its mana cost (an opponent controls a Forest and you control an Island)",
			ManaCost:  "",
			Condition: b41SubmergeCondition,
		}},
		Targets: TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return ctx.Game.TuckToLibraryForEffect(item.Targets[0].ID, false)
		},
	})
}

// b41SubmergeCondition is Submerge's clause: an OPPONENT of the caster
// controls a Forest, and the caster controls an Island. One walk of
// the battlefield, because both halves read the same slice and the
// answer is only "yes" when both are satisfied.
//
// Kept beside the card rather than in batch41_helpers.go because it is
// one card's sentence, not a template: no other card in the batch asks
// a question of this shape.
func b41SubmergeCondition(g *game.Game, caster uuid.UUID) bool {
	opponentForest, yourIsland := false, false
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == caster {
			if hasSubtype(c, "Island") {
				yourIsland = true
			}
			continue
		}
		if hasSubtype(c, "Forest") {
			opponentForest = true
		}
	}
	return opponentForest && yourIsland
}
