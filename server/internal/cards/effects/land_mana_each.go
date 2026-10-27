package effects

import (
	"slices"
	"strconv"
	"strings"

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
//
// Then is a registered tail (RegisterLandManaThen), and the walk is data:
// the lands still to visit ride the option pick that asks about the
// current one (landManaPick), so a table waiting on the question is a
// restore point (#4253 slice 10c).
type ActivateAManaAbilityOfEachLand struct {
	Player uuid.UUID
	// Card is the card name the questions are asked in ("Drain Power").
	Card string
	// Then is the rest of the card, run once with the mana lost. The zero
	// ref is none.
	Then game.TailRef
}

// RegisterLandManaThen registers the rest of an ActivateAManaAbilityOfEachLand
// card: a tail run against a Context whose controller is the card's and
// whose source is the card, handed the mana the player lost (LostMana). The
// key is an on-disk identity (ledgered "tail <key>"): never renamed, never
// reused. Call it once, from a package-level var.
func RegisterLandManaThen(key string, body func(ctx *Context, lost []game.ManaToken) error) game.TailRef {
	return CardTail(key, func(ctx *Context, o game.TailOutcome) error {
		return body(ctx, lostManaOf(o.Tail.Word))
	})
}

// lostManaWord writes the mana a player lost as the tail's Word, the
// colours in the order they were lost: "R:2|G:1". Restrictions and riders
// are not carried: the mana the tail adds is new mana (Drain Power).
func lostManaWord(lost []game.ManaToken) string {
	var parts []string
	for _, t := range lost {
		parts = append(parts, t.Color+":"+strconv.Itoa(t.N()))
	}
	return strings.Join(parts, "|")
}

// lostManaOf reads lostManaWord back, one run per colour written.
func lostManaOf(word string) []game.ManaToken {
	var out []game.ManaToken
	for _, part := range strings.Split(word, "|") {
		color, n, ok := strings.Cut(part, ":")
		if !ok || color == "" {
			continue
		}
		count, err := strconv.Atoi(n)
		if err != nil || count <= 0 {
			continue
		}
		t := game.ManaToken{Color: color}
		if count > 1 {
			t.Count = count
		}
		out = append(out, t)
	}
	return out
}

func (a ActivateAManaAbilityOfEachLand) Apply(ctx *Context) error {
	w := landManaWalk{player: a.Player, card: a.Card, then: a.Then}
	if p := ctx.Game.PlayerByIDForEffect(a.Player); p != nil && !p.Eliminated {
		w.lands = ctx.Game.LandsControlledForEffect(a.Player)
	}
	return w.step(ctx)
}

// landManaWalk is where the walk stands: the lands it visits in order, the
// next one, the lands whose abilities wait on mana (deferred) and whether
// the walk is on its second pass over them. A plain value.
type landManaWalk struct {
	player   uuid.UUID
	card     string
	then     game.TailRef
	lands    []uuid.UUID
	next     int
	deferred []uuid.UUID
	retrying bool
}

// step activates the next land's mana ability, asking when there is a
// choice; the answer's continuation (landManaPick) calls step again with
// the walk it read off the question.
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
		opts, waiting := ctx.Game.LandManaOptionsForEffect(w.player, land)
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
			_ = ctx.Game.ActivateLandManaForEffect(w.player, opts[0])
			continue
		}
		return w.ask(ctx, land, opts, nil)
	}
}

// ask puts the land's options to the player. The walk is written on the
// question: Carry holds the land asked about, then the lands still to
// visit, then the deferred ones; Nums holds how many are still to visit and
// whether this is the second pass; Words holds the card, the tail's key and
// the labels of options already refused. A refused activation (the board
// moved under it) asks again without that option.
func (w landManaWalk) ask(ctx *Context, land uuid.UUID, opts []game.LandManaOption, refused []string) error {
	choices := make([]game.ChoiceOption, len(opts))
	for i, o := range opts {
		choices[i] = game.ChoiceOption{Label: o.Label}
	}
	remaining := w.lands[w.next:]
	carry := append([]uuid.UUID{land}, remaining...)
	carry = append(carry, w.deferred...)
	retrying := 0
	if w.retrying {
		retrying = 1
	}
	words := append([]string{w.card, w.then.Key()}, refused...)
	return PickOption{
		Player:   w.player,
		Question: w.card + " — activate a mana ability of this land",
		Options:  choices,
		ThenKey:  landManaPick,
		Carry:    carry,
		Words:    words,
		Nums:     []int{len(remaining), retrying},
	}.Apply(ctx)
}

// landManaPick is the answer to a land's question: the chosen way is
// activated and the walk goes on. The options are asked of the land again,
// less the ones already refused, since the question was asked of the board
// as it stands and the board has not moved. An on-disk key. Registered in
// init, not a var initialiser, because the walk refers to it and it refers
// to the walk.
var landManaPick game.OptionPickThen

func init() {
	landManaPick = OptionPickThen("option-pick/land-mana-each", landManaAnswer)
}

func landManaAnswer(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 || len(r.Words) < 2 || len(r.Nums) < 2 {
		return nil
	}
	n := r.Nums[0]
	if n < 0 || 1+n > len(r.Carry) {
		return nil
	}
	w := landManaWalk{
		player:   r.Chooser,
		card:     r.Words[0],
		lands:    append([]uuid.UUID(nil), r.Carry[1:1+n]...),
		deferred: append([]uuid.UUID(nil), r.Carry[1+n:]...),
		retrying: r.Nums[1] != 0,
	}
	if ref, ok := game.TailRefByKey(r.Words[1]); ok {
		w.then = ref
	}
	land, refused := r.Carry[0], r.Words[2:]
	if r.Index == game.NoChoiceIndex || r.Option == nil {
		// Nobody could answer (the player left): stop asking.
		w.next, w.deferred, w.retrying = len(w.lands), nil, true
		return w.step(ctx)
	}
	all, _ := ctx.Game.LandManaOptionsForEffect(w.player, land)
	var opts []game.LandManaOption
	for _, o := range all {
		if !slices.Contains(refused, o.Label) {
			opts = append(opts, o)
		}
	}
	chosen := -1
	if r.Index >= 0 && r.Index < len(opts) && opts[r.Index].Label == r.Option.Label {
		chosen = r.Index
	} else {
		for i, o := range opts {
			if o.Label == r.Option.Label {
				chosen = i
				break
			}
		}
	}
	if chosen < 0 || ctx.Game.ActivateLandManaForEffect(w.player, opts[chosen]) != nil {
		rest := opts
		refusedNow := append([]string(nil), refused...)
		if chosen >= 0 {
			rest = append(append([]game.LandManaOption(nil), opts[:chosen]...), opts[chosen+1:]...)
			refusedNow = append(refusedNow, opts[chosen].Label)
		}
		switch len(rest) {
		case 0:
		case 1:
			_ = ctx.Game.ActivateLandManaForEffect(w.player, rest[0])
		default:
			return w.ask(ctx, land, rest, refusedNow)
		}
	}
	return w.step(ctx)
}

func (w landManaWalk) finish(ctx *Context) error {
	lost := ctx.Game.LoseAllUnspentManaForEffect(w.player)
	if w.then.Key() == "" {
		return nil
	}
	t := game.KeyedTail{Ref: w.then, Player: ctx.Controller(), Source: ctx.Source(), Word: lostManaWord(lost)}
	return t.RunTold(ctx.Game, nil, game.TotalMana(lost))
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
