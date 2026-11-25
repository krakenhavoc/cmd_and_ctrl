package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Custodi Soulcaller — Creature — Human Cleric {1}{W}{W}, 1/2 (Conspiracy:
// Take the Crown, S82 older sets batch 270, #5540):
//
//	"Melee (Whenever this creature attacks, it gets +1/+1 until end of turn
//	 for each opponent you attacked this combat.)
//	 Whenever this creature attacks, return target creature card with mana
//	 value X or less from your graveyard to the battlefield, where X is the
//	 number of players you attacked this combat."
//
// Melee is the engine's keyword (CR 702.121); its pump and this trigger are
// two triggers of the one attack and resolve in the order the controller
// picks. X is the count melee itself reads (each opponent attacked with a
// creature this combat, once each) read by the target's mana value
// predicate as the trigger's target is chosen and again as it resolves
// (CR 608.2b); the count includes the declaration still being announced. The creature returns under its owner's control, as
// Eddie Brock's does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dc25bdde-1c73-4c61-99d8-b772aae278cd",
		Name:            "Custodi Soulcaller",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordMelee},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks("Custodi Soulcaller — return target creature card with mana value X or less from your graveyard to the battlefield, where X is the number of players you attacked this combat",
				effects.ReturnTheFirstClauseTargetToBattlefield),
				effects.TargetCardInGraveyard("target creature card with mana value X or less in your graveyard, X the number of players you attacked this combat",
					effects.Creature(), effects.YouOwn(), custodiSoulcallerManaValue)),
		},
	})
}

// custodiSoulcallerManaValue passes a card whose mana value is at most the
// number of players its caster attacked this combat.
func custodiSoulcallerManaValue(g *game.Game, caster uuid.UUID, c game.Card) bool {
	mv, ok := g.ManaValueForEffect(c)
	return ok && mv <= custodiSoulcallerPlayersAttacked(g, caster)
}

// custodiSoulcallerPlayersAttacked is the number of players `you` attacked
// this combat: the opponents Game.OpponentsAttackedThisCombat counts from
// the announced attacks, plus those the staged declaration is attacking
// that no announcement has recorded yet. The trigger's target is chosen as
// the declaration is announced, one attacker at a time, so the record alone
// would miss attackers that are declared with the Soulcaller but announced
// after it (CR 508.1, 603.3d).
func custodiSoulcallerPlayersAttacked(g *game.Game, you uuid.UUID) int {
	if g.Turn.ActiveSeat < 0 || g.Turn.ActiveSeat >= len(g.Seats) || g.Seats[g.Turn.ActiveSeat].ID != you {
		return 0
	}
	seats := map[uuid.UUID]bool{}
	for _, p := range g.Seats {
		if p != nil && p.ID != you {
			seats[p.ID] = true
		}
	}
	attacked := map[uuid.UUID]bool{}
	for _, a := range g.TurnTally.Attacks {
		if a.PhaseID == g.Turn.PhaseID && seats[a.Defender] {
			attacked[a.Defender] = true
		}
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == you && c.IsCreature() && seats[c.AttackingTarget] {
			attacked[c.AttackingTarget] = true
		}
	}
	return len(attacked)
}
