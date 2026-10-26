package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Acererak the Archlich — Legendary Creature — Zombie Wizard {2}{B}, 5/5:
//
//	"When Acererak enters, if you haven't completed Tomb of
//	 Annihilation, return Acererak to its owner's hand and venture into
//	 the dungeon.
//	 Whenever Acererak attacks, for each opponent, you create a 2/2 black
//	 Zombie creature token unless that player sacrifices a creature of
//	 their choice."
//
// The enter trigger is an intervening "if" (CR 603.4) on
// CompletedDungeon(…, Tomb of Annihilation). The venture happens even
// when Acererak has already left the battlefield: "return Acererak … and
// venture" are two instructions, and the first doing nothing does not
// stop the second.
//
// The attack trigger asks each opponent in turn order (CR 101.4), each
// knowing the answers before theirs, through a keyed option pick (a
// restore point). An opponent with no creature to sacrifice is not
// asked: the Zombie is created (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "31b4e85d-89fc-4712-8f05-7ec097b1d3b6",
		Name:         "Acererak the Archlich",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			IfVentureCond(condYouHaventCompletedTomb, WhenThisEnters(
				"Acererak the Archlich — return it to its owner's hand and venture into the dungeon", acererakReturn)),
			WheneverThisAttacks("Acererak the Archlich — for each opponent, create a 2/2 black Zombie unless that player sacrifices a creature",
				acererakAttack),
		},
	})
	acererakThen = OptionPickThen("option-pick/acererak-the-archlich-zombie-or-sacrifice", acererakAnswered)
	acererakSacrificed = RunThen("prompt-run/acererak-the-archlich-sacrificed", func(ctx *Context, r game.RunDone) error {
		return acererakStep(ctx, r.Carry)
	})
}

// acererakSacrificed is the next opponent's question once an opponent's
// sacrifice has landed, keyed (#4290): Carry is the opponents still to be
// asked. Registered in init because it asks the next opponent, which names
// the key again.
var acererakSacrificed game.PromptRunThen

// acererakThen is the attack trigger's keyed answer: Carry is the
// opponents still to be asked, in turn order.
var acererakThen game.OptionPickThen

// acererakReturn returns Acererak to its owner's hand and ventures.
func acererakReturn(g *game.Game, item *game.StackItem) error {
	return BounceToHand{
		Target: item.SourceCardID,
		Then: func(ctx *Context, _ bool) error {
			return VentureIntoTheDungeon{}.Apply(ctx)
		},
	}.Apply(NewContext(g, item))
}

// acererakAttack asks each opponent in turn order.
func acererakAttack(g *game.Game, item *game.StackItem) error {
	var opps []uuid.UUID
	for _, id := range apnapPlayers(g) {
		if id != item.Controller {
			opps = append(opps, id)
		}
	}
	return acererakStep(NewContext(g, item), opps)
}

// acererakSacrificeSpec is "a creature of their choice".
func acererakSacrificeSpec() *game.TargetSpec { return sacrificeSpec("a creature", Creature()) }

// acererakStep asks the first opponent of `remaining` and hands the
// rest on.
func acererakStep(ctx *Context, remaining []uuid.UUID) error {
	for len(remaining) > 0 {
		opp, rest := remaining[0], remaining[1:]
		p := ctx.Game.PlayerByIDForEffect(opp)
		if p == nil || p.Eliminated {
			remaining = rest
			continue
		}
		if len(ctx.Game.EffectSacrificeCandidatesForEffect(opp, acererakSacrificeSpec())) == 0 {
			if err := acererakZombie(ctx); err != nil {
				return err
			}
			remaining = rest
			continue
		}
		return PickOption{
			Player:   opp,
			Question: "Acererak the Archlich — sacrifice a creature, or its controller creates a 2/2 black Zombie?",
			Options: []game.ChoiceOption{
				{Label: "Let them create a 2/2 black Zombie"},
				{Label: "Sacrifice a creature"},
			},
			ThenKey: acererakThen,
			Carry:   rest,
		}.Apply(ctx)
	}
	return nil
}

// acererakAnswered applies one opponent's answer and asks the next. A
// board that emptied between the question and the answer creates the
// Zombie.
func acererakAnswered(ctx *Context, r game.OptionPicked) error {
	opp, rest := r.Chooser, append([]uuid.UUID(nil), r.Carry...)
	if r.Index == 1 && len(ctx.Game.EffectSacrificeCandidatesForEffect(opp, acererakSacrificeSpec())) > 0 {
		return ctx.SacrificeThen(opp, acererakSacrificeSpec(),
			"Acererak the Archlich — sacrifice a creature", 1, acererakSacrificed, rest...)
	}
	if r.Index >= 0 {
		if err := acererakZombie(ctx); err != nil {
			return err
		}
	}
	return acererakStep(ctx, rest)
}

// acererakZombie is "you create a 2/2 black Zombie creature token".
func acererakZombie(ctx *Context) error {
	return CreateToken{Controller: ctx.Controller(), Template: TokenCard("2/2 black Zombie"), N: 1}.Apply(ctx)
}
