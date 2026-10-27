package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// each_player_in_turn.go: "each player mills ...", "exile the top card of
// each opponent's library" and the other multi-library instructions, done
// one library at a time (#4266).
//
// Why a chain and not a loop. A mill or an exile can stop to ask a question:
// a CR 616.1 ordering prompt between two replacement effects, a commander's
// CR 903.9 offer. A loop that calls the fire-and-forget form once per player
// starts every player's mill on one line, so every library that has something
// to ask opens its prompt at once and the table answers them in whatever order
// the UI shows them. Players who must choose at the same time do it in APNAP
// order (CR 101.4, and CR 616.1 for the replacement choices a mill asks): the
// active player first, then each other player in turn order. A chain makes
// that real. Each player's mill starts from the previous player's continuation, so at
// most one of these prompts is open at a time, the active player's first, and
// the next library is not touched until it is answered. Doomsday
// Excruciator (doomsdayExileFrom) does the same by hand; a card that does it for
// several players calls the helpers below instead of writing another copy.
//
// The chain is data (#4253): each step is a registered tail
// (registerEachPlayerStep) carrying the players still to go, the cards done
// so far and the card's own tail, so a prompt any library's instruction
// pauses on is a restore point.

// apnapOrder returns the members of `players` that are still in the game, in
// APNAP order (CR 101.4). A player named twice is listed once.
func apnapOrder(g *game.Game, players []uuid.UUID) []uuid.UUID {
	want := make(map[uuid.UUID]bool, len(players))
	for _, id := range players {
		want[id] = true
	}
	var out []uuid.UUID
	for _, id := range apnapPlayers(g) {
		if want[id] {
			out = append(out, id)
		}
	}
	return out
}

// eachPlayerStepFunc starts one player's share of a keyed each-player
// chain: `n` is that player's number (its entry in the chain's Nums, 0 when
// the chain carries none), and `step` is the tail the instruction must end
// on, told the cards it did the instruction to, from wherever it ends:
// inline on a clean run, from the answer when a prompt paused it. The ctx
// is rebuilt from the step (tailContext).
type eachPlayerStepFunc func(ctx *Context, player uuid.UUID, n int, step game.KeyedTail) error

// registerEachPlayerStep registers one step of a keyed each-player chain
// under key, an on-disk identity: never renamed, never reused, ledgered
// "tail <key>". The step's Carry is the players still to go and the cards
// done so far (packIDLists), its Nums the numbers of the players still to
// go (empty for a chain with none), and its Next the card's own tail, told
// every card the chain did its instruction to, library by library in APNAP
// order, once the last player is done. Every other field (Player, Source,
// N, Asked) is the card's, carried unchanged. Call it once, from a
// package-level var.
func registerEachPlayerStep(key string, act eachPlayerStepFunc) game.TailRef {
	return game.RegisterTail(key, func(g *game.Game, o game.TailOutcome) error {
		remaining, landed := unpackIDPair(o.Tail.Carry)
		landed = append(append([]uuid.UUID(nil), landed...), o.IDs...)
		return eachPlayerStepFrom(g, o.Tail, remaining, o.Tail.Nums, landed, act)
	})
}

// eachPlayerInTurn starts a keyed each-player chain: `act` is done for each
// of `players` in APNAP order (CR 101.4), each from the previous one's
// continuation, and then `done` is told every card the chain did the
// instruction to. `step` is the chain's registered step
// (registerEachPlayerStep) with the card's own fields set (Player, Source,
// N, Asked). `nums`, when not nil, is each player's number, read by `act`.
//
// The cards are carried forward BY VALUE on the step's tail, never through
// a variable the continuations share: an undo across a prompt replays the
// tail, and a list built up in a captured slice would count the first
// run's cards twice (the property Dread Summons, #893, needs).
func eachPlayerInTurn(g *game.Game, step game.KeyedTail, players []uuid.UUID, nums func(uuid.UUID) int, done game.KeyedTail, act eachPlayerStepFunc) error {
	order := apnapOrder(g, players)
	var ns []int
	if nums != nil {
		for _, id := range order {
			ns = append(ns, nums(id))
		}
	}
	step.Next = nil
	if !done.IsZero() {
		next := done
		step.Next = &next
	}
	return eachPlayerStepFrom(g, step, order, ns, nil, act)
}

// eachPlayerStepFrom does the instruction to the head of `remaining`, the
// step's tail carrying the rest, or hands `landed` to the card's tail when
// nobody is left.
func eachPlayerStepFrom(g *game.Game, step game.KeyedTail, remaining []uuid.UUID, nums []int, landed []uuid.UUID, act eachPlayerStepFunc) error {
	if len(remaining) == 0 {
		if step.Next == nil {
			return nil
		}
		return step.Next.Run(g, landed)
	}
	head := remaining[0]
	n := 0
	var rest []int
	if len(nums) > 0 {
		n, rest = nums[0], append([]int(nil), nums[1:]...)
	}
	step.Carry = packIDLists(remaining[1:], landed)
	step.Nums = rest
	return act(tailContext(g, step), head, n, step)
}

// millEachPlayerInTurn is "each of these players mills N cards": each library
// in APNAP order, one prompt at a time, then `done`, a registered tail, is
// told every card that reached a graveyard, library by library. `n` is asked
// once per player BEFORE any of them mills, so a count that reads the table
// ("cards equal to the number of cards in their hand", Folio of Fancies) is
// read as the instruction begins, as the paper instruction does, and not
// after an earlier player's mill. A player whose N is not positive mills
// nothing. `done` may be zero.
func millEachPlayerInTurn(ctx *Context, players []uuid.UUID, n func(player uuid.UUID) int, done game.KeyedTail) error {
	return eachPlayerInTurn(ctx.Game, ctx.ItemTail(millEachStep), players, n, done, millEachAct)
}

// millEachStep is one step of millEachPlayerInTurn's chain. An on-disk key.
var millEachStep = registerEachPlayerStep("each-player/mill-step", millEachAct)

// millEachAct mills `n` cards from one player's library, the keyed mill
// (#4253 slice 9), so a leg's CR 616 prompt is a restore point.
func millEachAct(ctx *Context, player uuid.UUID, n int, step game.KeyedTail) error {
	return ctx.Game.MillToZoneKeyedForEffect(player, n, game.ZoneGraveyard, game.MillUntil{}, step)
}

// exileTopEachPlayerInTurn is "exile the top card of each of these players'
// libraries, and <grant>": template names the grant (its From is overwritten
// with each player in turn), and `done`, a registered tail, is told every
// card that reached exile, library by library in APNAP order. Face up and
// unbudgeted only: the face-down and budgeted forms are refused.
//
// The chain is data (#4253): each library's exile is the keyed impulse
// exile (Game.ExileTopWithPermissionKeyedForEffect), whose tail is the next
// step, carrying the players still to go and the cards exiled so far. A CR
// 616.1 prompt any library's exile pauses on is a restore point.
func exileTopEachPlayerInTurn(ctx *Context, players []uuid.UUID, template ExileTopWithPermission, done game.KeyedTail) error {
	if template.FaceDown || template.TotalManaValueAtMost != nil || template.EachManaValueAtMost != nil {
		return fmt.Errorf("exileTopEachPlayerInTurn: the face-down and budgeted forms have no continuation")
	}
	n := template.N
	if n <= 0 {
		n = 1
	}
	perm := template.permission()
	step := game.KeyedTail{Ref: exileTopEachStep, Source: ctx.Source(), N: n, Permission: &perm}
	if !template.GrantToOwner {
		// uuid.Nil is "each library's owner".
		step.Player = template.GrantTo
	}
	if !done.IsZero() {
		next := done
		step.Next = &next
	}
	return exileTopEachFrom(ctx.Game, step, apnapOrder(ctx.Game, players), nil)
}

// exileTopEachStep is one step of exileTopEachPlayerInTurn's chain, told the
// cards that reached exile from the library just done: its Carry is the
// players still to go and the cards exiled before it (packIDLists), its
// Permission, N and Player the grant, its count and its holder (uuid.Nil
// for each library's owner), and its Next the caller's tail, told every
// card once the last library is done. An on-disk key.
var exileTopEachStep = game.RegisterTail("each-player/exile-top-step", func(g *game.Game, o game.TailOutcome) error {
	lists := unpackIDLists(o.Tail.Carry)
	var remaining, landed []uuid.UUID
	if len(lists) > 0 {
		remaining = lists[0]
	}
	if len(lists) > 1 {
		landed = lists[1]
	}
	landed = append(append([]uuid.UUID(nil), landed...), o.IDs...)
	return exileTopEachFrom(g, o.Tail, remaining, landed)
})

// exileTopEachFrom exiles from the head of `remaining`, the step's tail
// carrying the rest, or hands `landed` to the caller's tail when no library
// is left.
func exileTopEachFrom(g *game.Game, step game.KeyedTail, remaining, landed []uuid.UUID) error {
	if len(remaining) == 0 {
		if step.Next == nil {
			return nil
		}
		return step.Next.Run(g, landed)
	}
	head := remaining[0]
	grantee := step.Player
	if grantee == uuid.Nil {
		grantee = head
	}
	var perm game.CastPermission
	if step.Permission != nil {
		perm = *step.Permission
	}
	step.Carry = packIDLists(remaining[1:], landed)
	return g.ExileTopWithPermissionKeyedForEffect(head, grantee, step.N, perm, step)
}

// constantMill is the N of a mill that names one number for every player.
func constantMill(n int) func(uuid.UUID) int {
	return func(uuid.UUID) int { return n }
}
