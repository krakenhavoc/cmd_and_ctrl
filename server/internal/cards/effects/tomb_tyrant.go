package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tomb Tyrant — Creature — Zombie Noble {3}{B}, 3/3 (Midnight Hunt
// Commander):
//
//	"Other Zombies you control get +1/+1.
//	 {2}{B}, {T}, Sacrifice a creature: Return a Zombie creature card at
//	 random from your graveyard to the battlefield. Activate only during
//	 your turn and only if there are at least three Zombie creature cards
//	 in your graveyard."
//
// The lord is Bladestitched Skaab's. The condition is read when the
// ability is activated, before its cost (CR 602.2b): the sacrificed
// creature may itself be a Zombie that then joins the pool the random pick
// draws from, as the effect resolves. The pick is one of the engine's
// keyed random draws (ChooseAtRandomForEffect), never a private RNG.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0fb2c61a-9e74-4ee2-810a-b77a3b45f446",
		Name:         "Tomb Tyrant",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Zombie"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Activated: []ActivatedAbility{{
			Label:     "{2}{B}, {T}, Sacrifice a creature: Return a Zombie creature card at random from your graveyard to the battlefield. Activate only during your turn and only if there are at least three Zombie creature cards in your graveyard.",
			Purpose:   game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:      Plus(ManaCost("{2}{B}"), TapCost(), SacrificeN(1, "a creature", Creature())),
			Condition: AllConditions(DuringYourTurn(), GraveyardAtLeast(3, tombTyrantZombieCard)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				pick := g.ChooseAtRandomForEffect(randomDraw(ctx), graveyardIDs(g, ctx.Controller(), tombTyrantZombieCard), 1)
				if len(pick) == 0 {
					return nil
				}
				return ReturnFromGraveyard{Target: pick[0], Dest: game.ZoneBattlefield, Controller: ctx.Controller()}.Apply(ctx)
			},
		}},
	})
}

// tombTyrantZombieCard is "Zombie creature card" in a graveyard.
func tombTyrantZombieCard(c game.Card) bool {
	return c.IsCreature() && c.HasSubtypeWord("Zombie")
}
