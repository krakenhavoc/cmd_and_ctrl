package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gluntch, the Bestower — Legendary Creature — Jellyfish, {1}{G}{W},
// 3/5:
//
//	"Flying
//	 At the beginning of your end step, choose a player. They put two
//	 +1/+1 counters on a creature they control. Choose a second player
//	 to draw a card. Then choose a third player to create two Treasure
//	 tokens."
//
// The card #929 was written for: three "choose a player" clauses in
// one sentence, none of them a target, each of which has to name a
// DIFFERENT player from the ones already named. It is the whole
// primitive's contract in one trigger — the pool, the exclusion, the
// chain, and a sub-choice made by somebody who is not the controller.
//
// Three notes on what the words mean:
//
//   - "Choose a player" is every seated player, the controller
//     included (CR 102.1). Gluntch is a politics card; handing
//     yourself the counters is a legal and common line.
//   - "A second player" / "a third player" mean a player not already
//     chosen. That is ChoosePlayer.Except fed from
//     ctx.ChosenPlayers(), which is the list the payload keeps.
//   - "A creature THEY control" is the chosen player's own choice,
//     not the controller's — so the counters wait on a second prompt,
//     addressed to that seat, and the rest of the trigger waits on it
//     in turn. A chosen player with no creature is skipped and the
//     trigger carries on (CR 608.2, "as much as it can").
//
// At two or three seats the later clauses simply run out of eligible
// players and do nothing, which is the printed outcome rather than a
// simplification.
//
// Flying is printed card data and arrives from the deck import, not
// from this entry.
func init() {
	Register(Spec{
		OracleID:     "0222dc7c-459b-4909-a037-72b2eb248599",
		Name:         "Gluntch, the Bestower",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Gluntch, the Bestower — bestow on three players", gluntchBestow),
		},
	})
}

// gluntchBestow opens the chain: the first player, then their own
// pick of a creature.
func gluntchBestow(g *game.Game, item *game.StackItem) error {
	return ChoosePlayer{
		Among:    Players,
		Question: "Gluntch, the Bestower — choose a player to put two +1/+1 counters on a creature they control",
		Then:     gluntchCounters,
	}.Apply(NewContext(g, item))
}

// gluntchCounters asks the chosen player which of their creatures
// takes the counters, then moves on to the second clause.
//
// The counters are placed by the CHOSEN player's answer, so the second
// clause is queued from inside that answer rather than after this
// function returns — the Scry.Then rule: a prompt only queues.
func gluntchCounters(ctx *Context) error {
	chosen := ctx.ChosenPlayer()
	creatures := creaturesControlledByPlayer(ctx.Game, chosen)
	if chosen == uuid.Nil || len(creatures) == 0 {
		return gluntchSecondPlayer(ctx)
	}
	item := ctx.Item
	queued := ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    chosen,
		FromPlayer: chosen,
		Source:     ctx.Source(),
		Question:   "Gluntch, the Bestower — put two +1/+1 counters on a creature you control",
		Cards:      creatures,
		Min:        1,
		Max:        1,
		// Re-checked against the live battlefield on submit: the
		// prompt is asynchronous and a creature can leave between the
		// question and the answer.
		Zone: game.ZoneBattlefield,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			next := NewContext(g, item)
			for _, id := range picked {
				if err := (AddCounter{Target: id, Kind: "+1/+1", N: 2}).Apply(next.asGroupMember()); err != nil {
					return err
				}
			}
			return gluntchSecondPlayer(next)
		},
	})
	if queued != uuid.Nil {
		return nil
	}
	// The chosen player has left between the two prompts (CR 800.4a).
	// Nobody puts the counters anywhere and the rest of the trigger
	// still resolves.
	return gluntchSecondPlayer(ctx)
}

// gluntchSecondPlayer is "Choose a second player to draw a card."
func gluntchSecondPlayer(ctx *Context) error {
	return ChoosePlayer{
		Among:    Players,
		Except:   ctx.ChosenPlayers(),
		Question: "Gluntch, the Bestower — choose a second player to draw a card",
		Then:     gluntchDraw,
	}.Apply(ctx)
}

// gluntchDraw draws for the second player and asks for the third.
func gluntchDraw(ctx *Context) error {
	if second := ctx.ChosenPlayer(); second != uuid.Nil {
		if err := (DrawCards{Player: second, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return ChoosePlayer{
		Among:    Players,
		Except:   ctx.ChosenPlayers(),
		Question: "Gluntch, the Bestower — choose a third player to create two Treasure tokens",
		Then:     gluntchTreasure,
	}.Apply(ctx)
}

// gluntchTreasure is the last clause.
func gluntchTreasure(ctx *Context) error {
	third := ctx.ChosenPlayer()
	if third == uuid.Nil {
		return nil
	}
	return CreateToken{Controller: third, Template: TreasureToken(), N: 2}.Apply(ctx)
}

// creaturesControlledByPlayer lists the creatures a player controls,
// in battlefield order — the candidate set behind "a creature they
// control".
//
// Here rather than in the card file because it is the shape of the
// clause, not of the card, and a player-choice prompt over somebody
// else's board is exactly the case that has to build its candidate
// list from what that seat can actually do (CR 608.2).
//
// Caller must hold g.mu — it is an effect-time read.
func creaturesControlledByPlayer(g *game.Game, playerID uuid.UUID) []uuid.UUID {
	if playerID == uuid.Nil {
		return nil
	}
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != playerID || !c.IsCreature() {
			continue
		}
		out = append(out, c.InstanceID)
	}
	return out
}
