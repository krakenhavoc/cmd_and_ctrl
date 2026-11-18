package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mid_b10_helpers.go — Innistrad: Midnight Hunt, batch 10 (#5497).

// s82Mid10EachCreatureOrEnchantmentYouControlEntered is the permanent an
// EventETB names, while it is a creature or an enchantment the delayed
// trigger's controller controls (Rite of Harmony). Types are read as the
// permanent is on the battlefield (CR 603.10: an enters trigger looks at the
// object as it entered).
func s82Mid10EachCreatureOrEnchantmentYouControlEntered(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) []uuid.UUID {
	if ev.Kind != game.EventETB || ev.CardID == uuid.Nil {
		return nil
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || c.Controller != dt.Controller || !(c.IsCreature() || c.IsEnchantment()) {
		return nil
	}
	return []uuid.UUID{ev.CardID}
}

// Rite of Harmony's trigger condition and body. On-disk keys: never renamed,
// never reused.
var (
	s82Mid10CreatureOrEnchantmentEntersCondition = game.DelayedEachCondition("etb/each-creature-or-enchantment-you-control", s82Mid10EachCreatureOrEnchantmentYouControlEntered)
	s82Mid10DrawACardBody                        = game.SimpleDelayedBody("rite-of-harmony/draw-a-card", Do(DrawCards{N: 1}))
)

// s82Mid10ReturnedHasteAndExile is the tail of "return it to the battlefield
// under your control. It gains haste. Exile it at the beginning of the next
// end step" (Wake to Slaughter): the entered card, if it is still that
// permanent on the battlefield, gains haste for as long as it stays and is
// exiled at the next end step. An on-disk key: never renamed, never reused.
var s82Mid10ReturnedHasteAndExile = EnteredTail("wake-to-slaughter/entered", func(ctx *Context, entered uuid.UUID, _ game.TailOutcome) error {
	g := ctx.Game
	if entered == uuid.Nil || !onBattlefield(g, entered) {
		return nil
	}
	if err := (ScopedEffectFor{
		Target:   entered,
		Mods:     []game.Mod{game.AddKeywordsMod("haste")},
		Duration: g.PinnedTo(game.IndefiniteDuration(), entered),
		Label:    "Wake to Slaughter — it gains haste",
	}).Apply(ctx); err != nil {
		return err
	}
	return ScheduleDelayedTrigger{
		Label: "Wake to Slaughter — exile it",
		Cards: []uuid.UUID{entered},
		Body:  exileListedCardsBody,
	}.Apply(ctx)
})

// s82Mid10ReturnedExileIfItWouldDie is the tail of Can't Stay Away's return:
// the entered creature gains "If this creature would die, exile it instead"
// for as long as it is that permanent. An on-disk key.
var s82Mid10ReturnedExileIfItWouldDie = EnteredTail("cant-stay-away/entered", func(ctx *Context, entered uuid.UUID, _ game.TailOutcome) error {
	if entered == uuid.Nil || !onBattlefield(ctx.Game, entered) {
		return nil
	}
	ctx.Game.ExileIfItWouldDieForEffect(ctx.Source(), entered, ctx.Controller(),
		"Can't Stay Away — if this creature would die, exile it instead")
	return nil
})
