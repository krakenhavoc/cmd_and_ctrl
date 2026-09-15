package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Quest for Renewal — Enchantment {1}{G} (EDHREC rank 3383):
//
//	"Whenever a creature you control becomes tapped, you may put a
//	 quest counter on this enchantment.
//	 As long as there are four or more quest counters on this
//	 enchantment, untap all creatures you control during each other
//	 player's untap step."
//
// The budget Seedborn Muse. The counter trigger is Beastmaster
// Ascension's shape with Magda's condition: it reads two event kinds,
// because the engine taps an attacker without an EventTapCard — see
// b11DwarfYouControlBecameTapped for why an EventAttack whose
// creature is now tapped is "became tapped", and why a vigilance
// attacker is not. Tapping for mana, convoking and attacking all
// count, as printed, and each is its own yes/no prompt.
//
// The untap half was a declared simplification until #74: the engine
// had no hook a card could add to the untap step's turn-based action,
// so it shipped as a triggered ability at the following upkeep —
// a step late, on the stack, where a tap effect in response left the
// creatures tapped. Spec.UntapStep is that hook, and the clause is
// now exact: nothing is announced, nothing can respond, and the
// creatures untap during the other player's untap step.
//
// "As long as there are four or more quest counters" is an ordinary
// continuous condition and belongs inside the permission's AppliesTo
// rather than beside it. Read at the instant the step asks, it has
// no announce-to-resolve gap for the count to change across — which
// is the second thing the trigger version got wrong and had to
// re-check by hand at both ends.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cba4ab80-09e8-4868-a082-a9a3bade9571",
		Name:         "Quest for Renewal",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b32CreatureYouControlBecameTapped(ev, source, g)
			}, "Quest for Renewal — put a quest counter", func(g *game.Game, item *game.StackItem) error {
				if !b09SourceStillOnBattlefield(g, item) {
					return nil
				}
				return AddCounter{Target: item.SourceCardID, Kind: "quest", N: 1}.Apply(NewContext(g, item))
			}), "Quest for Renewal — put a quest counter on it?"),
		},
		UntapStep: []game.UntapStepPermission{{
			Label: "Quest for Renewal — untap all creatures you control",
			AppliesTo: func(g *game.Game, source *game.Card, activePlayer uuid.UUID) bool {
				return source.Counters["quest"] >= 4 && eachOtherPlayersUntapStep(g, source, activePlayer)
			},
			Untaps: func(_ *game.Game, source, target *game.Card) bool {
				return target.Controller == source.Controller && target.IsCreature()
			},
		}},
	})
}
