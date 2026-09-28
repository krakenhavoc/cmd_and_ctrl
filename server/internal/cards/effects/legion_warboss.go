package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Legion Warboss — Creature — Goblin Soldier, {2}{R}, 2/2:
//
//	"Mentor (Whenever this creature attacks, put a +1/+1 counter on
//	 target attacking creature with lesser power.)
//	 At the beginning of combat on your turn, create a 1/1 red
//	 Goblin creature token. That token gains haste until end of turn
//	 and attacks this combat if able."
//
// #1599: the token's haste and attack requirement are the SOURCE's
// own scoped record (attack_requirements.go's own doc comment names
// this card for exactly this shape) rather than a lord's static,
// because they belong to one token for one turn, not to "Goblin
// creatures you control" — a second Goblin that enters later this
// turn is never required to attack by this ability. The token is not
// sacrificed by this ability (that line belongs to other "make a
// temporary attacker" cards, not this one) — Legion Warboss's tokens
// stick around.
//
// Mentor (CR 702.136) is new machinery for the catalog: a
// TargetsFrom clause built fresh at trigger time from the mentoring
// creature's OWN current power (read off the `source` value TargetsFrom
// is handed, never a captured one), offering only attacking creatures
// with strictly lesser power — which excludes Legion Warboss itself
// without an explicit "other" clause, since nothing has less power
// than its own. CR 603.3d drops the trigger without a prompt when no
// attacker qualifies (Legion Warboss attacking alone, or every other
// attacker at least as big).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0ccb9af3-6902-4130-a59c-4c882dc3a2fd",
		Name:         "Legion Warboss",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventAttack},
				AppliesTo: ThisAttacked,
				TargetsFrom: func(_ game.TriggerContext, source *game.Card, _ *game.Game) *game.TargetSpec {
					power := source.CurrentPower()
					return &game.TargetSpec{
						Mode: "creature", Label: "target attacking creature with lesser power",
						Zones: []game.ZoneKind{game.ZoneBattlefield},
						CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
							return c.IsCreature() && c.AttackingTarget != uuid.Nil && c.CurrentPower() < power
						},
						Min: 1, Max: 1,
					}
				},
				Key: "Legion Warboss — mentor",
				// b36CounterOnChosenAnimal is the shared "+1/+1 counter
				// on the announced target" body (Animal Sanctuary,
				// Benevolent Hydra, Hall of Oracles) — Mentor's payoff
				// is the identical shape, so it is reused rather than
				// duplicated (TestNoNewExactClonesInTheCatalog).
				Effect: b36CounterOnChosenAnimal,
			},
			AtBeginningOfYourCombat("Legion Warboss — create a 1/1 red Goblin that attacks this combat if able",
				legionWarbossToken),
		},
	})
}

// legionWarbossToken creates the 1/1 red Goblin and pins the printed
// clause to that ONE token: haste and "attacks this combat if able",
// both until end of turn (ScopedEffectFor, not a static — see the
// card comment).
func legionWarbossToken(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	cursor := b25LastEventSeq(g)
	if err := (CreateToken{Template: RedGoblinToken(), N: 1}).Apply(ctx); err != nil {
		return err
	}
	tokens := b27TokensCreatedByAfter(g, item.Controller, cursor)
	if len(tokens) == 0 {
		return nil
	}
	token := tokens[0]
	if err := (ScopedEffectFor{
		Target:   token,
		Mods:     []game.Mod{game.AddKeywordsMod("haste"), game.AddAttackRequirementMod(uuid.Nil)},
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    "Legion Warboss — the token has haste and attacks this combat if able",
	}).Apply(ctx); err != nil {
		return err
	}
	return nil
}
