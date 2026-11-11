package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Convert to Slime — Sorcery {3}{B}{G} (Duskmourn Commander):
//
//	"Destroy up to one target artifact, up to one target creature, and up to
//	 one target enchantment.
//	 Delirium — Then if there are four or more card types among cards in your
//	 graveyard, create an X/X green Ooze creature token, where X is the total
//	 mana value of permanents destroyed this way."
//
// Three optional target clauses. Each target still legal is destroyed in
// turn; a permanent counts as destroyed this way only if it landed in a
// graveyard (CR 701.8a), so an indestructible or regenerated one does not,
// and its mana value is read as it last existed on the battlefield (a token
// is 0). "Then" makes the delirium check come after the destruction, so the
// destroyed cards themselves count towards it. A permanent targeted by two
// clauses cannot occur: the types are checked one clause at a time and a
// permanent that is both an artifact and a creature is destroyed once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b4429e71-434e-480c-a8d1-2ad11de8b2d9",
		Name:         "Convert to Slime",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetPermanent("up to one target artifact", Artifact()).WithCount(0, 1),
			TargetCreature("up to one target creature").WithCount(0, 1),
			TargetPermanent("up to one target enchantment", Enchantment()).WithCount(0, 1),
		),
		Purpose: game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			total := 0
			seen := map[uuid.UUID]bool{}
			for slot := 0; slot < 3; slot++ {
				t, ok := ctx.ClauseTarget(slot)
				if !ok || t.Kind != game.TargetCard || seen[t.ID] {
					continue
				}
				seen[t.ID] = true
				card, found := ctx.Game.LookupCardForEffect(t.ID)
				if !found {
					continue
				}
				if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				if z := ctx.Game.FindCardZoneForEffect(t.ID); z != nil && (z.Kind == game.ZoneGraveyard || z.Kind == game.ZoneCommand) {
					total += card.ManaValue()
				}
			}
			if !drcDeliriumMet(ctx.Game, ctx.Controller()) {
				return nil
			}
			tok := ctx.TextToken(TokenCard("2/2 green Ooze"))
			tok.Power, tok.Toughness = total, total
			return ctx.Game.CreateTokenForEffect(ctx.Controller(), tok, 1)
		},
	})
}
