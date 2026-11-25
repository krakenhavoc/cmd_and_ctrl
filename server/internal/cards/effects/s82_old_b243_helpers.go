package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b243_helpers.go — shared vocabulary of Amonkhet batch 243 (S82 older
// sets, #5535): the Aftermath split cards.

// OldB243EachOpponentChoosesACreatureYouGainControlOfIt is "Each opponent
// chooses a creature they control. You gain control of those creatures"
// (Riches): each opponent with a creature is asked once, in turn order
// (CR 101.4), for one they control; when every answer is in, the caster gains
// control of every chosen creature, with no end (CR 611.2a).
func OldB243EachOpponentChoosesACreatureYouGainControlOfIt(label string) func(item *game.StackItem, ctx *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		g := ctx.Game
		var legs []game.ChooseCardsPrompt
		for _, seat := range ApnapPlayers(g) {
			if seat == ctx.Controller() {
				continue
			}
			p := g.PlayerByIDForEffect(seat)
			if p == nil || p.Eliminated {
				continue
			}
			ids := PermanentsControlledByMatching(g, seat, Creature())
			if len(ids) == 0 {
				continue
			}
			legs = append(legs, game.ChooseCardsPrompt{
				Chooser:    seat,
				FromPlayer: seat,
				Source:     item.SourceCardID,
				Question:   label + " — choose a creature you control; its caster gains control of it",
				Cards:      ids,
				Min:        1,
				Max:        1,
				Zone:       game.ZoneBattlefield,
				ThenKey:    game.ChooseCardsRunLeg,
			})
		}
		if len(legs) == 0 {
			return nil
		}
		_, err := g.ChooseCardsRunThenKeyedForEffect(legs, oldB243TakeChosen, item.SourceCardID, game.RunData{Carry: []uuid.UUID{ctx.Controller()}})
		return err
	}
}

// oldB243TakeChosen gives the caster control of each creature chosen. An
// on-disk key: never renamed, never reused.
var oldB243TakeChosen = RunThen("prompt-run/s82-old-b243-riches-take", func(ctx *Context, r game.RunDone) error {
	for _, id := range r.Picks().Cards() {
		if !OnBattlefield(ctx.Game, id) {
			continue
		}
		if err := (GainControl{Target: id, Controller: ctx.Controller(), Duration: game.IndefiniteDuration(), Label: "Riches — gain control"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

// OldB243ChooseACardNameCastBan is "Choose a card name. Until your next turn,
// your opponents can't cast spells with the chosen name" (Comply): the name is
// asked for as the spell resolves, and each opponent is barred from casting
// spells with it until the caster's next turn begins (CR 611.2b).
func OldB243ChooseACardNameCastBan(label string) func(item *game.StackItem, ctx *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		ctx.Game.QueueCardNameChoiceKeyedForEffect(ctx.Controller(), ctx.Source(),
			label+" — choose a card name", ctx.ItemTail(oldB243NamedBan))
		return nil
	}
}

// oldB243NamedBan is the ban once the name (the tail's Word) is known. An
// on-disk key.
var oldB243NamedBan = CardTail("then/s82-old-b243-comply-named", func(ctx *Context, o game.TailOutcome) error {
	name := o.Tail.Word
	if name == "" {
		return nil
	}
	d := DurationUntilYourNextTurn(ctx, ctx.Controller())
	for _, opp := range ctx.Opponents() {
		if err := (RestrictCasting{
			Player:   opp,
			Rule:     game.CastBanRule{Kind: game.CastBanNamed, Filter: game.PermissionFilter{Named: name}},
			Label:    "Comply — can't cast spells named " + name,
			Duration: d,
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
