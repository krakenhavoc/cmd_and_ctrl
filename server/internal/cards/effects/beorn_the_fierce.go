package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Beorn the Fierce — Legendary Creature — Bear Shapeshifter Warrior
// {3}{G}{G}, 6/6 (The Hobbit):
//
//	"Trample
//	 Other Bears you control get +2/+2.
//	 At the beginning of combat on your turn, put a trample counter on up
//	 to one target creature you control. It becomes a Bear in addition to
//	 its other types. Then if you control three or more Bears, draw two
//	 cards."
//
// The anthem is Canopy Tactician's lord shape over Bears. The trigger
// targets "up to one" creature, so it may be declined; with a target, the
// counter is a trample counter (a keyword counter, CR 122.1b) and "becomes
// a Bear" is a layer 4 addition with no stated duration (CR 611.2a), pinned
// to that object so it ends when the creature leaves (Ascent of the
// Worthy's shape). The Bear count is read last, after the creature is one,
// so the creature just made a Bear counts. With no target chosen the
// ability still counts the Bears; a chosen target that is gone makes the
// whole ability do nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "30ca8e92-5955-46a1-86c1-094a873f518f",
		Name:            "Beorn the Fierce",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Bear"}, Others: true, YoursOnly: true}, 2, 2),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtBeginningOfYourCombat("Beorn the Fierce — put a trample counter on up to one target creature you control; it becomes a Bear; with three or more Bears, draw two cards",
					beornTheFierceCombat),
				TargetCreature("up to one target creature you control", YouControl()).WithCount(0, 1)),
		},
	})
}

// beornTheFierceCombat is the beginning-of-combat body.
func beornTheFierceCombat(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	// The creatures just made Bears: the layers have not been recomputed
	// yet, so they are counted here rather than read back.
	madeBears := map[uuid.UUID]bool{}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		madeBears[t.ID] = true
		if err := (AddCounter{Target: t.ID, Kind: game.CounterTrample, N: 1}).Apply(ctx); err != nil {
			return err
		}
		if err := (ScopedEffectFor{
			Target:   t.ID,
			Mods:     []game.Mod{game.AddSubtypesMod(ctx.TextSubtypes("Bear")...)},
			Duration: g.PinnedTo(game.IndefiniteDuration(), t.ID),
			Label:    "Beorn the Fierce — it becomes a Bear in addition to its other types",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	if beornBearsControlled(ctx, madeBears) < 3 {
		return nil
	}
	return DrawCards{N: 2}.Apply(ctx)
}

// beornBearsControlled is the number of Bears the ability's controller
// controls, counting `madeBears` as Bears too.
func beornBearsControlled(ctx *Context, madeBears map[uuid.UUID]bool) int {
	// "Bear" is a word of this card's text (CR 612.2).
	bear := ctx.TextSubtype("Bear")
	return len(permanentsControlledByMatching(ctx.Game, ctx.Controller(),
		func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
			return madeBears[c.InstanceID] || c.HasSubtype(bear)
		}))
}
