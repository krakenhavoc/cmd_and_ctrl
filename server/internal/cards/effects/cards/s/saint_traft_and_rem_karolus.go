package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Saint Traft and Rem Karolus — Legendary Creature — Spirit Human
// {U}{R}{W}, 3/4 (March of the Machine Commander, #5647):
//
//	"Whenever Saint Traft and Rem Karolus becomes tapped, create a 1/1
//	 red Human creature token if this is the first time this ability has
//	 resolved this turn. If it's the second time, create a 1/1 blue
//	 Spirit creature token with flying. If it's the third time, create a
//	 4/4 white Angel creature token with flying.
//	 Whenever you cast a spell that has convoke, untap Saint Traft and
//	 Rem Karolus."
//
// "Becomes tapped" is any tap, an attack and a convoke tap included
// (ThisBecameTapped). Which token is Vito, Fanatic of Aclazotz's read:
// the engine's per-object tally of how many times this ability has
// resolved this turn, counting the resolution in progress, so a fourth
// or later resolution makes nothing. The untap asks whether the spell
// had convoke as it became cast (YouCastASpellThatHasConvoke); tapping
// the pair to convoke that very spell is paid before the spell becomes
// cast, so the trigger untaps them again.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3a1be8b3-16e2-4fae-bea7-3a98f2f9db38",
		Name:         "Saint Traft and Rem Karolus",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
				saintTraftAndRemLabel, saintTraftAndRemResolves),
			effects.On(game.EventCast, effects.YouCastASpellThatHasConvoke,
				"Saint Traft and Rem Karolus — untap Saint Traft and Rem Karolus",
				func(g *game.Game, item *game.StackItem) error {
					return effects.UntapTarget{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

const saintTraftAndRemLabel = "Saint Traft and Rem Karolus — a 1/1 Human, then a 1/1 Spirit with flying, then a 4/4 Angel with flying"

// saintTraftAndRemResolves picks the token by how many times the
// ability has resolved this turn.
func saintTraftAndRemResolves(g *game.Game, item *game.StackItem) error {
	var token string
	switch g.ResolvedThisTurn(item.SourceCardID, saintTraftAndRemLabel) {
	case 1:
		token = "1/1 red Human"
	case 2:
		token = "1/1 blue Spirit with flying"
	case 3:
		token = "4/4 white Angel with flying"
	default:
		return nil
	}
	return effects.CreateToken{Template: effects.TokenCard(token), N: 1}.Apply(effects.NewContext(g, item))
}
