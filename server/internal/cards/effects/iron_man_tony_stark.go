package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Iron Man, Tony Stark — Legendary Artifact Creature — Human Hero
// {3}{R}{R}, 3/3 (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Attacking creatures you control get +1/+0.
//	 Whenever you cast a red spell, create a 2/1 colorless Robot Hero
//	 artifact creature token with flying."
//
// The anthem is a layer 7c +1/+0 over the controller's attacking
// creatures (b16Anthem, with Ohran Frostfang's attacking-status read and
// DependsOnAttackingStatus so entering or leaving combat rebuilds the
// layer cache). The cast trigger is Runaway Steam-Kin's "red spell"
// watch with a CreateToken body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "188e1516-5766-406b-aacb-468f2d9a530a",
		Name:            "Iron Man, Tony Stark",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{ironManTonyStarkAnthem()},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(OfColor("R"), "Iron Man — create a 2/1 Robot Hero with flying",
				Do(CreateToken{Template: TokenCard("2/1 colorless Robot Hero artifact with flying"), N: 1})),
		},
	})
}

func ironManTonyStarkAnthem() game.StaticAbility {
	st := b16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsCreature() && target.Controller == source.Controller && target.AttackingTarget != uuid.Nil
	}, 1, 0)
	st.Label = "Attacking creatures you control get +1/+0."
	st.DependsOnAttackingStatus = true
	return st
}
