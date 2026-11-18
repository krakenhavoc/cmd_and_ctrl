package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Molten Echoes — Enchantment {2}{R}{R}:
//
//	"As this enchantment enters, choose a creature type.
//	 Whenever a nontoken creature you control of the chosen type enters,
//	 create a token that's a copy of that creature. That token gains haste.
//	 Exile it at the beginning of the next end step."
//
// The type is chosen as the enchantment enters (CR 614.12) and kept on the
// permanent. The trigger fires for each nontoken creature of that type that
// enters under your control, judged on the entering card's layered types; the
// token is a copy of it with haste until end of turn, and a delayed trigger
// exiles exactly that token at the beginning of the next end step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "26486a42-4dc6-472b-9a9b-3c3ca78cc3b5",
		Name:         "Molten Echoes",
		Completeness: CompletenessFull,
		AsEnters:     ChooseCreatureTypeAsEnters("Molten Echoes"),
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && source.NamedTribe != "" && !c.IsToken() && c.IsCreature() && c.HasSubtype(source.NamedTribe)
			}, "Molten Echoes — create a token that's a copy of that creature; it gains haste; exile it at the beginning of the next end step",
				moltenEchoesCopy),
		},
	})
}

func moltenEchoesCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	cursor := b25LastEventSeq(g)
	if err := (CreateTokenCopy{
		Controller: ctx.Controller(),
		Copy:       item.Trigger.Event.CardID,
		N:          1,
	}).Apply(ctx); err != nil {
		return err
	}
	tokens := b27TokensCreatedByAfter(g, ctx.Controller(), cursor)
	for _, token := range tokens {
		if err := (ScopedEffectFor{
			Target:   token,
			Mods:     []game.Mod{game.AddKeywordsMod("haste")},
			Duration: DurationUntilEndOfTurn(ctx),
			Label:    "Molten Echoes — the token gains haste",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	if len(tokens) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label: "Molten Echoes — exile the token",
		Cards: tokens,
		Body:  exileListedCardsBody,
	}.Apply(ctx)
}
