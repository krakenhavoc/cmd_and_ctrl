package effects

import (
	"slices"
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dack Fayden, Helping Hand — Legendary Creature — Human Advisor
// {4}{W}{W}, 4/6:
//
//	"When Dack Fayden enters, reveal cards from the top of your library
//	 until you reveal X creature cards, where X is the number of
//	 opponents you have. Put those creature cards onto the battlefield,
//	 then shuffle. They're goaded for the rest of the game. For each of
//	 those permanents, choose a different opponent. Each opponent gains
//	 control of the permanent for which they were chosen."
//
// Three things worth knowing.
//
//   - The reveal is revealUntilCount, the reveal-until Selvala's Stampede
//     uses, with X read as the live opponent count (eliminated seats are
//     not opponents). A library that runs out first reveals all of
//     itself and puts the creatures it had; every revealed card that
//     wasn't put on the battlefield is in the library when it is
//     shuffled.
//   - "Goaded for the rest of the game" is the goad marker with no end:
//     game.GoadForRestOfGameForEffect (#3522, ADR 0045's 2026-10-14
//     amendment). The goader is Dack's controller and stays so after the
//     permanent changes hands, so each stolen creature must attack each
//     combat and may not attack the player who goaded it if it can attack
//     anyone else (CR 701.15b). The goad ends only when the permanent
//     leaves the battlefield.
//   - "A different opponent" for each permanent is a chain of single-pick
//     prompts, one per permanent, each leaving out the opponents already
//     chosen, so the client and the bot need no new prompt (a bot takes
//     the first opponent offered, which is always a different one). The
//     continuation is a registered key, so a table waiting on an answer
//     is still a restore point. Control is given as each answer arrives
//     rather than after the last; nothing can happen between the two (no
//     one has priority inside a resolution), so the result is the same
//     as the printed "each opponent gains control" at the end. A
//     permanent with only one opponent left to take it is given to that
//     opponent without a prompt: there is nothing to choose. At two
//     seats that is the whole ability: one creature, the one opponent.
//
// No simplification.
func init() {
	dackFaydenAnswer = OptionPickThen("option-pick/dack-fayden-helping-hand-opponent", dackFaydenChosen)
	Register(Spec{
		OracleID:     "3b21abb0-9a58-4e62-9650-4b04ce2bfb12",
		Name:         "Dack Fayden, Helping Hand",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dack Fayden, Helping Hand — reveal until X creature cards, goad them, and give them to opponents", dackFaydenEnters),
		},
	})
}

// dackFaydenAnswer is the registered continuation of each "choose a
// different opponent" question. Its key is an on-disk identity: never
// renamed, never reused. Assigned in init because the continuation asks
// the next question, which names this key.
var dackFaydenAnswer game.OptionPickThen

func dackFaydenEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Controller()
	run, hits := revealUntilCount(ctx, player, func(c game.Card) bool { return c.IsCreature() },
		len(ctx.Opponents()), "Dack Fayden, Helping Hand — reveal until X creature cards")
	return PutFromLibraryOntoBattlefield{
		Player:   player,
		Cards:    run,
		Match:    func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return selvalasStampedeIsHit(hits, c.InstanceID) },
		All:      true,
		ThenTail: dackFaydenShuffleAndGoad,
	}.Apply(ctx)
}

// dackFaydenShuffleAndGoad is the rider as a keyed tail (#4253 slice 6):
// shuffle, goad what entered, then hand the permanents out. An on-disk
// key.
var dackFaydenShuffleAndGoad = PutFromLibraryTail("dack-fayden-helping-hand/shuffle-goad-and-give",
	func(ctx *Context, res PutFromLibraryResult) error {
		g := ctx.Game
		// "Then shuffle": the rest are still in the library.
		if err := g.ShuffleLibraryForEffect(res.Player); err != nil {
			return err
		}
		var perms []uuid.UUID
		for _, id := range res.Entered {
			if g.GoadForRestOfGameForEffect(id, res.Player) {
				perms = append(perms, id)
			}
		}
		// Rebuilt from the live game: this runs after any entry
		// question, and undo restores a clone.
		return dackFaydenAsk(NewContext(g, &game.StackItem{
			Kind:         game.StackItemSpell,
			Controller:   res.Player,
			Owner:        res.Player,
			SourceCardID: res.Source,
		}), perms, nil)
	})

// dackFaydenCarry packs the permanents still to hand out and the
// opponents already chosen into one ID list: the permanents, a
// uuid.Nil, then the opponents.
func dackFaydenCarry(perms, used []uuid.UUID) []uuid.UUID {
	out := append([]uuid.UUID(nil), perms...)
	out = append(out, uuid.Nil)
	return append(out, used...)
}

func dackFaydenUnpack(carry []uuid.UUID) (perms, used []uuid.UUID) {
	for i, id := range carry {
		if id == uuid.Nil {
			return carry[:i], carry[i+1:]
		}
	}
	return carry, nil
}

// dackFaydenAsk hands out the first of `perms` that is still on the
// battlefield: to the only opponent left when there is just one, else by
// asking the controller. `used` are the opponents who already took one.
func dackFaydenAsk(ctx *Context, perms, used []uuid.UUID) error {
	g := ctx.Game
	for i, id := range perms {
		c, ok := g.LookupCardForEffect(id)
		if !ok || !onBattlefield(g, id) {
			continue
		}
		var options []game.ChoiceOption
		for _, opp := range ctx.Opponents() {
			if slices.Contains(used, opp) {
				continue
			}
			if p := ctx.PlayerByID(opp); p != nil {
				name := p.Name
				if name == "" {
					name = "Seat " + strconv.Itoa(p.Seat+1)
				}
				options = append(options, game.ChoiceOption{Label: name, Player: opp})
			}
		}
		switch len(options) {
		case 0:
			return nil
		case 1:
			return dackFaydenGive(ctx, perms[i:], used, options[0].Player)
		}
		return PickOption{
			Player:   ctx.Controller(),
			Question: "Dack Fayden, Helping Hand — choose a different opponent to gain control of " + c.Name,
			Options:  options,
			ThenKey:  dackFaydenAnswer,
			Carry:    dackFaydenCarry(perms[i:], used),
		}.Apply(ctx)
	}
	return nil
}

// dackFaydenChosen is the answer: the first carried permanent goes to
// the opponent chosen, and the next is asked about. Nobody choosing (the
// controller left the game) stops the chain.
func dackFaydenChosen(ctx *Context, r game.OptionPicked) error {
	if r.Option == nil || r.Option.Player == uuid.Nil {
		return nil
	}
	perms, used := dackFaydenUnpack(r.Carry)
	return dackFaydenGive(ctx, perms, used, r.Option.Player)
}

// dackFaydenGive gives perms[0] to `opp` and asks about the rest.
func dackFaydenGive(ctx *Context, perms, used []uuid.UUID, opp uuid.UUID) error {
	if len(perms) == 0 {
		return nil
	}
	ctx.Game.GainControlForEffect(ctx.Source(), perms[0], opp, game.IndefiniteDuration(),
		"Dack Fayden, Helping Hand — gain control")
	return dackFaydenAsk(ctx, perms[1:], append(append([]uuid.UUID(nil), used...), opp))
}
