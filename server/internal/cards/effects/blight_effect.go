package effects

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blight_effect.go — #4683: "blight N" as the EFFECT of a resolving spell
// or ability (CR 701.68a), the card side of game.BlightKeyedForEffect.
//
//	"When this creature enters, you may blight 2. If you do, create two
//	 1/1 black and red Goblin creature tokens." (Sourbread Auntie)
//	"Whenever this creature attacks, blight 1." (Shadow Urchin)
//	"… each opponent blights 1." (High Perfect Morcant)
//
// Blight{…}.Apply asks each blighting player, as the ability resolves
// (CR 608.2c), which creature they control gets the counters: one
// choose_cards prompt per player, in APNAP order, each seeing the earlier
// choices (CR 101.4, ADR 0175's ordered collector), and then the counters
// are put on every chosen creature. The choice is not a target, so a
// hexproof creature can be blighted and a creature with protection too.
// A player who controls no creature that can have counters put on it is
// not asked and does not blight (CR 701.68b); a mandatory blight with one
// candidate needs no prompt.
//
// The counters are an effect's, not a cost's (game.BlightKeyedForEffect),
// so Doubling Season doubles them (CR 614.16) and the blighting player is
// their placer.
//
// Everything printed after the blight is a registered Blighted
// (RegisterBlighted), told who blighted which creature (CR 701.68c's
// "blighted creature"): "if you do" is BlightDone.Did, "if you don't" its
// negation. It runs once every counter has landed, after any CR 616
// ordering prompt, and it runs when nobody could blight too. The prompts,
// the placements and the rest of the card are data, so a table waiting on
// any of them is a restore point.
//
// "Blight N" paid as an activated ability's COST (#4684) is not this: a
// cost is chosen when the ability is activated and settles at once
// (game.blightLocked). What the two share is game.BlightOptionsForEffect,
// the creatures a player may blight.

// Blight is "<players> blight N" as an effect.
type Blight struct {
	// N is the number of -1/-1 counters. Required, at least one.
	N int

	// Players blight, each choosing one creature they control. Empty means
	// the resolving effect's controller ("you may blight 2"). A player
	// named more than once blights once; the others are asked in APNAP
	// order whatever order they are given in.
	Players []uuid.UUID

	// May is "you may blight N": each player may decline, and choosing no
	// creature is not blighting.
	May bool

	// Question is each prompt's header. Empty gets one built from N.
	Question string

	// Then is the rest of the card (RegisterBlighted). Zero means nothing
	// follows ("each opponent blights 1.").
	Then Blighted

	// Carry and Nums are plain data handed to Then (BlightDone.Carry /
	// BlightDone.Nums).
	Carry []uuid.UUID
	Nums  []int
}

// Blighted is a registered rest of the card after a Blight. Hold one only
// from RegisterBlighted.
type Blighted struct{ run game.PromptRunThen }

// BlightDone is what a Blighted body is told.
type BlightDone struct {
	// Players are the players who blighted, in the order they were asked,
	// and Creatures the creature each of them blighted (CR 701.68c), in the
	// same order. A player who declined, or could not blight, is in
	// neither.
	Players   []uuid.UUID
	Creatures []uuid.UUID
	// Placed is how many -1/-1 counters landed in all, after the
	// replacement effects (zero when Vizier of Remedies took the only one).
	Placed int
	// Carry and Nums are the Blight's own.
	Carry []uuid.UUID
	Nums  []int
}

// Did reports whether anyone blighted — "if you do" on a blight only the
// controller makes. A blight whose counters a replacement removed was
// still done: the creature was chosen.
func (d BlightDone) Did() bool { return len(d.Creatures) > 0 }

// By is the creature `player` blighted, or uuid.Nil.
func (d BlightDone) By(player uuid.UUID) uuid.UUID {
	for i, p := range d.Players {
		if p == player && i < len(d.Creatures) {
			return d.Creatures[i]
		}
	}
	return uuid.Nil
}

// RegisterBlighted registers the rest of a card after its blight under key
// (ledgered as "run <key>" and "tail blighted/<key>": never renamed, never
// reused). Call it once, from a package-level var. The body runs against a
// Context rebuilt from the asking item.
func RegisterBlighted(key string, then func(ctx *Context, d BlightDone) error) Blighted {
	tailKey := "blighted/" + key
	if rest, ok := strings.CutPrefix(key, "test/"); ok {
		// A test's blight keeps its tail in the ledger's test namespace.
		tailKey = "test/blighted-" + rest
	}
	done := CardTail(tailKey, func(ctx *Context, o game.TailOutcome) error {
		lists := unpackIDLists(o.Tail.Carry)
		for len(lists) < 3 {
			lists = append(lists, nil)
		}
		return then(ctx, BlightDone{
			Players:   lists[1],
			Creatures: lists[2],
			Placed:    o.Amount,
			Carry:     lists[0],
			Nums:      append([]int(nil), o.Tail.Nums...),
		})
	})
	return Blighted{run: RunThen(key, func(ctx *Context, r game.RunDone) error {
		if len(r.Nums) == 0 {
			return nil
		}
		n, nums := r.Nums[0], r.Nums[1:]
		var players, creatures []uuid.UUID
		for _, sc := range r.Picks() {
			if len(sc.Cards) == 0 {
				continue
			}
			players = append(players, sc.Seat)
			creatures = append(creatures, sc.Cards[0])
		}
		next := game.KeyedTail{Ref: done, Player: ctx.Controller(), Source: r.Source, Asked: r.Asked}
		chain := game.KeyedTail{
			Ref:    blightPlaceNext,
			Player: ctx.Controller(),
			Source: r.Source,
			Carry:  packIDLists(r.Carry, players, creatures),
			Nums:   append([]int{n, 0, 0}, nums...),
			Next:   &next,
		}
		return chain.Run(ctx.Game, nil)
	})}
}

// blightNothingAfter is the Blighted of a blight nothing follows. An
// on-disk key.
var blightNothingAfter = RegisterBlighted("blight/nothing-after", func(*Context, BlightDone) error { return nil })

// blightPlaceNext puts the counters on the chosen creatures one at a time,
// each placement's settled count added up, and then runs the card's tail
// (its Next) told the creatures and the total. The tail's Carry is
// packIDLists(the Blight's carry, the players, their creatures); its Nums
// are N, the next creature to place on, the counters placed so far, and
// then the Blight's own numbers. Each placement may pause on a CR 616
// ordering prompt and resumes here, so the chain is data throughout. An
// on-disk key.
var blightPlaceNext = game.RegisterTail("blight/place-next", func(g *game.Game, o game.TailOutcome) error {
	t := o.Tail
	if len(t.Nums) < 3 || t.Next == nil {
		return nil
	}
	n, next, placed := t.Nums[0], t.Nums[1], t.Nums[2]+o.Amount
	lists := unpackIDLists(t.Carry)
	for len(lists) < 3 {
		lists = append(lists, nil)
	}
	carry, players, creatures := lists[0], lists[1], lists[2]
	if next < len(creatures) && next < len(players) {
		step := t
		step.Nums = append([]int{n, next + 1, placed}, t.Nums[3:]...)
		return g.BlightKeyedForEffect(players[next], creatures[next], n, step)
	}
	done := *t.Next
	done.Carry = packIDLists(carry, players, creatures)
	done.Nums = append([]int(nil), t.Nums[3:]...)
	return done.RunTold(g, creatures, placed)
})

// Apply asks the blight and puts the counters.
func (b Blight) Apply(ctx *Context) error {
	if b.N <= 0 {
		return fmt.Errorf("%w: blight of %d", game.ErrInvalidParam, b.N)
	}
	then := b.Then
	if !then.run.Valid() {
		then = blightNothingAfter
	}
	who := b.Players
	if len(who) == 0 {
		who = []uuid.UUID{ctx.Controller()}
	}
	named := make(map[uuid.UUID]bool, len(who))
	for _, id := range who {
		named[id] = true
	}
	question := b.Question
	if question == "" {
		question = blightQuestion(ctx, b.N, b.May)
	}
	floor := 1
	if b.May {
		floor = 0
	}
	var legs []game.ChooseCardsPrompt
	for _, seat := range apnapPlayers(ctx.Game) {
		if !named[seat] {
			continue
		}
		// CR 701.68b: a player who can't put the counters on a creature
		// they control can't blight, and is not asked.
		options := ctx.Game.BlightOptionsForEffect(seat)
		if len(options) == 0 {
			continue
		}
		legs = append(legs, game.ChooseCardsPrompt{
			Chooser:     seat,
			FromPlayer:  seat,
			Source:      ctx.Source(),
			Question:    question,
			Cards:       options,
			Min:         floor,
			Max:         1,
			Zone:        game.ZoneBattlefield,
			Destination: game.ChooseBlight,
		})
	}
	nums := append([]int{b.N}, b.Nums...)
	return ctx.Game.ChoosePermanentsInOrderKeyedForEffect(legs, then.run, ctx.Source(), ctx.RunData(append([]uuid.UUID(nil), b.Carry...), nums...))
}

// blightQuestion is a blight prompt's default header: the source card's
// name and what is asked.
func blightQuestion(ctx *Context, n int, may bool) string {
	what := "a -1/-1 counter"
	if n > 1 {
		what = fmt.Sprintf("%d -1/-1 counters", n)
	}
	verb := fmt.Sprintf("blight %d: choose a creature you control to put %s on", n, what)
	if may {
		verb = fmt.Sprintf("you may blight %d: choose a creature you control to put %s on, or choose nothing", n, what)
	}
	if c, ok := ctx.Game.LookupCardForEffect(ctx.Source()); ok && c.Name != "" {
		return c.Name + " — " + verb
	}
	return verb
}
