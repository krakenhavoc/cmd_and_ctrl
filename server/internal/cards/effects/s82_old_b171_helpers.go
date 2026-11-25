package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b171_helpers.go — shared bodies for the S82 older-sets batch 171
// (War of the Spark, #5520).

// S82OldB171OtherCreaturesGainUntilEOT is "other creatures you control gain
// <keyword> until end of turn" (Loxodon Sergeant, Thundering Ceratok): the
// creatures you control as the ability resolves, the source left out (CR 611.2c).
func S82OldB171OtherCreaturesGainUntilEOT(name, keyword string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return GrantKeywordUntilEOT{
			Match:    And(Creature(), YouControl(), OtherThan(item.SourceCardID)),
			Keywords: []string{keyword},
			Label:    name + " — other creatures you control gain " + keyword,
		}.Apply(NewContext(g, item))
	}
}

// S82OldB171RhonasDoubles is God-Eternal Rhonas's enters ability: double the
// power of each other creature you control, then those creatures gain
// vigilance, both over the set read once as it resolves (CR 611.2c, 701.10b).
func S82OldB171RhonasDoubles(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	set := map[uuid.UUID]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() && c.InstanceID != item.SourceCardID {
			ids = append(ids, c.InstanceID)
			set[c.InstanceID] = true
		}
	}
	if len(ids) == 0 {
		return nil
	}
	if err := DoublePowerOfEachUntilEOT(ctx, ids, "God-Eternal Rhonas — double power"); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{
		Match:    func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return set[c.InstanceID] },
		Keywords: []string{"vigilance"},
		Label:    "God-Eternal Rhonas — those creatures gain vigilance",
	}.Apply(ctx)
}
