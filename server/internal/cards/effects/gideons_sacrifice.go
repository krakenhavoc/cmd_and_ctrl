package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gideon's Sacrifice — Instant {W}:
//
//	"Choose a creature or planeswalker you control. All damage that would
//	 be dealt this turn to you and permanents you control is dealt to the
//	 chosen permanent instead (if it's still on the battlefield)."
//
// ADR 0108 §9 (#1905): the permanent is chosen as the spell resolves, not
// targeted. Damage to the chosen permanent itself stays where it is.
// The rulings: no creature or planeswalker to choose is no effect; a
// chosen permanent that has left, or is no longer a creature or
// planeswalker, redirects nothing (CR 614.9).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "1af63a5e-2bec-4f8d-a373-d9ce43a7d242",
		Name:         "Gideon's Sacrifice",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ChoosePermanents{
				Question:   "Gideon's Sacrifice — choose a creature or planeswalker you control",
				Candidates: creaturesOrPlaneswalkersOf,
				Then: func(ctx *Context, picked game.PromptedPicks) error {
					cards := picked.Cards()
					if len(cards) != 1 {
						return nil
					}
					return RedirectDamage{Protect: ShieldYouAndPermanentsYouControl, To: RedirectToObject(cards[0])}.Apply(ctx)
				},
			}.Apply(ctx)
		},
	})
}

// creaturesOrPlaneswalkersOf offers one of the creatures and planeswalkers
// `of` controls.
func creaturesOrPlaneswalkersOf(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == of && (c.IsCreature() || c.IsPlaneswalker()) {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) == 0 {
		return nil, 0, 0
	}
	return ids, 1, 1
}
