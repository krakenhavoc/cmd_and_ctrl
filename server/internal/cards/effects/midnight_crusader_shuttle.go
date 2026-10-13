package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Midnight Crusader Shuttle — Artifact — Vehicle {4}, 3/4:
//
//	"Midnight Entity — Whenever this Vehicle attacks, defending player
//	 faces a villainous choice — That player sacrifices a creature of
//	 their choice, or you gain control of a creature of your choice that
//	 player controls until end of turn. If you gain control of a creature
//	 this way, tap it, and it's attacking that player.
//	 Crew 2"
//
// A villainous choice (CR 701.55, #3143) faced by the defending player: the
// player the Shuttle is attacking, or the controller of the planeswalker
// or battle it is attacking.
//
// Option A is the shared sacrifice body, so a player with no creature may
// pick it and lose nothing (CR 701.55b).
//
// Option B is the controller's choice among the facing player's
// creatures, asked of the controller as a pick over that board. It is not
// targeting, so hexproof and shroud do not stop it. They gain control
// until end of turn (a pinned scoped effect, so a flickered creature is
// not the one taken), and only when control really changed (the layer
// pass runs first and the new controller is checked) they tap it and put
// it into combat attacking the facing player. That last step is
// CR 506.3c's "put onto the battlefield attacking" door
// (Game.PutOnBattlefieldAttackingForEffect): the creature was never
// DECLARED as an attacker, so nothing triggers "whenever ~ attacks", and
// the summoning-sickness and tapped checks of an attack declaration are
// not made. With no creature to take, option B does nothing.
//
// Crew 2 is the ordinary CR 702.122 ability.
//
// No simplifications.
var shuttleChoice = NewVillainousChoice("midnight-crusader-shuttle",
	"Midnight Crusader Shuttle — sacrifice a creature, or the controller takes one of yours?",
	VillainousOption{
		Label:    "Sacrifice a creature of your choice",
		Harm:     2,
		Possible: villainousCanSacrifice(Creature()),
		Body:     villainousSacrifice("a creature", Creature()),
	},
	VillainousOption{
		Label: "The controller of Midnight Crusader Shuttle gains control of a creature of their choice that you control until end of turn, taps it, and it's attacking you",
		Harm:  2,
		Body: func(ctx *Context, f Facing) error {
			return ChoosePermanents{
				Player:   ctx.Controller(),
				Of:       []uuid.UUID{f.Player},
				Question: "Midnight Crusader Shuttle — choose a creature that player controls to gain control of",
				Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
					var ids []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == of && c.IsCreature() {
							ids = append(ids, c.InstanceID)
						}
					}
					return ids, 1, 1
				},
				Then: func(ctx *Context, picked game.PromptedPicks) error {
					for _, id := range picked.Cards() {
						return shuttleTakeAndAttack(ctx, id, f.Player)
					}
					return nil
				},
			}.Apply(ctx)
		},
	})

// shuttleTakeAndAttack gains control of `id` until end of turn and, if
// that worked, taps it and puts it into combat attacking `defender`.
func shuttleTakeAndAttack(ctx *Context, id, defender uuid.UUID) error {
	me := ctx.Controller()
	if err := (GainControl{
		Target:     id,
		Controller: me,
		Duration:   DurationUntilEndOfTurn(ctx),
		Label:      "Midnight Crusader Shuttle — gain control",
	}).Apply(ctx); err != nil {
		return err
	}
	ctx.Game.RecomputeLayersIfStaleLocked()
	if c, ok := ctx.Game.LookupCardForEffect(id); !ok || c.Controller != me {
		return nil
	}
	if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	ctx.Game.PutOnBattlefieldAttackingForEffect(id, defender)
	return nil
}

func init() {
	Register(Spec{
		OracleID:     "22fa2860-d746-4429-b7e3-3403bef27f7b",
		Name:         "Midnight Crusader Shuttle",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Midnight Entity — the defending player faces a villainous choice",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					defender := g.DefendingPlayerForAttackForEffect(item.Trigger.Event.Target)
					if defender == uuid.Nil {
						return nil
					}
					return FaceVillainousChoice{
						Players: []uuid.UUID{defender},
						Choice:  shuttleChoice,
					}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Midnight Crusader Shuttle"),
		}},
	})
}
