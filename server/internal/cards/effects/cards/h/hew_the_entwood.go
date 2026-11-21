package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hew the Entwood — Sorcery {3}{R}{R} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"Sacrifice any number of lands. Reveal the top X cards of your library,
//	 where X is the number of lands sacrificed this way. Choose any number of
//	 artifact and/or land cards revealed this way. Put all nonland cards chosen
//	 this way onto the battlefield, then put all land cards chosen this way onto
//	 the battlefield tapped, then put the rest on the bottom of your library in
//	 a random order."
//
// Pitiless Carnage's "any number" sacrifice (one simultaneous exit; X is how
// many left), then the reveal. The choice is one prompt over the revealed cards
// whose set rule admits only artifact and land cards (any number, none
// included). The three puts are then three steps chained as data: the chosen
// nonland (artifact) cards enter together, then the chosen land cards enter
// together tapped (an artifact land is a land card), then everything not chosen,
// revealed cards of any kind included, goes on the bottom in a random order. A card whose
// entry a replacement stops stays in the library and is not "the rest". With
// nothing sacrificed nothing is revealed (CR 701.20, X = 0).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8818bbf8-58ec-4070-a81a-c7ce248057ab",
		Name:         "Hew the Entwood",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ChoosePermanents{
				Question: "Hew the Entwood — sacrifice any number of lands",
				Keyed:    hewTheEntwoodSacrificed,
				Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
					var out []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == of && c.IsLand() {
							out = append(out, c.InstanceID)
						}
					}
					return out, 0, 0
				},
			}.Apply(ctx)
		},
	})
}

var hewTheEntwoodSacrificed = effects.RegisterPermanentsPick("prompt-run/hew-the-entwood-sacrifice", effects.PermanentsDef{
	Sacrifice: true,
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		n := picked.Count()
		if n <= 0 {
			return nil
		}
		var revealed []uuid.UUID
		if err := (effects.RevealTopOfLibrary{N: n, Reason: "Hew the Entwood — revealed from the top of the library", Revealed: &revealed}).Apply(ctx); err != nil {
			return err
		}
		if len(revealed) == 0 {
			return nil
		}
		ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    ctx.Controller(),
			FromPlayer: ctx.Controller(),
			Source:     ctx.Source(),
			Question:   "Hew the Entwood — choose any number of artifact and/or land cards revealed this way",
			Cards:      revealed,
			Min:        0,
			Max:        len(revealed),
			LiveZone:   game.ZoneLibrary,
			ThenKey:    hewTheEntwoodChosen,
			Asked:      game.NewAskedItem(ctx.Item),
		})
		return nil
	},
})

// hewTheEntwoodChosen is the pick over the revealed cards: only artifact and
// land cards may be chosen. An on-disk key: never renamed, never reused.
var hewTheEntwoodChosen = effects.AskedPickRule("choose-cards/hew-the-entwood-choose",
	func(picked []game.Card) bool {
		for _, c := range picked {
			if !c.IsArtifact() && !c.IsLand() {
				return false
			}
		}
		return true
	},
	func(ctx *effects.Context, r game.CardsChosen) error {
		chosen := map[uuid.UUID]bool{}
		var nonlands, lands, rest []uuid.UUID
		for _, id := range r.Picked {
			chosen[id] = true
			if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsLand() {
				lands = append(lands, id)
			} else {
				nonlands = append(nonlands, id)
			}
		}
		for _, id := range r.Candidates {
			if !chosen[id] {
				rest = append(rest, id)
			}
		}
		if len(nonlands) == 0 {
			return hewTheEntwoodPutLands(ctx, lands, rest)
		}
		carry := append(append([]uuid.UUID(nil), lands...), rest...)
		next := game.KeyedTail{Ref: hewTheEntwoodLandsTail, Player: ctx.Controller(), Source: ctx.Source(), Carry: carry, N: len(lands)}
		return ctx.Game.PutCardsFromLibraryOntoBattlefieldKeyedForEffect(nonlands, game.LibraryEntryOptions{Controller: ctx.Controller()}, next)
	})

// hewTheEntwoodLandsTail is the step after the nonland cards have entered: the
// chosen land cards, then the rest. The carry is the lands followed by the rest,
// N the number of lands. An on-disk key: never renamed, never reused.
var hewTheEntwoodLandsTail = effects.CardTail("then/hew-the-entwood-lands", func(ctx *effects.Context, o game.TailOutcome) error {
	n := min(o.Tail.N, len(o.Tail.Carry))
	return hewTheEntwoodPutLands(ctx, o.Tail.Carry[:n], o.Tail.Carry[n:])
})

// hewTheEntwoodPutLands puts the chosen land cards onto the battlefield tapped,
// together, and then the rest on the bottom.
func hewTheEntwoodPutLands(ctx *effects.Context, lands, rest []uuid.UUID) error {
	if len(lands) == 0 {
		return hewTheEntwoodPutRest(ctx, rest)
	}
	next := game.KeyedTail{Ref: hewTheEntwoodRestTail, Player: ctx.Controller(), Source: ctx.Source(), Carry: rest}
	return ctx.Game.PutCardsFromLibraryOntoBattlefieldKeyedForEffect(lands, game.LibraryEntryOptions{Controller: ctx.Controller(), Tapped: true}, next)
}

// hewTheEntwoodRestTail is the last step. An on-disk key: never renamed, never
// reused.
var hewTheEntwoodRestTail = effects.CardTail("then/hew-the-entwood-rest", func(ctx *effects.Context, o game.TailOutcome) error {
	return hewTheEntwoodPutRest(ctx, o.Tail.Carry)
})

// hewTheEntwoodPutRest puts the revealed cards that were not chosen on the bottom
// of the library in a random order. A card that has left the library is skipped.
func hewTheEntwoodPutRest(ctx *effects.Context, rest []uuid.UUID) error {
	var inLibrary []uuid.UUID
	for _, id := range rest {
		if z := ctx.Game.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneLibrary {
			inLibrary = append(inLibrary, id)
		}
	}
	if len(inLibrary) == 0 {
		return nil
	}
	return ctx.Game.PutOnBottomInRandomOrderForEffect(ctx.Controller(), game.ZoneLibrary, inLibrary)
}
