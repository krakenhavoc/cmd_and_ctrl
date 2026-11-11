package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Foreboding Steamboat — Artifact — Vehicle {3}{B}{B}, 5/7 (Murders at Karlov
// Manor Commander):
//
//	"When this Vehicle enters, each player chooses two nontoken, non-Vehicle
//	 creatures they control. Exile them until this Vehicle leaves the
//	 battlefield.
//	 Whenever this Vehicle attacks, put a card exiled with it into its
//	 owner's graveyard. If you do, investigate.
//	 Crew 2"
//
// The enters trigger asks each player in turn order from the Steamboat's
// controller for exactly two of their own nontoken, non-Vehicle creatures
// (all of them when they control fewer than two), the way The Legend of
// Yangchen's chapter chains its picks, and only then exiles everything
// chosen, each through ExileUntil{ThisLeaves} (CR 610.3) with the
// CR 607.2a link stamped. A Steamboat gone before the trigger resolves
// exiles nothing (CR 610.3b). The attack trigger lists the cards in exile
// linked to this object; its controller picks one when there are several,
// the shared cmdr2S32PutExiledCardIntoGraveyard moves it, and the Clue
// follows only if a card moved. A card put into a graveyard this way does
// not return when the Steamboat leaves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d97a0e41-9dd7-4466-afb7-3d8bd6cd0eed",
		Name:         "Foreboding Steamboat",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Foreboding Steamboat — each player chooses two nontoken, non-Vehicle creatures they control; exile them until this Vehicle leaves the battlefield",
				forebodingSteamboatEnters),
			WheneverThisAttacks("Foreboding Steamboat — put a card exiled with it into its owner's graveyard; if you do, investigate",
				forebodingSteamboatAttacks),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Foreboding Steamboat"),
		}},
	})
}

func forebodingSteamboatEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return forebodingSteamboatChoose(ctx, seatsStartingWith(g, ctx.Controller()), 0, nil)
}

// forebodingSteamboatChoose asks players[i] for two of their creatures, then
// the next player; after the last it exiles the lot.
func forebodingSteamboatChoose(ctx *Context, players []uuid.UUID, i int, chosen []uuid.UUID) error {
	if i >= len(players) {
		for _, id := range chosen {
			if err := (ExileUntil{
				Target:     id,
				ThisLeaves: true,
				Linked:     true,
				Label:      "Foreboding Steamboat — returns when Foreboding Steamboat leaves the battlefield",
			}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}
	taken := append([]uuid.UUID(nil), chosen...)
	return ChoosePermanents{
		Player:   players[i],
		Of:       []uuid.UUID{players[i]},
		Question: "Foreboding Steamboat — choose two nontoken, non-Vehicle creatures you control to exile until it leaves the battlefield",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var pool []uuid.UUID
			for _, c := range g.Battlefield.Cards {
				if c.Controller != of || !c.IsCreature() || IsToken(c) || c.HasSubtype("Vehicle") || slices.Contains(taken, c.InstanceID) {
					continue
				}
				pool = append(pool, c.InstanceID)
			}
			n := min(2, len(pool))
			return pool, n, n
		},
		Keyed: forebodingSteamboatChosen,
		Carry: packIDLists(players, taken),
		Nums:  []int{i},
	}.Apply(ctx)
}

var forebodingSteamboatChosen PermanentsThen

func init() {
	forebodingSteamboatChosen = RegisterPermanentsPick("prompt-run/foreboding-steamboat-exile", PermanentsDef{
		ThenRun: func(ctx *Context, picked game.PromptedPicks, r game.RunDone) error {
			players, taken := unpackIDPair(r.Carry)
			i := 0
			if len(r.Nums) > 0 {
				i = r.Nums[0]
			}
			return forebodingSteamboatChoose(ctx, players, i+1, append(taken, picked.Cards()...))
		},
	})
}

func forebodingSteamboatAttacks(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	src, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var cands []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(src)) {
		cands = append(cands, c.InstanceID)
	}
	switch len(cands) {
	case 0:
		return nil
	case 1:
		return forebodingSteamboatBin(ctx, cands[0])
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Foreboding Steamboat — put a card exiled with it into its owner's graveyard",
		Cards:    cands,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneExile,
		ThenKey:  forebodingSteamboatPick,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

var forebodingSteamboatPick = askedPick("choose-cards/foreboding-steamboat-grave", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return forebodingSteamboatBin(ctx, r.Picked[0])
})

// forebodingSteamboatBin puts the card into its owner's graveyard, then
// investigates if it moved.
func forebodingSteamboatBin(ctx *Context, id uuid.UUID) error {
	moved, err := cmdr2S32PutExiledCardIntoGraveyard(ctx.Game, id)
	if err != nil || !moved {
		return err
	}
	return Investigate{N: 1}.Apply(ctx)
}
