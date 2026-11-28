package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Planar Overlay — Sorcery {2}{U} (Planeshift, S82 OLD b643, #5615):
//
//	"Each player chooses a land they control of each basic land type. Return those
//	 lands to their owners' hands."
//
// Printed as one choice per basic land type, so it is asked as one: each player in
// turn order from the active player (CR 101.4) names a land of each basic land type
// they control, one prompt per type in the order Plains, Island, Swamp, Mountain,
// Forest, each seeing the earlier picks. A land with two types (a dual) may be
// named for both, and is returned once. A type no land of theirs has is skipped
// (nothing to choose, CR 609.3). The picks are carried as data on the run, and only
// when the last is in are all the chosen lands returned to their owners' hands
// together, so nothing leaves before everyone has chosen.
//
// No simplification.
func init() {
	planarOverlayChosen = effects.RegisterPermanentsPick("prompt-run/planar-overlay-return", effects.PermanentsDef{
		ThenRun: func(ctx *effects.Context, picked game.PromptedPicks, r game.RunDone) error {
			carry := append(append([]uuid.UUID(nil), r.Carry...), picked.Cards()...)
			seat, typ := 0, 0
			if len(r.Nums) >= 2 {
				seat, typ = r.Nums[0], r.Nums[1]+1
			}
			return planarOverlayNext(ctx, carry, seat, typ)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "a7be1624-04cc-4ed5-8b05-63b633893dee",
		Name:         "Planar Overlay",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return planarOverlayNext(ctx, nil, 0, 0)
		},
	})
}

// planarOverlayChosen is the keyed pick every prompt of the card is asked with.
var planarOverlayChosen effects.PermanentsThen

// planarOverlayLands lists the lands `of` controls that have the basic land type.
func planarOverlayLands(cards []game.Card, of uuid.UUID, landType string) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range cards {
		if c.Controller == of && c.IsLand() && c.HasSubtype(landType) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// planarOverlayNext asks the next (seat, type) that has a land to choose from, or,
// once every pick is in, returns the chosen lands to their owners' hands.
func planarOverlayNext(ctx *effects.Context, carry []uuid.UUID, seat, typ int) error {
	seats := effects.SeatsFromActive(ctx.Game)
	for ; seat < len(seats); seat, typ = seat+1, 0 {
		for ; typ < len(game.BasicLandTypes); typ++ {
			landType, of := game.BasicLandTypes[typ], seats[seat]
			if len(planarOverlayLands(ctx.Game.BattlefieldCardsForEffect(), of, landType)) == 0 {
				continue
			}
			return effects.ChoosePermanents{
				Player:   of,
				Of:       []uuid.UUID{of},
				Question: "Planar Overlay — choose a " + landType + " you control to return to its owner's hand",
				Candidates: func(g *game.Game, who uuid.UUID) ([]uuid.UUID, int, int) {
					return planarOverlayLands(g.Battlefield.Cards, who, landType), 1, 1
				},
				Keyed: planarOverlayChosen,
				Carry: carry,
				Nums:  []int{seat, typ},
			}.Apply(ctx)
		}
	}
	seen := map[uuid.UUID]bool{}
	for _, id := range carry {
		if seen[id] || !effects.OnBattlefield(ctx.Game, id) {
			continue
		}
		seen[id] = true
		if err := (effects.BounceToHand{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
