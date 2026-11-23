package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Woodcaller Automaton — Artifact Creature — Construct {10}, 8/8 (The
// Brothers' War, #5656): "Prototype {2}{G}{G} — 3/3", "When this creature
// enters, if you cast it, untap target land you control. It becomes a
// Treefolk creature with haste and base power and toughness equal to this
// creature's power and toughness. It's still a land."
//
// Prototype is the card's second face (CR 718, game/prototype.go), built
// by the deck importer from the inset; nothing here spells it out. "If you
// cast it" is an intervening if (CR 603.4), rechecked as the ability
// resolves. The land untaps, then becomes a creature for good (no
// duration): one record pinned to that land object (CR 611.2c, CR 400.7)
// adds the Creature type and the Treefolk subtype (layer 4), haste (layer
// 6) and sets base power and toughness (layer 7b), the shape awaken uses.
// Nothing is removed, so it keeps its land types and abilities. The size is
// the Automaton's power and toughness as the ability resolves (3/3
// prototyped, 8/8 otherwise, counters included), or as it last existed
// (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "124c3af4-8545-479a-b44c-446da1c8b535",
		Name:         "Woodcaller Automaton",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.S82BroB21EntersIfCast("Woodcaller Automaton — untap target land you control; it becomes a Treefolk creature with haste and base power and toughness equal to this creature's power and toughness",
					woodcallerAutomatonAnimate),
				effects.TargetPermanent("target land you control", effects.Land(), effects.YouControl())),
		},
	})
}

func woodcallerAutomatonAnimate(ctx *effects.Context) error {
	land := effects.FirstLegalCardTarget(ctx)
	if land == uuid.Nil {
		return nil
	}
	if err := (effects.UntapTarget{Target: land}).Apply(ctx); err != nil {
		return err
	}
	info, ok := ctx.SourcePermanent()
	if !ok {
		return nil
	}
	return effects.ScopedEffectFor{
		Target: land,
		Mods: []game.Mod{
			game.AddTypesMod("Creature"),
			game.AddSubtypesMod(ctx.TextSubtypes("Treefolk")...),
			game.SetBasePowerMod(info.Power),
			game.SetBaseToughnessMod(info.Toughness),
			game.AddKeywordsMod("haste"),
		},
		Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), land),
		Label:    "Woodcaller Automaton — a Treefolk creature with haste, still a land",
	}.Apply(ctx)
}
