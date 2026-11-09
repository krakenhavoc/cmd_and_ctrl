package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Figure of Fable — Creature — Kithkin {G/W} (Lorwyn Eclipsed, #4458):
//
//	"{G/W}: This creature becomes a Kithkin Scout with base power and
//	 toughness 2/3.
//	 {1}{G/W}{G/W}: If this creature is a Scout, it becomes a Kithkin Soldier
//	 with base power and toughness 4/5.
//	 {3}{G/W}{G/W}{G/W}: If this creature is a Soldier, it becomes a Kithkin
//	 Avatar with base power and toughness 7/8 and protection from each of
//	 your opponents."
//
// None of the three has a duration: each is a scoped effect pinned to this
// object, lasting until it leaves the battlefield (the ruling), and a later
// one overwrites an earlier type or base P/T set (layers 4 and 7b, by
// timestamp). The type is a SET (CR 205.1a): "Scout" is gone once it is a
// Soldier. The "if this creature is a …" is read as the ability resolves, so
// each can be activated whatever it is now (the ruling).
func init() {
	Register(Spec{
		OracleID:     "5d4ea910-3b16-4f98-89be-c1b2cfff4bf0",
		Name:         "Figure of Fable",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			figureOfFableStage("{G/W}: This creature becomes a Kithkin Scout with base power and toughness 2/3.",
				"{G/W}", "", []string{"Kithkin", "Scout"}, 2, 3, ""),
			figureOfFableStage("{1}{G/W}{G/W}: If this creature is a Scout, it becomes a Kithkin Soldier with base power and toughness 4/5.",
				"{1}{G/W}{G/W}", "Scout", []string{"Kithkin", "Soldier"}, 4, 5, ""),
			figureOfFableStage("{3}{G/W}{G/W}{G/W}: If this creature is a Soldier, it becomes a Kithkin Avatar with base power and toughness 7/8 and protection from each of your opponents.",
				"{3}{G/W}{G/W}{G/W}", "Soldier", []string{"Kithkin", "Avatar"}, 7, 8, game.ProtectionFromEachOfYourOpponents),
		},
	})
}

// figureOfFableStage is one of the three abilities. needs is the creature
// type the source must be as it resolves ("" for none); protection is the
// keyword the last stage adds.
func figureOfFableStage(label, cost, needs string, types []string, power, toughness int, protection string) ActivatedAbility {
	return ActivatedAbility{
		Label:   label,
		Purpose: game.Purpose{Answers: game.AnswerPump},
		Cost:    ManaCost(cost),
		Effect: func(g *game.Game, item *game.StackItem) error {
			if !sourceIsStillThisPermanent(g, item) {
				return nil
			}
			ctx := NewContext(g, item)
			if needs != "" {
				src, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok || !src.HasSubtype(ctx.TextSubtype(needs)) {
					return nil
				}
			}
			mods := append([]game.Mod{game.SetCreatureTypesMod(ctx.TextSubtypes(types...)...)},
				game.SetBasePTMods(power, toughness)...)
			if protection != "" {
				mods = append(mods, game.AddKeywordsMod(protection))
			}
			return ScopedEffectFor{
				Target:   item.SourceCardID,
				Mods:     mods,
				Duration: g.PinnedTo(game.IndefiniteDuration(), item.SourceCardID),
				Label:    "Figure of Fable — becomes a " + types[0] + " " + types[1],
			}.Apply(ctx)
		},
	}
}
