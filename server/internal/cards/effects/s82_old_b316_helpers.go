package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b316_helpers.go — shared pieces of OLD batch 316, Duel Decks:
// Zendikar vs. Eldrazi (S82, #5548).

// S82OldB316Rally is "Whenever this creature or another Ally you control
// enters, <effect>" (rally): one trigger per entering Ally, this creature
// included.
func S82OldB316Rally(label string, effect Effect) game.TriggeredAbility {
	return bfzB1Rally(label, effect)
}

// S82OldB316ArcherShot is Tajuru Archer's "you may have this creature deal
// damage to target creature with flying equal to the number of Allies you
// control": the count is read as the trigger resolves, Allies read through
// the layered view (a changeling counts).
func S82OldB316ArcherShot(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := OldB10CountSubtype(g, item.Controller, ctx.TextSubtype("Ally"))
	for _, t := range ctx.LegalTargets() {
		return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
	}
	return nil
}

// S82OldB316SacrificeThenTokens is "Sacrifice all creatures you control, then
// create that many <token>" (Hellion Eruption): the count is what was really
// sacrificed, told by a keyed tail once the move has landed.
func S82OldB316SacrificeThenTokens(g *game.Game, item *game.StackItem) error {
	ids := make([]uuid.UUID, 0)
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return g.SacrificeAllKeyedForEffect(item.SourceCardID, ids,
		game.KeyedTail{Ref: s82OldB316HellionTokens, Player: item.Controller, Source: item.SourceCardID})
}

var s82OldB316HellionTokens = game.RegisterTail("then/s82-old-b316-hellion-eruption-tokens", func(g *game.Game, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	ctx, _ := AskedContext(g, nil, o.Tail.Player, o.Tail.Source, nil)
	return CreateToken{Template: TokenCard("4/4 red Hellion"), N: len(o.IDs)}.Apply(ctx)
})
