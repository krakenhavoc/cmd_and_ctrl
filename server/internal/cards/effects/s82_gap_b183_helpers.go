package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_gap_b183_helpers.go — shared pieces of GAP batch 183 (#5948, Starter
// Commander Decks).

// GapB183RegisterTappedDual registers a land that "enters tapped" and
// taps for either of two colours (the Guildgates' shape), for the four lands
// of this batch that print exactly those two lines.
func GapB183RegisterTappedDual(oracleID, name, a, b string) {
	Register(Spec{
		OracleID:      oracleID,
		Name:          name,
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{DualManaAbility(a, b)},
	})
}

// GapB183CastSubtypeSpellsFromGraveyardThisTurn is "you may cast <subtype>
// spells from your graveyard this turn": a standing cast permission over the
// controller's graveyard for spells with the subtype, which the engine ends
// with the turn. Unlike Yawgmoth's Will it replaces nothing.
type GapB183CastSubtypeSpellsFromGraveyardThisTurn struct {
	Subtype string
}

func (e GapB183CastSubtypeSpellsFromGraveyardThisTurn) Apply(ctx *Context) error {
	perm := game.CastPermission{
		Player: ctx.Controller(),
		Zone:   game.ZoneGraveyard,
		Scope:  game.ScopeStanding,
		Source: ctx.Source(),
		Label:  "Cast a " + e.Subtype + " spell from your graveyard",
	}
	perm.Filter.Subtype = e.Subtype
	ctx.Game.GrantCastPermissionForEffect(perm)
	return nil
}
