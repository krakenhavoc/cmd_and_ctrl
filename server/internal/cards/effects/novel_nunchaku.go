package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Novel Nunchaku — {2}{G} Artifact — Equipment (Teenage Mutant Ninja
// Turtles, #4457):
//
//	"When this Equipment enters, attach it to target creature you control.
//	 When you do, equipped creature fights up to one target creature an
//	 opponent controls.
//	 Equipped creature gets +1/+1 and has trample.
//	 Equip {3}"
//
// The enters trigger is Biorganic Carapace's attach. "When you do" is a
// CR 603.12 reflexive trigger, created only when the Nunchaku really
// became attached to that creature (a target gone in response, or a
// Nunchaku that left, attaches nothing and fights nothing). Its "up to one
// target creature an opponent controls" is chosen as it goes on the stack
// (CR 603.3d). The fight is Fight the _____ Fight's: the Nunchaku rides
// the payload and "equipped creature" is read as the reflexive trigger
// resolves, so the creature it equips then fights, with the +1/+1 already
// counted (CR 701.14).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6d819742-3d02-4d60-9b10-82caa49a9d49",
		Name:         "Novel Nunchaku",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("trample"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Novel Nunchaku — attach it to target creature you control", novelNunchakuAttach),
				PermanentYouControl("target creature you control", Creature()),
			),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}

// novelNunchakuAttach attaches the Nunchaku and, once it is attached to the
// target, creates the reflexive fight.
func novelNunchakuAttach(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := g.AttachSourceForEffect(item, t); err != nil {
			return err
		}
		self, ok := g.LookupCardForEffect(item.SourceCardID)
		if !ok || !onBattlefield(g, self.InstanceID) || self.AttachedTo.Kind != game.TargetCard || self.AttachedTo.ID != t.ID {
			return nil
		}
		return ReflexiveTrigger{
			Label: "Novel Nunchaku — equipped creature fights up to one target creature an opponent controls",
			Cards: []uuid.UUID{item.SourceCardID},
			Body:  novelNunchakuFightBody,
		}.Apply(ctx)
	}
	return nil
}

// novelNunchakuFightBody is the reflexive half. An on-disk key.
var novelNunchakuFightBody = game.ReflexiveBody("novel-nunchaku/fight", simpleBody(fightTheBlankFightFight),
	constTargets(func() *game.TargetSpec {
		return TargetCreature("up to one target creature an opponent controls", OpponentControls()).WithCount(0, 1)
	}))
