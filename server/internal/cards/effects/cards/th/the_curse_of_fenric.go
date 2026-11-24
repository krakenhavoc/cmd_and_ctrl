package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Curse of Fenric — Enchantment — Saga {2}{G}{W} (Doctor Who, #6100):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — For each player, destroy up to one target creature that player
//	 controls. For each creature destroyed this way, its controller
//	 creates a 3/3 green Mutant creature token with deathtouch.
//	 II — Target nontoken creature becomes a 6/6 legendary Horror creature
//	 named Fenric and loses all abilities.
//	 III — Target Mutant fights another target creature named Fenric."
//
// Chapter I is the per-player clause list Battle at the Helvault uses: one
// "up to one" clause per seat, each bound to its own player, so the CR
// 608.2b re-check asks "does THAT player still control it?". The picks are
// destroyed together and the tokens follow the ones really destroyed (CR
// 701.8a: put into a graveyard; an indestructible or regenerated creature
// was not), each for the player who controlled that creature as it last
// existed on the battlefield (CR 608.2h).
//
// Chapter II is one effect with no duration (CR 611.2a), its mods at one
// timestamp: the name in layer 3 (CR 612.8), "a legendary Horror
// creature" in layer 4 with CR 205.1a's meaning (its other card types and
// creature types go, the supertype is gained), every ability lost in layer
// 6 and base 6/6 in layer 7b. It stays after the Saga is gone.
//
// Chapter III fights the two picks (CR 701.14a); if either is no longer a
// legal target, neither deals damage (CR 701.14b). "Named Fenric" reads the
// name after the layers, so a creature chapter II renamed qualifies.
//
// No simplification.
const fenricName = "The Curse of Fenric"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef7fce1b-0d95-4f92-a8ad-09d9756fa0e4",
		Name:         fenricName,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1,
				effects.SagaChapterLabel(fenricName, 1, "for each player, destroy up to one target creature that player controls; for each creature destroyed this way, its controller creates a 3/3 green Mutant creature token with deathtouch"),
				effects.Gap110PerPlayerClauses("creature", effects.Creature()), fenricDestroy),
			effects.ChapterTriggerTargeting(2,
				effects.SagaChapterLabel(fenricName, 2, "target nontoken creature becomes a 6/6 legendary Horror creature named Fenric and loses all abilities"),
				effects.TargetCreature("target nontoken creature", effects.Not(effects.IsTokenPredicate())), fenricBecome),
			effects.ChapterTriggerTargeting(3,
				effects.SagaChapterLabel(fenricName, 3, "target Mutant fights another target creature named Fenric"),
				effects.Clauses(
					effects.TargetCreature("target Mutant", effects.Subtype("Mutant")),
					effects.Distinct(effects.TargetCreature("another target creature named Fenric", fenricNamedFenric)),
				),
				func(g *game.Game, item *game.StackItem) error {
					return effects.FightTargetClauses(effects.NewContext(g, item))
				}),
		},
	})
}

// fenricNamedFenric is "named Fenric", on the name after the layers.
func fenricNamedFenric(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	name := c.Effective().Name
	if name == "" {
		name = c.Name
	}
	return name == "Fenric"
}

// fenricDestroy is chapter I: destroy every still-legal pick together, then
// make a Mutant for the controller of each one really destroyed.
func fenricDestroy(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var doomed []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			doomed = append(doomed, t.ID)
		}
	}
	if len(doomed) == 0 {
		return nil
	}
	return g.DestroyPermanentsKeyedForEffect(doomed, ctx.ItemTail(fenricMutants))
}

// fenricMutants is "for each creature destroyed this way, its controller
// creates a 3/3 green Mutant creature token with deathtouch". An on-disk
// key: never renamed, never reused.
var fenricMutants = effects.CardTail("then/the-curse-of-fenric-mutants", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		info, ok := ctx.Game.LastKnownPermanentForEffect(id)
		if !ok || info.Controller == uuid.Nil {
			continue
		}
		if err := (effects.CreateToken{
			Controller: info.Controller,
			Template:   effects.TokenCard("3/3 green Mutant with deathtouch"),
			N:          1,
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

// fenricBecome is chapter II.
func fenricBecome(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	mods := []game.Mod{
		game.SetNameMod("Fenric"),
		game.SetTypesMod([]string{"Creature"}, []string{"Horror"}, "Legendary"),
		game.LoseAllAbilitiesMod(),
	}
	return effects.ScopedEffectFor{
		Target:   t.ID,
		Mods:     append(mods, game.SetBasePTMods(6, 6)...),
		Duration: game.IndefiniteDuration(),
		Label:    fenricName + " — a 6/6 legendary Horror creature named Fenric with no abilities",
	}.Apply(ctx)
}
