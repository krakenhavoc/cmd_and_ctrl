package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b800_helpers.go — S82 OLD batch 800 (Mirage, #5633).

// SapphireCharmDrawBody is Sapphire Charm's first bullet, "target player
// draws a card at the beginning of the next turn's upkeep": the delayed
// trigger is the spell's controller's (CR 603.7d) and the player it names, as
// plain data, is who draws. An on-disk key.
var SapphireCharmDrawBody = game.DelayedBody("s82-old-b800/sapphire-charm-target-player-draws", func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if p.Player == uuid.Nil {
		return nil
	}
	return DrawCards{Player: p.Player, N: 1}.Apply(NewContext(g, item))
})

// OldB800SapphireCharmDraw schedules that draw for the first legal player
// target.
func OldB800SapphireCharmDraw(_ *game.StackItem, ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		return ScheduleDelayedTrigger{
			At:     game.StepUpkeep,
			Label:  "Sapphire Charm — that player draws a card",
			Body:   SapphireCharmDrawBody,
			Params: game.EffectParams{Player: t.ID},
		}.Apply(ctx)
	}
	return nil
}

// OldB800PhaseOutTheTarget is "target creature an opponent controls phases
// out": the first legal target phases out until its controller's next untap
// step (CR 702.26).
func OldB800PhaseOutTheTarget(_ *game.StackItem, ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return PhaseOut{Targets: []uuid.UUID{t.ID}}.Apply(ctx)
		}
	}
	return nil
}

// OldB800InfernalContract is "draw four cards. You lose half your life,
// rounded up": the draws first, the life read as the spell resolves and lost
// (not damage, CR 119.3), so no prevention sees it.
func OldB800InfernalContract(_ *game.StackItem, ctx *Context) error {
	if err := (DrawCards{N: 4}).Apply(ctx); err != nil {
		return err
	}
	p := ctx.PlayerByID(ctx.Item.Controller)
	if p == nil {
		return nil
	}
	if n := game.HalfLifeRoundedUp(p.Life); n > 0 {
		return ctx.Game.ChangePlayerLifeForEffect(ctx.Item.SourceCardID, ctx.Item.Controller, -n)
	}
	return nil
}

// zirilanFetched is the rest of Zirilan's ability, told the Dragon that
// entered: haste until end of turn, and a delayed exile at the beginning of
// the next end step. An on-disk key.
var zirilanFetched = CardTail("then/s82-old-b800-zirilan-haste-and-exile", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		c, ok := BattlefieldCardForEffect(ctx.Game, id)
		if !ok {
			continue
		}
		if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Zirilan of the Claw — haste"}).Apply(ctx); err != nil {
			return err
		}
		if err := (ScheduleDelayedTrigger{
			Label: "Zirilan of the Claw — exile " + c.Name,
			Cards: []uuid.UUID{id},
			Body:  ExileListedCardsBody,
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

// OldB800ZirilanSearch is "search your library for a Dragon permanent card,
// put that card onto the battlefield, then shuffle": one Dragon, with the
// haste and the delayed exile as the tail that runs once it has entered.
func OldB800ZirilanSearch(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return SearchLibrary{
		Player: ctx.Controller(),
		Predicate: func(c game.Card) bool {
			return (c.IsPermanent()) && c.HasSubtypeWord("Dragon")
		},
		Dest:     game.ZoneBattlefield,
		Limit:    1,
		Shuffle:  true,
		Reason:   "Zirilan of the Claw — search for a Dragon permanent card",
		Source:   ctx.Source(),
		ThenTail: ctx.ItemTail(zirilanFetched),
	}.Apply(ctx)
}

// oldB800NamedCreature is "a creature named <name>" as a sacrifice kind.
func oldB800NamedCreature(name string) game.SacrificeKind {
	return game.SacrificeKind{Label: "a creature named " + name, CardTypes: []string{"creature"}, Names: []string{name}}
}

// OldB800UrborgPantherSummonCost is "sacrifice a creature named Feral Shadow,
// a creature named Breathstealer, and this creature": one pick for each named
// kind (a set rule, each kind filled by its own permanent) and the Panther
// itself.
func OldB800UrborgPantherSummonCost() game.AbilityCost {
	return Plus(SacrificeThis(), SacrificeEach("a creature named Feral Shadow and a creature named Breathstealer",
		oldB800NamedCreature("Feral Shadow"), oldB800NamedCreature("Breathstealer")))
}

// OldB800SearchSpiritOfTheNight is "search your library for a card named
// Spirit of the Night, put that card onto the battlefield, then shuffle".
func OldB800SearchSpiritOfTheNight(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return SearchLibrary{
		Player:    ctx.Controller(),
		Predicate: func(c game.Card) bool { return game.HasName(c, "Spirit of the Night") },
		Dest:      game.ZoneBattlefield,
		Limit:     1,
		Shuffle:   true,
		Reason:    "Urborg Panther — search for a card named Spirit of the Night",
		Source:    ctx.Source(),
	}.Apply(ctx)
}

// merfolkSeerDraw is the paid branch of Merfolk Seer's "you may pay {1}{U}.
// If you do, draw a card". An on-disk key.
var merfolkSeerDraw = PayUnlessBranches("pay-unless/s82-old-b800-merfolk-seer-draw",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return DrawCards{N: 1}.Apply(ctx)
	}, nil)

// OldB800MerfolkSeerDies asks for {1}{U} as the dies trigger resolves.
func OldB800MerfolkSeerDies(g *game.Game, item *game.StackItem) error {
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{1}{U}",
		Question: "Merfolk Seer — pay {1}{U} to draw a card?",
		Then:     merfolkSeerDraw,
	}.Apply(NewContext(g, item))
}

// OldB800ReparationsCastTargetsYou is Reparations' trigger condition: an
// opponent of its controller cast a spell that targets that player or a
// creature they control (CR 603.2, judged as the spell is cast).
func OldB800ReparationsCastTargetsYou(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || ev.Actor == uuid.Nil || ev.Actor == source.Controller {
		return false
	}
	item := g.StackItemForEffect(ev.CardID)
	if item == nil || item.Kind != game.StackItemSpell {
		return false
	}
	return B10ItemTargetsYouOrACreatureYouControl(g, source.Controller, item)
}
