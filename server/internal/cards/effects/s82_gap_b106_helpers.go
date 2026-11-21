package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_gap_b106_helpers.go — shared bodies for S82 GAP batch 106 (#5948).

// gapB106CopyTheCastSpell is "copy it. You may choose new targets for the
// copy" for a trigger on a spell's cast (Codie, Ravenous Codex; Melek, Izzet
// Paragon): the copy is of the spell the trigger's event names, under the
// trigger's controller.
func gapB106CopyTheCastSpell(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return CopySpell{
		StackID:          ctx.Trigger().Event.CardID,
		Controller:       item.Controller,
		ChooseNewTargets: true,
	}.Apply(ctx)
}

// gapB106OneDamageAndOneLife is "it deals 1 damage to <the target> and you
// gain 1 life" (Fear of Lost Teeth; Lavabelly Sliver's granted trigger): one
// damage from the source to each legal target, then the life gain.
func gapB106OneDamageAndOneLife(g *game.Game, item *game.StackItem) error {
	if err := sourceDealsDamageToEachLegalTarget(1)(g, item); err != nil {
		return err
	}
	return GainLife{Amount: 1}.Apply(NewContext(g, item))
}
