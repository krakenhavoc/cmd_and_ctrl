package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Conqueror's Flail — Artifact — Equipment {2}:
//
//	"Equipped creature gets +1/+1 for each color among permanents you
//	 control.
//	 As long as this Equipment is attached to a creature, your
//	 opponents can't cast spells during your turn.
//	 Equip {2}"
//
// The pump is PumpAttachedPer's shape (Blackblade Reforged's counted
// bonus), with the count read live off the board by
// colorsAmongPermanentsControlledBy.
//
// The lockout is Dragonlord Dromoka's CastRestriction
// (OpponentsCantCastDuringYourTurn), CONDITIONED on the Flail actually
// being attached — Dromoka's version is unconditional (she IS the
// source), the Flail's only holds "as long as" it is on a creature.
// game.CastQuery.Source is a fresh copy of the permanent contributing
// the restriction, so reading its AttachedTo field is a live check,
// not a cached one: unequip the Flail and the very next cast query
// sees it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9aace0d3-89e6-4254-b96f-ee3a878f2f91",
		Name:         "Conqueror's Flail",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 1, effects.ColorsAmongPermanentsControlledBy),
		},
		CastRestrictions: []game.CastRestriction{
			{
				Label: "Conqueror's Flail — your opponents can't cast spells during your turn while it's attached to a creature",
				Forbids: func(q game.CastQuery) bool {
					if q.Source.AttachedTo.Kind != game.TargetCard {
						return false
					}
					if q.Source.Controller == q.Controller {
						return false
					}
					if !effects.IsActivePlayer(q.Game, q.Source.Controller) {
						return false
					}
					return effects.MatchCastCard(q.Game, nil, q.Controller, q.Card)
				},
			},
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{2}"),
		},
	})
}
