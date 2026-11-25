package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Verdant Mastery — Sorcery {5}{G}:
//
//	"You may pay {3}{G} rather than pay this spell's mana cost.
//	 Search your library for up to four basic land cards and reveal them.
//	 Put one of them onto the battlefield tapped under an opponent's
//	 control if the {3}{G} cost was paid. Put two of them onto the
//	 battlefield tapped under your control and the rest into your hand.
//	 Then shuffle."
//
// The Mastery cycle's alternative cost (Baleful Mastery's bare
// game.AlternativeCost) over a search whose found cards wait, revealed, in
// the library while the caster splits them (#6156, HoldFound):
//
//   - With the {3}{G} cost paid, the caster picks one of the found lands
//     and then which opponent gets it (ChoosePlayer: "an opponent" names
//     no seat), and it enters tapped under that player's control (CR
//     110.2a's "under another player's control").
//   - Of the lands left, the caster picks two to enter tapped under their
//     own control, and the rest go to their hand (a put, not a draw).
//   - Then the library is shuffled.
//
// Every pick is clamped to what is left (CR 608.2's "as much as
// possible"), so a search that finds one land with the cost paid hands it
// to the opponent. No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     VerdantMasteryOracleID,
		Name:         "Verdant Mastery",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 2},
		AlternativeCosts: []game.AlternativeCost{
			{
				Key:      VerdantMasteryAltCost,
				Label:    "Pay {3}{G} rather than pay this spell's mana cost",
				ManaCost: "{3}{G}",
			},
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.SearchLibraryThenForEffect(game.SearchLibrarySpec{
				Player:    ctx.Controller(),
				Source:    ctx.Source(),
				Pred:      effects.IsBasicLand,
				Dest:      game.ZoneLibrary,
				HoldFound: true,
				Limit:     4,
				Reveal:    true,
				Reason:    "Verdant Mastery — search your library for up to four basic land cards",
				// N is 1 when the {3}{G} cost was paid, read as the spell
				// resolves.
				ThenTail: effects.WithN(ctx.ItemTail(verdantMasteryFoundTail), verdantMasteryPaidN(ctx)),
			})
		},
	})
}

const (
	// VerdantMasteryOracleID is the card's oracle_id.
	VerdantMasteryOracleID = "b392fabc-51f2-4ae4-b21c-1c662e4c6953"
	// VerdantMasteryAltCost is the {3}{G} alternative cost's key.
	VerdantMasteryAltCost = "verdant-mastery"
)

// verdantMasteryPaidN is 1 when the spell was cast for {3}{G}, else 0.
func verdantMasteryPaidN(ctx *effects.Context) int {
	if ctx.PaidAltCost(VerdantMasteryAltCost) {
		return 1
	}
	return 0
}

// verdantMasteryFoundTail splits the found lands. An on-disk key.
var verdantMasteryFoundTail = effects.CardTail("verdant-mastery/found", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return ctx.Game.ShuffleLibraryForEffect(ctx.Controller())
	}
	if o.Tail.N != 1 {
		return verdantMasteryPickTwo(ctx, o.IDs)
	}
	return effects.RevealPick{
		Player:   ctx.Controller(),
		Owner:    ctx.Controller(),
		Question: "Verdant Mastery — choose the land that enters tapped under an opponent's control",
		Cards:    o.IDs,
		Min:      1,
		Max:      1,
		Keyed:    verdantMasteryOpponentsLand,
	}.Apply(ctx)
})

// verdantMasteryOpponentsLand is the land picked for an opponent: the
// caster then names the opponent. An on-disk key.
var verdantMasteryOpponentsLand = effects.RegisterRevealPick("choose-cards/verdant-mastery-opponents-land", func(ctx *effects.Context, picked, left []uuid.UUID, _ []int) error {
	if len(picked) == 0 {
		return verdantMasteryPickTwo(ctx, left)
	}
	return effects.ChoosePlayer{
		Among:    effects.Opponents,
		Question: "Verdant Mastery — choose an opponent; that land enters tapped under their control",
		ThenKey:  verdantMasteryOpponent,
		Carry:    append(append([]uuid.UUID(nil), picked[0]), left...),
	}.Apply(ctx)
})

// verdantMasteryOpponent puts the first carried land onto the battlefield
// tapped under the chosen opponent's control, then splits the rest. With
// no opponent to name the land stays among the rest. An on-disk key.
var verdantMasteryOpponent = effects.ChoosePlayerThen("option-pick/verdant-mastery-opponent", func(ctx *effects.Context, chosen uuid.UUID, carry []uuid.UUID) error {
	if len(carry) == 0 {
		return ctx.Game.ShuffleLibraryForEffect(ctx.Controller())
	}
	if chosen == uuid.Nil {
		return verdantMasteryPickTwo(ctx, carry)
	}
	tail := ctx.ItemTail(verdantMasteryAfterOpponent)
	tail.Carry = append([]uuid.UUID(nil), carry[1:]...)
	return ctx.Game.PutCardsFromLibraryOntoBattlefieldKeyedForEffect(carry[:1],
		game.LibraryEntryOptions{Controller: chosen, Tapped: true}, tail)
})

// verdantMasteryAfterOpponent is the rest of the split once the
// opponent's land has entered. An on-disk key.
var verdantMasteryAfterOpponent = effects.CardTail("verdant-mastery/after-opponent", func(ctx *effects.Context, o game.TailOutcome) error {
	return verdantMasteryPickTwo(ctx, o.Tail.Carry)
})

// verdantMasteryPickTwo has the caster pick two of the lands left for the
// battlefield.
func verdantMasteryPickTwo(ctx *effects.Context, left []uuid.UUID) error {
	left = effects.LiveCards(ctx.Game, left)
	if len(left) == 0 {
		return ctx.Game.ShuffleLibraryForEffect(ctx.Controller())
	}
	return effects.RevealPick{
		Player:   ctx.Controller(),
		Owner:    ctx.Controller(),
		Question: "Verdant Mastery — choose two lands to put onto the battlefield tapped; the rest go to your hand",
		Cards:    left,
		Min:      2,
		Max:      2,
		Keyed:    verdantMasteryMine,
	}.Apply(ctx)
}

// verdantMasteryMine puts the two picked lands onto the battlefield tapped
// under the caster's control, then the rest into their hand. An on-disk
// key.
var verdantMasteryMine = effects.RegisterRevealPick("choose-cards/verdant-mastery-mine", func(ctx *effects.Context, picked, left []uuid.UUID, _ []int) error {
	tail := ctx.ItemTail(verdantMasteryToHand)
	tail.Carry = append([]uuid.UUID(nil), left...)
	return ctx.Game.PutCardsFromLibraryOntoBattlefieldKeyedForEffect(picked,
		game.LibraryEntryOptions{Controller: ctx.Controller(), Tapped: true}, tail)
})

// verdantMasteryToHand puts the lands left into the caster's hand, then
// shuffles. An on-disk key.
var verdantMasteryToHand = effects.CardTail("verdant-mastery/to-hand", func(ctx *effects.Context, o game.TailOutcome) error {
	return ctx.Game.TakeFromLibraryToHandKeyedForEffect(ctx.Controller(), o.Tail.Carry, ctx.Tail(effects.ThenShuffle))
})
