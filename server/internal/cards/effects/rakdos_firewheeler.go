package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rakdos Firewheeler — Creature — Human Rogue {B}{B}{R}{R}, 4/3 (Ravnica
// Remastered, #5948):
//
//	"When this creature enters, it deals 2 damage to target opponent and 2
//	 damage to up to one target creature or planeswalker."
//
// Two target clauses on one trigger (Clauses), the opponent and an optional
// creature or planeswalker, each judged again as the trigger resolves
// (CR 608.2b): a clause whose target has become illegal is skipped and the
// other still resolves. The Firewheeler deals both.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f81327c2-a0e4-475d-bb89-8a42b1ddc93c",
		Name:         "Rakdos Firewheeler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Rakdos Firewheeler — 2 damage to target opponent and 2 damage to up to one target creature or planeswalker",
					rakdosFirewheelerDamage),
				Clauses(
					TargetPlayer("target opponent", Opponent()),
					TargetPermanent("up to one target creature or planeswalker", Or(Creature(), Planeswalker())).WithCount(0, 1),
				)),
		},
	})
}

func rakdosFirewheelerDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var player, permanent uuid.UUID
	for _, t := range ctx.LegalTargets() {
		switch t.Kind {
		case game.TargetPlayer:
			player = t.ID
		case game.TargetCard:
			permanent = t.ID
		}
	}
	// One sentence deals both, so both are one instance of damage (CR 615.8).
	return g.DamageInstanceForEffect(func() error {
		for _, victim := range []uuid.UUID{player, permanent} {
			if victim == uuid.Nil {
				continue
			}
			if err := (DealDamage{Source: item.SourceCardID, Target: victim, Amount: 2}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}
