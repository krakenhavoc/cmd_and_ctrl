package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Infamous Cruelclaw — Legendary Creature — Weasel Mercenary {1}{B}{R},
// 3/3:
//
//	"Menace
//	 Whenever The Infamous Cruelclaw deals combat damage to a player,
//	 exile cards from the top of your library until you exile a nonland
//	 card. You may cast that card by discarding a card rather than paying
//	 its mana cost."
//
// ADR 0135 §2 (#2412): the cast is a per-card CastPermission priced by
// DiscardCard, which synthesises an offer with no mana and one card from
// your hand DISCARDED as a cost (CR 118.9, 701.9a), so a discard payoff
// sees it and a caster with an empty hand cannot claim it (CR 118.3).
// Amped Raptor's grant otherwise: flash timing (the rulings ignore the
// card's own timing), an {X} in its cost is 0 (CR 107.3b), and a window
// that closes when you pass priority, leaving an uncast card in exile.
// The lands exiled on the way stay in exile.
//
// #4250: the cast is made DURING the trigger's resolution (CR 608.2g), through the
// resolution_cast prompt over the card (offerResolutionCastOver), and
// the decline leaves it where it is. Nobody receives priority in between.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e229d55c-9513-437d-bd67-015ac46aaa18",
		Name:            "The Infamous Cruelclaw",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("The Infamous Cruelclaw — exile until a nonland card; cast it by discarding a card", CruelclawHits),
		},
	})
}

// cruelclawAltCostKey is the claim the permission synthesises. An
// on-disk identity (it lands on StackItem.AltCost): never renamed.
const CruelclawAltCostKey = "infamous_cruelclaw"

func CruelclawHits(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	return effects.ExileUntilNonlandGrantingCast(ctx, game.CastPermission{
		AltCostKey:  CruelclawAltCostKey,
		DiscardCard: true,
		Timing:      game.TimingFlash,
		CastOnly:    true,
		Duration:    ctx.Game.UntilEndOfTurnDuration(),
		LapseOnPass: game.LapseStaysInExile,
		Source:      ctx.Source(),
		SourceName:  "The Infamous Cruelclaw",
		Label:       "Discard a card (The Infamous Cruelclaw)",
	})
}
