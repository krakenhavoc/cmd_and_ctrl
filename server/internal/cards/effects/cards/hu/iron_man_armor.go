package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Iron Man Armor — Artifact — Equipment {3} (Marvel Super Heroes, #4455):
//
//	"When this Equipment enters, attach it to target creature you control.
//	 Equipped creature gets +2/+1 and has flying.
//	 {2}: If this Equipment isn't a creature, it becomes a 0/0 Construct
//	 Hero artifact creature with flying and "This creature gets +1/+1 for
//	 each artifact you control" until end of turn.
//	 Equip {2}"
//
// Celestial Armor's shape for the enters-attach and the equipped bonus. The
// {2} ability is one effect with one timestamp (CR 613.7): the type change
// (Artifact Creature, Construct Hero; Equipment stays, an artifact subtype of
// a type it keeps, CR 205.1a), the 0/0 base, flying, and a granted static
// ("This creature gets +1/+1 for each artifact you control", counted live,
// the Armor itself included once it is a creature). The check is on
// resolution: if it is already a creature nothing happens, and an Armor that
// left or came back as a new object gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "958d81c2-ca09-4275-890e-3bb595317325",
		Name:         "Iron Man Armor",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 1),
			effects.GrantToAttached("flying"),
		},
		Grants: []effects.AbilityGrant{{
			Key: ironManArmorPump,
			Static: []game.StaticAbility{{
				Label:    "This creature gets +1/+1 for each artifact you control.",
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return source != nil && target.InstanceID == source.InstanceID
				},
				Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
					n := effects.ArtifactsControlledBy(g, source)
					c.Power += n
					c.Toughness += n
				},
			}},
			Text: "This creature gets +1/+1 for each artifact you control.",
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Iron Man Armor — attach it to target creature you control", effects.AttachSourceToTarget),
				effects.PermanentYouControl("target creature you control", effects.Creature()),
			),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}: If this Equipment isn't a creature, it becomes a 0/0 Construct Hero artifact creature with flying and \"This creature gets +1/+1 for each artifact you control\" until end of turn.",
				Cost:    effects.ManaCost("{2}"),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  ironManArmorAnimate,
			},
			effects.EquipAbility("{2}"),
		},
	})
}

// ironManArmorPump is the granted static's bundle key. An on-disk identity:
// never renamed.
const ironManArmorPump = "iron-man-armor/artifact-pump"

// ironManArmorAnimate is the {2} ability's resolution.
func ironManArmorAnimate(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	self, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) || self.IsCreature() {
		return nil
	}
	return effects.GrantAbilitiesFor{
		Target: item.SourceCardID,
		Keys:   []string{ironManArmorPump},
		Also: []game.Mod{
			game.SetTypesMod([]string{"Artifact", "Creature"}, ctx.TextSubtypes("Construct", "Hero")),
			game.SetBasePowerMod(0),
			game.SetBaseToughnessMod(0),
			game.AddKeywordsMod("flying"),
		},
		Label: "Iron Man Armor — becomes a 0/0 Construct Hero artifact creature with flying until end of turn",
	}.Apply(ctx)
}
