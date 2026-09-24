package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Eternal Wanderer — Legendary Planeswalker {4}{W}{W}, loyalty 5:
//
//	"No more than one creature can attack The Eternal Wanderer each
//	 combat.
//	 +1: Exile up to one target artifact or creature. Return that card
//	     to the battlefield under its owner's control at the beginning
//	     of that player's next end step.
//	 0: Create a 2/2 white Samurai creature token with double strike.
//	 −4: For each player, choose a creature that player controls. Each
//	     player sacrifices all creatures they control not chosen this
//	     way."
//
// The proof card for #1534's per-permanent attack scope
// (AttackLimitAttackingThis, ADR 0045 amendment of 2026-09-24,
// Decision 47): only an attack that names the Wanderer herself counts,
// so her controller — and any other planeswalker they control — may
// still be attacked by any number of creatures in the same combat.
//
// The loyalty abilities:
//
//   - The 0 is CreateToken.
//   - The −4 is Tragic Arrogance's shape (ChoosePermanents with a
//     non-owner chooser, one leg per player, APNAP from the active
//     player): the Wanderer's controller picks one creature on each
//     board, then every creature not picked is sacrificed as one
//     event. A player with no creatures is skipped.
//   - The +1 is the flicker.go delayed blink
//     (exileTargetsThenScheduleReturn). DECLARED WEAKER THAN PRINTED:
//     the return fires at the NEXT end step, not "that player's" —
//     a CR 603.7 delayed trigger can be bound to its controller's turn
//     (ControllerTurnOnly) but not to a third player's, and binding it
//     by making the owner its controller would misstate CR 603.7d. So
//     an opponent's creature exiled on your turn comes back at your
//     end step instead of theirs; your own comes back exactly when
//     printed, since the ability is activated on your turn. The gap is
//     a DelayedTrigger field (a named turn-owner), not this card.
func init() {
	Register(Spec{
		OracleID:     "20a1671d-e8a4-4cf1-87a7-f2f6319f4b9e",
		Name:         "The Eternal Wanderer",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The +1 returns the exiled card at the next end step, even when that isn't its owner's turn.",
		},
		// The fallback for tokens, fixtures and the dev spawner; an
		// imported deck reads printed loyalty (ADR 0032 §1).
		StartingLoyalty: 5,
		AttackLimits:    []game.AttackLimit{NoMoreThanNCanAttackThisEachCombat(1)},
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Exile up to one target artifact or creature. Return that card to the battlefield under its owner's control at the beginning of that player's next end step.",
				Cost:    LoyaltyCost(1),
				Targets: TargetPermanent("up to one target artifact or creature", Or(Artifact(), Creature())).WithCount(0, 1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					// "Up to one" with nothing chosen resolves and does
					// nothing.
					if len(item.Targets) == 0 {
						return nil
					}
					return exileTargetsThenScheduleReturn(NewContext(g, item), "The Eternal Wanderer — return the exiled card")
				},
			},
			{
				Label: "0: Create a 2/2 white Samurai creature token with double strike.",
				Cost:  LoyaltyCost(0),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return CreateToken{
						Controller: ctx.Controller(),
						Template:   TokenCard("2/2 white Samurai with double strike"),
						N:          1,
					}.Apply(ctx)
				},
			},
			{
				Label: "−4: For each player, choose a creature that player controls. Each player sacrifices all creatures they control not chosen this way.",
				Cost:  LoyaltyCost(-4),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return ChoosePermanents{
						Question:   "The Eternal Wanderer — choose the creature that player keeps",
						Of:         seatsFromActive(g),
						Candidates: eternalWandererCandidates,
						Then:       eternalWandererSacrificeTheRest,
					}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

// eternalWandererCandidates is one player's creatures, exactly one to
// be chosen. A player with none is skipped.
//
// Caller holds g.mu.
func eternalWandererCandidates(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
	var ids []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == of && c.IsCreature() {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) == 0 {
		return nil, 0, 0
	}
	return ids, 1, 1
}

// eternalWandererSacrificeTheRest is "each player sacrifices all
// creatures they control not chosen this way", as one event.
//
// Caller holds g.mu.
func eternalWandererSacrificeTheRest(ctx *Context, picked game.PromptedPicks) error {
	g := ctx.Game
	var doomed []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if !c.IsCreature() || picked.Contains(c.InstanceID) {
			continue
		}
		doomed = append(doomed, c.InstanceID)
	}
	if len(doomed) == 0 {
		return nil
	}
	return g.SacrificeAllThenForEffect(ctx.Source(), doomed, nil)
}
