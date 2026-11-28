package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b743_helpers.go — Exodus, S82 OLD batch 743 (#5626).

// WithMoreLandsThanTheChooser narrows a player clause the triggering player
// chooses (OfThatPlayersChoice) to "who controls more lands than they do and
// is their opponent", they being that chooser (Oath of Lieges,
// game.TargetSpec.MoreLandsThanChooser).
func WithMoreLandsThanTheChooser(spec *game.TargetSpec) *game.TargetSpec {
	spec.MoreLandsThanChooser = true
	return spec
}

// S82OldB743EachOpponentGainsThatMuch is Grollub's "each opponent gains that
// much life": the amount is the damage the Grollub was dealt (the batch
// total, ThatMuchDamage).
func S82OldB743EachOpponentGainsThatMuch(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Event.Amount <= 0 {
		return nil
	}
	ctx := NewContext(g, item)
	for _, opp := range ctx.Opponents() {
		if err := (GainLife{Player: opp, Amount: item.Trigger.Event.Amount}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// S82OldB743AllNonlandPermanentsYouControlAreWhite is Zealots en-Dal's
// intervening "if": every nonland permanent `you` controls is white, by its
// colors after the layers. With no nonland permanent it holds vacuously.
func S82OldB743AllNonlandPermanentsYouControlAreWhite(g *game.Game, you uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == you && !c.IsLand() && !c.HasColorWord("W") {
			return false
		}
	}
	return true
}

// S82OldB743LifeAtMostTen is Convalescence's intervening "if": you have 10 or
// less life.
func S82OldB743LifeAtMostTen(g *game.Game, you uuid.UUID) bool {
	p := g.PlayerByIDForEffect(you)
	return p != nil && p.Life <= 10
}

// S82OldB743TargetOpponentWithMoreLife is Keeper of the Light's "target
// opponent who has more life than you do as you activate this ability". The
// life totals are judged as the ability is activated; the life gain does not
// depend on the opponent, so the ability still resolves if the target stops
// qualifying (StillResolves).
func S82OldB743TargetOpponentWithMoreLife() *game.TargetSpec {
	spec := TargetPlayer("target opponent who has more life than you do",
		Opponent(),
		func(g *game.Game, caster uuid.UUID, p *game.Player) bool {
			me := g.PlayerByIDForEffect(caster)
			return me != nil && p.Life > me.Life
		})
	return spec.StillResolves()
}

// S82OldB743EnchantedCreature is Song of Serenity's "creatures that are
// enchanted": a creature with an Aura attached to it.
func S82OldB743EnchantedCreature(target *game.Card, g *game.Game) bool {
	if target == nil || !target.IsCreature() {
		return false
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.HasSubtype("Aura") && c.AttachedTo.Kind == game.TargetCard && c.AttachedTo.ID == target.InstanceID {
			return true
		}
	}
	return false
}

// S82OldB743CarnophageUpkeep is Carnophage's "tap this creature unless you pay
// 1 life": a controller who can't pay taps it, otherwise they choose.
func S82OldB743CarnophageUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !OnBattlefield(g, item.SourceCardID) || SourceIsNewObject(g, item) {
		return nil
	}
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil || !g.CanPayLifeLocked(p, 1) {
		return TapTarget{Target: item.SourceCardID}.Apply(ctx)
	}
	return PickOption{
		Question: "Carnophage — pay 1 life, or tap this creature",
		Options:  []game.ChoiceOption{{Label: "Tap this creature"}, {Label: "Pay 1 life", LifeCost: 1}},
		ThenKey:  s82OldB743CarnophageChose,
		WithItem: true,
	}.Apply(ctx)
}

// s82OldB743CarnophageChose is the keyed answer: paying life leaves it
// untapped, anything else taps it. An on-disk key: never renamed, never
// reused.
var s82OldB743CarnophageChose = OptionPickThen("option-pick/s82-old-b743-carnophage-pay-or-tap", func(ctx *Context, r game.OptionPicked) error {
	item := ctx.Item
	if !OnBattlefield(ctx.Game, item.SourceCardID) || SourceIsNewObject(ctx.Game, item) {
		return nil
	}
	if r.Option != nil && r.Option.Label == "Pay 1 life" {
		if err := ctx.Game.PayLifeForEffect(item.SourceCardID, item.Controller, 1); err == nil {
			return nil
		}
	}
	return TapTarget{Target: item.SourceCardID}.Apply(ctx)
})

// S82OldB743ManaBreach is "that player returns a land they control to its
// owner's hand": the caster named by the cast event picks one of their own
// lands; with none, nothing happens.
func S82OldB743ManaBreach(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Event.Actor == uuid.Nil {
		return nil
	}
	player := item.Trigger.Event.Actor
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Eliminated || len(s82OldB743LandsOf(g, player)) == 0 {
		return nil
	}
	return ChoosePermanents{
		Player:   player,
		Question: "Mana Breach — return a land you control to its owner's hand",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			return s82OldB743LandsOf(g, of), 1, 1
		},
		Of:    []uuid.UUID{player},
		Keyed: s82OldB743ManaBreachBounced,
	}.Apply(NewContext(g, item))
}

func s82OldB743LandsOf(g *game.Game, player uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsLand() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

var s82OldB743ManaBreachBounced = RegisterPermanentsPick("prompt-run/s82-old-b743-mana-breach-return", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if !OnBattlefield(ctx.Game, id) {
				continue
			}
			if err := (BounceToHand{Target: id}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	},
})

// S82OldB743APlayerCastASpell is the When of "whenever a player casts a
// spell": any cast, anyone's, with a caster named.
func S82OldB743APlayerCastASpell(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventCast && ev.Actor != uuid.Nil
}

// S82OldB743OathOfLiegesAsk asks the upkeep player whether to search for a
// basic land card.
func S82OldB743OathOfLiegesAsk(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	first := item.Trigger.Event.Actor
	if first == uuid.Nil || g.PlayerByIDForEffect(first) == nil {
		return nil
	}
	return MayChoice{
		Player:   first,
		Question: "Oath of Lieges — search your library for a basic land card and put it onto the battlefield?",
		YesLabel: "Search",
		NoLabel:  "Don't search",
		Then:     s82OldB743OathOfLiegesSearch,
	}.Apply(NewContext(g, item))
}

// s82OldB743OathOfLiegesSearch is the upkeep player's yes: search for a basic
// land card, put it onto the battlefield (untapped), then shuffle. An on-disk
// key: never renamed, never reused.
var s82OldB743OathOfLiegesSearch = MayYes("confirm/s82-old-b743-oath-of-lieges-search", func(ctx *Context, r game.ConfirmAnswer) error {
	return SearchLibrary{
		Player:    r.Chooser,
		Predicate: IsBasicLand,
		Dest:      game.ZoneBattlefield,
		Limit:     1,
		Shuffle:   true,
		Reason:    "Oath of Lieges — a basic land card, onto the battlefield",
	}.Apply(ctx)
})
