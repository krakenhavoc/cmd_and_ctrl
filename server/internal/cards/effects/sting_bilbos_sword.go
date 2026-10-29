package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sting, Bilbo's Sword — Legendary Artifact — Equipment {2} (#4500, The
// Hobbit):
//
//	"Flash
//	 When Sting enters, put a hone counter on Sting for each creature
//	 target opponent controls. Attach Sting to up to one target creature
//	 you control. (Each hone counter on an Equipment grants +1/+0 to
//	 equipped creature.)
//	 Equip {3}"
//
// Two target clauses: the opponent (slot 0) and up to one creature you
// control (slot 1). The creatures are counted as the trigger resolves
// (CR 608.2), so a creature that died in response is not counted, and an
// opponent who is no longer a legal target gives a count of nothing. The
// counters go on first and the attach follows (the printed order), though
// neither depends on the other. The hone counters' +1/+0 for the equipped
// creature is the engine's: CR 122.1j, a layer 7c effect read off the
// Equipment (game/hone_counters.go), so it follows Sting when it is
// equipped to another creature and stops when Sting is unattached.
//
// "Up to one" with no creature picked attaches nothing and Sting still
// gets its counters (an unattached Equipment with hone counters is
// legal). An attach that cannot happen does nothing (CR 701.3b), and the
// attach checks Sting is still the permanent that asked (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9779f32c-b1a2-42a3-8e78-14c28c3ad254",
		Name:            "Sting, Bilbo's Sword",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Sting, Bilbo's Sword — hone counters for each creature target opponent controls, then attach Sting to up to one target creature you control",
					stingEnters),
				Clauses(
					TargetPlayer("target opponent", Opponent()),
					PermanentYouControl("up to one target creature you control", Creature()).WithCount(0, 1),
				)),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{3}"),
		},
	})
}

// stingEnters puts a hone counter on Sting for each creature the target
// opponent controls, then attaches Sting to the chosen creature.
func stingEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if p, ok := ctx.ClauseTarget(0); ok && p.Kind == game.TargetPlayer {
		n := 0
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == p.ID && c.IsCreature() {
				n++
			}
		}
		if n > 0 {
			if err := (AddCounter{Target: ctx.Source(), Kind: game.CounterHone, N: n}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	if c, ok := ctx.ClauseTarget(1); ok && c.Kind == game.TargetCard {
		return g.AttachSourceForEffect(item, c)
	}
	return nil
}
