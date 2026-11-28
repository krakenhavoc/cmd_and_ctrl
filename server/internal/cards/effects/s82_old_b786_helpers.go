package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b786_helpers.go — shared pieces of the Visions cards of batch 786
// (S82 OLD, #5631).

// OldB786EachEndStepBegan is the When of "at the beginning of each end step":
// any player's end step (CR 513.1).
func OldB786EachEndStepBegan(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventBeginEndStep
}

// OldB786PlayersWhoTappedALandForMana lists, in APNAP order from the active
// player, the living players who tapped a land for mana this turn (CR 106.12,
// the #6217 turn record).
func OldB786PlayersWhoTappedALandForMana(g *game.Game) []uuid.UUID {
	var out []uuid.UUID
	n := len(g.Seats)
	for i := 0; i < n; i++ {
		p := g.Seats[g.SeatInTurnOrderForEffect(g.Turn.ActiveSeat, i)]
		if p == nil || p.Eliminated || !g.TappedLandForManaThisTurnForEffect(p.ID) {
			continue
		}
		out = append(out, p.ID)
	}
	return out
}

// OldB786SacrificedAPlains is "the land `id` the player sacrificed was a
// Plains" as it last existed on the battlefield (CR 608.2h): its last-known
// subtypes, so a land that was a Plains only through an effect counts and a
// Plains that had stopped being one does not.
func OldB786SacrificedAPlains(g *game.Game, id uuid.UUID) bool {
	ref, ok := g.PermanentRefForEffect(id)
	if !ok {
		return false
	}
	info, ok := g.PermanentForEffect(ref)
	return ok && slices.Contains(info.Characteristic.Subtypes, "Plains")
}

// OldB786DesolationDamage is the tail of Desolation: 2 damage to each player
// who sacrificed a Plains this way, dealt as one instruction. On disk:
// never renamed.
var OldB786DesolationDamage = RunThen("prompt-run/s82-old-b786-desolation-damage", func(ctx *Context, r game.RunDone) error {
	sacrificed := r.Sacrifices()
	return ctx.Game.DamageInstanceForEffect(func() error {
		for _, e := range sacrificed {
			if !slices.ContainsFunc(e.Cards, func(id uuid.UUID) bool { return OldB786SacrificedAPlains(ctx.Game, id) }) {
				continue
			}
			if err := (DealDamage{Source: ctx.Source(), Target: e.Seat, Amount: 2}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
})

// OldB786DesolationSacrifice is Desolation's end-step body: each player who
// tapped a land for mana this turn sacrifices a land of their choice, all
// together after the last has chosen (CR 101.4), then Plains sacrificers take
// 2 damage.
func OldB786DesolationSacrifice(g *game.Game, item *game.StackItem) error {
	players := OldB786PlayersWhoTappedALandForMana(g)
	if len(players) == 0 {
		return nil
	}
	ctx := NewContext(g, item)
	return ctx.PlayersSacrificeThen(players, SacrificeSpec("a land", Land()),
		"Desolation — sacrifice a land", OldB786DesolationDamage, ctx.RunData(nil))
}
