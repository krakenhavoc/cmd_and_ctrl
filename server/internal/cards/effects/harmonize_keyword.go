package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// harmonize_keyword.go — #3931 (the keyword; harmonize.go is the card) (CR 702.180, ADR 0135 amendment 2026-10-16).

// AltCostKeyHarmonize is harmonize's offer key, printed or granted.
const AltCostKeyHarmonize = game.AltCostKeyHarmonize

// Harmonize is "Harmonize [cost] (You may cast this card from your
// graveyard for its harmonize cost. You may tap a creature you control
// to reduce that cost by an amount of generic mana equal to its power.
// Then exile this spell.)" — CR 702.180a:
//
//	CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
//	AlternativeCosts: []game.AlternativeCost{Harmonize("{4}{G}")},
//
// Flashback's place and exile (FromZone graveyard, ExileOnLeavingStack)
// with a tap of up to one untapped creature you control as its card
// component (game.HarmonizeTapCost), flagged ReducedByTappedPower: the
// pricer takes the named creature's power off the harmonize cost's
// generic part (an {X} included, once the printed generic is gone),
// after any increase and never below {0} (CR 601.2f). The creature is
// chosen with the cost and tapped as it is paid, with the spell on the
// stack (CR 702.180b); it may have arrived this turn (CR 302.6), and it
// cannot also tap for the spell's mana. Naming none is a legal payment.
func Harmonize(cost string) game.AlternativeCost {
	return game.AlternativeCost{
		Key:                  AltCostKeyHarmonize,
		Label:                "Harmonize " + cost,
		ManaCost:             cost,
		FromZone:             game.ZoneGraveyard,
		ExileOnLeavingStack:  true,
		TapOthers:            game.HarmonizeTapCost(),
		ReducedByTappedPower: true,
		PayLabel:             game.HarmonizeTapLabel,
	}
}

// GrantHarmonizeToCard is Songcrafter Mage's clause: "target instant or
// sorcery card in your graveyard gains harmonize until end of turn. Its
// harmonize cost is equal to its mana cost" (CR 702.180).
//
// GrantFlashbackToCard's shape under the harmonize key: the permission
// names the card OBJECT (CR 400.7), leaves its cost empty ("equal to its
// mana cost"), and sets ExileOnResolution for CR 702.180a's exile. The
// key is what gives the granted offer harmonize's tap and reduction
// (CastPermission.AlternativeCostFor), so a granted harmonize prices
// exactly as a printed one does.
type GrantHarmonizeToCard struct {
	// Target is the card in a graveyard that gains harmonize.
	Target uuid.UUID
	// Label is the clause as printed, for the client's cost picker.
	Label string
}

func (e GrantHarmonizeToCard) Apply(ctx *Context) error {
	if e.Target == uuid.Nil {
		return nil
	}
	label := e.Label
	if label == "" {
		label = "Harmonize — its mana cost"
	}
	ctx.Game.GrantCastPermissionOverCardForEffect(e.Target, game.CastPermission{
		Player:            ctx.Controller(),
		Zone:              game.ZoneGraveyard,
		AltCostKey:        AltCostKeyHarmonize,
		ExileOnResolution: true,
		// Duration left zero: "until end of turn", as for flashback.
		Source: ctx.Source(),
		Label:  label,
	})
	return nil
}
