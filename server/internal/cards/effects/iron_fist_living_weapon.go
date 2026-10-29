package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Iron Fist, Living Weapon — {2}{R} Legendary Creature — Human Warrior Hero
// 2/3 (Marvel Super Heroes, #4455):
//
//	"Whenever you cast a spell that targets a creature you control, Iron Fist
//	 gains "{T}: Iron Fist deals damage equal to his power to any other
//	 target" until end of turn."
//
// The trigger is Colleen Wing's (a spell of yours with a creature of yours
// among its targets), and the grant is a duration grant (ADR 0093 PR 4): a
// bundle declared in Spec.Grants and handed to Iron Fist for the turn. The
// granted ability is his own, so {T} needs him to have been under your
// control since the turn began (CR 302.6), and "any other target" excludes
// him. The damage is his power as the ability resolves; if he has left by
// then, as he last was (CR 608.2h). Casting a second such spell gives him a
// second copy of the ability, which is a second tap-ability he can only use
// once while he stays tapped.
//
// A new object of the same card (it left and came back) gets nothing: the
// grant is only given to the Iron Fist that triggered it.
//
// No simplification.
const ironFistTapDamage = "iron-fist-living-weapon/tap-damage"

func init() {
	Register(Spec{
		OracleID:     "bbf873c2-94c1-46b8-912e-e29e48e0212c",
		Name:         "Iron Fist, Living Weapon",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: ironFistTapDamage,
			Activated: []ActivatedAbility{{
				Label:   "{T}: Iron Fist deals damage equal to his power to any other target",
				Cost:    TapCost(),
				Targets: Another(TargetAny()),
				Effect:  ironFistDamage,
			}},
			Text: "{T}: Iron Fist deals damage equal to his power to any other target.",
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, mshB4CastSpellTargetingYourCreature,
				"Iron Fist, Living Weapon — he gains \"{T}: He deals damage equal to his power to any other target\" until end of turn",
				ironFistGrant),
		},
	})
}

func ironFistGrant(g *game.Game, item *game.StackItem) error {
	if !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
		return nil
	}
	return GrantAbilitiesFor{
		Target: item.SourceCardID,
		Keys:   []string{ironFistTapDamage},
		Label:  "Iron Fist, Living Weapon — {T}: damage equal to his power to any other target",
	}.Apply(NewContext(g, item))
}

func ironFistDamage(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 {
		return nil
	}
	return DealDamage{
		Source: item.SourceCardID,
		Target: item.Targets[0].ID,
		Amount: powerNowOrLastKnown(g, item.SourceCardID),
	}.Apply(NewContext(g, item))
}
