package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tifa, Martial Artist — Legendary Creature — Human Monk {1}{R}{G}{W},
// 4/4 (#3656):
//
//	"Melee (Whenever this creature attacks, it gets +1/+1 until end of
//	 turn for each opponent you attacked this combat.)
//	 Whenever one or more creatures you control with power 7 or greater
//	 deal combat damage to a player, untap all creatures you control. If
//	 it's the first combat phase of your turn, there is an additional
//	 combat phase after this phase."
//
// Melee has no engine keyword, so it is written out as its own attack
// trigger: it counts the distinct players the controller's attackers are
// aimed at, read as the trigger resolves (the whole declaration is made
// by then), and Tifa gets that many +1/+1. The damage trigger is the
// shared one-per-player shape over power 7 or greater, read off the
// damage source. The extra combat is added only in the turn's first
// combat phase, so an added phase does not add another.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0836c8d1-c453-4a3a-a52a-eb66839b8542",
		Name:         "Tifa, Martial Artist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Tifa, Martial Artist — Melee: +1/+1 until end of turn for each opponent you attacked this combat",
				func(g *game.Game, item *game.StackItem) error {
					attacked := map[uuid.UUID]bool{}
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller != item.Controller || c.AttackingTarget == uuid.Nil {
							continue
						}
						if p := g.DefendingPlayerForAttackForEffect(c.AttackingTarget); p != uuid.Nil && p != item.Controller {
							attacked[p] = true
						}
					}
					n := len(attacked)
					if n == 0 || !onBattlefield(g, item.SourceCardID) {
						return nil
					}
					return BoostUntilEOT{Target: item.SourceCardID, Power: n, Toughness: n, Label: "Tifa, Martial Artist — Melee"}.Apply(NewContext(g, item))
				}),
			WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(PowerGE(7),
				"Tifa, Martial Artist — untap all creatures you control; if it's the first combat phase of your turn, an additional combat phase",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (UntapAllCreaturesYouControl{}).Apply(ctx); err != nil {
						return err
					}
					if !g.IsFirstCombatPhaseForEffect() {
						return nil
					}
					return ExtraCombatAfterThisPhase().Apply(ctx)
				}),
		},
	})
}
