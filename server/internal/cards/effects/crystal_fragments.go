package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crystal Fragments // Summon: Alexander — a transforming Equipment
// whose back face is a Saga creature.
//
// Front face, Artifact — Equipment {W}:
//
//	"Equipped creature gets +1/+1.
//	 {5}{W}{W}: Exile this Equipment, then return it to the battlefield
//	 transformed under its owner's control. Activate only as a sorcery.
//	 Equip {1}"
//
// Back face, Enchantment Creature — Saga Construct, 4/3:
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I, II — Prevent all damage that would be dealt to creatures you
//	 control this turn.
//	 III — Tap all creatures your opponents control.
//	 Flying"
//
// The flip is ADR 0079's second verb (Sorin of House Markov's): two
// zone changes and a new object (CR 400.7), so the Equipment falls off
// and the Saga enters with its first lore counter and chapter I. "Under
// its owner's control" leaves the Controller zero. The back face
// registers under "<oracle_id>#1" (game.CatalogKey).
//
// Chapters I and II are #2045's recipient set: creatures you control as
// the damage would be dealt (CR 611.2c), and not you. Chapter III taps
// every creature an opponent controls as it resolves.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     crystalFragmentsOracleID,
		Name:         "Crystal Fragments",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(1, 1)},
		Activated: []ActivatedAbility{
			{
				Label:        "{5}{W}{W}: Exile this Equipment, then return it to the battlefield transformed under its owner's control. Activate only as a sorcery.",
				Cost:         ManaCost("{5}{W}{W}"),
				SorcerySpeed: true,
				Effect: func(g *game.Game, item *game.StackItem) error {
					return ExileAndReturnTransformed{Target: item.SourceCardID}.Apply(NewContext(g, item))
				},
			},
			EquipAbility("{1}"),
		},
	})

	creaturesYouControl := func(g *game.Game, item *game.StackItem) error {
		return PreventDamageFromSource{Protect: ShieldCreaturesYouControl}.Apply(NewContext(g, item))
	}
	Register(Spec{
		OracleID:        crystalFragmentsOracleID + "#1",
		Name:            "Summon: Alexander",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel("Summon: Alexander", 1, "prevent all damage that would be dealt to creatures you control this turn"), creaturesYouControl),
			ChapterTrigger(2, SagaChapterLabel("Summon: Alexander", 2, "prevent all damage that would be dealt to creatures you control this turn"), creaturesYouControl),
			ChapterTrigger(3, SagaChapterLabel("Summon: Alexander", 3, "tap all creatures your opponents control"), tapAllCreaturesYourOpponentsControl),
		},
	})
}

// crystalFragmentsOracleID is shared by both faces.
const crystalFragmentsOracleID = "b291d046-7649-4a05-98c4-224ecaece912"

// tapAllCreaturesYourOpponentsControl is "Tap all creatures your
// opponents control", read as it resolves.
func tapAllCreaturesYourOpponentsControl(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, opp := range ctx.Opponents() {
		if opp == uuid.Nil {
			continue
		}
		if err := b29TapAllCreaturesControlledBy(ctx, opp); err != nil {
			return err
		}
	}
	return nil
}
