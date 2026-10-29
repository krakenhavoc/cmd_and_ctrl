package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_spell_b_helpers.go — shared bodies for the Reality
// Fracture instants and sorceries of slice fra-spell-b (tracker #2795).

// rfSpellBMillThenMayTakePermanent is Something Worth Saving: mill n
// cards, then the controller may put a permanent card from among them
// into their hand; the `then` milled was registered with
// (rfSpellBTakeFromMilledTail) runs once that is settled (also when
// nothing was milled or nothing is eligible).
//
// #4253 slice 10a: the mill's "then" is a keyed tail, the question its
// Word. The pick after it is keyed too (rfSpellBTakeFromMilledTail).
func rfSpellBMillThenMayTakePermanent(ctx *Context, n int, question string, milled game.TailRef) error {
	return ctx.Game.MillToZoneKeyedForEffect(ctx.Controller(), n, game.ZoneGraveyard, game.MillUntil{},
		withWord(ctx.ItemTail(milled), question))
}

// rfSpellBTakeFromMilledTail registers the mill's tail for one card under
// key (an on-disk identity): the pick over what was milled, then `then`.
// Call it once, from a package-level var.
//
// #4253: the pick is keyed too, under "choose-cards/" plus the tail's key
// with its slashes dashed and "-pick" after it (an on-disk key, ledgered):
// the asking item rides the prompt as data and `then` runs against it.
func rfSpellBTakeFromMilledTail(key string, then func(ctx *Context) error) game.TailRef {
	pick := askedPick("choose-cards/"+strings.ReplaceAll(key, "/", "-")+"-pick", func(ctx *Context, r game.CardsChosen) error {
		if len(r.Picked) > 0 {
			if err := (ReturnFromGraveyard{Target: r.Picked[0], Dest: game.ZoneHand}).Apply(ctx); err != nil {
				return err
			}
		}
		return then(ctx)
	})
	return CardTail(key, func(ctx *Context, o game.TailOutcome) error {
		return rfSpellBMayTakeFromMilled(ctx, pick, o.Tail.Word, then, o.IDs)
	})
}

// rfSpellBGreatestPowerYouControl is the greatest power among creatures
// `player` controls, zero with none.
func rfSpellBGreatestPowerYouControl(ctx *Context, player uuid.UUID) int {
	best := 0
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() && c.CurrentPower() > best {
			best = c.CurrentPower()
		}
	}
	return best
}

// rfSpellBCadetWithHaste is "Create a 2/2 colorless Wizard Soldier
// creature token named Cadet. It gains haste until end of turn."
func rfSpellBCadetWithHaste(_ *game.StackItem, ctx *Context, _ int) error {
	made, err := ctx.Game.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(TokenCard("2/2 colorless Wizard Soldier named Cadet")), 1, game.TokenEntryOptions{})
	if err != nil {
		return err
	}
	for _, id := range made {
		if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Cadet gains haste until end of turn"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// rfSpellBCounterNoncreatureUnlessPays is Theorix Charm's first bullet.
func rfSpellBCounterNoncreatureUnlessPays(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	return CounterUnlessPaid{
		StackID:  t.ID,
		Cost:     "{2}",
		Question: "Theorix Charm — pay {2} or your spell is countered",
	}.Apply(ctx)
}

// rfSpellBMayTakeFromMilled offers a permanent card from among the milled
// ones, then runs `then`.
func rfSpellBMayTakeFromMilled(ctx *Context, pick game.ChooseCardsThen, question string, then func(ctx *Context) error, milled []uuid.UUID) error {
	item := ctx.Item
	var eligible []uuid.UUID
	for _, id := range milled {
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsPermanent() {
			eligible = append(eligible, id)
		}
	}
	if len(eligible) == 0 {
		return then(ctx)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: question,
		Cards:    eligible,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneGraveyard,
		ThenKey:  pick,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}
