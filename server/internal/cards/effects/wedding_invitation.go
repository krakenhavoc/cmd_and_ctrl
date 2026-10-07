package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wedding Invitation — Artifact {2}:
//
//	"When this artifact enters, draw a card.
//	 {T}, Sacrifice this artifact: Target creature can't be blocked this
//	 turn. If it's a Vampire, it also gains lifelink until end of turn."
//
// The sacrifice is a cost, so the Invitation is gone before the ability
// resolves and the effect reads only the target. "If it's a Vampire" is
// read at resolution off the creature's effective subtypes (so a
// changeling counts). A target that left in response is skipped
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d9e752f2-d552-4d72-babd-29eca3508820",
		Name:         "Wedding Invitation",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{samiDrawOnETB("Wedding Invitation")},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice this artifact: Target creature can't be blocked this turn. If it's a Vampire, it also gains lifelink until end of turn",
			Cost:    Plus(TapCost(), SacrificeThis()),
			Targets: TargetCreature("target creature"),
			Effect:  weddingInvitationResolve,
		}},
	})
}

func weddingInvitationResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (RestrictUntilEOT{
			Target:       t.ID,
			Restrictions: game.CantBeBlocked,
			Label:        "Wedding Invitation — can't be blocked",
		}).Apply(ctx); err != nil {
			return err
		}
		if c, ok := g.LookupCardForEffect(t.ID); ok && c.HasSubtype("Vampire") {
			return GrantKeywordUntilEOT{
				Target:   t.ID,
				Keywords: []string{"lifelink"},
				Label:    "Wedding Invitation — lifelink",
			}.Apply(ctx)
		}
		return nil
	}
	return nil
}
