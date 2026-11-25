package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bellowing Mauler's punisher: "each player loses 4 life unless they
// sacrifice a nontoken creature of their choice." Every player in turn
// order starting with the controller (the active player at the beginning
// of the controller's end step, so APNAP, CR 101.4) is asked in turn, each
// question queued by the answer to the one before it (punisherRepeat's
// chain), so a player who sacrificed their last creature changes what the
// next one is asked.

const (
	bellowingMaulerName     = "Bellowing Mauler"
	bellowingMaulerLife     = 4
	bellowingMaulerLoseText = "Lose 4 life"
	bellowingMaulerSacText  = "Sacrifice a nontoken creature"
)

// bellowingMaulerPlayers lists the living players in seat order starting
// with the effect's controller.
func BellowingMaulerPlayers(ctx *Context) []uuid.UUID {
	me := ctx.Controller()
	n := len(ctx.Game.Seats)
	start := 0
	for i, p := range ctx.Game.Seats {
		if p != nil && p.ID == me {
			start = i
			break
		}
	}
	var out []uuid.UUID
	for i := 0; i < n; i++ {
		p := ctx.Game.Seats[ctx.Game.SeatInTurnOrderForEffect(start, i)]
		if p == nil || p.Eliminated {
			continue
		}
		out = append(out, p.ID)
	}
	return out
}

// bellowingMaulerCandidates is the nontoken creatures a player controls.
func bellowingMaulerCandidates(g *game.Game, player uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() && !c.IsToken() && !g.CantBeSacrificedForEffect(c.InstanceID) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// bellowingMaulerStep asks the first player still owed. A player with no
// nontoken creature just loses the life, with no question.
//
// Caller holds g.mu.
func BellowingMaulerStep(ctx *Context, remaining []uuid.UUID) error {
	for len(remaining) > 0 {
		victim, rest := remaining[0], remaining[1:]
		pl := ctx.Game.PlayerByIDForEffect(victim)
		if pl == nil || pl.Eliminated {
			remaining = rest
			continue
		}
		if len(bellowingMaulerCandidates(ctx.Game, victim)) == 0 {
			if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), victim, -bellowingMaulerLife); err != nil {
				return err
			}
			remaining = rest
			continue
		}
		return PickOption{
			Player:   victim,
			Question: bellowingMaulerName + " — lose 4 life, or sacrifice a nontoken creature?",
			Options: []game.ChoiceOption{
				{Label: bellowingMaulerLoseText, LifeCost: bellowingMaulerLife},
				{Label: bellowingMaulerSacText},
			},
			ThenKey: bellowingMaulerAnswered,
			Carry:   rest,
		}.Apply(ctx)
	}
	return nil
}

// bellowingMaulerSacrificed continues the chain once the sacrifice has
// settled. An on-disk key: never renamed, never reused.
var bellowingMaulerSacrificed game.TailRef

// bellowingMaulerAnswered is a player's answer. An on-disk key: never
// renamed, never reused.
var bellowingMaulerAnswered game.OptionPickThen

// The keyed continuations are registered in init, not at var
// initialisation: the step and the answers refer to each other.
func init() {
	bellowingMaulerSacrificed = CardTail("then/bellowing-mauler-sacrificed", func(ctx *Context, o game.TailOutcome) error {
		return BellowingMaulerStep(ctx, o.Tail.Carry)
	})
	bellowingMaulerAnswered = OptionPickThen("option-pick/bellowing-mauler-lose-or-sacrifice", bellowingMaulerAnswer)
}

func bellowingMaulerAnswer(ctx *Context, r game.OptionPicked) error {
	victim, rest := r.Chooser, r.Carry
	if r.Option == nil {
		return BellowingMaulerStep(ctx, rest)
	}
	if r.Option.Label == bellowingMaulerSacText {
		return SacrificeChoice{
			Player:     victim,
			Candidates: bellowingMaulerCandidates(ctx.Game, victim),
			Question:   bellowingMaulerName + " — sacrifice a nontoken creature",
			ThenTail:   WithCarry(ctx.Tail(bellowingMaulerSacrificed), rest...),
		}.Apply(ctx)
	}
	if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), victim, -bellowingMaulerLife); err != nil {
		return err
	}
	return BellowingMaulerStep(ctx, rest)
}

// sncB13YouSacrificedAnotherCreature is "you sacrifice another creature":
// the sacrifice event (it fires before the zone move, so the permanent is
// still a creature to read) was made by the source's controller and is not
// the source itself. Body Dropper and Forge Boss read it.
func SncB13YouSacrificedAnotherCreature(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Actor != source.Controller || ev.CardID == source.InstanceID {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature()
}
