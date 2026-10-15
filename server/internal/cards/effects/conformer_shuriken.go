package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "8e7681ba-6412-4165-b374-5636feec216d",
		Name:         "Conformer Shuriken",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: conformerShurikenGrant,
			Triggered: []game.TriggeredAbility{
				withTargetsFrom(
					On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
						return attackDeclared(ev, source)
					}, "Conformer Shuriken — tap target creature defending player controls",
						func(g *game.Game, item *game.StackItem) error {
							ctx := NewContext(g, item)
							target, ok := b16FirstLegalTargetCard(ctx)
							if !ok {
								return nil
							}
							if err := (TapTarget{Target: target}).Apply(ctx); err != nil {
								return err
							}
							if !onBattlefield(g, item.SourceCardID) || !onBattlefield(g, target) {
								return nil
							}
							diff := b31CurrentPowerOf(g, target) - b31CurrentPowerOf(g, item.SourceCardID)
							if diff <= 0 {
								return nil
							}
							return AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: diff}.Apply(ctx)
						}),
					TargetOfDefendingPlayer("creature", "target creature defending player controls", Creature())),
			},
			Text: "Whenever this creature attacks, tap target creature defending player controls. If that creature has greater power than this creature, put a number of +1/+1 counters on this creature equal to the difference.",
		}},
		Static: []game.StaticAbility{
			GrantAbilitiesToAttached(conformerShurikenGrant),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
