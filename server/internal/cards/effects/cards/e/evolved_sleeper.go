package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Evolved Sleeper — Creature — Human {B}, 1/1 (Dominaria United):
//
//	"{B}: This creature becomes a Human Cleric with base power and toughness
//	 2/2.
//	 {1}{B}: If this creature is a Cleric, put a deathtouch counter on it and
//	 it becomes a Phyrexian Human Cleric with base power and toughness 3/3.
//	 {1}{B}{B}: If this creature is a Phyrexian, put a +1/+1 counter on it,
//	 then you draw a card and you lose 1 life."
//
// Figure of Fable's shape: the first two abilities each add a scoped effect
// pinned to this object with no duration (until it leaves the battlefield),
// setting the creature types (CR 205.1a) and base power and toughness
// (layer 7b). Each "if this creature is a …" is read as the ability
// resolves, so any of them can be activated whatever the creature is now and
// does nothing when the condition is false (the deathtouch counter included).
// The third has no type change: it puts a +1/+1 counter on this creature,
// then draws and loses 1 life, all only if it is a Phyrexian as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "12367806-5924-485c-814a-5b67dd79235e",
		Name:         "Evolved Sleeper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{B}: This creature becomes a Human Cleric with base power and toughness 2/2.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{B}"),
				Effect:  evolvedSleeperBecomes("", []string{"Human", "Cleric"}, 2, 2, false),
			},
			{
				Label:   "{1}{B}: If this creature is a Cleric, put a deathtouch counter on it and it becomes a Phyrexian Human Cleric with base power and toughness 3/3.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{1}{B}"),
				Effect:  evolvedSleeperBecomes("Cleric", []string{"Phyrexian", "Human", "Cleric"}, 3, 3, true),
			},
			{
				Label:   "{1}{B}{B}: If this creature is a Phyrexian, put a +1/+1 counter on it, then you draw a card and you lose 1 life.",
				Purpose: game.Purpose{Draws: 1, Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{1}{B}{B}"),
				Effect:  evolvedSleeperGrow,
			},
		},
	})
}

// evolvedSleeperBecomes is one of the first two abilities: needs is the type
// the source must have as it resolves ("" for none), types the creature types
// it becomes, and counter whether it first gets a deathtouch counter.
func evolvedSleeperBecomes(needs string, types []string, power, toughness int, counter bool) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		if !effects.SourceIsStillThisPermanent(g, item) {
			return nil
		}
		ctx := effects.NewContext(g, item)
		if needs != "" {
			src, ok := g.LookupCardForEffect(item.SourceCardID)
			if !ok || !src.HasSubtype(ctx.TextSubtype(needs)) {
				return nil
			}
		}
		if counter {
			if err := (effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterDeathtouch, N: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		mods := append([]game.Mod{game.SetCreatureTypesMod(ctx.TextSubtypes(types...)...)},
			game.SetBasePTMods(power, toughness)...)
		return effects.ScopedEffectFor{
			Target:   item.SourceCardID,
			Mods:     mods,
			Duration: g.PinnedTo(game.IndefiniteDuration(), item.SourceCardID),
			Label:    "Evolved Sleeper — becomes a " + types[0] + " " + types[len(types)-1],
		}.Apply(ctx)
	}
}

// evolvedSleeperGrow is the third ability: nothing happens unless this
// creature is a Phyrexian as it resolves.
func evolvedSleeperGrow(g *game.Game, item *game.StackItem) error {
	if !effects.SourceIsStillThisPermanent(g, item) {
		return nil
	}
	ctx := effects.NewContext(g, item)
	src, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || !src.HasSubtype(ctx.TextSubtype("Phyrexian")) {
		return nil
	}
	if err := (effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	if err := (effects.DrawCards{N: 1}).Apply(ctx); err != nil {
		return err
	}
	return g.ChangePlayerLifeForEffect(item.SourceCardID, ctx.Controller(), -1)
}
