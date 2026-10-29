package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Avengers Disassembled — Sorcery {1}{R}{R} (Marvel Super Heroes):
//
//	"Choose one or both —
//	 • Avengers Disassembled deals 3 damage to each creature.
//	 • Destroy target land. Its controller may search their library for a
//	   basic land card, put it onto the battlefield tapped, then shuffle."
//
// Brutal Expulsion's "one or both" shape (ChooseN 1..2). The first bullet is
// Anger of the Gods's untargeted sweep without the exile rider, declared as a
// sweep for the bots. The second is Sundering Eruption's destruction and
// search: the land's controller is read before it dies, the search is an
// optional prompt addressed to them and is offered whether or not the land
// really died (indestructible, regeneration), and it is skipped only when the
// target left before resolution, because then there is no controller to ask.
// With both bullets chosen the damage comes first, in printed order
// (CR 608.2c), and the land is destroyed afterwards.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "34493cf5-7586-48ef-806d-57273b2ab034",
		Name:         "Avengers Disassembled",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose one or both", 1, 2,
			ModeWithPurpose(ModeDoing("Avengers Disassembled deals 3 damage to each creature.", nil, avengersDisassembledSweep),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3}}),
			ModeDoing("Destroy target land. Its controller may search their library for a basic land card, put it onto the battlefield tapped, then shuffle.",
				TargetPermanent("target land", Land()), avengersDisassembledLand),
		),
	})
}

// avengersDisassembledSweep is the first bullet: 3 damage to each creature.
func avengersDisassembledSweep(_ *game.StackItem, ctx *Context, _ int) error {
	return damageEachMatching(ctx, Creature(), 3)
}

// avengersDisassembledLand is the second bullet.
func avengersDisassembledLand(_ *game.StackItem, ctx *Context, occurrence int) error {
	target, ok := ModeTarget(ctx, occurrence)
	if !ok || target.Kind != game.TargetCard {
		return nil
	}
	victim, ok := controllerOfTarget(ctx, target.ID)
	if !ok {
		return nil
	}
	if err := (DestroyTarget{Target: target.ID}).Apply(ctx); err != nil {
		return err
	}
	return SearchLibrary{
		Player:        victim,
		Predicate:     IsBasicLand,
		Dest:          game.ZoneBattlefield,
		Limit:         1,
		Shuffle:       true,
		TappedOnEntry: true,
		Optional:      true,
		Reason:        "Avengers Disassembled — you may search for a basic land",
	}.Apply(ctx)
}
