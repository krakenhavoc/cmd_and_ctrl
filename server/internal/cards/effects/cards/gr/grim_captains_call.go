package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grim Captain's Call — Sorcery {2}{B} (Ixalan, S82 older sets):
//
//	"Return a Pirate card from your graveyard to your hand, then do the same
//	 for Vampire, Dinosaur, and Merfolk."
//
// Four steps in printed order (CR 608.2c), Pirate, Vampire, Dinosaur, Merfolk.
// Each asks for one card of that creature type in the graveyard right now
// (Queen Kayla bin-Kroog's one-prompt-per-step pattern) and returns it to the
// hand before the next step looks, so a card that is two of the types can be
// taken only once. A step with no candidate is skipped; every pick is
// mandatory when there is one. Nothing targets.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "94d3de8c-c8a2-4775-9096-d518f227f09a",
		Name:         "Grim Captain's Call",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return grimCaptainsCallAsk(ctx, 0)
		},
	})
}

var grimCaptainsCallTypes = []string{"Pirate", "Vampire", "Dinosaur", "Merfolk"}

// grimCaptainsCallAsk asks for the card of the type at index step, skipping
// types with no card in the graveyard.
func grimCaptainsCallAsk(ctx *effects.Context, step int) error {
	g, you := ctx.Game, ctx.Controller()
	for ; step < len(grimCaptainsCallTypes); step++ {
		p := g.PlayerByIDForEffect(you)
		if p == nil || p.Graveyard == nil {
			return nil
		}
		var ids []uuid.UUID
		for _, c := range p.Graveyard.Cards {
			if c.IsToken() {
				continue
			}
			if lc, ok := g.LookupCardForEffect(c.InstanceID); ok && lc.HasSubtype(grimCaptainsCallTypes[step]) {
				ids = append(ids, c.InstanceID)
			}
		}
		if len(ids) == 0 {
			continue
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    you,
			FromPlayer: you,
			Source:     ctx.Source(),
			Question:   "Grim Captain's Call — return a " + grimCaptainsCallTypes[step] + " card from your graveyard to your hand",
			Cards:      ids,
			Min:        1,
			Max:        1,
			LiveZone:   game.ZoneGraveyard,
			ThenKey:    grimCaptainsCallPicked,
			CarryNums:  []int{step},
			Asked:      game.NewAskedItem(ctx.Item),
		})
		return nil
	}
	return nil
}

// grimCaptainsCallPicked returns the pick to the hand, then asks for the next
// type. An on-disk key.
var grimCaptainsCallPicked game.ChooseCardsThen

func init() {
	grimCaptainsCallPicked = effects.AskedPick("choose-cards/grim-captains-call-type", func(ctx *effects.Context, r game.CardsChosen) error {
		for _, id := range r.Picked {
			if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
				return err
			}
		}
		step := 0
		if len(r.Nums) > 0 {
			step = r.Nums[0]
		}
		return grimCaptainsCallAsk(ctx, step+1)
	})
}
