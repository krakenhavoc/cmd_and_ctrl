package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// villainous_choice.go — "[a player] faces a villainous choice —
// [option A], or [option B]" (CR 701.55, #3140).
//
// CR 701.55a: the player chooses A or B, "then all actions in the
// chosen option are performed". CR 701.55b: they may choose an option
// that is illegal or impossible and "perform as much of the action as
// is possible" — an exception to 608.2d, so BOTH options are always
// offered, and "sacrifice a creature" with no creature is a legal pick
// that sacrifices nothing. Nothing here prunes an option; the body of
// an option is written to do as much as it can.
//
// CR 701.55c: a replacement effect (The Valeyard) may make a player
// face the choice additional times, "one at a time". CR 701.55d: with
// several players it is performed for each one at a time in APNAP
// order, which is an exception to 608.2e's usual "each player chooses
// in turn order and then it happens".
//
// The question is the #568 option pick, addressed to the facing player
// and not to the controller, with a KEYED continuation (#2854) so a
// table waiting on it is still a restore point. The two options are
// bodies held by the registered VillainousChoice, never captured: the
// prompt carries only plain data.
//
//	var thisIsHowItEnds = NewVillainousChoice("this-is-how-it-ends",
//	    "This Is How It Ends — lose 5 life or shuffle another creature?",
//	    VillainousOption{Label: "You lose 5 life", LifeCost: 5, Body: …},
//	    VillainousOption{Label: "You shuffle another creature you own into your library", Harm: 2, Body: …})
//	…
//	FaceVillainousChoice{Players: opponents, Choice: thisIsHowItEnds}.Apply(ctx)
//
// # Carry
//
// The prompt's Carry (after the controller, which PickOption puts first
// and OptionPickThen takes off again) is the players still to face the
// choice, uuid.Nil, then the card's own Carry. A player who must face
// it twice appears twice, so "the entire process … one at a time" is
// the list order and no count is stored. Facing.Carry is the card's
// part.
//
// # What a body must not assume
//
// The bodies run when the facing player answers, not when the effect
// resolved: read the board then, capture nothing, and treat the
// facing player as Facing.Player and the effect's controller as
// ctx.Controller(). A body that asks a follow-up question (sacrifice a
// creature of your choice) leaves it open beside the next question in
// the chain: the engine has no "after every prompt this body opened"
// seam, and a second question for the same player is answered
// independently of the first.

// VillainousOption is one branch of a villainous choice.
type VillainousOption struct {
	// Label is the option as printed, from the facing player's point of
	// view ("You lose 5 life"). Required.
	Label string

	// LifeCost is the life the option makes the facing player lose,
	// when it is a flat number a bot can weigh against its life total.
	// It is a hint: the Body does the losing.
	LifeCost int

	// Harm is the rest of what it costs them, in cards (1 is a card
	// from hand, 2 a creature), for a bot choosing the lesser evil.
	// A hint; 0 when the card has nothing to price.
	Harm int

	// Possible reports whether the option would do anything to the
	// player facing it right now (nil: always). It prices the hints
	// only: an option that cannot be carried out costs them nothing,
	// so the bot takes it (CR 701.55b), and it stays on offer either
	// way.
	Possible func(g *game.Game, facing uuid.UUID) bool

	// Body performs the option. It runs with the game lock held, may
	// queue prompts of its own, and must capture nothing.
	Body func(ctx *Context, f Facing) error
}

// Facing is what an option's body is told about the choice it is part of.
type Facing struct {
	// Player is the player who faced the choice and picked this option.
	Player uuid.UUID
	// Carry is the plain data the card put on the choice.
	Carry []uuid.UUID
	// Remaining is how many more times this same FaceVillainousChoice
	// will ask a question after this one (CR 701.55c/d: the other
	// players, and the additional times a Valeyard gives one of them).
	// Zero on the last. A card that chains a SECOND choice off the end
	// of the first (Hunted by The Family's next target) starts it only
	// when Remaining is zero, so the chain is not begun once per
	// additional time.
	Remaining int
}

// VillainousChoice is one card's villainous choice, registered once
// under an on-disk key.
type VillainousChoice struct {
	question string
	a, b     VillainousOption
	then     game.OptionPickThen
}

// NewVillainousChoice registers a villainous choice. Call it once per
// printed choice, from a package-level var or an init. The key
// ("this-is-how-it-ends") is an on-disk identity ("option-pick/
// villainous-<key>" in testdata/effect_keys.txt): never renamed, never
// reused. Panics on a missing label or body, as the registries do.
func NewVillainousChoice(key, question string, a, b VillainousOption) *VillainousChoice {
	for _, o := range []VillainousOption{a, b} {
		if o.Label == "" || o.Body == nil {
			panic("effects: villainous choice " + key + " needs a label and a body on both options")
		}
	}
	c := &VillainousChoice{question: question, a: a, b: b}
	c.then = OptionPickThen("option-pick/villainous-"+key, c.answered)
	return c
}

// FaceVillainousChoice makes each of Players face Choice: "each
// opponent faces a villainous choice", "target opponent faces…", "its
// controller faces…".
type FaceVillainousChoice struct {
	// Players face the choice. They are asked in APNAP order (CR
	// 701.55d) whatever order they are listed in; one who has left the
	// game is skipped, and a player listed twice faces it twice.
	Players []uuid.UUID
	Choice  *VillainousChoice
	// Carry is the card's own data, handed to both bodies as
	// Facing.Carry.
	Carry []uuid.UUID
}

// Apply asks the first player's first question. The rest follow as each
// is answered.
func (f FaceVillainousChoice) Apply(ctx *Context) error {
	if f.Choice == nil {
		return nil
	}
	listed := map[uuid.UUID]int{}
	for _, p := range f.Players {
		listed[p]++
	}
	var waiting []uuid.UUID
	for _, p := range apnapPlayers(ctx.Game) {
		n := listed[p]
		if n == 0 {
			continue
		}
		// CR 701.55c: every listed instruction, each faced the
		// additional times the table gives them.
		for i := 0; i < n*ctx.Game.VillainousChoiceTimesForEffect(p); i++ {
			waiting = append(waiting, p)
		}
	}
	return f.Choice.ask(ctx, waiting, f.Carry)
}

// ask queues the question for the first player in `waiting` who is
// still in the game. With nobody left it does nothing: the chain is over.
func (c *VillainousChoice) ask(ctx *Context, waiting, carry []uuid.UUID) error {
	for i, p := range waiting {
		pl := ctx.Game.PlayerByIDForEffect(p)
		if pl == nil || pl.Eliminated {
			continue
		}
		packed := append(append([]uuid.UUID(nil), waiting[i+1:]...), uuid.Nil)
		packed = append(packed, carry...)
		return PickOption{
			Player:   p,
			Question: c.question,
			Options:  []game.ChoiceOption{c.a.offer(ctx.Game, p), c.b.offer(ctx.Game, p)},
			ThenKey:  c.then,
			Carry:    packed,
		}.Apply(ctx)
	}
	return nil
}

// offer is the option as the facing player is shown it: both options
// are always offered (CR 701.55b), and the price hints are zeroed for
// one that would do nothing to them.
func (o VillainousOption) offer(g *game.Game, facing uuid.UUID) game.ChoiceOption {
	opt := game.ChoiceOption{Label: o.Label, LifeCost: o.LifeCost, Harm: o.Harm}
	if o.Possible != nil && !o.Possible(g, facing) {
		opt.LifeCost, opt.Harm = 0, 0
	}
	return opt
}

// answered is the registered continuation: perform the chosen option,
// then ask the next question in the chain. A player who chose nothing
// (they left the game, CR 800.4a) does nothing and the chain goes on.
func (c *VillainousChoice) answered(ctx *Context, r game.OptionPicked) error {
	sep := -1
	for i, id := range r.Carry {
		if id == uuid.Nil {
			sep = i
			break
		}
	}
	if sep < 0 {
		return nil
	}
	waiting, carry := r.Carry[:sep], append([]uuid.UUID(nil), r.Carry[sep+1:]...)
	var err error
	if r.Option != nil {
		opt := c.a
		if r.Index == 1 {
			opt = c.b
		}
		err = opt.Body(ctx, Facing{Player: r.Chooser, Carry: carry, Remaining: stillToAsk(ctx.Game, waiting)})
	}
	if nextErr := c.ask(ctx, waiting, carry); err == nil {
		err = nextErr
	}
	return err
}

// stillToAsk is how many of the waiting players are still in the game:
// the ones ask will actually put a question to. A player who has left
// is skipped there (CR 800.4a), so Facing.Remaining must not count them.
func stillToAsk(g *game.Game, waiting []uuid.UUID) int {
	n := 0
	for _, p := range waiting {
		if pl := g.PlayerByIDForEffect(p); pl != nil && !pl.Eliminated {
			n++
		}
	}
	return n
}

// villainousSacrifice is the body of "[that player] sacrifices a <label>
// of their choice": one sacrifice prompt for the facing player, none
// when they have nothing that qualifies (CR 701.21a, 701.55b).
func villainousSacrifice(label string, match CardPredicate) func(*Context, Facing) error {
	return func(ctx *Context, f Facing) error {
		ctx.Game.PlayerSacrificesForEffect(ctx.Source(), f.Player, sacrificeSpec(label, match), "Sacrifice "+label)
		return nil
	}
}

// villainousCanSacrifice is the Possible check that goes with
// villainousSacrifice.
func villainousCanSacrifice(match CardPredicate) func(*game.Game, uuid.UUID) bool {
	return func(g *game.Game, facing uuid.UUID) bool {
		return len(permanentsControlledByMatching(g, facing, match)) > 0
	}
}

// villainousDiscard is the body of "[that player] discards N cards": the
// facing player's own pick from their hand (CR 701.9a), as many as they
// have when the hand is smaller, and no prompt for an empty hand.
func villainousDiscard(n int) func(*Context, Facing) error {
	return func(ctx *Context, f Facing) error {
		ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: f.Player, Source: ctx.Source(), N: n})
		return nil
	}
}

// villainousHandSize is the number of cards in a player's hand.
func villainousHandSize(g *game.Game, player uuid.UUID) int {
	if p := g.PlayerByIDForEffect(player); p != nil {
		return p.Hand.Size()
	}
	return 0
}
