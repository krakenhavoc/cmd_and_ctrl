package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Outpace Oblivion — Enchantment {2}{R}:
//
//	"Start your engines!
//	 When this enchantment enters, it deals 5 damage to up to one target
//	 creature or planeswalker.
//	 {2}, Sacrifice this enchantment: It deals 2 damage to each player who
//	 doesn't have max speed."
//
// ADR 0138 (#2711). The ETB is a real "up to one" target clause. The
// activated ability reads each player's speed as it resolves (CR 608.2h)
// and deals the damage as one simultaneous instance; the source's
// controller is hit too unless they have max speed. Eliminated seats are
// skipped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4d92f48d-c851-4b71-b7fa-916459928d59",
		Name:            "Outpace Oblivion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{StartYourEngines},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Outpace Oblivion — 5 damage to up to one target creature or planeswalker", outpaceOblivionBlast),
				TargetPermanent("up to one target creature or planeswalker", Or(Creature(), Planeswalker())).WithCount(0, 1)),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, Sacrifice this enchantment: It deals 2 damage to each player who doesn't have max speed.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}"), SacrificeThis()),
			Effect:  outpaceOblivionSweep,
		}},
	})
}

// outpaceOblivionBlast is the ETB: 5 damage to the chosen creature or
// planeswalker, if it is still a legal target (CR 608.2b).
func outpaceOblivionBlast(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 5}.Apply(ctx)
		}
	}
	return nil
}

// outpaceOblivionSweep is the sacrifice ability: 2 damage to each player
// without max speed, dealt together.
func outpaceOblivionSweep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var hit []uuid.UUID
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || YouHaveMaxSpeed(g, p.ID) {
			continue
		}
		hit = append(hit, p.ID)
	}
	return g.DamageInstanceForEffect(func() error {
		for _, id := range hit {
			if err := (DealDamage{Source: item.SourceCardID, Target: id, Amount: 2}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}
