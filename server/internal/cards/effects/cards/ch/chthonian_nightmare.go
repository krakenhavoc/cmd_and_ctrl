package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chthonian Nightmare — Enchantment {1}{B}:
//
//	"When this enchantment enters, you get {E}{E}{E} (three energy
//	 counters).
//	 Pay X {E}, Sacrifice a creature, Return this enchantment to its
//	 owner's hand: Return target creature card with mana value X from
//	 your graveyard to the battlefield. Activate only as a sorcery."
//
// ADR 0129 §2: "Pay X {E}" is PayXEnergy. X is announced with the
// activation (CR 107.3a), before the target, and may not exceed the
// activator's energy (CR 118.3). The target clause reads the announced X
// (WithManaValueEqualsX), checked at announce and again as the ability
// resolves (CR 608.2b). The return-this cost is #2028's ReturnThis: the
// enchantment goes to its owner's hand as the cost is paid, so the
// ability resolves with it already gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c7a360f5-725e-4864-93cc-87f96f95975e",
		Name:         "Chthonian Nightmare",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Energy: 3},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouGetEnergy("Chthonian Nightmare", 3),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "Pay X {E}, Sacrifice a creature, Return this enchantment to its owner's hand: Return target creature card with mana value X from your graveyard to the battlefield. Activate only as a sorcery.",
			Cost:  effects.Plus(effects.PayXEnergy(), effects.SacrificeACreature(), effects.ReturnThis()),
			Targets: effects.TargetCardInGraveyard("target creature card with mana value X from your graveyard",
				effects.Creature(), effects.YouOwn()).WithManaValueEqualsX(),
			SorcerySpeed: true,
			Effect:       effects.OldB775ReturnTargetCreatureCardToBattlefield,
		}},
	})
}
