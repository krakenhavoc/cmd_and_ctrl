package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blastoneuron — Creature — Jellyfish Leech {3}{B}, 2/2 (Star Trek):
//
//	Flying
//	When this creature enters, exile target creature card with mana value 3
//	or less from your graveyard. Create a token that's a copy of it. That
//	token gains haste. Sacrifice it at the beginning of the next end step.
//	Whenever one or more cards leave your graveyard, each opponent loses 1
//	life and you gain 1 life.
//
// The copy is made of the card that really reached exile (a card that left
// the graveyard in response makes the trigger do nothing, CR 608.2b). The
// haste is an effect on the token, not part of the copy (CR 707.2), and the
// sacrifice is a delayed trigger at the next end step (CR 603.7).
//
// The second trigger watches every way a card leaves the graveyard, a cast
// from it included, once per batch: the enter trigger's own exile sets it
// off.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "87bcc935-e7e7-4e98-a23c-01c5a9b7cfcd",
		Name:            "Blastoneuron",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Blastoneuron — exile a creature card from your graveyard and create a hasty token copy of it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return ExileTarget{Target: id, ThenTail: ctx.ItemTail(blastoneuronCopy)}.Apply(ctx)
				}),
				TargetCardInGraveyard("target creature card with mana value 3 or less from your graveyard",
					YouOwn(), Creature(), ManaValueLE(3))),
			OncePerBatch(OnAny(leftGraveyardEvents, whenCardsLeaveYourGraveyard,
				"Blastoneuron — each opponent loses 1 life and you gain 1 life",
				Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1}))),
		},
	})
}

// blastoneuronCopy is "create a token that's a copy of it", once the card
// has landed in exile. An on-disk key.
var blastoneuronCopy = CardTail("then/blastoneuron-copy", func(ctx *Context, o game.TailOutcome) error {
	if !oneWentThisWay(o) {
		return nil
	}
	tmpl, ok := TokenCopyTemplate(ctx.Game, o.IDs[0])
	if !ok {
		return nil
	}
	return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: ctx.Controller(),
		Source:     ctx.Item.SourceCardID,
		Groups:     []game.TokenGroup{{Template: copiedToken(tmpl), Count: 1}},
	}, ctx.ItemTail(blastoneuronHasteAndSacrifice))
})

// blastoneuronHasteAndSacrifice is "that token gains haste. Sacrifice it at
// the beginning of the next end step." An on-disk key.
var blastoneuronHasteAndSacrifice = CardTail("then/blastoneuron-haste-and-sacrifice", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	for _, id := range o.IDs {
		if err := (ScopedEffectFor{Target: id, Mods: []game.Mod{game.AddKeywordsMod("haste")},
			Duration: game.IndefiniteDuration(), Label: "Blastoneuron — the token gains haste"}).Apply(ctx); err != nil {
			return err
		}
	}
	return ScheduleDelayedTrigger{
		At:         game.StepEnd,
		Controller: ctx.Item.Controller,
		Label:      "Blastoneuron — sacrifice the token",
		Cards:      o.IDs,
		Body:       sacrificeListedCardsBody,
	}.Apply(ctx)
})
