package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sothera, the Supervoid — Legendary Enchantment {2}{B}{B} (Edge of
// Eternities):
//
//	"Whenever a creature you control dies, each opponent chooses a
//	 creature they control and exiles it.
//	 At the beginning of your end step, if a player controls no creatures,
//	 sacrifice Sothera, then put a creature card exiled with it onto the
//	 battlefield under your control with two additional +1/+1 counters on
//	 it."
//
// The seam is an "each opponent chooses a creature and exiles it" edict
// whose exile is linked (CR 607.2a: Card.ExiledWith names this Sothera).
// The opponents are asked in turn order, one chooseCards prompt each
// (Braids' chain: the opponents still to be asked ride the prompt's Carry),
// so every opponent chooses from the creatures they control as their turn
// of the chain comes; one with a single creature is not asked, one with
// none exiles nothing. The creature goes to exile linked to the Sothera
// that triggered, and a Sothera that left and came back claims nothing
// (CR 400.7). The first ability triggers for every creature that dies with
// Sothera (the rulings).
//
// The end-step trigger has a CR 603.4 intervening "if" over the whole
// table, checked as it triggers and again as it resolves. The sacrifice
// and the put are one sentence: the put happens even if the Sothera has
// already left, choosing among the creature cards exiled with it (read by
// the object, so a Sothera that has gone still names them). With several
// the controller chooses; the counters are on top of any it enters with.
//
// No simplification.
func init() {
	dies := effects.WheneverACreatureYouControlDies("Sothera, the Supervoid — each opponent chooses a creature they control and exiles it", sotheraEdict)
	noCreatures := effects.CondBoard{Name: "a player controls no creatures", Test: func(ctx *effects.Context) bool { return sotheraSomeoneHasNoCreatures(ctx.Game) }}
	endStep := effects.InterveningIf(noCreatures,
		effects.On(game.EventBeginEndStep, effects.AllOf(effects.ByYou, func(_ game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
			return sotheraSomeoneHasNoCreatures(g)
		}), "Sothera, the Supervoid — sacrifice it, then put a creature card exiled with it onto the battlefield with two additional +1/+1 counters", sotheraReanimate))
	effects.Register(effects.Spec{
		OracleID:     "760ede92-a826-4144-9de8-84ba9b8743d1",
		Name:         "Sothera, the Supervoid",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{dies, endStep},
	})
}

// sotheraSomeoneHasNoCreatures is "a player controls no creatures": any
// player still in the game, the Sothera's controller included.
func sotheraSomeoneHasNoCreatures(g *game.Game) bool {
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		if effects.CountControlled(g, p.ID, func(c game.Card) bool { return c.IsCreature() }) == 0 {
			return true
		}
	}
	return false
}

func sotheraEdict(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	return sotheraAsk(ctx, effects.OpponentsInTurnOrder(g, item.Controller))
}

// sotheraAsk walks the opponents still to choose. The first one with a
// real choice is asked; the rest ride its prompt.
func sotheraAsk(ctx *effects.Context, remaining []uuid.UUID) error {
	g := ctx.Game
	for len(remaining) > 0 {
		opp, rest := remaining[0], remaining[1:]
		remaining = rest
		if p := g.PlayerByIDForEffect(opp); p == nil || p.Eliminated {
			continue
		}
		var cands []uuid.UUID
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == opp && c.IsCreature() {
				cands = append(cands, c.InstanceID)
			}
		}
		switch len(cands) {
		case 0:
			continue
		case 1:
			if err := (effects.ExileTarget{Target: cands[0], Linked: true}).Apply(ctx); err != nil {
				return err
			}
			continue
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  opp,
			Source:   ctx.Source(),
			Question: "Sothera, the Supervoid — choose a creature you control to exile",
			Cards:    cands,
			Min:      1,
			Max:      1,
			Zone:     game.ZoneBattlefield,
			LiveZone: game.ZoneBattlefield,
			ThenKey:  sotheraExile,
			Carry:    append([]uuid.UUID(nil), rest...),
			Asked:    game.NewAskedItem(ctx.Item),
		})
		return nil
	}
	return nil
}

// sotheraExile is one opponent's choice, then the next opponent. Carry is
// the opponents still to be asked. An on-disk key: never renamed, never
// reused.
var sotheraExile game.ChooseCardsThen

func init() {
	sotheraExile = effects.AskedPick("choose-cards/sothera-exile-creature", sotheraExilePicked)
}

func sotheraExilePicked(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 1 {
		if c, ok := ctx.Game.LookupCardForEffect(r.Picked[0]); ok && c.Controller == r.Chooser && c.IsCreature() && effects.OnBattlefield(ctx.Game, r.Picked[0]) {
			if err := (effects.ExileTarget{Target: r.Picked[0], Linked: true}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return sotheraAsk(ctx, append([]uuid.UUID(nil), r.Carry...))
}

func sotheraReanimate(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	return effects.SacrificePermanent{Target: ctx.Source(), ThenTail: ctx.ItemTail(sotheraPut)}.Apply(ctx)
}

// sotheraPut is "then put a creature card exiled with it onto the
// battlefield", whether or not the sacrifice happened. An on-disk key:
// never renamed, never reused.
var sotheraPut = effects.CardTail("sothera-the-supervoid/put-exiled-creature", func(ctx *effects.Context, _ game.TailOutcome) error {
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range ctx.Game.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		if c.IsCreature() && !c.IsToken() {
			ids = append(ids, c.InstanceID)
		}
	}
	switch len(ids) {
	case 0:
		return nil
	case 1:
		return sotheraEnter(ctx, ids[0])
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Sothera, the Supervoid — choose a creature card exiled with it to put onto the battlefield",
		Cards:    ids,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneExile,
		ThenKey:  sotheraChoose,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
})

var sotheraChoose = effects.AskedPick("choose-cards/sothera-put-exiled-creature", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) != 1 {
		return nil
	}
	return sotheraEnter(ctx, r.Picked[0])
})

// sotheraEnter puts the chosen card onto the battlefield under the
// ability's controller with two +1/+1 counters.
func sotheraEnter(ctx *effects.Context, id uuid.UUID) error {
	return effects.PutCardsFromZoneTogetherWith(ctx, game.ZoneExile, []uuid.UUID{id}, game.ZoneEntryOptions{
		Controller:         ctx.Controller(),
		EntersWithCounters: map[string]int{game.CounterPlusOne: 2},
	}, game.TailRef{}, 0)
}
