package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// rprB04Outlast is "Outlast {cost}" (CR 702.107a): "{cost}, {T}: Put a +1/+1
// counter on this creature. Activate only as a sorcery."
func rprB04Outlast(cost string) ActivatedAbility {
	return ActivatedAbility{
		Label:        "Outlast " + cost + " (" + cost + ", {T}: Put a +1/+1 counter on this creature. Outlast only as a sorcery.)",
		Cost:         Plus(ManaCost(cost), TapCost()),
		SorcerySpeed: true,
		Purpose:      game.Purpose{},
		Effect: func(g *game.Game, item *game.StackItem) error {
			return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
		},
	}
}

// rprB04AnotherCreatureOfTheChosenTypeEntered is "another creature you
// control of the chosen type enters": the entering object is read through
// the layered view, so a changeling is every type (CR 702.73a).
func rprB04AnotherCreatureOfTheChosenTypeEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source.NamedTribe == "" {
		return false
	}
	c, ok := enteredUnderYourControl(ev, source, g, true)
	if !ok {
		return false
	}
	return (TribeFilter{Chosen: true, Others: true, YoursOnly: true}).Matches(&c, g, source)
}

// rprB04PreventDividedDamage is "prevent the next N damage that would be
// dealt this turn to any number of targets, divided as you choose": each
// target still legal gets its own shield for the share announced for it
// (CR 601.2d, 615.7). A target that has gone takes no shield and its share
// is not moved to another (CR 608.2b).
func rprB04PreventDividedDamage(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			n := ctx.Item.Distribution[t.ID]
			if n <= 0 {
				continue
			}
			if err := (PreventNextDamage{Target: t.ID, Amount: n, Label: label}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}
