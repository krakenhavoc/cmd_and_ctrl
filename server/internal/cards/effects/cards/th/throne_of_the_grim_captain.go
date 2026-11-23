package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
const ThroneOfTheGrimCaptainOracleID = "a7c96442-706b-4fd8-9b38-052872dafe58"

func init() {
	effects.Register(effects.Spec{
		OracleID:     ThroneOfTheGrimCaptainOracleID,
		Name:         "Throne of the Grim Captain",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Mill two cards.",
				Cost:    effects.TapCost(),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.MillCards{N: 2}.Apply(effects.NewContext(g, item))
				},
			},
			effects.Craft("Craft with a Dinosaur, a Merfolk, a Pirate, and a Vampire {4}", "{4}",
				effects.CraftWithEachOf("Dinosaur", "Merfolk", "Pirate", "Vampire")),
		},
	})

	effects.Register(effects.Spec{
		OracleID:        ThroneOfTheGrimCaptainOracleID + "#1",
		Name:            "The Grim Captain",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace", "trample", "lifelink", "hexproof"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerEdict(effects.WheneverThisAttacks("The Grim Captain — each opponent sacrifices a nonland permanent", grimCaptainAttacks)),
		},
	})
}

// grimCaptainAttacks is the attack trigger: the edict, then the
// optional put.
func grimCaptainAttacks(g *game.Game, item *game.StackItem) error {
	return effects.EachPlayerSacrifices{
		ExceptController: true,
		Match:            effects.Nonland(),
		Label:            "a nonland permanent",
		Keyed:            grimCaptainEdicted,
	}.Apply(effects.NewContext(g, item))
}

// grimCaptainEdicted is the optional put once the edict has landed, keyed
// (#4290).
var grimCaptainEdicted = effects.RegisterEdictThen("prompt-run/throne-of-the-grim-captain-put", func(ctx *effects.Context, _ game.PromptedSacrifices) error {
	return grimCaptainMayPut(ctx)
})

// grimCaptainMayPut offers a creature card among the Captain's
// materials still in exile.
func grimCaptainMayPut(ctx *effects.Context) error {
	var creatures []uuid.UUID
	for _, c := range effects.CraftMaterials(ctx) {
		if c.IsCreature() {
			creatures = append(creatures, c.InstanceID)
		}
	}
	if len(creatures) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "The Grim Captain — you may put an exiled creature card used to craft it onto the battlefield tapped and attacking",
		Cards:    creatures,
		Min:      0,
		Max:      1,
		ThenKey:  GrimCaptainPut,
		LiveZone: game.ZoneExile,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
}

// grimCaptainPut is the optional put's answer, keyed (#4253 slice 8): the
// creature card goes onto the battlefield tapped and attacking what its
// controller picks.
var GrimCaptainPut = game.RegisterChooseCardsThen("choose-cards/the-grim-captain-put",
	func(g *game.Game, r game.CardsChosen) error {
		if !r.Answered || len(r.Picked) == 0 {
			return nil
		}
		ctx, _ := effects.AskedContext(g, r.Asked, r.Chooser, r.Source, nil)
		return effects.PutAttackingYourChoice(ctx, r.Picked[0], game.ZoneExile, "The Grim Captain")
	})
