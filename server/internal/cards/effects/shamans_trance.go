package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shaman's Trance — Instant {2}{R} (#1987):
//
//	"Other players can't play lands or cast spells from their graveyards
//	 this turn. You may play lands and cast spells from other players'
//	 graveyards this turn as though those cards were in your graveyard."
//
// ADR 0066 amendment 2026-10-16. Two halves, both until end of turn
// (CR 514.2), written from one resolution:
//
//   - The ban is one ModCantPlayFromOwnGraveyard record naming every
//     other player still in the game
//     (Game.CantPlayFromOwnGraveyardThisTurnForEffect). Both the cast
//     gate and the land-play gate read it, and only for a card in the
//     banned player's OWN graveyard (CR 101.2).
//   - The permission is one stored ScopeStanding CastPermission per
//     other player, over that player's graveyard (ZoneOwner), at the
//     printed cost: Yawgmoth's Will's grant pointed at another pile, the
//     shape Xanathar's library grant already uses (#1035). The caster
//     plays lands at their normal timing, spending a land drop
//     (CR 305.1, 305.2), and casts spells at their normal timing; a
//     spell cast this way is theirs to control (CR 601.2a) and goes to
//     its owner's graveyard afterwards (CR 400.3).
//
// "As though those cards were in your graveyard" is what lets a card's
// own graveyard text work for the caster: a flashback card in an
// opponent's graveyard may be flashed back (the 2004-10-04 ruling) as
// well as cast for its mana cost, and so may any card that prints a way
// to cast itself from its owner's graveyard. It opens plays only: the
// abilities of cards in other players' graveyards (unearth, embalm,
// scavenge) are not opened (the other 2004-10-04 ruling).
//
// Simplification, declared: a permission the CASTER holds over "your
// graveyard" (Underworld Breach's escape, Lurrus, Muldrotha) does not
// reach the other graveyards. Shaman's Trance itself already lets those
// cards be cast at their printed cost, so what is lost is the extra
// price such a permanent would offer. That is the weaker direction.
func init() {
	Register(Spec{
		OracleID:     shamansTranceOracle,
		Name:         shamansTranceName,
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Your own cards that give the cards in your graveyard a way to be cast, such as Underworld Breach, don't reach the cards in other players' graveyards.",
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			shamansTranceResolve(ctx.Game, ctx.Source(), ctx.Controller())
			return nil
		},
	})
}

const (
	shamansTranceOracle = "619d8d81-9e10-4353-b30d-a0f9c3451cd2"
	shamansTranceName   = "Shaman's Trance"
)

// shamansTranceResolve bans every other player's own graveyard and opens
// each of those graveyards to the caster, until end of turn.
func shamansTranceResolve(g *game.Game, source, controller uuid.UUID) {
	var others []uuid.UUID
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == controller {
			continue
		}
		others = append(others, p.ID)
	}
	g.CantPlayFromOwnGraveyardThisTurnForEffect(source, others,
		shamansTranceName+" — other players can't play lands or cast spells from their graveyards this turn")
	for _, other := range others {
		g.GrantCastPermissionForEffect(game.CastPermission{
			Player:    controller,
			Zone:      game.ZoneGraveyard,
			Scope:     game.ScopeStanding,
			ZoneOwner: other,
			Source:    source,
			Label:     "Play from other players' graveyards (" + shamansTranceName + ")",
		})
	}
}
