package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const corporealProjectionGrant = "corporeal-projection/myriad" // an on-disk identity: never renamed

// Corporeal Projection — Sorcery {U}{R} (Ravnica: Clue Edition):
//
//	"Target creature you control gains myriad until end of turn. (Whenever
//	 it attacks, for each opponent other than defending player, you may
//	 create a token copy that's tapped and attacking that player or a
//	 planeswalker they control. Exile the tokens at end of combat.)
//	 Overload {3}{U}{U}{R}{R} (You may cast this spell for its overload
//	 cost. If you do, change "target" in its text to "each.")"
//
// Myriad is the shared granted trigger (myriad.go), a layer-6 grant until end
// of turn (ADR 0093). Overloaded, it has no target and every creature you
// control as it resolves gains myriad (CR 611.2c: a creature that enters
// later gets none).
func init() {
	Register(Spec{
		OracleID:     "09038989-3c1f-4426-a150-e0a9374f34c3",
		Name:         "Corporeal Projection",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0, "myriad")),
		Grants: []AbilityGrant{{
			Key:       corporealProjectionGrant,
			Triggered: []game.TriggeredAbility{GrantedMyriad("Corporeal Projection")},
			Text:      "Myriad",
		}},
		AlternativeCosts: []game.AlternativeCost{Overload("{3}{U}{U}{R}{R}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			grant := GrantAbilitiesFor{
				Keys:  []string{corporealProjectionGrant},
				Label: "Corporeal Projection — gains myriad until end of turn",
			}
			if ctx.PaidAltCost("overload") {
				grant.Match = And(Creature(), YouControl())
				return grant.Apply(ctx)
			}
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			grant.Target = id
			return grant.Apply(ctx)
		},
	})
}
