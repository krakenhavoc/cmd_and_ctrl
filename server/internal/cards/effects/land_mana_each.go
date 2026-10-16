package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// land_mana_each.go — ADR 0200, #3952: "<player> activates a mana ability
// of each land they control. Then that player loses all unspent mana"
// (Drain Power, Pygmy Hippo).

// ActivateAManaAbilityOfEachLand makes Player activate a mana ability of
// each land they control, then lose all unspent mana, and hands Then the
// mana they lost (CR 605.3a, CR 106.4).
//
// The lands are the ones Player controls as the instruction begins, taken
// in battlefield order. For each one the player chooses among the ways
// they can activate one of its mana abilities right now
// (game.LandManaOptionsForEffect), colours included, through an
// option_pick; a land with a single way activates it without a question,
// and a land with none (tapped, say) adds nothing. A land with an ability
// whose mana cost the pool cannot pay yet (a filter land's "{B/R}, {T}") is
// asked after the others, so their mana can pay for it if the player
// chooses that ability.
//
// Then runs once, after the last land, with the mana lost: every mana in
// the pool at that moment, including mana that was there before. It runs
// even when Player has left the game (with nothing lost), so the rest of
// the card still happens.
type ActivateAManaAbilityOfEachLand struct {
	Player uuid.UUID
	// Card is the card name the questions are asked in ("Drain Power").
	Card string
	Then func(ctx *Context, lost []game.ManaToken) error
}

func (a ActivateAManaAbilityOfEachLand) Apply(ctx *Context) error {
	w := landManaWalk{a: a}
	if p := ctx.Game.PlayerByIDForEffect(a.Player); p != nil && !p.Eliminated {
		w.lands = ctx.Game.LandsControlledForEffect(a.Player)
	}
	return w.step(ctx)
}

type landManaWalk struct {
	a        ActivateAManaAbilityOfEachLand
	lands    []uuid.UUID
	next     int
	deferred []uuid.UUID
	retrying bool
}

// step activates the next land's mana ability, asking when there is a
// choice; the answer's continuation calls step again. The walk is a value
// and every continuation holds its own copy, so an undo that brings a
// question back also brings back the walk as it stood when it was asked.
func (w landManaWalk) step(ctx *Context) error {
	for {
		if w.next >= len(w.lands) {
			if w.retrying || len(w.deferred) == 0 {
				return w.finish(ctx)
			}
			w.lands, w.next, w.deferred, w.retrying = w.deferred, 0, nil, true
		}
		land := w.lands[w.next]
		w.next++
		opts, waiting := ctx.Game.LandManaOptionsForEffect(w.a.Player, land)
		// A land with an ability its pool cannot pay for yet waits for
		// the others, whatever else it offers, so the player may still
		// choose that ability once the other lands' mana is in the pool.
		if waiting && !w.retrying {
			w.deferred = append(append([]uuid.UUID(nil), w.deferred...), land)
			continue
		}
		switch {
		case len(opts) == 0:
			continue
		case len(opts) == 1:
			// A refusal leaves nothing paid; the land adds nothing.
			_ = ctx.Game.ActivateLandManaForEffect(w.a.Player, opts[0])
			continue
		}
		return w.ask(ctx, opts)
	}
}

// ask puts the land's options to the player. A refused activation (the
// board moved under it) asks again without that option.
func (w landManaWalk) ask(ctx *Context, opts []game.LandManaOption) error {
	choices := make([]game.ChoiceOption, len(opts))
	for i, o := range opts {
		choices[i] = game.ChoiceOption{Label: o.Label}
	}
	return PickOption{
		Player:   w.a.Player,
		Question: w.a.Card + " — activate a mana ability of this land",
		Options:  choices,
		Then: func(ctx *Context, index int) error {
			if index < 0 || index >= len(opts) {
				// Nobody could answer (the player left): stop asking.
				done := w
				done.next, done.deferred, done.retrying = len(w.lands), nil, true
				return done.step(ctx)
			}
			if err := ctx.Game.ActivateLandManaForEffect(w.a.Player, opts[index]); err != nil {
				rest := append(append([]game.LandManaOption(nil), opts[:index]...), opts[index+1:]...)
				switch len(rest) {
				case 0:
				case 1:
					_ = ctx.Game.ActivateLandManaForEffect(w.a.Player, rest[0])
				default:
					return w.ask(ctx, rest)
				}
			}
			return w.step(ctx)
		},
	}.Apply(ctx)
}

func (w landManaWalk) finish(ctx *Context) error {
	lost := ctx.Game.LoseAllUnspentManaForEffect(w.a.Player)
	if w.a.Then == nil {
		return nil
	}
	return w.a.Then(ctx, lost)
}

// addManaLike adds to `player` one mana of each type in `tokens`: the same
// colours, as new mana this effect makes (Drain Power's "you add the mana
// lost this way"). Restrictions and spend riders belonged to the ability
// that first made the mana and do not come with it.
func addManaLike(ctx *Context, player uuid.UUID, tokens []game.ManaToken) error {
	counts := map[string]int{}
	var order []string
	for _, t := range tokens {
		if _, seen := counts[t.Color]; !seen {
			order = append(order, t.Color)
		}
		counts[t.Color] += t.N()
	}
	for _, c := range order {
		if err := (AddMana{Player: player, Amount: ManaN(counts[c], c)}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
