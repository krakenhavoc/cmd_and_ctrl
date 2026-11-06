package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Secret Tunnel — Land — Cave (Avatar: The Last Airbender, #4459):
//
//	"This land can't be blocked.
//	 {T}: Add {C}.
//	 {4}, {T}: Two target creatures you control that share a creature
//	 type can't be blocked this turn."
//
// "Can't be blocked" is the S24 restriction bit on the land itself
// (RestrictSelf, Blighted Agent's): it does nothing until something
// animates the land, and then the land creature can't be blocked.
//
// The activation's two targets are two clauses, the second related to the
// first (SharesACreatureTypeWithTarget, the shape Burglar's Plot's "share a
// card type" uses): the pair must share a creature type when it is chosen
// and again as the ability resolves, on creature types after the layers, so
// Ally and Lemur count and artifact, legendary or snow never do (the
// ruling). Distinct, because "two target creatures" is one instance of the
// word "target" (CR 115.3). If the first creature leaves in response, the
// second still compares against the creature types the first had as it
// left (CR 608.2h, the ruling); if the second leaves, the first is still
// legal. Each creature still legal gets the restriction for the turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "632e2979-d88a-482e-9bb8-57b683c5310f",
		Name:          "Secret Tunnel",
		Completeness:  CompletenessFull,
		Static:        []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
		ManaAbilities: []ManaAbility{colorlessTapRow()},
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}: Two target creatures you control that share a creature type can't be blocked this turn.",
			Purpose: ForTargets(CantBeBlockedUntilEOT(0, 0, 0), CantBeBlockedUntilEOT(1, 0, 0)),
			Cost:    Plus(ManaCost("{4}"), TapCost()),
			Targets: Clauses(
				TargetCreature("target creature you control", YouControl()),
				Distinct(SharesACreatureTypeWithTarget(
					TargetCreature("another target creature you control that shares a creature type with it", YouControl()), 0)),
			),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (RestrictUntilEOT{
						Target:       t.ID,
						Restrictions: game.CantBeBlocked,
						Label:        "Secret Tunnel — can't be blocked",
					}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
