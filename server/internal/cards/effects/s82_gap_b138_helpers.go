package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b138_helpers.go — Dominaria Remastered creatures, GAP batch 138
// (#5948).

// GapB138ExileACreatureOrSacrifice is Wormfang Drake's linked pair: "When
// this creature enters, sacrifice it unless you exile a creature you
// control other than this creature. When this creature leaves the
// battlefield, return the exiled card to the battlefield under its owner's
// control."
//
// It is Champion's shape (see NewChampion) with "a creature" for the
// subtype: the enters trigger exiles the chosen creature until the Drake
// leaves the battlefield (ExileUntil, CR 610.3), so the return is the
// engine's record of that exile. The choice is made as the trigger resolves
// (a choose_cards prompt of one creature or none; none, or having no other
// creature, sacrifices the Drake). A Drake that has already left by then
// exiles nothing and is not sacrificed again (CR 400.7).
//
// Each card registers its own pick key: an on-disk identity, never renamed
// or reused.
type GapB138ExileACreatureOrSacrifice struct {
	cardName string
	question string
	label    string
	pick     game.ChooseCardsThen
}

// NewGapB138ExileACreatureOrSacrifice registers the pick under pickKey.
// Call it once, from a package-level var.
func NewGapB138ExileACreatureOrSacrifice(cardName, pickKey string) *GapB138ExileACreatureOrSacrifice {
	c := &GapB138ExileACreatureOrSacrifice{
		cardName: cardName,
		question: cardName + " — exile a creature you control other than " + cardName + " (or sacrifice " + cardName + ")",
		label:    cardName + " — the exiled card returns when " + cardName + " leaves the battlefield",
	}
	c.pick = AskedPick(pickKey, c.picked)
	return c
}

// Trigger is the enters trigger.
func (c *GapB138ExileACreatureOrSacrifice) Trigger() game.TriggeredAbility {
	return WhenThisEnters(c.cardName+" — sacrifice it unless you exile a creature you control other than it", c.enters)
}

func (c *GapB138ExileACreatureOrSacrifice) enters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var candidates []uuid.UUID
	for _, p := range g.BattlefieldCardsForEffect() {
		if p.Controller == ctx.Controller() && p.InstanceID != item.SourceCardID && p.IsCreature() {
			candidates = append(candidates, p.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return SacrificePermanent{Target: item.SourceCardID}.Apply(ctx)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    ctx.Controller(),
		FromPlayer: ctx.Controller(),
		Source:     item.SourceCardID,
		Question:   c.question,
		Cards:      candidates,
		Min:        0,
		Max:        1,
		LiveZone:   game.ZoneBattlefield,
		ThenKey:    c.pick,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

func (c *GapB138ExileACreatureOrSacrifice) picked(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 1 && OnBattlefield(ctx.Game, r.Picked[0]) {
		return ExileUntil{Target: r.Picked[0], ThisLeaves: true, Label: c.label}.Apply(ctx)
	}
	return SacrificePermanent{Target: ctx.Source()}.Apply(ctx)
}

// GapB138SpellsCastThisTurnByAnyone is the number of spells cast this turn
// by every player, the entering card's own cast included when it was cast.
func GapB138SpellsCastThisTurnByAnyone(g *game.Game) int {
	n := 0
	for _, p := range g.Seats {
		if p != nil {
			n += g.CastTallyFor(p.ID).Total
		}
	}
	return n
}

// GapB138WasCast reports whether the card's most recent cast event is its
// entry as a spell: the last EventCast naming it is newer than its last
// move onto the battlefield, so the per-turn tally holds that spell.
func GapB138WasCast(g *game.Game, cardID uuid.UUID) bool {
	for i := len(g.Events) - 1; i >= 0; i-- {
		ev := g.Events[i]
		if ev.CardID != cardID {
			continue
		}
		switch {
		case ev.Kind == game.EventCast:
			return true
		case ev.Kind == game.EventZoneMove && ev.NewZone == game.ZoneBattlefield:
			return false
		}
	}
	return false
}
