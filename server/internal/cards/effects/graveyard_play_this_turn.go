package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// GraveyardPlayThisTurn is the whole printed clause pair of Yawgmoth's
// Will (ADR 0108 §4, #1823):
//
//	"Until end of turn, you may play lands and cast spells from your
//	 graveyard. If a card would be put into your graveyard from anywhere
//	 this turn, exile that card instead."
//
// It writes both halves from one place, so neither can ship without the
// other. The permission alone would be stronger than printed: the
// graveyard would refill as it was played from.
//
//   - The permission is a stored ScopeStanding CastPermission over the
//     controller's graveyard, until end of turn, paying the printed
//     cost (CR 601.2). It is "a rule over the zone", not a set locked at
//     resolution, because the cards that arrive later are exiled anyway
//     and the ones already there are the ones it means.
//   - The replacement is a ModExileInsteadOfYourGraveyard record
//     (CR 614.1a): it applies to costs and effects alike, and to the
//     spell that made it, which goes to the graveyard after its effect
//     starts (the Yawgmoth's Will ruling).
//
// SpellsOnly is Forgotten Cellar's "you may cast spells from your
// graveyard this turn", which opens no land play.
type GraveyardPlayThisTurn struct {
	SpellsOnly bool
	// Label names the permission in the client's cost picker.
	Label string
}

// YawgmothsWillThisTurn is "you may play lands and cast spells from
// your graveyard this turn, and cards that would go there are exiled
// instead", the shape of three of the four cards that print it.
func YawgmothsWillThisTurn() Applier { return GraveyardPlayThisTurn{} }

func (e GraveyardPlayThisTurn) Apply(ctx *Context) error {
	label := e.Label
	if label == "" {
		label = "Play from your graveyard"
		if e.SpellsOnly {
			label = "Cast from your graveyard"
		}
	}
	perm := game.CastPermission{
		Player: ctx.Controller(),
		Zone:   game.ZoneGraveyard,
		Scope:  game.ScopeStanding,
		Source: ctx.Source(),
		Label:  label,
	}
	perm.Filter.NonLandOnly = e.SpellsOnly
	ctx.Game.GrantCastPermissionForEffect(perm)
	ctx.Game.ExileInsteadOfYourGraveyardThisTurnForEffect(ctx.Source(), ctx.Controller(),
		"exile instead of your graveyard this turn")
	return nil
}
