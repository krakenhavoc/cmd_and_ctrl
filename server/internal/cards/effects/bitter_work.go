package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bitterWorkDrawLabel is the attack trigger's stack label and its
// CR 603.2c batch key.
const bitterWorkDrawLabel = "Bitter Work — draw a card (you attacked a player with a creature with power 4 or greater)"

// Bitter Work — Enchantment {1}{R}{G}:
//
//	"Whenever you attack a player with one or more creatures with
//	 power 4 or greater, draw a card.
//	 Exhaust — {4}: Earthbend 4. Activate only during your turn.
//	 (Target land you control becomes a 0/0 creature with haste that's
//	 still a land. Put four +1/+1 counters on it. When it dies or is
//	 exiled, return it to the battlefield tapped. Activate each exhaust
//	 ability only once.)"
//
// One of the two cards #1178 stopped at: earthbend shipped with #1180
// and the only thing left between this card and one line was the
// "only once" half of exhaust, which #1181 built. Both halves of the
// activated ability are now declarations — `Exhaust: true` for the
// keyword and `DuringYourTurn()` for the printed activation
// instruction — and the engine checks them in that order, so a second
// activation on your own turn is refused as exhausted rather than as
// badly timed.
//
// The two gates are genuinely different and this card is why the
// engine keeps them apart: "Activate only during your turn" is false
// on Tuesday and true on Wednesday, and "activate each exhaust ability
// only once" is never true again while this object is on the
// battlefield. The client says the two differently for the same
// reason.
//
// The attack trigger is the "you attack A PLAYER" batch (CR 603.2c):
// one declaration is one batch, so however many big creatures swing at
// one player they draw one card, and swinging at a SECOND player is a
// second occurrence and a second card — OncePerBatchPerPlayer, the
// same reading Horizon Explorer's ruling states. "With one or more
// creatures with power 4 or greater" is answered per attacker as the
// declaration is announced: the first attacker at that player whose
// power reaches 4 fires the ability and every other attacker at the
// same player is collapsed into it, which is what "one or more"
// means. Power is read post-layer, so a pump that resolved before the
// declaration counts and one that resolves after it does not.
//
// The earthbend is the keyword action and nothing else — the
// animation, the haste, the four counters and the delayed return all
// live in game.EarthbendForEffect, and this file is the count plus the
// shared target clause (earthbend.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e9a24ed8-e844-4c50-b386-bc28da9989b1",
		Name:         "Bitter Work",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatchPerPlayer(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !b16YouAttackedAPlayer(ev, source, g) {
					return false
				}
				return b31CurrentPowerOf(g, ev.CardID) >= 4
			}, bitterWorkDrawLabel, Do(DrawCards{N: 1}))),
		},
		Activated: []ActivatedAbility{{
			Label:     "Exhaust — {4}: Earthbend 4. Activate only during your turn.",
			Exhaust:   true,
			Cost:      ManaCost("{4}"),
			Condition: DuringYourTurn(),
			Targets:   EarthbendTargets(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				target := FirstLegalBattlefieldTarget(ctx)
				if target == uuid.Nil {
					return nil
				}
				return Earthbend{Target: target, N: 4}.Apply(ctx)
			},
		}},
	})
}
