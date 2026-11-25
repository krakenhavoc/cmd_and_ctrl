package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tayam, Luminous Enigma — {1}{W}{B}{G} Legendary Creature — Nightmare Beast
// 3/3 (Commander 2020):
//
//	"Each other creature you control enters with an additional vigilance
//	 counter on it.
//	 {3}, Remove three counters from among creatures you control: Mill three
//	 cards, then return a permanent card with mana value 3 or less from your
//	 graveyard to the battlefield."
//
// The first ability is Tromell's entry replacement with a keyword counter
// (CR 122.1b) and without the token exclusion: a token enters with it too.
// The cost is Tekuthal's any-kind "from among" removal, over creatures
// (Tayam's own counters included). The effect mills, then asks for a
// permanent card with mana value 3 or less from the graveyard, the milled
// cards among them; the pick is mandatory and a lone candidate is taken.
func init() {
	tayamPick = effects.AskedPick("choose-cards/tayam-return-permanent", func(ctx *effects.Context, r game.CardsChosen) error {
		if len(r.Picked) == 0 {
			return nil
		}
		return effects.ReturnFromGraveyard{Target: r.Picked[0], Dest: game.ZoneBattlefield}.Apply(ctx)
	})
	tayamAfterMill = effects.CardTail("tail/tayam-after-mill", func(ctx *effects.Context, _ game.TailOutcome) error {
		p := ctx.PlayerByID(ctx.Controller())
		if p == nil || p.Graveyard == nil {
			return nil
		}
		var cands []uuid.UUID
		for _, c := range p.Graveyard.Cards {
			if c.IsPermanent() && !c.IsToken() && c.ManaValue() <= 3 {
				cands = append(cands, c.InstanceID)
			}
		}
		switch len(cands) {
		case 0:
			return nil
		case 1:
			return effects.ReturnFromGraveyard{Target: cands[0], Dest: game.ZoneBattlefield}.Apply(ctx)
		}
		ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  ctx.Controller(),
			Source:   ctx.Source(),
			Question: "Tayam — return a permanent card with mana value 3 or less from your graveyard to the battlefield",
			Cards:    cands,
			Min:      1,
			Max:      1,
			LiveZone: game.ZoneGraveyard,
			ThenKey:  tayamPick,
			Asked:    game.NewAskedItem(ctx.Item),
		})
		return nil
	})
	effects.Register(effects.Spec{
		OracleID:     "84be8354-0bfa-4df1-868a-13197bd49191",
		Name:         "Tayam, Luminous Enigma",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if src == nil || ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || ev.CardID == src.InstanceID {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && c.Controller == src.Controller
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.AddCounterAtETB(game.CounterVigilance, 1)
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Tayam, Luminous Enigma — each other creature you control enters with an additional vigilance counter",
		}},
		Activated: []effects.ActivatedAbility{{
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Label:   "{3}, Remove three counters from among creatures you control: Mill three cards, then return a permanent card with mana value 3 or less from your graveyard to the battlefield.",
			Cost: effects.Plus(effects.ManaCost("{3}"),
				effects.RemoveCountersAmong("", 3, "creatures you control", effects.Creature())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.MillToZone{N: 3, ThenTail: ctx.ItemTail(tayamAfterMill)}.Apply(ctx)
			},
		}},
	})
}

var (
	tayamPick      game.ChooseCardsThen
	tayamAfterMill game.TailRef
)
