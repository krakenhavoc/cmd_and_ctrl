package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Conformer Shuriken — Legendary Artifact — Equipment {2} (#3656):
//
//	"Equipped creature has "Whenever this creature attacks, tap target
//	 creature defending player controls. If that creature has greater
//	 power than this creature, put a number of +1/+1 counters on this
//	 creature equal to the difference."
//	 Equip {2}"
//
// The trigger is the equipped creature's own (ADR 0093, Medic's
// Kitesail's shape), with Floral Spuzzem's "defending player controls"
// target clause read off the attacker. Powers are compared as the
// ability resolves, after the tap, as the sentence reads. A target that
// left in response, or an attacker that left, does nothing (CR 608.2b).
//
// No simplification.
const conformerShurikenGrant = "conformer-shuriken/tap-and-grow"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "8e7681ba-6412-4165-b374-5636feec216d",
		Name:         "Conformer Shuriken",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: conformerShurikenGrant,
			Triggered: []game.TriggeredAbility{
				effects.WithTargetsFrom(
					effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
						return effects.AttackDeclared(ev, source)
					}, "Conformer Shuriken — tap target creature defending player controls",
						func(g *game.Game, item *game.StackItem) error {
							ctx := effects.NewContext(g, item)
							target, ok := effects.B16FirstLegalTargetCard(ctx)
							if !ok {
								return nil
							}
							if err := (effects.TapTarget{Target: target}).Apply(ctx); err != nil {
								return err
							}
							if !effects.OnBattlefield(g, item.SourceCardID) || !effects.OnBattlefield(g, target) {
								return nil
							}
							diff := effects.B31CurrentPowerOf(g, target) - effects.B31CurrentPowerOf(g, item.SourceCardID)
							if diff <= 0 {
								return nil
							}
							return effects.AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: diff}.Apply(ctx)
						}),
					effects.TargetOfDefendingPlayer("creature", "target creature defending player controls", effects.Creature())),
			},
			Text: "Whenever this creature attacks, tap target creature defending player controls. If that creature has greater power than this creature, put a number of +1/+1 counters on this creature equal to the difference.",
		}},
		Static: []game.StaticAbility{
			effects.GrantAbilitiesToAttached(conformerShurikenGrant),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
