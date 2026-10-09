package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_spell_b_helpers.go — shared bodies for the Reality
// Fracture instants and sorceries of slice fra-spell-b (tracker #2795).

// rfSpellBMillThenMayTakePermanent is Something Worth Saving: mill n
// cards, then the controller may put a permanent card from among them
// into their hand; `then` runs once that is settled (also when nothing
// was milled or nothing is eligible).
func rfSpellBMillThenMayTakePermanent(ctx *Context, n int, question string, then func(ctx *Context) error) error {
	item := ctx.Item
	return MillToZone{N: n, Then: func(ctx *Context, milled []uuid.UUID) error {
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
			Zone:     game.ZoneGraveyard,
			Then: func(g *game.Game, picked []uuid.UUID) error {
				c := NewContext(g, item)
				if len(picked) > 0 {
					if err := (ReturnFromGraveyard{Target: picked[0], Dest: game.ZoneHand}).Apply(c); err != nil {
						return err
					}
				}
				return then(c)
			},
		})
		return nil
	}}.Apply(ctx)
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
	made, err := ctx.Game.CreateTokensForEffect(ctx.Controller(), TokenCard("2/2 colorless Wizard Soldier named Cadet"), 1, game.TokenEntryOptions{})
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
