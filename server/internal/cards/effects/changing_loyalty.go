package effects

import (
	"errors"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Changing Loyalty — Enchantment — Aura {1}{B} (Secrets of Strixhaven
// Commander, #4834):
//
//	"Flash
//	 Replicate {2} (When you cast this spell, copy it for each time you
//	 paid its replicate cost. You may choose new targets for the copies.
//	 Copies become tokens.)
//	 Enchant creature
//	 When enchanted creature dies, return it to the battlefield under
//	 your control."
//
// Replicate is the keyword's two halves (CR 702.56a): the optional mana
// cost paid any number of times (replicateMana) and the shared
// Replicate() cast trigger, which copies the Aura spell once per payment
// and offers each copy a new creature to enchant. A copy of a permanent
// spell becomes a token as it resolves (CR 707.10f), so each copy is a
// token Aura on its own creature.
//
// The dies trigger is Kaya's Ghostform's return under "your" control,
// "your" being the Aura's controller as the ability triggered. "It" is
// the creature card that went to the graveyard; if it has left the
// graveyard by the time the ability resolves it is a new object and is
// not returned (CR 400.7). A token that died has no card to return.
func init() {
	Register(Spec{
		OracleID:        "cdfa9de4-7d7d-4638-8ac4-793ece6372dd",
		Name:            "Changing Loyalty",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(),
		OptionalCosts:   []game.AdditionalCost{replicateMana("{2}", 10)},
		Triggered: []game.TriggeredAbility{
			Replicate(),
			WhenEnchantedCreatureDies("Changing Loyalty — return that creature to the battlefield under your control", changingLoyaltyReturn),
		},
	})
}

func changingLoyaltyReturn(g *game.Game, item *game.StackItem) error {
	dead := item.Trigger.Event.CardID
	if z := g.FindCardZoneForEffect(dead); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	_, err := g.ReturnToBattlefieldForEffect(dead, item.Controller, false)
	if errors.Is(err, game.ErrCardNotFound) {
		return nil
	}
	return err
}
