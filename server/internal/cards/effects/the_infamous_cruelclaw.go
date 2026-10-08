package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
// SANDBOX SIMPLIFICATION, cascade's (cascade.go): the cast is made right
// after the trigger finishes resolving rather than during it. The
// pass-closed window keeps it no stronger than printed.
func init() {
	Register(Spec{
		OracleID:        "e229d55c-9513-437d-bd67-015ac46aaa18",
		Name:            "The Infamous Cruelclaw",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("The Infamous Cruelclaw — exile until a nonland card; cast it by discarding a card", cruelclawHits),
		},
	})
}

// cruelclawAltCostKey is the claim the permission synthesises. An
// on-disk identity (it lands on StackItem.AltCost): never renamed.
const cruelclawAltCostKey = "infamous_cruelclaw"

func cruelclawHits(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return exileUntilNonlandGrantingCast(ctx, game.CastPermission{
		AltCostKey:  cruelclawAltCostKey,
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
