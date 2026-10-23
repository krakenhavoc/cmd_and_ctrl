package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chorale of the Void — Enchantment — Aura {3}{B} (Edge of Eternities):
//
//	"Enchant creature you control
//	 Whenever enchanted creature attacks, put target creature card from
//	 defending player's graveyard onto the battlefield under your control
//	 tapped and attacking.
//	 Void — At the beginning of your end step, sacrifice this Aura unless
//	 a nonland permanent left the battlefield this turn or a spell was
//	 warped this turn."
//
// The attack trigger belongs to the Aura, so "you" is the Aura's
// controller (CR 109.5). Its target clause is built as the trigger goes on
// the stack (TargetsFrom), from the player behind what the enchanted
// creature attacked, a planeswalker's or battle's protector included, and
// is re-checked as it resolves (CR 608.2b). The creature enters tapped and
// attacking (CR 508.4): the Aura's controller says what it attacks when
// there is more than one choice, and it was never declared as an attacker,
// so no "whenever a creature attacks" trigger sees it.
//
// The end-step sacrifice is NOT an intervening if: the ability triggers
// every end step and void is checked as it resolves (AtYourEndStepUnlessVoid,
// void.go), which is why a nonland permanent leaving in response to it saves
// the Aura.
//
// No simplification.
func init() {
	raise := WheneverEnchantedCreatureAttacks("Chorale of the Void — put target creature card from defending player's graveyard onto the battlefield tapped and attacking",
		choraleRaise)
	raise.TargetsFrom = choraleTargets
	Register(Spec{
		OracleID:     "cd50bff0-c461-4023-b62a-6746eaaeab0e",
		Name:         "Chorale of the Void",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			raise,
			AtYourEndStepUnlessVoid("Chorale of the Void — sacrifice this Aura unless void", SacrificeThisIfStillOnBattlefield),
		},
	})
}

// choraleTargets is "target creature card from defending player's
// graveyard", the defending player read off the attack that triggered it.
// It reads the trigger's event and the player behind the attacked object,
// and nothing else.
func choraleTargets(tc game.TriggerContext, _ *game.Card, g *game.Game) *game.TargetSpec {
	defender := g.DefendingPlayerForAttackForEffect(tc.Event.Target)
	if defender == uuid.Nil {
		return nil
	}
	return TargetCardInGraveyard("target creature card in defending player's graveyard", Creature(), OwnedBy(defender))
}

// choraleRaise puts the still-legal target onto the battlefield under the
// resolving controller's control, tapped and attacking.
func choraleRaise(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range legalTargetIDs(ctx) {
		return putAttackingYourChoice(ctx, id, game.ZoneGraveyard, "Chorale of the Void")
	}
	return nil
}
