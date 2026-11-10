package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Naga Fleshcrafter — Creature — Snake Shapeshifter {3}{U}, 0/0 (Tarkir:
// Dragonstorm, #4463):
//
//	"You may have this creature enter as a copy of any creature on the
//	 battlefield.
//	 Renew — {2}{U}, Exile this card from your graveyard: Put a +1/+1
//	 counter on target nonlegendary creature you control. Each other
//	 creature you control becomes a copy of that creature until end of
//	 turn. Activate only as a sorcery."
//
// The first line is Clone's (EntersAsCopyOf): not targeting, optional, and a
// declined copy is the printed 0/0, which dies (CR 704.5f). Renew is
// renew.go. As it resolves the target takes its +1/+1 counter, then every
// OTHER creature its controller controls becomes a copy of it until end of
// turn (BecomeCopy, CR 707.2, 611.2): a copy takes only copiable values, so
// the counter, damage and effects on the target are not copied, a token is
// copied as the token it was created as, and a target that is itself a copy
// is copied as what it copies. A target that is gone or no longer a
// nonlegendary creature you control does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a4f8012e-9880-45b0-9eba-d74c207d3242",
		Name:         "Naga Fleshcrafter",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf("Naga Fleshcrafter", anyCreatureOnBattlefield, nil),
		},
		Activated: []ActivatedAbility{
			Renew("{2}{U}", "Put a +1/+1 counter on target nonlegendary creature you control. Each other creature you control becomes a copy of that creature until end of turn.",
				TargetCreature("target nonlegendary creature you control", Not(Legendary()), YouControl()),
				nagaFleshcrafterRenew),
		},
	})
}

// nagaFleshcrafterRenew puts the counter on the target, then turns each
// other creature the controller controls into a copy of it.
func nagaFleshcrafterRenew(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (AddCounter{Target: target, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	var others []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.Controller == ctx.Controller() && c.InstanceID != target {
			others = append(others, c.InstanceID)
		}
	}
	if len(others) == 0 {
		return nil
	}
	return BecomeCopy{Targets: others, Of: target}.Apply(ctx)
}
