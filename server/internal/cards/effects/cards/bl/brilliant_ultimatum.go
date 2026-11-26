package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brilliant Ultimatum — Sorcery {W}{W}{U}{U}{U}{B}{B} (Shards of Alara, S82 OLD
// batch 518, #5588):
//
//	"Exile the top five cards of your library. An opponent separates those
//	 cards into two piles. You may play lands and cast spells from one of those
//	 piles. If you cast a spell this way, you cast it without paying its mana
//	 cost."
//
// Boneyard Parley's flow over the exiled cards: the top five go to exile face up
// (so the piles are public), the caster names the opponent who separates them
// (either pile may be empty), and the caster takes one pile. The other pile
// stays exiled. A land in the chosen pile is offered to be played now, as part
// of the resolution, with a land drop (CR 305.2a, hideaway's offer; with no drop
// left nothing is asked and the land stays exiled); a spell in it is handed the
// {0} cast permission over the exiled card, castable at any speed until end of
// turn (ADR 0066's posture on every "you may cast it" a resolution offers, the
// window hideaway's spell uses).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a9d58cad-c1c2-4bd2-80f2-da9ce60801df",
		Name:         "Brilliant Ultimatum",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.ExileTopKeyedForEffect(ctx.Controller(), ctx.Controller(), 5, ctx.ItemTail(brilliantUltimatumExiled))
		},
	})
}

// brilliantUltimatumExiled asks for the opponent once the five cards are in
// exile. An on-disk key: never renamed, never reused.
var brilliantUltimatumExiled = effects.CardTail("then/s82-old-b518-brilliant-ultimatum-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	if len(ctx.Opponents()) == 0 {
		// Nobody left to separate them (CR 800.4a): one pile, taken.
		return brilliantUltimatumOffer(ctx, o.IDs)
	}
	return effects.ChoosePlayer{
		Among:    effects.Opponents,
		Question: "Brilliant Ultimatum — choose an opponent to separate these cards into two piles",
		ThenKey:  brilliantUltimatumChosen,
		Carry:    append([]uuid.UUID(nil), o.IDs...),
	}.Apply(ctx)
})

// brilliantUltimatumChosen starts the split with the chosen opponent as the
// splitter. An on-disk key: never renamed, never reused.
var brilliantUltimatumChosen = effects.ChoosePlayerThen("option-pick/s82-old-b518-brilliant-ultimatum-chosen", func(ctx *effects.Context, chosen uuid.UUID, carry []uuid.UUID) error {
	if len(carry) == 0 {
		return nil
	}
	if chosen == uuid.Nil {
		return brilliantUltimatumOffer(ctx, carry)
	}
	return effects.PileSplit{
		Splitter:      chosen,
		Chooser:       ctx.Controller(),
		Owner:         ctx.Controller(),
		SplitQuestion: "Brilliant Ultimatum — separate the cards into two piles",
		PickQuestion:  "Brilliant Ultimatum — choose a pile: you may play its lands and cast its spells without paying their mana costs",
		Cards:         carry,
		ThenTail:      ctx.ItemTail(brilliantUltimatumDealt),
	}.Apply(ctx)
})

// brilliantUltimatumDealt takes the chosen pile; the other stays in exile. An
// on-disk key: never renamed, never reused.
var brilliantUltimatumDealt = effects.PileTail("pile/s82-old-b518-brilliant-ultimatum-dealt", func(ctx *effects.Context, taken, _ []uuid.UUID, _ game.TailOutcome) error {
	return brilliantUltimatumOffer(ctx, taken)
})

// brilliantUltimatumOffer plays and casts from the taken pile: each land is
// offered now, each spell is granted a free cast until end of turn.
func brilliantUltimatumOffer(ctx *effects.Context, pile []uuid.UUID) error {
	player := ctx.Controller()
	for _, id := range pile {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || ctx.Game.FindCardZoneForEffect(id) == nil || ctx.Game.FindCardZoneForEffect(id).Kind != game.ZoneExile {
			continue
		}
		if c.IsLand() {
			if !ctx.Game.CanPlayLandDuringResolutionForEffect(player, id) {
				continue
			}
			if err := ctx.Game.QueueMayCastKeyedForEffect(game.MayCastPrompt{
				Chooser:  player,
				Source:   ctx.Source(),
				Card:     id,
				Question: "Brilliant Ultimatum — play " + c.Name + "?",
			}, brilliantUltimatumPlayLand, nil); err != nil {
				return err
			}
			continue
		}
		ctx.Game.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
			Player:   player,
			Zone:     game.ZoneExile,
			Cost:     "{0}",
			Timing:   game.TimingFlash,
			CastOnly: true,
			Source:   ctx.Source(),
			Label:    "Brilliant Ultimatum — cast " + c.Name + " without paying its mana cost",
		})
	}
	return nil
}

// brilliantUltimatumPlayLand is the land's offer: accepting plays it as part of
// the resolution if a land drop is still left. An on-disk key: never renamed,
// never reused.
var brilliantUltimatumPlayLand = game.RegisterMayCastThen("may-cast/s82-old-b518-brilliant-ultimatum-play-land", func(g *game.Game, r game.MayCastAnswer) error {
	if !r.Accepted || !g.CanPlayLandDuringResolutionForEffect(r.Chooser, r.Card) {
		return nil
	}
	_, err := g.PlayLandDuringResolutionForEffect(r.Chooser, r.Card)
	return err
})
