package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Niv-Mizzet, Supreme — Legendary Creature — Dragon Avatar
// {W}{U}{B}{R}{G}, 5/5:
//
//	"Flying, hexproof from monocolored
//	 Each instant and sorcery card in your graveyard that's exactly two
//	 colors has jump-start. (You may cast that card from your graveyard
//	 by discarding a card in addition to paying its other costs. Then
//	 exile it.)"
//
// Hexproof from monocolored is game.HexproofFromMonocolored (CR 702.11d,
// Dragonfire Blade's keyword).
//
// The second line is a STANDING cast permission (ADR 0066), Lier's shape
// with jump-start's terms (CR 702.133a, #5645): derived off the
// battlefield, so a card that reaches the graveyard while Niv-Mizzet is
// out has jump-start at once and every one loses it when he leaves. It
// names no Cost, so the cast pays the card's own mana cost;
// DiscardCardInAddition is the card discarded as an additional cost, and
// ExileOnResolution is the exile whenever the spell would leave the stack.
// The filter judges the card in the graveyard (CardInZone): an instant or
// sorcery card that is exactly two colours there (ColorCountExactly), so
// a hybrid card and a split card of two colours qualify and a monocolored,
// colorless or three-colour card does not.
//
// A card that prints its own jump-start has two under Niv-Mizzet
// (CR 113.2c) at the same price, its mana cost and a discard, so the cast
// picker lists one (#3938: a granted instance that only repeats the
// printed price is not a second row).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4c3a865a-88b1-43e0-b326-17c45670da54",
		Name:            "Niv-Mizzet, Supreme",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", game.HexproofFromMonocolored},
		CastPermissions: []game.CastPermission{{
			Zone:                  game.ZoneGraveyard,
			Filter:                game.PermissionFilter{CardInZone: true, InstantOrSorceryOnly: true, ColorCountExactly: 2},
			AltCostKey:            game.AltCostKeyJumpStart,
			DiscardCardInAddition: true,
			ExileOnResolution:     true,
			Label:                 "Jump-start — discard a card (Niv-Mizzet, Supreme)",
		}},
	})
}
