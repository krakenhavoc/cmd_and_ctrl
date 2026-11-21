package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Urza's Sylex — Legendary Artifact {3} (The Brothers' War):
//
//	"{2}{W}{W}, {T}, Exile Urza's Sylex: Each player chooses six lands
//	 they control. Destroy all other permanents. Activate only as a
//	 sorcery.
//	 When Urza's Sylex is put into exile from the battlefield, you may
//	 pay {2}. If you do, search your library for a planeswalker card,
//	 reveal it, put it into your hand, then shuffle."
//
// The first ability is Planetary Annihilation's choice (a player with six
// lands or fewer keeps them all and is not asked) followed by a
// destroy-all of everything not kept, so an indestructible permanent
// survives (CR 702.12b). The exile is its cost, so the Sylex itself is
// already gone.
//
// The second is its own trigger on the Sylex's own move from the
// battlefield to exile, whatever caused it, so the cost of the first
// ability triggers it and so does an opponent's Swords to Plowshares.
// The {2} is paid on resolution (MayPay).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fb4d99c2-daec-430a-9ca2-5a178a96618b",
		Name:         "Urza's Sylex",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepDestroy, Partial: true}},
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}{W}{W}, {T}, Exile Urza's Sylex: Each player chooses six lands they control. Destroy all other permanents. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{2}{W}{W}"), effects.TapCost(), effects.ExileThis()),
			SorcerySpeed: true,
			Effect:       urzasSylexResolve,
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, urzasSylexExiled,
				"Urza's Sylex — you may pay {2}; if you do, search your library for a planeswalker card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}",
						Question: "Urza's Sylex — pay {2} to search your library for a planeswalker card?",
						Then:     urzasSylexPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// urzasSylexExiled is "when this is put into exile from the battlefield".
func urzasSylexExiled(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventLTB && ev.CardID == source.InstanceID && ev.NewZone == game.ZoneExile
}

// urzasSylexPaid is the search after the {2}. An on-disk key.
var urzasSylexPaid = effects.PayUnlessBranches("pay-unless/urzas-sylex-search",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.SearchLibrary{
			Player:    ctx.Controller(),
			Predicate: func(c game.Card) bool { return c.IsPlaneswalker() },
			Dest:      game.ZoneHand,
			Limit:     1,
			Reveal:    true,
			Shuffle:   true,
			Reason:    "Urza's Sylex — a planeswalker card",
		}.Apply(ctx)
	}, nil)

// urzasSylexResolve asks each player with more than six lands which six
// they keep, then destroys everything else.
func urzasSylexResolve(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var legs []game.ChooseCardsPrompt
	var asked []uuid.UUID
	for _, seat := range effects.ApnapPlayers(g) {
		lands := effects.PermanentsControlledByMatching(g, seat, effects.Land())
		if len(lands) <= 6 {
			continue
		}
		asked = append(asked, seat)
		legs = append(legs, game.ChooseCardsPrompt{
			Chooser:    seat,
			FromPlayer: seat,
			Source:     item.SourceCardID,
			Question:   "Urza's Sylex — choose six lands you control; all other permanents are destroyed",
			Cards:      lands,
			Min:        6,
			Max:        6,
			Zone:       game.ZoneBattlefield,
			ThenKey:    game.ChooseCardsRunLeg,
		})
	}
	if len(legs) == 0 {
		return urzasSylexDestroy(ctx, nil, nil)
	}
	_, err := g.ChooseCardsRunThenKeyedForEffect(legs, urzasSylexRun, item.SourceCardID, ctx.RunData(asked))
	return err
}

// urzasSylexRun destroys every permanent that is not one of the six
// lands its controller chose. An on-disk key.
var urzasSylexRun = effects.RunThen("prompt-run/urzas-sylex-keep-lands", func(ctx *effects.Context, r game.RunDone) error {
	return urzasSylexDestroy(ctx, r.Picks().Cards(), r.Carry)
})

// urzasSylexDestroy destroys all permanents but the chosen lands, and the
// lands of every player who had six or fewer (nobody chose for them).
func urzasSylexDestroy(ctx *effects.Context, chosen, askedSeats []uuid.UUID) error {
	kept := map[uuid.UUID]bool{}
	for _, id := range chosen {
		kept[id] = true
	}
	asked := map[uuid.UUID]bool{}
	for _, id := range askedSeats {
		asked[id] = true
	}
	return effects.DestroyAllMatching{Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		if kept[c.InstanceID] {
			return false
		}
		return !(c.IsLand() && !asked[c.Controller])
	}}.Apply(ctx)
}
