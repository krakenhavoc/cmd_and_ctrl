package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hunter's Axe — Artifact — Equipment {G} (Reality Fracture):
//
//	"Equipped creature gets +2/+0 and has "Whenever this creature
//	 attacks, it gains your choice of trample or deathtouch until end
//	 of turn."
//	 Equip {2}"
//
// The granted trigger is the equipped creature's: it asks the creature's
// controller for the keyword as the trigger resolves, and grants it to
// the creature until end of turn. If the creature has left the
// battlefield by then, nothing happens (CR 608.2b).
//
// No simplification.
const huntersAxeGrant = "hunters-axe/trample-or-deathtouch"

var huntersAxeKeywords = []string{"trample", "deathtouch"}

func init() {
	Register(Spec{
		OracleID:     "d7d59fef-1401-464b-b1bb-ee5da92cde51",
		Name:         "Hunter's Axe",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: huntersAxeGrant,
			Triggered: []game.TriggeredAbility{
				On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclared(ev, source)
				}, "Hunter's Axe — it gains your choice of trample or deathtouch until end of turn", huntersAxeChoose),
			},
			Text: "Whenever this creature attacks, it gains your choice of trample or deathtouch until end of turn.",
		}},
		Static: []game.StaticAbility{
			PumpAttached(2, 0),
			GrantAbilitiesToAttached(huntersAxeGrant),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}

func huntersAxeChoose(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	creature := item.SourceCardID
	if !onBattlefield(g, creature) {
		return nil
	}
	options := make([]game.ChoiceOption, len(huntersAxeKeywords))
	for i, kw := range huntersAxeKeywords {
		options[i] = game.ChoiceOption{Label: "Gains " + kw + " until end of turn"}
	}
	return PickOption{
		Question: "Hunter's Axe — the creature gains your choice of trample or deathtouch until end of turn",
		Options:  options,
		Then: func(ctx *Context, index int) error {
			if index < 0 || index >= len(huntersAxeKeywords) {
				return nil
			}
			return GrantKeywordUntilEOT{
				Target:   creature,
				Keywords: []string{huntersAxeKeywords[index]},
				Label:    "Hunter's Axe — gains " + huntersAxeKeywords[index] + " until end of turn",
			}.Apply(ctx)
		},
	}.Apply(ctx)
}
