package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Throne of the Grim Captain // The Grim Captain — a transforming
// legendary artifact (#2709, ADR 0137 and its 2026-10-10 amendment):
//
//	Throne of the Grim Captain — Legendary Artifact {2}
//	  "{T}: Mill two cards.
//	   Craft with a Dinosaur, a Merfolk, a Pirate, and a Vampire {4}"
//	The Grim Captain — Legendary Creature — Skeleton Spirit Pirate, 7/7
//	  "Menace, trample, lifelink, hexproof
//	   Whenever The Grim Captain attacks, each opponent sacrifices a
//	   nonland permanent of their choice. Then you may put an exiled
//	   creature card used to craft The Grim Captain onto the battlefield
//	   under your control tapped and attacking."
//
// The craft is a set rule: four materials, one for each subtype, none
// filling two (CraftWithEachOf — a changeling is one of them, never
// all four, per the ruling). The Captain's attack trigger is the edict
// with its "then" as the continuation, which runs once every opponent
// has chosen. The put is optional: a creature card among the materials
// still in exile (CR 702.167c), then what it attacks — any player,
// planeswalker or battle its controller could attack, not necessarily
// the Captain's (the ruling) — and it enters through the CR 506.3c door
// tapped and attacking. It was never declared as an attacker, so its
// own "whenever this attacks" does not trigger.
//
// No simplification.
const throneOfTheGrimCaptainOracleID = "a7c96442-706b-4fd8-9b38-052872dafe58"

func init() {
	Register(Spec{
		OracleID:     throneOfTheGrimCaptainOracleID,
		Name:         "Throne of the Grim Captain",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{T}: Mill two cards.",
				Cost:    TapCost(),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return MillCards{N: 2}.Apply(NewContext(g, item))
				},
			},
			Craft("Craft with a Dinosaur, a Merfolk, a Pirate, and a Vampire {4}", "{4}",
				CraftWithEachOf("Dinosaur", "Merfolk", "Pirate", "Vampire")),
		},
	})

	Register(Spec{
		OracleID:        throneOfTheGrimCaptainOracleID + "#1",
		Name:            "The Grim Captain",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace", "trample", "lifelink", "hexproof"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("The Grim Captain — each opponent sacrifices a nonland permanent", grimCaptainAttacks),
		},
	})
}

// grimCaptainAttacks is the attack trigger: the edict, then the
// optional put.
func grimCaptainAttacks(g *game.Game, item *game.StackItem) error {
	return EachPlayerSacrifices{
		ExceptController: true,
		Match:            Nonland(),
		Label:            "a nonland permanent",
		Then: func(ctx *Context, _ game.PromptedSacrifices) error {
			return grimCaptainMayPut(ctx)
		},
	}.Apply(NewContext(g, item))
}

// grimCaptainMayPut offers a creature card among the Captain's
// materials still in exile.
func grimCaptainMayPut(ctx *Context) error {
	var creatures []uuid.UUID
	for _, c := range CraftMaterials(ctx) {
		if c.IsCreature() {
			creatures = append(creatures, c.InstanceID)
		}
	}
	if len(creatures) == 0 {
		return nil
	}
	item := ctx.Item
	me := ctx.Controller()
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  me,
		Source:   ctx.Source(),
		Question: "The Grim Captain — you may put an exiled creature card used to craft it onto the battlefield tapped and attacking",
		Cards:    creatures,
		Min:      0,
		Max:      1,
		Zone:     game.ZoneExile,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) == 0 {
				return nil
			}
			return putAttackingYourChoice(NewContext(g, item), picked[0], "The Grim Captain")
		},
	})
	return nil
}

// putAttackingYourChoice asks what `card` (in exile) attacks — any
// player, planeswalker or battle the controller could attack — and puts
// it onto the battlefield under the controller's control tapped and
// attacking that (CR 506.3c). With one choice there is no question.
func putAttackingYourChoice(ctx *Context, card uuid.UUID, name string) error {
	me := ctx.Controller()
	refs := ctx.Game.AttackTargetsForEffect(me)
	if len(refs) == 0 {
		return nil
	}
	put := func(g *game.Game, target uuid.UUID) error {
		if z := g.FindCardZoneForEffect(card); z == nil || z.Kind != game.ZoneExile {
			return nil
		}
		return g.PutOntoBattlefieldTogetherThenForEffect(
			[]game.BatchEntry{{CardID: card, From: game.ZoneExile}},
			game.ZoneEntryOptions{Controller: me, Tapped: true, Attacking: target}, nil)
	}
	if len(refs) == 1 {
		return put(ctx.Game, refs[0].ID)
	}
	options := make([]game.ChoiceOption, len(refs))
	for i, r := range refs {
		options[i] = game.ChoiceOption{Label: "Attack " + attackTargetName(ctx.Game, r.ID)}
		if r.Kind != game.AttackTargetPlayer {
			options[i].Cards = []uuid.UUID{r.ID}
		}
	}
	return PickOption{
		Question: name + " — what does it attack?",
		Options:  options,
		Then: func(ctx *Context, index int) error {
			if index < 0 || index >= len(refs) {
				index = 0
			}
			return put(ctx.Game, refs[index].ID)
		},
	}.Apply(ctx)
}
