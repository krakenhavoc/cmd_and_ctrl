package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ursine Monstrosity — Creature — Bear Mutant {2}{G}, 3/3 (Duskmourn
// Commander):
//
//	"Trample
//	 At the beginning of combat on your turn, mill a card and choose an
//	 opponent at random. This creature attacks that player this combat if
//	 able. Until end of turn, this creature gains indestructible and gets +1/+1
//	 for each card type among cards in your graveyard."
//
// Knight Rampager's "attacks that player this combat if able" after a one-card mill: the opponent is
// drawn through the keyed random stream and logged, and the Monstrosity gets
// game.ModAttackPlayerIfAble until end of combat. The card types are counted
// after the mill, once, as the trigger resolves (CR 611.2c), and the +1/+1 is
// a fixed amount from then on. A Monstrosity that left the battlefield is a
// new object, so none of it follows.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "647feda9-2876-4430-902f-10dc1f4f99a9",
		Name:            "Ursine Monstrosity",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Ursine Monstrosity — mill a card, choose an opponent at random; this creature attacks that player this combat if able; indestructible and +1/+1 for each card type among cards in your graveyard until end of turn",
				ursineMonstrosityCombat),
		},
	})
}

func ursineMonstrosityCombat(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (MillCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
		return err
	}
	pick := g.ChooseAtRandomForEffect(randomDraw(ctx), ctx.Opponents(), 1)
	if len(pick) != 1 {
		return nil
	}
	name := "an opponent"
	if p := g.PlayerByIDForEffect(pick[0]); p != nil && p.Name != "" {
		name = p.Name
	}
	g.EmitEvent(game.Event{
		Kind:   game.EventOptionChosen,
		Actor:  item.Controller,
		CardID: item.SourceCardID,
		Label:  "Ursine Monstrosity: " + name + " (at random)",
	})
	if sourceIsNewObject(g, item) {
		return nil
	}
	d, ok := DurationUntilEndOfCombat(ctx)
	if !ok {
		return nil
	}
	if err := (ScopedEffectFor{
		Target:   item.SourceCardID,
		Mods:     []game.Mod{game.AttackPlayerIfAbleMod(pick[0])},
		Duration: d,
		Label:    "Ursine Monstrosity — attacks that player this combat if able",
	}).Apply(ctx); err != nil {
		return err
	}
	if err := (GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"indestructible"},
		Label: "Ursine Monstrosity — indestructible until end of turn"}).Apply(ctx); err != nil {
		return err
	}
	n := b16CardTypesInGraveyard(g, item.Controller)
	if n == 0 {
		return nil
	}
	return BoostUntilEOT{Target: item.SourceCardID, Power: n, Toughness: n,
		Label: "Ursine Monstrosity — +1/+1 for each card type among cards in your graveyard"}.Apply(ctx)
}
