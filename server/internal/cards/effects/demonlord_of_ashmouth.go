package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Demonlord of Ashmouth — Creature — Demon {2}{B}{B}, 5/4:
//
//	"Flying
//	 When this creature enters, exile it unless you sacrifice another
//	 creature.
//	 Undying"
//
// The trigger fires on every entry, the undying return included
// (ruling). You may always choose the exile, even with another creature
// to sacrifice (ruling); with none the exile happens without a
// question. A Demonlord that has already left, or come back as a new
// object, is not exiled (CR 400.7). Flying and undying are
// PrintedKeywords (#2075).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "049777d0-910d-46d2-9ca1-10277b2fb845",
		Name:            "Demonlord of Ashmouth",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordUndying},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Demonlord of Ashmouth — exile it unless you sacrifice another creature", demonlordExileUnlessSacrifice),
		},
	})
}

func demonlordExileUnlessSacrifice(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
		return nil
	}
	var others []uuid.UUID
	for _, id := range permanentsControlledByMatching(g, item.Controller, Creature()) {
		if id != item.SourceCardID {
			others = append(others, id)
		}
	}
	if len(others) == 0 {
		return demonlordExileItself(ctx)
	}
	return PickOption{
		Question: "Demonlord of Ashmouth — sacrifice another creature, or exile Demonlord of Ashmouth",
		Options:  []game.ChoiceOption{{Label: "Exile Demonlord of Ashmouth"}, {Label: "Sacrifice another creature"}},
		Then: func(ctx *Context, index int) error {
			if index == 1 {
				return SacrificeChoice{Player: item.Controller, Candidates: others,
					Question: "Demonlord of Ashmouth — sacrifice another creature"}.Apply(ctx)
			}
			return demonlordExileItself(ctx)
		},
	}.Apply(ctx)
}

// demonlordExileItself exiles the trigger's source, the Demonlord.
func demonlordExileItself(ctx *Context) error {
	return ExileTarget{Target: ctx.Item.SourceCardID}.Apply(ctx)
}
