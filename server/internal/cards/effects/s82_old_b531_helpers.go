package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b531_helpers.go — shared pieces for the S82 OLD batch 531 cards
// (Shadowmoor activated abilities, #5591).

// OldB531UntapCost is "<mana>, {Q}" (CR 107.6): the mana, and the source must
// be tapped to untap it (AbilityCost.UntapSelf), so a creature that has not been
// under its controller's control since their turn began cannot pay it (CR 302.6).
func OldB531UntapCost(mana string) game.AbilityCost {
	return game.AbilityCost{Mana: mana, UntapSelf: true}
}

// OldB531ThisBecomesAllColorsUntilEOT is "this creature becomes all colors until
// end of turn" (Scrapbasket): Tam's colour-setting layer-5 effect on the source.
func OldB531ThisBecomesAllColorsUntilEOT(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if !OnBattlefield(g, item.SourceCardID) {
			return nil
		}
		return ScopedEffectFor{
			Target:   item.SourceCardID,
			Mods:     []game.Mod{game.SetColorsMod(game.AllColors...)},
			Duration: DurationUntilEndOfTurn(ctx),
			Label:    label,
		}.Apply(ctx)
	}
}

// OldB531ThisGainsUntilEOT is "this creature gains <keyword> until end of turn".
func OldB531ThisGainsUntilEOT(label string, keywords ...string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !OnBattlefield(g, item.SourceCardID) {
			return nil
		}
		return GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: keywords, Label: label}.Apply(NewContext(g, item))
	}
}

// OldB531TargetCreatureProtectionOfItsControllersChoice is "target creature gains
// protection from the color of its controller's choice until end of turn" (Pale
// Wayfarer): the colour is asked at resolution (Mother of Runes' order) and
// asked of the creature's controller, not of the ability's.
func OldB531TargetCreatureProtectionOfItsControllersChoice(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		c, ok := g.LookupCardForEffect(id)
		if !ok || !OnBattlefield(g, id) {
			return nil
		}
		ChooseProtectionColorFor(ctx, c.Controller, id, name, name+" — choose a color for protection")
		return nil
	}
}

// OldB531LookAtTopMayExile is "look at the top card of your library. You may
// exile that card" (Puresight Merrow): a look (CR 701.20a, only you learn what it
// is), a may, and an exile that is face up, so everyone sees the card once it is
// exiled.
func OldB531LookAtTopMayExile(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		looked := g.LookAtTopOfLibraryForEffect(item.Controller, 1)
		if len(looked) == 0 {
			return nil
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  item.Controller,
			Source:   item.SourceCardID,
			Question: name + " — exile that card?",
			Cards:    looked,
			Min:      0,
			Max:      1,
			LiveZone: game.ZoneLibrary,
			ThenKey:  oldB531ExilePicked,
		})
		return nil
	}
}

// oldB531ExilePicked exiles the card the player chose to exile. An on-disk key:
// never renamed, never reused.
var oldB531ExilePicked = game.RegisterChooseCardsThen("choose-cards/s82-old-b531-exile-picked", func(g *game.Game, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if err := g.ExileCardForEffect(id); err != nil {
			return err
		}
	}
	return nil
})

// OldB531ImpromptuRaid is "Reveal the top card of your library. If it isn't a
// creature card, put it into your graveyard. Otherwise, put that card onto the
// battlefield. That creature gains haste. Sacrifice it at the beginning of the
// next end step." The creature enters through the keyed library entry, so haste
// and the delayed sacrifice follow once it has really arrived, even when a
// replacement window asked a question first (CR 614).
func OldB531ImpromptuRaid(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var revealed []uuid.UUID
	if err := (RevealTopOfLibrary{N: 1, Reason: "Impromptu Raid", Revealed: &revealed}).Apply(ctx); err != nil {
		return err
	}
	if len(revealed) == 0 {
		return nil
	}
	top, ok := g.LookupCardForEffect(revealed[0])
	if !ok {
		return nil
	}
	if !top.IsCreature() {
		return g.PutIntoGraveyardForEffect(top.InstanceID)
	}
	return g.PutCardsFromLibraryOntoBattlefieldKeyedForEffect([]uuid.UUID{top.InstanceID},
		game.ZoneEntryOptions{Controller: item.Controller}, ctx.ItemTail(oldB531RaidArrived))
}

// oldB531RaidArrived gives the creature that arrived haste, with no duration as
// printed, and schedules its sacrifice for the beginning of the next end step,
// bound to that object (CR 603.7a; a creature that left and came back is new).
// An on-disk key: never renamed, never reused.
var oldB531RaidArrived = CardTail("then/s82-old-b531-impromptu-raid-arrived", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		c, ok := BattlefieldCardForEffect(ctx.Game, id)
		if !ok {
			continue
		}
		if err := (ScopedEffectFor{
			Target:   id,
			Mods:     []game.Mod{game.AddKeywordsMod("haste")},
			Duration: game.IndefiniteDuration(),
			Label:    "Impromptu Raid — gains haste",
		}).Apply(ctx); err != nil {
			return err
		}
		if err := (ScheduleDelayedTrigger{
			Label:  "Impromptu Raid — sacrifice the creature",
			Body:   StdB66MeekAttackSacrifice,
			Params: game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
