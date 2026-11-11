package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Theater of Horrors — Enchantment {1}{B}{R} (Duskmourn Commander):
//
//	"At the beginning of your upkeep, exile the top card of your library.
//	 During your turn, if an opponent lost life this turn, you may play lands
//	 and cast spells from among cards exiled with this enchantment.
//	 {3}{R}: This enchantment deals 1 damage to target opponent or
//	 planeswalker."
//
// Evendo Brushrazer's shape: the exile is linked to this object (ExileTarget
// with Linked), and a gated standing permission over exile, narrowed to the
// cards exiled with this source (ExiledWithSource) and to your own turn, is
// derived off the battlefield on every query. Its gate is "an opponent lost
// life this turn", read live off the turn tally as the play is tried
// (opponentLifeLostThisTurn); it opens when the life is lost and a card stays
// exiled and playable on later turns. The permission is not cast-only, so a
// land is played (using the land drop) and a spell cast for its mana cost. A
// Theater that left stops granting it: the cards are stranded (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b391a4ce-52ce-4a8c-bbe5-6053951ee1d4",
		Name:         "Theater of Horrors",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Theater of Horrors — exile the top card of your library", Do(Instruction(theaterOfHorrorsExile))),
		},
		GatedCastPermissions: []game.CastPermissionGate{{
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return opponentLifeLostThisTurn(g, controller) > 0
			},
			Permission: game.CastPermission{
				Zone:   game.ZoneExile,
				Filter: game.PermissionFilter{ExiledWithSource: true},
				Timing: game.TimingYourTurnOnly,
				Label:  "During your turn, if an opponent lost life this turn, play lands and cast spells from among cards exiled with this enchantment (Theater of Horrors)",
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{3}{R}: This enchantment deals 1 damage to target opponent or planeswalker.",
			Cost:    ManaCost("{3}{R}"),
			Targets: targetOpponentOrPlaneswalker(),
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Effect:  sourceDealsDamageToEachLegalTarget(1),
		}},
	})
}

// theaterOfHorrorsExile is "exile the top card of your library", linked to
// this Theater.
func theaterOfHorrorsExile(ctx *Context) error {
	p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
	if p == nil || p.Library == nil {
		return nil
	}
	top, err := p.Library.Top()
	if err != nil {
		return nil
	}
	return ExileTarget{Target: top.InstanceID, Linked: true}.Apply(ctx)
}
