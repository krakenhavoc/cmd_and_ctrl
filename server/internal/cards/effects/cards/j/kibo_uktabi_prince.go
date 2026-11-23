package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kibo, Uktabi Prince — Legendary Creature — Monkey Noble {2}{G}, 2/2
// (Jumpstart 2022, #5948):
//
//	"{T}: Each player creates a colorless artifact token named Banana with
//	 "{T}, Sacrifice this token: Add {R} or {G}. You gain 2 life."
//	 Whenever an artifact an opponent controls is put into a graveyard from
//	 the battlefield, put a +1/+1 counter on each creature you control
//	 that's an Ape or a Monkey.
//	 Whenever Kibo attacks, defending player sacrifices an artifact of their
//	 choice."
//
// The Banana is a token with its own mana ability (the life gain is part of
// it, CR 605.1a), made for every player still in the game. The second
// ability is Party Dude's watch on an artifact, token or card, that left an
// opponent's control for a graveyard, read as it last existed there (CR
// 603.10a); the counters go on each Ape and Monkey as the ability resolves,
// Kibo included (he is a Monkey). "Defending player" is the player attacked
// (Barret Wallace's read): they choose which artifact to sacrifice as the
// ability resolves, and a player with none sacrifices nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0b9be4fa-5238-4afd-a9f6-f9022e67e5ab",
		Name:         "Kibo, Uktabi Prince",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Each player creates a colorless artifact token named Banana with \"{T}, Sacrifice this token: Add {R} or {G}. You gain 2 life.\"",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1},
			Cost:    effects.TapCost(),
			Effect:  kiboBananas,
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.S82MomB14OpponentArtifactPutIntoGraveyard,
				"Kibo, Uktabi Prince — put a +1/+1 counter on each Ape and Monkey you control", kiboCounters),
			effects.WheneverThisAttacks("Kibo, Uktabi Prince — defending player sacrifices an artifact of their choice", kiboSacrifice),
		},
	})
}

// kiboBananas gives each player still in the game a Banana, in turn order
// from the active player (CR 101.4).
func kiboBananas(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, id := range effects.TablePlayers(ctx) {
		if err := (effects.CreateToken{Controller: id, Template: effects.GapB154BananaToken(), N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// kiboCounters puts a +1/+1 counter on each creature its controller controls
// that is an Ape or a Monkey, as the ability resolves.
func kiboCounters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() && (c.HasSubtype(ctx.TextSubtype("Ape")) || c.HasSubtype(ctx.TextSubtype("Monkey"))) {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

// kiboSacrifice has the defending player sacrifice an artifact of their
// choice.
func kiboSacrifice(g *game.Game, item *game.StackItem) error {
	defender := effects.B17DefendingPlayer(g, item.Trigger.Event)
	if defender == uuid.Nil || g.PlayerByIDForEffect(defender) == nil {
		return nil
	}
	return effects.SacrificeChoice{
		Player:     defender,
		Candidates: effects.PermanentsControlledByMatching(g, defender, effects.Artifact()),
		Question:   "Kibo, Uktabi Prince — sacrifice an artifact",
	}.Apply(effects.NewContext(g, item))
}
