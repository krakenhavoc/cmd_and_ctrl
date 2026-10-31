package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "958d81c2-ca09-4275-890e-3bb595317325",
		Name:         "Iron Man Armor",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 1),
			GrantToAttached("flying"),
		},
		Grants: []AbilityGrant{{
			Key: ironManArmorPump,
			Static: []game.StaticAbility{{
				Label:    "This creature gets +1/+1 for each artifact you control.",
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return source != nil && target.InstanceID == source.InstanceID
				},
				Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
					n := artifactsControlledBy(g, source)
					c.Power += n
					c.Toughness += n
				},
			}},
			Text: "This creature gets +1/+1 for each artifact you control.",
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Iron Man Armor — attach it to target creature you control", AttachSourceToTarget),
				PermanentYouControl("target creature you control", Creature()),
			),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "{2}: If this Equipment isn't a creature, it becomes a 0/0 Construct Hero artifact creature with flying and \"This creature gets +1/+1 for each artifact you control\" until end of turn.",
				Cost:    ManaCost("{2}"),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  ironManArmorAnimate,
			},
			EquipAbility("{2}"),
		},
	})
}

// ironManArmorPump is the granted static's bundle key. An on-disk identity:
// never renamed.
const ironManArmorPump = "iron-man-armor/artifact-pump"

// ironManArmorAnimate is the {2} ability's resolution.
func ironManArmorAnimate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	self, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) || self.IsCreature() {
		return nil
	}
	return GrantAbilitiesFor{
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
