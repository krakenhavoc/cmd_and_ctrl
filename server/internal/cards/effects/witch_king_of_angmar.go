package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Witch-king of Angmar — Legendary Creature — Wraith Noble
// {3}{B}{B}, 5/3:
//
//	"Flying
//	 Whenever one or more creatures deal combat damage to you, each
//	 opponent sacrifices a creature of their choice that dealt combat
//	 damage to you this turn. The Ring tempts you.
//	 Discard a card: Witch-king of Angmar gains indestructible until
//	 end of turn. Tap him."
//
// The trigger is "one or more … to you" (CR 603.2c): OncePerBatch, so
// every creature that connects in one combat damage step is one
// trigger, and a first-strike step and a regular step are two. The
// damage is combat damage to Witch-king's controller by any creature.
//
// The edict's candidate list is the per-turn record of WHICH creature
// objects dealt combat damage to its controller (#2149,
// TurnTally.DamageDealers). It is read as the edict resolves, so a
// creature that dealt damage in an earlier step of the same turn
// counts (it "dealt combat damage to you this turn"), and one that
// has left and returned since is a new object that did not (CR 400.7).
// An opponent with no such creature on the battlefield sacrifices
// nothing, and the Ring still tempts you: the tempt is the sentence
// after the edict, not a part of it. It runs once every opponent has
// answered (EachPlayerSacrifices.Then), so a sacrificed creature is
// gone before the Ring-bearer is chosen.
//
// "Tap him" is a second sentence of the activated ability, not a
// cost: the discard is the whole cost, so he gains indestructible
// even when already tapped.
func init() {
	Register(Spec{
		OracleID:        "2eb1b429-8d11-42f3-8815-5a60b6a4e2d5",
		Name:            "Witch-king of Angmar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventDealDamage, creatureDealtCombatDamageToYou,
				"Witch-king of Angmar — each opponent sacrifices a creature that dealt combat damage to you this turn; the Ring tempts you",
				witchKingEdict)),
		},
		Activated: []ActivatedAbility{{
			Label:   "Discard a card: Witch-king of Angmar gains indestructible until end of turn. Tap him",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    DiscardACard(),
			Effect:  witchKingShield,
		}},
	})
}

// creatureDealtCombatDamageToYou is "a creature deals combat damage to
// you": the event is combat damage to the source's controller from a
// creature, whoever controls it.
func creatureDealtCombatDamageToYou(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || ev.Target != source.Controller {
		return false
	}
	dealer, ok := g.LookupCardForEffect(ev.Source)
	return ok && dealer.IsCreature()
}

func witchKingEdict(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	you := item.Controller
	return EachPlayerSacrifices{
		ExceptController: true,
		Match:            And(Creature(), dealtCombatDamageToPlayerThisTurn(you)),
		Label:            "a creature that dealt combat damage to you this turn",
		Then: func(ctx *Context, _ game.PromptedSacrifices) error {
			return TheRingTemptsYou{Player: you}.Apply(ctx)
		},
	}.Apply(ctx)
}

func witchKingShield(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (GrantKeywordUntilEOT{
		Target:   item.SourceCardID,
		Keywords: []string{"indestructible"},
		Label:    "Witch-king of Angmar — indestructible",
	}).Apply(ctx); err != nil {
		return err
	}
	return TapTarget{Target: item.SourceCardID}.Apply(ctx)
}
