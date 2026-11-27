package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b623_helpers.go — shared pieces of Judgment batch 623 (#5611),
// damage.

// OldB623GainLifePerCardInYourGraveyard is Ancestor's Chosen's "you gain 1
// life for each card in your graveyard", counted as the trigger resolves.
func OldB623GainLifePerCardInYourGraveyard(g *game.Game, item *game.StackItem) error {
	return GainLife{Amount: B31GraveyardSize(g, item.Controller)}.Apply(NewContext(g, item))
}

// OldB623DwarvenScorcher is "This creature deals 1 damage to target creature
// unless that creature's controller has this creature deal 2 damage to them."
// The creature's controller (read as the ability resolves) is asked, as any
// "unless" payment: "yes" has the source deal 2 damage to them, "no" has it
// deal 1 damage to the creature. The source was sacrificed as the cost, so its
// damage is dealt as it last existed (CR 608.2h).
func OldB623DwarvenScorcher(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	creature, found := g.LookupCardForEffect(id)
	if !found {
		return nil
	}
	words, action := DamagePayment(2).Prompt()
	return UpkeepPayUnless{
		Chooser:  creature.Controller,
		Cost:     words,
		Action:   action,
		Question: "Dwarven Scorcher — have it deal 2 damage to you, or it deals 1 damage to " + creature.Name + "?",
		Else:     oldB623ScorcherDeals1,
		Carry:    []uuid.UUID{id},
	}.Apply(ctx)
}

// oldB623ScorcherDeals1 is the decline: the source deals 1 damage to the
// creature it targeted. An on-disk key.
var oldB623ScorcherDeals1 = PayUnlessBranches("pay-unless/old-b623-dwarven-scorcher-1", nil,
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if len(r.Carry) == 0 {
			return nil
		}
		return DealDamage{Source: ctx.Source(), Target: r.Carry[0], Amount: 1}.Apply(ctx)
	})

// OldB623TargetUpToThatManyPermanentsThatPlayerControls is Cephalid
// Constable's clause: "up to that many target permanents that player controls",
// the number being the combat damage the creature just dealt and the player the
// one it was dealt to, both read off the damage event as the trigger is put on
// the stack (CR 603.3d). With no damage or no player there is nothing to target.
func OldB623TargetUpToThatManyPermanentsThatPlayerControls(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
	player, n := tc.Event.Target, tc.Event.Amount
	if player == uuid.Nil || n <= 0 {
		return nil
	}
	return &game.TargetSpec{
		Mode: "permanent", Label: "up to that many target permanents that player controls",
		Zones: []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return c.Controller == player
		},
		Min: 0, Max: n,
	}
}

// OldB623ScalpelexisExile is Scalpelexis's process: the damaged player exiles
// the top four cards of their library; if two or more of those cards have the
// same name, they repeat it. The player rides the tail's Carry, the asking item
// its Asked. A library with fewer than four cards exiles what it has, and an
// empty one ends the process (nothing has a name to share).
func OldB623ScalpelexisExile(ctx *Context, player uuid.UUID) error {
	tail := ctx.ItemTail(oldB623ScalpelexisRepeat)
	tail.Carry = []uuid.UUID{player}
	return ctx.Game.ExileTopKeyedForEffect(player, player, 4, tail)
}

// oldB623ScalpelexisRepeat looks at the four cards that landed in exile and
// goes round again when two of them share a name. An on-disk key.
var oldB623ScalpelexisRepeat = CardTail("old-b623/scalpelexis-repeat", func(ctx *Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	seen := map[string]bool{}
	repeat := false
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok {
			continue
		}
		if seen[c.Name] {
			repeat = true
		}
		seen[c.Name] = true
	}
	if !repeat {
		return nil
	}
	return ctx.Game.ExileTopKeyedForEffect(o.Tail.Carry[0], o.Tail.Carry[0], 4, o.Tail)
})
