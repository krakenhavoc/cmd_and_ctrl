package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sheoldred // The True Scriptures — a transforming Praetor whose back face
// is a Saga.
//
// Front face, Legendary Creature — Phyrexian Praetor {3}{B}{B}, 4/5:
//
//	"Menace
//	 When Sheoldred enters, each opponent sacrifices a nontoken creature
//	 or planeswalker of their choice.
//	 {4}{B}: Exile Sheoldred, then return it to the battlefield
//	 transformed under its owner's control. Activate only as a sorcery
//	 and only if an opponent has eight or more cards in their graveyard."
//
// Back face, Enchantment — Saga:
//
//	"(As this Saga enters and after your draw step, add a lore counter.)
//	 I — For each opponent, destroy up to one target creature or
//	     planeswalker that player controls.
//	 II — Each opponent discards three cards, then mills three cards.
//	 III — Put all creature cards from all graveyards onto the battlefield
//	       under your control. Exile this Saga, then return it to the
//	       battlefield (front face up)."
//
// The ETB is an edict: each opponent chooses their own nontoken creature or
// planeswalker through the sacrifice prompt, nothing targets, and an
// opponent with none is skipped. The flip is ExileAndReturnTransformed, as
// Crystal Fragments; the back face registers under "<oracle_id>#1". The
// activation condition reads every opponent's graveyard live, so one
// opponent at eight is enough.
//
// Chapter I has one "up to one" clause per opponent, bound to that
// opponent's permanents (Molten Primordial's TargetsFrom), so the pick for
// one opponent can't be a permanent of another; a clause with nothing legal
// offers nothing. Chapter II fans the discard out to every opponent and
// mills three for each of them after all the discards are answered, which
// is the printed order. Chapter III returns every creature card in every
// graveyard together, under the Saga's controller, then flickers the Saga
// (front face up, a new object).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        sheoldredTrueScripturesOracleID,
		Name:            "Sheoldred",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sheoldred — each opponent sacrifices a nontoken creature or planeswalker", Do(EachPlayerSacrifices{
				ExceptController: true,
				Match:            And(Or(Creature(), Planeswalker()), Not(IsTokenPredicate())),
				Label:            "a nontoken creature or planeswalker",
			})),
		},
		Activated: []ActivatedAbility{{
			Label: "{4}{B}: Exile Sheoldred, then return it to the battlefield transformed under its owner's control. " +
				"Activate only as a sorcery and only if an opponent has eight or more cards in their graveyard.",
			Cost:         ManaCost("{4}{B}"),
			SorcerySpeed: true,
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				for _, p := range g.Seats {
					if p != nil && p.ID != controller && !p.Eliminated && p.Graveyard != nil && p.Graveyard.Size() >= 8 {
						return true
					}
				}
				return false
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return ExileAndReturnTransformed{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})

	chapterOne := ChapterTrigger(1, SagaChapterLabel("The True Scriptures", 1, "for each opponent, destroy up to one target creature or planeswalker that player controls"),
		func(g *game.Game, item *game.StackItem) error {
			return destroyEachLegalTarget(item, NewContext(g, item))
		})
	chapterOne.TargetsFrom = trueScripturesChapterOneClauses
	Register(Spec{
		OracleID:     sheoldredTrueScripturesOracleID + "#1",
		Name:         "The True Scriptures",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			chapterOne,
			ChapterTrigger(2, SagaChapterLabel("The True Scriptures", 2, "each opponent discards three cards, then mills three cards"),
				trueScripturesChapterTwo),
			ChapterTrigger(3, SagaChapterLabel("The True Scriptures", 3, "put all creature cards from all graveyards onto the battlefield under your control, then exile this Saga and return it front face up"),
				trueScripturesChapterThree),
		},
	})
}

const sheoldredTrueScripturesOracleID = "97652492-7906-4d79-983c-fa1dc1239eba"

// trueScripturesChapterOneClauses is one "up to one target creature or
// planeswalker <player> controls" clause per opponent of the Saga's
// controller, in turn order. It reads nothing the game can change but the
// seat list, so the row restores.
func trueScripturesChapterOneClauses(tc game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
	controller := source.Controller
	if obj := tc.Object; obj != nil && obj.ID == source.InstanceID && obj.Controller != uuid.Nil {
		controller = obj.Controller
	}
	var clauses []*game.TargetSpec
	for _, p := range g.Seats {
		if p == nil || p.ID == controller {
			continue
		}
		clauses = append(clauses, TargetPermanent("up to one target creature or planeswalker "+p.Name+" controls",
			And(Or(Creature(), Planeswalker()), ControlledBy(p.ID))).WithCount(0, 1))
	}
	if len(clauses) == 0 {
		return nil
	}
	return Clauses(clauses[0], clauses[1:]...)
}

func trueScripturesChapterTwo(g *game.Game, item *game.StackItem) error {
	return g.EachPlayerDiscardsThenForEffect(item.Controller,
		game.DiscardPrompt{Source: item.SourceCardID, N: 3},
		func(g *game.Game, _ game.PromptedDiscards) error {
			c := NewContext(g, item)
			for _, opp := range c.Opponents() {
				if err := (MillCards{Player: opp, N: 3}).Apply(c); err != nil {
					return err
				}
			}
			return nil
		})
}

func trueScripturesChapterThree(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ids := allGraveyardsCardIDs(ctx, func(c game.Card) bool { return c.IsCreature() })
	return ReturnFromGraveyardTogether{
		Targets:    ids,
		Controller: item.Controller,
		Then: func(ctx *Context, _ []uuid.UUID) error {
			return Flicker{Target: item.SourceCardID, Controller: item.Controller}.Apply(ctx)
		},
	}.Apply(ctx)
}
