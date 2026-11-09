package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b35_helpers.go — shared helpers for Standard-set leftover batches 35 to
// 38 (#4452).

// b35GolemToken is the 3/3 colorless Golem artifact creature token the Big Score
// prints (Sandstorm Salvager, Legion Extruder). The tokens table's "3/3
// colorless Golem artifact" row is an enchantment Golem, so it is not reused.
func b35GolemToken() game.Card {
	return game.Card{Name: "Golem", TypeLine: "Token Artifact Creature — Golem", Power: 3, Toughness: 3}
}

// b35CreatureTokensYouControl lists the creature tokens `controller` controls.
func b35CreatureTokensYouControl(g *game.Game, controller uuid.UUID) []uuid.UUID {
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && c.IsToken() {
			ids = append(ids, c.InstanceID)
		}
	}
	return ids
}

// b35IDSet is a CardPredicate over a fixed set of instance ids.
func b35IDSet(ids []uuid.UUID) CardPredicate {
	set := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return set[c.InstanceID] }
}

// b35YouDiscardLoseSacrifice is "you discard `discard` card(s), lose
// `life` life, and sacrifice `sac` creature(s)" (Greed's Gambit), in printed
// order: the discard prompt is queued first, the life is lost, then the
// sacrifice prompt is queued behind it.
func b35YouDiscardLoseSacrifice(discard, life, sac int, label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		p := item.Controller
		g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: p, Source: item.SourceCardID, N: discard, Question: label + " — discard"})
		if err := g.ChangePlayerLifeForEffect(item.SourceCardID, p, -life); err != nil {
			return err
		}
		g.PlayerSacrificesNForEffect(item.SourceCardID, p, sacrificeSpec("a creature", Creature()), label+" — sacrifice", sac)
		return nil
	}
}

// b35BatToken is the 2/1 black Bat creature token with flying (Greed's Gambit).
func b35BatToken() game.Card {
	return game.Card{Name: "Bat", TypeLine: "Token Creature — Bat", Power: 2, Toughness: 1, Colors: []string{"B"}, Keywords: []string{"flying"}}
}

// b35CopyTargetTokenModeBody is "create a token that's a copy of target
// <artifact> token you control" as a mode's body (Oltec Matterweaver); the
// target is read at resolution (CR 608.2b).
func b35CopyTargetTokenModeBody(_ *game.StackItem, ctx *Context, occ int) error {
	tr, ok := ModeTarget(ctx, occ)
	if !ok || tr.Kind != game.TargetCard {
		return nil
	}
	return CreateTokenCopy{Controller: ctx.Controller(), Copy: tr.ID, N: 1}.Apply(ctx)
}

// b35CopyFirstTargetToken is "create a token that's a copy of target
// <artifact> token you control" as a resolution (Worldwalker Helm).
func b35CopyFirstTargetToken(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return CreateTokenCopy{Controller: ctx.Controller(), Copy: t.ID, N: 1}.Apply(ctx)
		}
	}
	return nil
}

// b35ArtifactTokenYouControl is the clause "target artifact token you
// control".
func b35ArtifactTokenYouControl() *game.TargetSpec {
	return TargetPermanent("target artifact token you control", Artifact(), IsTokenPredicate(), YouControl())
}

// b35YouActivatedAnAbilityTargetingACreatureOrPlayer is Ertha Jo's trigger
// condition: "you activate an ability that targets a creature or player".
// It reads the ability's chosen targets off its stack item as the
// activation put it there (CR 602.2b); a mana ability never used the stack
// and never raises the event.
func b35YouActivatedAnAbilityTargetingACreatureOrPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventActivateAbility || source == nil || ev.Actor != source.Controller || ev.StackItemID == uuid.Nil {
		return false
	}
	it := g.StackItemForEffect(ev.StackItemID)
	if it == nil {
		return false
	}
	for _, t := range it.Targets {
		switch t.Kind {
		case game.TargetPlayer:
			return true
		case game.TargetCard:
			if c, ok := g.LookupCardForEffect(t.ID); ok && c.IsCreature() {
				return true
			}
		}
	}
	return false
}

// b35PlotFromHand is "you may exile a nonland card with mana value 3 or
// less from your hand. If you do, it becomes plotted" (Kellan Joins Up,
// Jace Reawakened's second +1): a pick of up to one over the controller's
// qualifying hand cards, whose exile hands the card to PlotExiled once it
// has reached exile (CR 702.170c). With nothing to pick nobody is asked.
func b35PlotFromHand(question string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		you := ctx.Controller()
		p := g.PlayerByIDForEffect(you)
		if p == nil || p.Hand == nil {
			return nil
		}
		var ids []uuid.UUID
		for _, c := range p.Hand.Cards {
			if !c.IsLand() && c.ManaValue() <= 3 {
				ids = append(ids, c.InstanceID)
			}
		}
		if len(ids) == 0 {
			return nil
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  you,
			Source:   ctx.Source(),
			Question: question,
			Cards:    ids,
			Min:      0,
			Max:      1,
			LiveZone: game.ZoneHand,
			ThenKey:  b35PlotFromHandPick,
			Asked:    game.NewAskedItem(ctx.Item),
		})
		return nil
	}
}

// b35PlotFromHandPick exiles the picked card and plots it once it is in
// exile. An on-disk key: never renamed, never reused.
var b35PlotFromHandPick = askedPick("choose-cards/exile-from-hand-and-plot", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return ExileTarget{Target: r.Picked[0], ThenTail: ctx.ItemTail(b35PlotTheExiled)}.Apply(ctx)
})

// b35PlotTheExiled is "if you do, it becomes plotted".
var b35PlotTheExiled = CardTail("exile-from-hand-and-plot/plot", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if err := (PlotExiled{Card: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

// b35ElkToken is the 3/3 green Elk creature token (Oko, the Ringleader).
func b35ElkToken() game.Card {
	return game.Card{Name: "Elk", TypeLine: "Token Creature — Elk", Power: 3, Toughness: 3, Colors: []string{"G"}}
}

// b35TokenCopyOfEachOtherNonland is "for each other nonland permanent you
// control, create a token that's a copy of that permanent" (Oko, the
// Ringleader's −5): the permanents are fixed first (CR 608.2c) and each
// copy takes the permanent's copiable values (CR 707.2).
func b35TokenCopyOfEachOtherNonland(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var refs []game.ObjectRef
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != item.Controller || c.IsLand() || c.InstanceID == item.SourceCardID {
			continue
		}
		if ref, ok := g.PermanentRefForEffect(c.InstanceID); ok {
			refs = append(refs, ref)
		}
	}
	for _, ref := range refs {
		tmpl, ok := g.TokenCopyTemplateForObjectForEffect(ref)
		if !ok {
			continue
		}
		if err := createTokenCopies(ctx, item.Controller, tmpl, 1); err != nil {
			return err
		}
	}
	return nil
}
