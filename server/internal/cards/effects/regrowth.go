package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Regrowth — "Return target card from your graveyard to your hand."
//
// The controller picks the card. Targets is "target card in your
// graveyard", validated at announce (CR 601.2c) and re-checked on
// resolution (CR 608.2b) like any other target.
//
// #338 stale-simplification sweep: this shipped in S14 auto-picking
// the MOST RECENTLY added card in the controller's graveyard,
// because there was no graveyard picker. S20 sub-PR 2 built one and
// converted Eternal Witness — the identical clause — to
// TargetCardInGraveyard, but Regrowth was left on the auto-pick. A
// player casting it got the top of the pile rather than the card
// they wanted, which is the whole point of the spell.
//
// Regrowth cannot return itself: at announce it is still on the
// stack, not in the graveyard, so it is never a legal candidate.
func init() {
	Register(Spec{
		OracleID: "e6e4a8bd-5c40-4654-8de1-0da9afed90fd",
		Name:     "Regrowth",
		Targets:  TargetCardInGraveyard("target card in your graveyard", YouOwn()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return ReturnFromGraveyard{
				Target: item.Targets[0].ID,
				Dest:   game.ZoneHand,
			}.Apply(ctx)
		},
	})
}
