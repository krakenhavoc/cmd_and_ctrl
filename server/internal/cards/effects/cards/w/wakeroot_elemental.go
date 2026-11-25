package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wakeroot Elemental — Creature — Elemental {4}{G}{G}, 5/5 (Core Set 2020):
//
//	"{G}{G}{G}{G}{G}: Untap target land you control. It becomes a 5/5
//	 Elemental creature with haste. It's still a land. (This effect lasts as
//	 long as that land remains on the battlefield.)"
//
// Woodcaller Automaton's animation: the land untaps, then one record pinned
// to that land object (CR 611.2c, CR 400.7) adds the Creature type and the
// Elemental subtype (layer 4), haste (layer 6) and sets base power and
// toughness to 5/5 (layer 7b). It has no other end: it lasts until the land
// leaves the battlefield. Nothing is removed, so the land keeps its land
// types and abilities.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "edc6db57-3519-4649-9c4a-5cf264cce26f",
		Name:         "Wakeroot Elemental",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}{G}{G}{G}{G}: Untap target land you control. It becomes a 5/5 Elemental creature with haste. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{G}{G}{G}{G}{G}"),
			Targets: effects.TargetPermanent("target land you control", effects.Land(), effects.YouControl()),
			Effect:  wakerootElementalAnimate,
		}},
	})
}

func wakerootElementalAnimate(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	land, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (effects.UntapTarget{Target: land}).Apply(ctx); err != nil {
		return err
	}
	return effects.ScopedEffectFor{
		Target: land,
		Mods: []game.Mod{
			game.AddTypesMod("Creature"),
			game.AddSubtypesMod(ctx.TextSubtypes("Elemental")...),
			game.SetBasePowerMod(5),
			game.SetBaseToughnessMod(5),
			game.AddKeywordsMod("haste"),
		},
		Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), land),
		Label:    "Wakeroot Elemental — a 5/5 Elemental creature with haste, still a land",
	}.Apply(ctx)
}
