package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b772_helpers.go — shared shapes of S82 OLD batch 772 (Weatherlight:
// cumulative upkeep, #5629).

// S82OldB772DamageToPlayerOrWalkerAndTheirCreatures is "<this> deals X damage to
// target player or planeswalker and each creature that player or that
// planeswalker's controller controls" (Lavalanche, Bonfire of the Damned, Heart
// of Bogardan) for an X the caller has worked out. The player is the target, or
// the controller of the targeted planeswalker, read before any damage lands; the
// creatures are snapshotted first and every point is one simultaneous damage
// instance (CR 608.2). With its one target illegal the effect does nothing (CR
// 608.2b); an X of 0 or less deals nothing.
func S82OldB772DamageToPlayerOrWalkerAndTheirCreatures(ctx *Context, x int) error {
	targets := ctx.LegalTargets()
	if x <= 0 || len(targets) == 0 {
		return nil
	}
	t := targets[0]
	player := t.ID
	if t.Kind == game.TargetCard {
		c, ok := ctx.Game.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		player = c.Controller
	}
	var creatures []uuid.UUID
	for _, c := range MatchingBattlefield(ctx, Creature()) {
		if c.Controller == player {
			creatures = append(creatures, c.InstanceID)
		}
	}
	return ctx.Game.DamageInstanceForEffect(func() error {
		if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: x}).Apply(ctx); err != nil {
			return err
		}
		for _, id := range creatures {
			if err := (DealDamage{Source: ctx.Source(), Target: id, Amount: x}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

// S82OldB772HeartOfBogardanDamage is Heart of Bogardan's "X damage, where X is
// twice the number of age counters on this enchantment minus 2": the age
// counters the event names, which is what it had as its cumulative upkeep went
// unpaid, before it was sacrificed (CR 608.2h).
func S82OldB772HeartOfBogardanDamage(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	return S82OldB772DamageToPlayerOrWalkerAndTheirCreatures(NewContext(g, item), 2*item.Trigger.Event.Amount-2)
}

// S82OldB772AgeCounters is the number of age counters on the ability's source as
// it resolves, or as it last existed once it has left (CR 608.2h).
func S82OldB772AgeCounters(ctx *Context) int {
	return CountersOnThisAsItResolves(game.CounterAge)(ctx)
}

// S82OldB772GainLifeForAgeCounters is Revered Unicorn's "you gain life equal to
// the number of age counters on it", counted as it last existed on the
// battlefield (CR 603.10a).
func S82OldB772GainLifeForAgeCounters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return GainLife{Amount: S82OldB772AgeCounters(ctx)}.Apply(ctx)
}

// S82OldB772WaveOfTerrorDestroy is "destroy each creature with mana value equal to
// the number of age counters on this enchantment. They can't be regenerated"
// (Wave of Terror), counted as the ability resolves.
func S82OldB772WaveOfTerrorDestroy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !OnBattlefield(g, item.SourceCardID) {
		return nil
	}
	return DestroyAllMatching{
		Match:             And(Creature(), ManaValueIs(S82OldB772AgeCounters(ctx))),
		CantBeRegenerated: true,
	}.Apply(ctx)
}

// S82OldB772AgeCountersPT is "power and toughness are each equal to 1 plus twice
// the number of age counters on it" (Mwonvuli Ooze), a characteristic-defining
// ability (CR 604.3, layer 7a) read as the layers are recomputed.
func S82OldB772AgeCountersPT(label string) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7A_CDA,
		Label:     label,
		AppliesTo: SelfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			n := 1 + 2*source.Counters[game.CounterAge]
			c.Power, c.Toughness = n, n
		},
	}
}

// S82OldB772LookAtTopTenExileAnyNumber is Ancestral Knowledge's "look at the top
// ten cards of your library, then exile any number of them and put the rest back
// on top of your library in any order": a look (only the looker learns them), a
// pick of any number from none to all, the picked cards exiled, and the rest put
// back as an ordering prompt. The cards' order is the pile's, top first.
func S82OldB772LookAtTopTenExileAnyNumber(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	cards := g.LookAtTopOfPlayersLibraryForEffect(ctx.Controller(), ctx.Controller(), 10)
	if len(cards) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Ancestral Knowledge — exile any number of the top ten cards of your library",
		Cards:    cards,
		Min:      0,
		Max:      len(cards),
		LiveZone: game.ZoneLibrary,
		ThenKey:  s82OldB772AncestralKnowledgeExiled,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// s82OldB772AncestralKnowledgeExiled exiles the picked cards and puts the rest
// back on top in any order. An on-disk key: never renamed, never reused.
var s82OldB772AncestralKnowledgeExiled = AskedPick("choose-cards/s82-old-b772-ancestral-knowledge", func(ctx *Context, r game.CardsChosen) error {
	picked := map[uuid.UUID]bool{}
	for _, id := range r.Picked {
		picked[id] = true
		if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	var rest []uuid.UUID
	for _, id := range r.Candidates {
		if !picked[id] {
			rest = append(rest, id)
		}
	}
	if len(rest) == 0 {
		return nil
	}
	return PutInLibraryInAnyOrder{
		Chooser:   ctx.Controller(),
		Cards:     rest,
		From:      game.ZoneLibrary,
		Placement: game.LibraryPlaceTop,
		Label:     "Ancestral Knowledge — put the rest back on top of your library in any order",
	}.Apply(ctx)
})

// S82OldB772SacrificeALandAndDiscardYourHand is Psychic Vortex's "sacrifice a land
// and discard your hand": the land is chosen and sacrificed, then the hand is
// discarded whether or not there was a land to sacrifice.
func S82OldB772SacrificeALandAndDiscardYourHand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return ctx.SacrificeThen(ctx.Controller(), SacrificeSpec("a land", Land()),
		"Psychic Vortex — sacrifice a land", 1, s82OldB772VortexDiscard)
}

// s82OldB772VortexDiscard is the second half of Psychic Vortex's end-step
// ability, once the land has gone: the controller discards their hand. An
// on-disk key: never renamed, never reused.
var s82OldB772VortexDiscard = RunThen("prompt-run/s82-old-b772-psychic-vortex-discard", func(ctx *Context, _ game.RunDone) error {
	_, err := DiscardWholeHand(ctx.Game, ctx.Controller())
	return err
})
