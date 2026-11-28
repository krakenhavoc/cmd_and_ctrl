package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b808_helpers.go — S82 set OLD, batch 808 (Mirage, activated
// abilities).

// OldB808FlyingIndefinitely is "This creature gains flying" / "This creature
// loses flying. (This effect lasts indefinitely.)" (Mist Dragon): a layer 6
// grant or removal of flying with no stated duration (CR 611.2a), pinned to
// the object the ability came from (CR 400.7). Each resolution takes its own
// timestamp (CR 613.7), so the later of a gain and a loss wins. A source that
// has left, or come back as a new object, gets nothing.
func OldB808FlyingIndefinitely(gain bool, label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !SourceIsStillThisPermanent(g, item) {
			return nil
		}
		mod := game.AddKeywordsMod("flying")
		if !gain {
			mod = game.RemoveKeywordsMod("flying")
		}
		return ScopedEffectFor{
			Target:   item.SourceCardID,
			Mods:     []game.Mod{mod},
			Duration: g.PinnedTo(game.IndefiniteDuration(), item.SourceCardID),
			Label:    label,
		}.Apply(NewContext(g, item))
	}
}

// OldB808RegenerateTheTarget is "Regenerate target creature" (Village Elder):
// one regeneration shield on the first target still legal as the ability
// resolves (CR 608.2b, 701.19a).
func OldB808RegenerateTheTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return Regenerate{Target: t.ID}.Apply(ctx)
		}
	}
	return nil
}

// OldB808TargetGainsForestwalk is "Target creature gains forestwalk until end
// of turn" (Unseen Walker): a layer 6 grant locked to the target as the ability
// resolves (CR 611.2c).
func OldB808TargetGainsForestwalk(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"forestwalk"}, Label: "Unseen Walker — forestwalk"}.Apply(ctx)
		}
	}
	return nil
}

// OldB808TargetGainsFlyingBaseToughnessOne is "Until end of turn, target
// creature you control gains flying and has base toughness 1" (Chariot of the
// Sun): one effect, one timestamp, a layer 6 grant and a layer 7b base-toughness
// setting, locked to the target as the ability resolves (CR 611.2c).
func OldB808TargetGainsFlyingBaseToughnessOne(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return UntilEndOfTurn(ctx, t.ID, nil, "Chariot of the Sun — flying and base toughness 1",
				game.AddKeywordsMod("flying"), game.SetBaseToughnessMod(1))
		}
	}
	return nil
}

// OldB808PTEqualToForestsYouControl is a characteristic-defining ability (CR
// 604.3) for "This creature's power and toughness are each equal to the number
// of Forests you control" (Uktabi Wildcats): every Forest its controller
// controls, whatever else it is, counts once. The type is a word of the
// source's text (CR 612.2).
func OldB808PTEqualToForestsYouControl(label string) game.StaticAbility {
	return game.StaticAbility{
		Label:     label,
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7A_CDA,
		AppliesTo: SelfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			forest := source.TextSubtype("Forest")
			n := 0
			for _, p := range g.BattlefieldCardsForEffect() {
				if p.Controller == source.Controller && p.HasSubtype(forest) {
					n++
				}
			}
			c.Power, c.Toughness = n, n
		},
	}
}

// OldB808GrinningTotemBody is the delayed trigger "At the beginning of your next
// upkeep, if you haven't played it, put it into its owner's graveyard": each
// listed card still in exile goes to its owner's graveyard. A card that was
// played has left exile, so it is left alone. An on-disk key: never renamed,
// never reused.
var OldB808GrinningTotemBody = game.SimpleDelayedBody("s82-old-b808/grinning-totem-graveyard",
	func(g *game.Game, item *game.StackItem) error {
		for _, t := range item.Targets {
			if t.Kind != game.TargetCard {
				continue
			}
			if _, err := Cmdr2S32PutExiledCardIntoGraveyard(g, t.ID); err != nil {
				return err
			}
		}
		return nil
	})

// OldB808GrinningTotemGrant is the rest of Grinning Totem's effect, run with
// the card the search exiled: its controller may play it until their next
// upkeep begins (CR 611.2), and a delayed trigger at that upkeep puts it into
// its owner's graveyard if it is still in exile. An on-disk key: never
// renamed, never reused.
var OldB808GrinningTotemGrant = CardTail("s82-old-b808/grinning-totem-grant",
	func(ctx *Context, o game.TailOutcome) error {
		var ids []uuid.UUID
		for _, id := range o.IDs {
			if err := ctx.Game.ExileCardWithPermissionForEffect(id, game.CastPermission{
				Player:   ctx.Controller(),
				Duration: DurationUntilYourNextUpkeep(ctx, ctx.Controller()),
			}); err != nil {
				return err
			}
			ids = append(ids, id)
		}
		if len(ids) == 0 {
			return nil
		}
		return ScheduleDelayedTrigger{
			At:                 game.StepUpkeep,
			Label:              "Grinning Totem — if you haven't played it, put it into its owner's graveyard",
			ControllerTurnOnly: true,
			Cards:              ids,
			Body:               OldB808GrinningTotemBody,
		}.Apply(ctx)
	})

// OldB808TargetSpellBecomesColorless is "Target spell becomes colorless" (Ersatz
// Gnomes): a layer 5 change to no colour with no stated duration (CR 611.2a),
// pinned to the spell on the stack (game.SetSpellColorsForEffect) and, for a
// permanent spell, carried to the permanent it becomes (CR 400.7a). A target
// that has resolved or been countered first does nothing (CR 608.2b).
func OldB808TargetSpellBecomesColorless(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard && g.StackItemForEffect(t.ID) != nil {
			g.SetSpellColorsForEffect(ctx.Source(), t.ID, nil, game.IndefiniteDuration(), "Ersatz Gnomes — colorless")
			return nil
		}
	}
	return nil
}
