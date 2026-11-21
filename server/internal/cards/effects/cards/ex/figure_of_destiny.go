package cardsex

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Figure of Destiny — Creature — Kithkin {R/W}, 1/1 (S82 GAP, #5948):
//
//	"{R/W}: This creature becomes a Kithkin Spirit with base power and
//	 toughness 2/2.
//	 {R/W}{R/W}{R/W}: If this creature is a Spirit, it becomes a Kithkin Spirit
//	 Warrior with base power and toughness 4/4.
//	 {R/W}{R/W}{R/W}{R/W}{R/W}{R/W}: If this creature is a Warrior, it becomes
//	 a Kithkin Spirit Warrior Avatar with base power and toughness 8/8, flying,
//	 and first strike."
//
// Figure of Fable's shape: no duration (each effect lasts while it is on the
// battlefield), a later type or base P/T set overwrites an earlier one, and
// the "if this creature is a …" is read as the ability resolves.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd51cc2d-0d39-4aba-b0ea-70e8118149bc",
		Name:         "Figure of Destiny",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			figureOfDestinyStage("{R/W}: This creature becomes a Kithkin Spirit with base power and toughness 2/2.",
				"{R/W}", "", []string{"Kithkin", "Spirit"}, 2, 2, nil),
			figureOfDestinyStage("{R/W}{R/W}{R/W}: If this creature is a Spirit, it becomes a Kithkin Spirit Warrior with base power and toughness 4/4.",
				"{R/W}{R/W}{R/W}", "Spirit", []string{"Kithkin", "Spirit", "Warrior"}, 4, 4, nil),
			figureOfDestinyStage("{R/W}{R/W}{R/W}{R/W}{R/W}{R/W}: If this creature is a Warrior, it becomes a Kithkin Spirit Warrior Avatar with base power and toughness 8/8, flying, and first strike.",
				"{R/W}{R/W}{R/W}{R/W}{R/W}{R/W}", "Warrior", []string{"Kithkin", "Spirit", "Warrior", "Avatar"}, 8, 8, []string{"flying", "first strike"}),
		},
	})
}

// figureOfDestinyStage is one of the three abilities. needs is the creature
// type the source must be as it resolves ("" for none); keywords are what the
// last stage adds.
func figureOfDestinyStage(label, cost, needs string, types []string, power, toughness int, keywords []string) effects.ActivatedAbility {
	return effects.ActivatedAbility{
		Label:   label,
		Purpose: game.Purpose{Answers: game.AnswerPump},
		Cost:    effects.ManaCost(cost),
		Effect: func(g *game.Game, item *game.StackItem) error {
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
			mods := append([]game.Mod{game.SetCreatureTypesMod(ctx.TextSubtypes(types...)...)},
				game.SetBasePTMods(power, toughness)...)
			if len(keywords) > 0 {
				mods = append(mods, game.AddKeywordsMod(keywords...))
			}
			return effects.ScopedEffectFor{
				Target:   item.SourceCardID,
				Mods:     mods,
				Duration: g.PinnedTo(game.IndefiniteDuration(), item.SourceCardID),
				Label:    "Figure of Destiny — becomes a " + types[len(types)-1] + " " + fmt.Sprintf("%d/%d", power, toughness),
			}.Apply(ctx)
		},
	}
}
